package apperror

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Map translates any Go error into a standardized domain AppError.
func Map(err error) *AppError {
	if err == nil {
		return nil
	}

	// 1. Existing AppError
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	// 2. pgx.ErrNoRows (matches Prisma P2025)
	if errors.Is(err, pgx.ErrNoRows) {
		return NotFound("Requested record was not found.")
	}

	// 3. PostgreSQL Database specific errors
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation (matches Prisma P2002)
			msg := "A record with this unique value already exists."
			if pgErr.Detail != "" {
				msg = fmt.Sprintf("Duplicate entry detected on unique field(s): %s", pgErr.Detail)
			}
			return &AppError{
				StatusCode: 409,
				ErrorName:  "Conflict",
				Message:    msg,
				Err:        err,
			}
		case "23503": // foreign_key_violation (matches Prisma P2003)
			msg := "Foreign key reference constraint failed."
			if pgErr.Detail != "" {
				msg = fmt.Sprintf("Foreign key reference constraint failed on: %s", pgErr.Detail)
			}
			return &AppError{
				StatusCode: 400,
				ErrorName:  "ForeignKeyViolation",
				Message:    msg,
				Err:        err,
			}
		default:
			return &AppError{
				StatusCode: 400,
				ErrorName:  "DatabaseError",
				Message:    "A database operation error occurred.",
				Err:        err,
			}
		}
	}

	// 4. Fallback: Generic unexpected internal error
	return Internal("An unexpected internal error occurred.", err)
}
