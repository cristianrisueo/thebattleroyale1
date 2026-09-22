// Package app da el runtime común de arranque/apagado de los servicios.
package app

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

// Component corre hasta que su contexto se cancela; un apagado limpio devuelve nil.
type Component interface {
	Name() string
	Run(ctx context.Context) error
}

// Run lanza los componentes y los apaga todos si uno falla o llega SIGINT/SIGTERM.
func Run(ctx context.Context, logger *slog.Logger, components ...Component) error {
	// Una señal del SO cancela ctx aquí.
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Si un componente falla, errgroup cancela ctx para los demás.
	g, ctx := errgroup.WithContext(ctx)

	// Arranca cada componente en su goroutine y registra inicio/parada.
	for _, c := range components {
		g.Go(func() error {
			logger.Info("component starting", "component", c.Name())
			err := c.Run(ctx)
			if err != nil {
				logger.Error("component stopped", "component", c.Name(), "error", err)
				return err
			}
			logger.Info("component stopped", "component", c.Name())
			return nil
		})
	}

	// Devuelve el primer error no nil, o nil si todos se apagaron limpios.
	return g.Wait()
}
