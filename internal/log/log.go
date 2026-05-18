// Package log provides a structured JSON logger with correlation ID support.
package log

import (
	"context"
	"log/slog"
	"os"
)

type contextKey struct{}

// New creates a JSON handler logger at the level specified by LOG_LEVEL env.
// Valid levels: debug, info, warn, error. Defaults to info.
func New() *slog.Logger {
	level := slog.LevelInfo
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
}

// WithCorrelationID returns a logger that includes the correlation ID in every log entry.
// If no correlation ID is in the context, the key is omitted.
func WithCorrelationID(ctx context.Context, logger *slog.Logger) *slog.Logger {
	if id, ok := ctx.Value(contextKey{}).(string); ok && id != "" {
		return logger.With("corr", id)
	}
	return logger
}

// SetCorrelationID stores a correlation ID in the context.
func SetCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, contextKey{}, id)
}

// CorrelationID returns the correlation ID from the context, or empty string if not set.
func CorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(contextKey{}).(string); ok {
		return id
	}
	return ""
}
