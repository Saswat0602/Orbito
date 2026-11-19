package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
)

type contextKey string

const loggerContextKey contextKey = "logger"

// New creates a new structured JSON logger writing to the provided writer (defaulting to os.Stdout).
func New(isProduction bool, w ...io.Writer) *slog.Logger {
	var out io.Writer = os.Stdout
	if len(w) > 0 && w[0] != nil {
		out = w[0]
	}

	level := slog.LevelDebug
	if isProduction {
		level = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level:     level,
		AddSource: !isProduction,
	})

	return slog.New(handler)
}

// WithContext returns a new context with the logger attached.
func WithContext(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, l)
}

// FromContext extracts the logger from the context or returns the default slog logger.
func FromContext(ctx context.Context) *slog.Logger {
	if ctx == nil {
		return slog.Default()
	}
	if l, ok := ctx.Value(loggerContextKey).(*slog.Logger); ok && l != nil {
		return l
	}
	return slog.Default()
}
