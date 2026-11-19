package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
)

// HTTP headers used for distributed tracing and correlation.
const (
	HeaderXCorrelationID = "x-correlation-id"
	HeaderXRequestID     = "x-request-id"
)

// RequestID middleware extracts or generates correlation and request IDs.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		correlationID := r.Header.Get(HeaderXCorrelationID)
		if correlationID == "" {
			correlationID = r.Header.Get(HeaderXRequestID)
		}
		if correlationID == "" {
			correlationID = uuid.NewString()
		}

		requestID := r.Header.Get(HeaderXRequestID)
		if requestID == "" {
			requestID = correlationID
		}

		ctx := r.Context()
		ctx = ctxhelper.WithCorrelationID(ctx, correlationID)
		ctx = ctxhelper.WithRequestID(ctx, requestID)

		// Mirror NestJS RequestContextInterceptor header behavior
		w.Header().Set(HeaderXCorrelationID, correlationID)
		w.Header().Set(HeaderXRequestID, requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// WithCorrelationID injects a correlation ID into context (primarily for testing and synthetic requests).
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return ctxhelper.WithCorrelationID(ctx, correlationID)
}

// GetCorrelationID extracts correlation ID from context.
func GetCorrelationID(ctx context.Context) string {
	return ctxhelper.GetCorrelationID(ctx)
}

// GetRequestID extracts request ID from context.
func GetRequestID(ctx context.Context) string {
	return ctxhelper.GetRequestID(ctx)
}
