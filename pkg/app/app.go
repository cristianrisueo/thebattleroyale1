// Package app gestiona el arranque y el apagado común de los componentes de un servicio
package app

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"
)

// Component representa una pieza del servicio que corre en segundo plano
type Component interface {
	// Name devuelve el nombre que identifica la pieza en los logs
	Name() string

	// Run bloquea hasta que ctx se cancela y devuelve nil solo si el apagado fue limpio
	Run(ctx context.Context) error
}

// Run arranca los componentes y espera a que todos paren
func Run(ctx context.Context, logger *slog.Logger, components ...Component) error {
	// Cancela ctx al recibir SIGINT o SIGTERM
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)

	// Libera la captura de señales al salir
	defer stop()

	// Cancela ctx también cuando falla cualquier componente
	g, ctx := errgroup.WithContext(ctx)

	// Lanza cada componente en su propia goroutine porque su Run bloquea
	for _, c := range components {
		g.Go(func() error {
			logger.Info("component starting", "component", c.Name())

			// Devuelve el error para que el grupo cancele al resto de componentes
			err := c.Run(ctx)
			if err != nil {
				logger.Error("component stopped", "component", c.Name(), "error", err)
				return err
			}

			logger.Info("component stopped", "component", c.Name())
			return nil
		})
	}

	// Espera a que todos terminen y devuelve el primer error no nil
	return g.Wait()
}
