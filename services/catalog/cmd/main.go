// Catalog sirve el catálogo de alumnos, armas y localizaciones por gRPC.
//
// Este fichero es el único del servicio que conoce a la vez la configuración,
// la base de datos y el servidor: el resto del código recibe ya montado lo que
// necesita. Todo lo genérico vive en pkg/; aquí solo queda el cableado propio
// de catalog.

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

// shutdownTimeout es el margen que se le da al servidor gRPC para terminar las
// RPC en curso antes de cortarlas en seco.
const shutdownTimeout = 10 * time.Second

// main es el punto de entrada del servicio catalog
func main() {
	// El logger se crea antes que nada para que hasta un fallo de config se vea en el log.
	log := logger.New("catalog")

	// La lógica vive en run porque os.Exit no ejecuta los defer y el pool quedaría abierto.
	if err := run(log); err != nil {
		log.Error("startup failed", "error", err)
		os.Exit(1)
	}
}

// run monta el servicio y bloquea hasta que termina de apagarse.
//
// El orden importa: cada dependencia se abre antes que quien la usa, y los
// defer las cierran en orden inverso. Por eso el pool se cierra después de que
// app.Run haya vuelto, cuando ya no queda ninguna RPC a medias que pudiera
// encontrarse la base de datos cerrada debajo.
//
// La configuración se lee del entorno y no tiene valores por defecto: un
// servicio que arranca contra la base de datos equivocada por un descuido es
// peor que uno que no arranca.
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

	// Un único pool para todo el servicio: lo comparten todos los handlers gRPC.
	pool, err := database.NewPool(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connecting to database: %w", err)
	}

	// Cierra el pool cuando el servidor gRPC termina.
	defer pool.Close()

	// Aún sin servicios propios registrados: eso llega con CatalogService.
	grpcServer := app.NewGRPCServer(grpcAddr, log, shutdownTimeout)

	// Bloquea hasta que llega una señal de parada y todos los componentes han recogido.
	return app.Run(ctx, log, grpcServer)
}
