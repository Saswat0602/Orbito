package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const (
	correlationIDKey contextKey = "correlation_id"
	requestIDKey     contextKey = "request_id"
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
		ctx = context.WithValue(ctx, correlationIDKey, correlationID)
		ctx = context.WithValue(ctx, requestIDKey, requestID)

		// Mirror NestJS RequestContextInterceptor header behavior
		w.Header().Set(HeaderXCorrelationID, correlationID)
		w.Header().Set(HeaderXRequestID, requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// WithCorrelationID injects a correlation ID into context (primarily for testing and synthetic requests).
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey, correlationID)
}

// GetCorrelationID extracts correlation ID from context.
func GetCorrelationID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx.Value(correlationIDKey).(string); ok {
		return val
	}
	return ""
}

// GetRequestID extracts request ID from context.
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx.Value(requestIDKey).(string); ok {
		return val
	}
	return ""
}
