package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

// pingTimeout limita la espera del Ping de arranque para no colgarse ante un Redis que no responde
const pingTimeout = 5 * time.Second

// NewClient crea un cliente de Redis instrumentado con trazas y comprueba con un Ping
//
// Param - ctx: Contexto de arranque del servicio, usado solo para el Ping
// Param - addr: Dirección host:puerto de Redis, sin esquema
// Returns - *redis.Client/error: El cliente listo para usar, que debe cerrarse al salir, error si falla
func NewClient(ctx context.Context, addr string) (*redis.Client, error) {
	// Crea el cliente sin conectar todavía: las conexiones se abren bajo demanda
	rdb := redis.NewClient(&redis.Options{Addr: addr})

	// Engancha al cliente un hook que abre un span por cada comando de Redis, dentro de la traza en curso
	if err := redisotel.InstrumentTracing(rdb); err != nil {
		rdb.Close() // Si hay errores cierra el cliente y devuelve el error
		return nil, fmt.Errorf("instrumenting redis tracing: %w", err)
	}

	// Crea un contexto que se cancela en 5 segundos
	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	// Realiza una conexión.
	if err := rdb.Ping(pingCtx).Err(); err != nil {
		rdb.Close() // Si hay errores cierra el cliente y devuelve el error
		return nil, fmt.Errorf("pinging redis: %w", err)
	}

	// Devuelve el cliente de redis
	return rdb, nil
}
