package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// UserChecker verifies the existence and status of a user from persistence.
type UserChecker func(ctx context.Context, id uuid.UUID) (*sqlc.User, error)

// TokenValidator parses and returns validated claims from a token string.
type TokenValidator interface {
	ValidateToken(tokenStr string) (sub string, email string, role string, err error)
}

// SystemUserID represents the system user UUID barred from authenticating.
var SystemUserID = uuid.MustParse("00000000-0000-0000-0000-000000000000")

// RequireAuth returns an HTTP middleware verifying JWT access tokens from Authorization header or cookie.
func RequireAuth(validator TokenValidator, checker UserChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenStr := extractToken(r)
			if tokenStr == "" {
				httpx.RespondError(w, r, apperror.Unauthorized("No authorization token provided"))
				return
			}

			sub, email, role, err := validator.ValidateToken(tokenStr)
			if err != nil {
				httpx.RespondError(w, r, apperror.Unauthorized("Invalid or expired token"))
				return
			}

			userUUID, err := uuid.Parse(sub)
			if err != nil {
				httpx.RespondError(w, r, apperror.Unauthorized("Invalid user identity"))
				return
			}

			if userUUID == SystemUserID {
				httpx.RespondError(w, r, apperror.Unauthorized("Invalid user account"))
				return
			}

			if checker != nil {
				user, checkErr := checker(r.Context(), userUUID)
				if checkErr != nil || user == nil {
					httpx.RespondError(w, r, apperror.Unauthorized("User not found"))
					return
				}

				if user.Status != sqlc.UserStatusACTIVE || user.DeletedAt.Valid {
					httpx.RespondError(w, r, apperror.Unauthorized("Account is not active"))
					return
				}
				role = string(user.Role)
				email = user.Email
			}

			ctx := r.Context()
			ctx = ctxhelper.WithUserID(ctx, sub)
			ctx = ctxhelper.WithUserEmail(ctx, email)
			ctx = ctxhelper.WithUserRole(ctx, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
			return strings.TrimSpace(parts[1])
		}
	}

	cookie, err := r.Cookie("access_token")
	if err == nil && cookie.Value != "" {
		return strings.TrimSpace(cookie.Value)
	}

	return ""
}

// AssertAccountActive checks user status and soft delete.
func AssertAccountActive(status sqlc.UserStatus, deletedAt pgtype.Timestamp) error {
	if status != sqlc.UserStatusACTIVE || deletedAt.Valid {
		return apperror.Unauthorized("Account is not active")
	}
	return nil
}
