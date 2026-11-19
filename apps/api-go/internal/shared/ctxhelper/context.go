package ctxhelper

import (
	"context"

	"github.com/orbito/api-go/internal/middleware"
)

type contextKey string

const (
	userIDKey    contextKey = "user_id"
	userEmailKey contextKey = "user_email"
	userRoleKey  contextKey = "user_role"
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

// GetCorrelationID delegates to middleware.GetCorrelationID.
func GetCorrelationID(ctx context.Context) string {
	return middleware.GetCorrelationID(ctx)
}
