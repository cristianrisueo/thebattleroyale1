// Package logger crea el logger JSON común a todos los servicios
package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New crea el logger raíz de un servicio con su nombre en cada línea
func NewLogger(service string) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level()})
	return slog.New(handler).With("service", service)
}

// level traduce LOG_LEVEL al nivel mínimo que se imprime, info si no está definido
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
	// Deja el servicio casi mudo ante un valor mal escrito: solo se ven los errores
	default:
		return slog.LevelError
	}
}
