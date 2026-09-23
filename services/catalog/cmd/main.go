package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	catalogv1 "github.com/cristianrisueo/thebattleroyale1/gen/catalog/v1"
	"github.com/cristianrisueo/thebattleroyale1/pkg/app"
	"github.com/cristianrisueo/thebattleroyale1/pkg/database"
	"github.com/cristianrisueo/thebattleroyale1/pkg/logger"
	"github.com/cristianrisueo/thebattleroyale1/services/catalog/internal"
	"github.com/jackc/pgx/v5/pgxpool"
)

// shutdownTimeout es el margen que se le da al servidor gRPC para terminar las RPC en curso
const shutdownTimeout = 10 * time.Second

// config agrupa los valores que el servicio lee del entorno al arrancar
type config struct {
	databaseURL string
	grpcAddr    string
}

// main es el punto de entrada del servicio catalog
func main() {
	// Crea el logger antes que nada para que hasta un fallo de configuración se vea en el log
	log := logger.New("catalog")

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

	// Un único pool para todo el servicio: lo comparten todas las RPC
	pool, err := database.NewPool(ctx, cfg.databaseURL)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}

	// Difiere el cierre del pool
	defer pool.Close()

	// Monta el dominio: repositorio, servicio y handler gRPC, devuelve solo lo que
	handler := createDependencies(pool)

	// Crea el servidor gRPC, que todavía no escucha
	grpcServer := app.NewGRPCServer(cfg.grpcAddr, log, shutdownTimeout)

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

	return config{databaseURL: databaseURL, grpcAddr: grpcAddr}, nil
}

// createDependencies monta el dominio sobre el pool y devuelve el handler gRPC
func createDependencies(pool *pgxpool.Pool) *internal.CatalogHandler {
	// Cada capa recibe la de debajo: el handler no sabe que existe Postgres
	repo := internal.NewCatalogRepository(pool)
	svc := internal.NewCatalogService(repo)

	return internal.NewCatalogHandler(svc)
}
