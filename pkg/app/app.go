// Package app da el runtime común de arranque y apagado de los servicios.
//
// Un servicio no es un único proceso en marcha, sino varias piezas corriendo a
// la vez: servidores gRPC y HTTP, consumidores de Kafka, el relay del outbox.
// Este paquete las trata a todas igual a través de Component, para que las
// señales del sistema y la propagación de fallos se escriban una sola vez.

package app

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

// Component es una pieza del servicio que corre en segundo plano.
//
// Run bloquea mientras la pieza trabaja y solo vuelve cuando ctx se cancela,
// momento en el que debe apagarse por su cuenta: un apagado limpio devuelve
// nil y cualquier otra cosa es un fallo que tumbará al resto de componentes.
// Name solo identifica la pieza en los logs.
type Component interface {
	Name() string
	Run(ctx context.Context) error
}

// Run arranca todos los componentes y no vuelve hasta que todos han parado.
//
// Se llama una vez por servicio, desde su main, y es lo único que conoce la
// lista completa de components. El apagado se ordena cancelando el ctx que reciben:
// señal del sistema (SIGINT al pulsar Ctrl+C, SIGTERM al parar un contenedor)
// o el fallo de otro componente. Devuelve el primer error que se produjo, o
// nil si todo se apagó limpio.
func Run(ctx context.Context, logger *slog.Logger, components ...Component) error {
	// Deriva un ctx que se cancela solo al recibir una señal de parada del SO.
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

	// Libera el registro de señales al salir: sin esto seguirían capturadas.
	defer stop()

	// El ctx del grupo añade la segunda vía de cancelación: el fallo de un componente.
	g, ctx := errgroup.WithContext(ctx)

	for _, c := range components {
		// Cada componente corre en su propia goroutine, porque su Run bloquea.
		g.Go(func() error {
			logger.Info("component starting", "component", c.Name())

			// Aquí se pasa la vida del componente: solo vuelve cuando se cancela ctx o falla.
			err := c.Run(ctx)
			if err != nil {
				// Devolver el error cancela el ctx del grupo y arrastra a los demás.
				logger.Error("component stopped", "component", c.Name(), "error", err)
				return err
			}

			logger.Info("component stopped", "component", c.Name())
			return nil
		})
	}

	// Espera a que todos terminen de recoger y devuelve el primer error no nil.
	return g.Wait()
}
