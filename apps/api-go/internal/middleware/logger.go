package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/orbito/api-go/internal/platform/logger"
)

// responseWriter is a wrapper around http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// Logger returns an HTTP middleware that logs requests using slog JSON.
func Logger(l *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapped := newResponseWriter(w)

			ctx := r.Context()
			reqLogger := l
			if reqLogger == nil {
				reqLogger = logger.FromContext(ctx)
			}

			correlationID := GetCorrelationID(ctx)
			requestID := GetRequestID(ctx)

			// Attach request-scoped logger to context
			scopedLogger := reqLogger.With(
				slog.String("correlation_id", correlationID),
				slog.String("request_id", requestID),
			)
			ctx = logger.WithContext(ctx, scopedLogger)

			next.ServeHTTP(wrapped, r.WithContext(ctx))

			duration := time.Since(start)

			scopedLogger.Info("http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", wrapped.statusCode),
				slog.Duration("latency", duration),
				slog.Float64("latency_ms", float64(duration.Microseconds())/1000.0),
				slog.Int64("bytes", wrapped.written),
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("user_agent", r.UserAgent()),
			)
		})
	}
}
