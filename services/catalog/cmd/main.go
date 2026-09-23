package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	catalogv1 "github.com/cristianrisueo/thebattleroyale1/gen/catalog/v1"
	"github.com/cristianrisueo/thebattleroyale1/pkg/app"
	"github.com/cristianrisueo/thebattleroyale1/pkg/cache"
	"github.com/cristianrisueo/thebattleroyale1/pkg/database"
	"github.com/cristianrisueo/thebattleroyale1/pkg/logger"
	"github.com/cristianrisueo/thebattleroyale1/pkg/otel"
	"github.com/cristianrisueo/thebattleroyale1/services/catalog/internal"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

// shutdownTimeout es el margen que se le da al servidor gRPC para terminar las RPC en curso
// tracerShutdownTimeout es el margen para vaciar los lotes de spans pendientes al apagar
const (
	shutdownTimeout       = 10 * time.Second
	tracerShutdownTimeout = 5 * time.Second
)

// config agrupa los valores que el servicio lee del entorno al arrancar
type config struct {
	databaseURL  string
	grpcAddr     string
	otlpEndpoint string
	redisAddr    string
	cacheTTL     time.Duration
}

// main es el punto de entrada del servicio catalog
func main() {
	// Crea el logger antes que nada para que hasta un fallo de configuración se vea en el log
	log := logger.NewLogger("catalog")

	// La lógica vive en run porque os.Exit no ejecuta los defer y el pool quedaría abierto
	if err := run(log); err != nil {
		log.Error("startup failed", "error", err)
		os.Exit(1)
	}
}

// run monta el servicio y bloquea hasta que termina de apagarse
//
// Param - log: Logger raíz del servicio, compartido por todos los componentes
// Returns - error: El primer fallo de arranque o de un componente, nil si todo se apagó limpio
func run(log *slog.Logger) error {
	ctx := context.Background()

	// Lee la configuración del entorno y falla si falta algo
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	// Registra el tracer global antes que el pool para que otelpgx lo encuentre
	shutdownTracer, err := otel.NewTracer(ctx, "catalog", cfg.otlpEndpoint)
	if err != nil {
		return fmt.Errorf("creating tracer: %w", err)
	}

	// Difiere el vaciado de spans con un ctx propio porque el del servicio ya estará cancelado
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), tracerShutdownTimeout)
		defer cancel()

		// Registra el fallo sin cambiar el error de salida: solo se pierden los últimos spans
		if err := shutdownTracer(shutdownCtx); err != nil {
			log.Error("tracer shutdown failed", "error", err)
		}
	}()

	// Un único pool para todo el servicio: lo comparten todas las RPC
	pool, err := database.NewPool(ctx, cfg.databaseURL)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}

	// Difiere el cierre del pool
	defer pool.Close()

	// Crea el cliente de redis
	rdb, err := cache.NewClient(ctx, cfg.redisAddr)
	if err != nil {
		return fmt.Errorf("connecting to redis: %w", err)
	}

	// Difiere el cierre del cliente de Redis
	defer rdb.Close()

	// Monta el dominio: repositorio, servicio y handler gRPC, devuelve solo lo que
	handler := createDependencies(pool, rdb, cfg.cacheTTL, log)

	// Crea el servidor gRPC, que todavía no escucha, con un span por cada RPC recibida
	grpcServer := app.NewGRPCServer(cfg.grpcAddr, log, shutdownTimeout,
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	// Registra el handler antes de servir: hacerlo después provoca un pánico
	catalogv1.RegisterCatalogServiceServer(grpcServer.Server(), handler)

	// Bloquea hasta que llega una señal de parada y todos los componentes han recogido
	return app.Run(ctx, log, grpcServer)
}

//* Funciones auxiliares para hacer más legible el método run

// loadConfig lee de las variables del entorno la configuración del servicio
func loadConfig() (config, error) {
	// La URL de la base de datos no tiene valor por defecto: arrancar contra la equivocada es peor que no arrancar
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return config{}, fmt.Errorf("missing DATABASE_URL")
	}

	// La dirección del servidor gRPC tampoco tiene valor por defecto
	grpcAddr := os.Getenv("GRPC_ADDR")
	if grpcAddr == "" {
		return config{}, fmt.Errorf("missing GRPC_ADDR")
	}

	// El recolector de trazas tampoco tiene valor por defecto
	otlpEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otlpEndpoint == "" {
		return config{}, fmt.Errorf("missing OTEL_EXPORTER_OTLP_ENDPOINT")
	}

	// La dirección de Redis tampoco tiene valor por defecto
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		return config{}, fmt.Errorf("missing REDIS_ADDR")
	}

	// El TTL de la caché tampoco tiene valor por defecto
	rawTTL := os.Getenv("CACHE_TTL")
	if rawTTL == "" {
		return config{}, fmt.Errorf("missing CACHE_TTL")
	}

	// Interpreta el TTL como duración de Go (5m, 30s...)
	cacheTTL, err := time.ParseDuration(rawTTL)
	if err != nil {
		return config{}, fmt.Errorf("parsing CACHE_TTL: %w", err)
	}

	// Rechaza un TTL no positivo: con 0 go-redis guardaría las claves sin caducidad
	if cacheTTL <= 0 {
		return config{}, fmt.Errorf("CACHE_TTL must be positive, got %s", cacheTTL)
	}

	return config{
		databaseURL:  databaseURL,
		grpcAddr:     grpcAddr,
		otlpEndpoint: otlpEndpoint,
		redisAddr:    redisAddr,
		cacheTTL:     cacheTTL,
	}, nil
}

// createDependencies monta el dominio sobre el pool y Redis y devuelve el handler gRPC
func createDependencies(pool *pgxpool.Pool, rdb *redis.Client, cacheTTL time.Duration, log *slog.Logger) *internal.CatalogHandler {
	// Cada capa recibe la de debajo: el handler no sabe que existe Postgres
	repo := internal.NewCatalogRepository(pool)

	// Pone la caché delante de Postgres: el servicio no distingue un repositorio del otro
	cached := internal.NewCachedRepository(repo, rdb, cacheTTL, log)

	svc := internal.NewCatalogService(cached)

	return internal.NewCatalogHandler(svc)
}
