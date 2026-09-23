// Package database crea el pool de conexiones a Postgres común a los servicios.
//
// Un pool no es una conexión, sino un grupo de conexiones ya abiertas que se
// reparte entre las peticiones concurrentes del servicio: abrir una conexión
// por petición sería caro (Postgres crea un proceso por conexión) y agotaría
// el límite del servidor en cuanto llegara un pico de tráfico.

package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// pingTimeout acota la espera del Ping de arranque: sin él, un Postgres que acepta
// la conexión pero no responde dejaría el servicio colgado sin arrancar ni fallar.
const pingTimeout = 5 * time.Second

// NewPool devuelve un pool ya comprobado contra la base de datos.
//
// pgxpool abre las conexiones de forma perezosa, así que un pool recién creado
// parece correcto aunque Postgres esté caído; el fallo aparecería en la primera
// consulta de un cliente. El Ping fuerza una conexión real ahora, para que una
// base de datos inaccesible sea un fallo de arranque y no un error en caliente.
func NewPool(ctx context.Context, url string) (*pgxpool.Pool, error) {
	// Solo valida la URL y prepara la estructura: todavía no habla con Postgres.
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}

	// El límite es solo para esta comprobación; el ctx del servicio no se toca.
	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	// Abre una conexión de verdad y espera respuesta: aquí se cae si Postgres no está.
	if err := pool.Ping(pingCtx); err != nil {
		// Sin este Close quedarían conexiones abiertas que ya nadie va a usar.
		pool.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return pool, nil
}
