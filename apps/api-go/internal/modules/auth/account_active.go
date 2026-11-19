package auth

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// AssertAccountActive verifies that an account exists, is ACTIVE, and is not soft-deleted.
// Matches exact NestJS error message: "Account is not active" with 401 Unauthorized.
func AssertAccountActive(status sqlc.UserStatus, deletedAt pgtype.Timestamp) error {
	if status != sqlc.UserStatusACTIVE || deletedAt.Valid {
		return apperror.Unauthorized("Account is not active")
	}
	return nil
}

// AssertNotSystemUser prevents the system user UUID from authenticating.
func AssertNotSystemUser(id uuid.UUID) error {
	if id == SystemUserID {
		return apperror.Unauthorized("Invalid user account")
	}
	return nil
}
