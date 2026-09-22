package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/cristianrisueo/thebattleroyale1/pkg/app"
	"github.com/cristianrisueo/thebattleroyale1/pkg/database"
	"github.com/cristianrisueo/thebattleroyale1/pkg/logger"
)

// shutdownTimeout es el margen para un apagado limpio del servidor gRPC.
const shutdownTimeout = 10 * time.Second

func main() {
	log := logger.New("catalog")

	// run() separada de main() para que os.Exit no se salte el defer del pool.
	if err := run(log); err != nil {
		log.Error("startup failed", "error", err)
		os.Exit(1)
	}
}

// run lee la config, abre el pool y arranca el servidor gRPC.
func run(log *slog.Logger) error {
	ctx := context.Background()

	// Falla si falta la URL de la base de datos.
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("missing DATABASE_URL")
	}

	// Falla si falta la dirección del servidor gRPC.
	grpcAddr := os.Getenv("GRPC_ADDR")
	if grpcAddr == "" {
		return fmt.Errorf("missing GRPC_ADDR")
	}

	// Pool compartido por todos los handlers gRPC del servicio.
	pool, err := database.NewPool(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}

	// Cierra el pool cuando el servidor gRPC termina.
	defer pool.Close()

	// Crea el servidor gRPC con el timeout de apagado ya fijado.
	grpcServer := app.NewGRPCServer(grpcAddr, log, shutdownTimeout)

	return app.Run(ctx, log, grpcServer)
}
