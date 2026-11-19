package ctxhelper

import (
	"context"
)

type contextKey string

const (
	userIDKey        contextKey = "user_id"
	userEmailKey     contextKey = "user_email"
	userRoleKey      contextKey = "user_role"
	correlationIDKey contextKey = "correlation_id"
	requestIDKey     contextKey = "request_id"
)

// WithUserID attaches the current user's ID to the request context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID extracts the current authenticated user's ID from context.
func GetUserID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx.Value(userIDKey).(string); ok {
		return val
	}
	return ""
}

// WithUserEmail attaches the authenticated user's email to context.
func WithUserEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, userEmailKey, email)
}

// GetUserEmail extracts the user's email from context.
func GetUserEmail(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx.Value(userEmailKey).(string); ok {
		return val
	}
	return ""
}

// WithUserRole attaches the user's role to context.
func WithUserRole(ctx context.Context, role string) context.Context {
	return context.WithValue(ctx, userRoleKey, role)
}

// GetUserRole extracts the user's role from context.
func GetUserRole(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx.Value(userRoleKey).(string); ok {
		return val
	}
	return ""
}

// WithCorrelationID attaches the correlation ID to the context.
func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationIDKey, id)
}

// GetCorrelationID extracts the correlation ID from context.
func GetCorrelationID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx.Value(correlationIDKey).(string); ok {
		return val
	}
	return ""
}

// WithRequestID attaches the request ID to the context.
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

// GetRequestID extracts the request ID from context.
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if val, ok := ctx.Value(requestIDKey).(string); ok {
		return val
	}
	return ""
}
