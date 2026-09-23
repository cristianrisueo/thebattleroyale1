// Package logger crea el *slog.Logger común para todos los servicios.
//
// La salida es JSON porque los logs se leen filtrando por campos, no leyendo
// líneas: con varias instancias de un mismo servicio consumiendo de Kafka, la
// salida de todas se mezcla y "service" es lo que permite separarlas.

package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New devuelve el logger raíz de un servicio, con su nombre en cada línea.
//
// El service que se le pasa aquí es el único atributo garantizado en toda la
// salida del proceso. A partir de la fase de observabilidad se le añadirá el
// identificador de traza, para poder saltar de un log a su traza distribuida.
func New(service string) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level()})
	return slog.New(handler).With("service", service)
}

// level traduce LOG_LEVEL al nivel mínimo que se imprime; sin definir, info.
func level() slog.Level {
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "", "info":
		return slog.LevelInfo
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		// Un valor mal escrito deja el servicio casi mudo: solo se verán los errores.
		return slog.LevelError
	}
}
