// Package database crea el pool de conexiones a Postgres común a los servicios
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pingTimeout limita la espera del Ping de arranque para no colgarse ante un Postgres que no responde
const pingTimeout = 5 * time.Second

// NewPool crea un pool y comprueba con un Ping que la base de datos responde
func NewPool(ctx context.Context, url string) (*pgxpool.Pool, error) {
	// Valida la URL sin abrir todavía ninguna conexión
	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("parsing database url: %w", err)
	}

	// Crea un span por consulta, colgado de la traza que venga en el ctx de la llamada
	config.ConnConfig.Tracer = otelpgx.NewTracer()

	// Prepara el pool con la configuración ya instrumentada
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}

	// Crea un contexto que se cancela en 5 segundos
	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	// Fuerza una conexión real para que un Postgres caído sea un fallo de arranque
	if err := pool.Ping(pingCtx); err != nil {
		// Cierra el pool para no dejar conexiones abiertas que nadie usará
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return pool, nil
}
