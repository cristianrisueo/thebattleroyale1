// Package database crea el pool de conexiones a Postgres común a los servicios
package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// pingTimeout limita la espera del Ping de arranque para no colgarse ante un Postgres que no responde
const pingTimeout = 5 * time.Second

// NewPool crea un pool y comprueba con un Ping que la base de datos responde
func NewPool(ctx context.Context, url string) (*pgxpool.Pool, error) {
	// Valida la URL sin abrir todavía ninguna conexión
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}

	// Acota solo la comprobación de arranque, sin tocar el ctx del servicio
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
