// Package database crea el pool de conexiones a Postgres común a los servicios.
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// pingTimeout acota cuánto se espera a que la base de datos responda al arrancar: 5s.
const pingTimeout = 5 * time.Second

// NewPool crea el pool y comprueba con un Ping que la base de datos responde.
func NewPool(ctx context.Context, url string) (*pgxpool.Pool, error) {
	// Crea el pool pero no abre conexión real todavía; por eso hace falta el Ping.
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}

	// Límite de tiempo solo para el ping, no para todo ctx.
	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	// Comprueba que la base de datos responde de verdad antes de devolver el pool.
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close() // sin esto, el pool quedaría abierto sin nadie que lo use.
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return pool, nil
}
