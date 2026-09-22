// Package logger crea el *slog.Logger común para todos los servicios.
package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New crea un logger JSON a stdout con el atributo fijo "service".
func New(service string) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level()})
	return slog.New(handler).With("service", service)
}

// level lee LOG_LEVEL.
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
		// Valor desconocido: error, para no silenciar en producción una
		// configuración mal escrita.
		return slog.LevelError
	}
}
