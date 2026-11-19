package users

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository defines data access methods for users module.
type Repository interface {
	ListUsers(ctx context.Context, limit, offset int32) ([]sqlc.ListUsersRow, error)
	CountUsers(ctx context.Context) (int64, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*sqlc.User, error)
	GetUserByEmail(ctx context.Context, email string) (*sqlc.User, error)
	GetUserByUsername(ctx context.Context, username string) (*sqlc.User, error)
	CreateUser(ctx context.Context, params sqlc.CreateUserParams) (*sqlc.User, error)
	CreateUserWithProfile(ctx context.Context, params sqlc.CreateUserWithProfileParams) (*sqlc.CreateUserWithProfileRow, error)
	UpdateUserProfile(ctx context.Context, params sqlc.UpdateUserProfileParams) (*sqlc.UpdateUserProfileRow, error)
	UpdateUserStatus(ctx context.Context, id uuid.UUID, status sqlc.UserStatus) (*sqlc.UpdateUserStatusRow, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, hash pgtype.Text) error
	DeleteUser(ctx context.Context, id, deletedBy uuid.UUID) error
	CheckUsersShareOrganization(ctx context.Context, u1, u2 uuid.UUID) (bool, error)
	GetPublicUserByID(ctx context.Context, id uuid.UUID) (*sqlc.GetPublicUserByIDRow, error)
	GetPublicUserByUsername(ctx context.Context, username string) (*sqlc.GetPublicUserByUsernameRow, error)
}

type pgRepository struct {
	queries *sqlc.Queries
}

// NewRepository creates a new PostgreSQL user repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{queries: db.Queries()}
}

func (r *pgRepository) ListUsers(ctx context.Context, limit, offset int32) ([]sqlc.ListUsersRow, error) {
	return r.queries.ListUsers(ctx, sqlc.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
}

func (r *pgRepository) CountUsers(ctx context.Context) (int64, error) {
	return r.queries.CountTotalUsers(ctx)
}

func (r *pgRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*sqlc.User, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &sqlc.User{
		ID:                    row.ID,
		Email:                 row.Email,
		Username:              row.Username,
		FirstName:             row.FirstName,
		LastName:              row.LastName,
		Avatar:                row.Avatar,
		Bio:                   row.Bio,
		Timezone:              row.Timezone,
		Language:              row.Language,
		MobileNumber:          row.MobileNumber,
		Role:                  row.Role,
		Status:                row.Status,
		Password:              row.Password,
		RefreshToken:          row.RefreshToken,
		ResetToken:            row.ResetToken,
		ResetTokenExpiry:      row.ResetTokenExpiry,
		DefaultOrganizationID: row.DefaultOrganizationID,
		LastLoginAt:           row.LastLoginAt,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
		DeletedAt:             row.DeletedAt,
	}, nil
}

func (r *pgRepository) GetUserByEmail(ctx context.Context, email string) (*sqlc.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &sqlc.User{
		ID:                    row.ID,
		Email:                 row.Email,
		Username:              row.Username,
		FirstName:             row.FirstName,
		LastName:              row.LastName,
		Avatar:                row.Avatar,
		Bio:                   row.Bio,
		Timezone:              row.Timezone,
		Language:              row.Language,
		MobileNumber:          row.MobileNumber,
		Role:                  row.Role,
		Status:                row.Status,
		Password:              row.Password,
		RefreshToken:          row.RefreshToken,
		ResetToken:            row.ResetToken,
		ResetTokenExpiry:      row.ResetTokenExpiry,
		DefaultOrganizationID: row.DefaultOrganizationID,
		LastLoginAt:           row.LastLoginAt,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
		DeletedAt:             row.DeletedAt,
	}, nil
}

func (r *pgRepository) GetUserByUsername(ctx context.Context, username string) (*sqlc.User, error) {
	row, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &sqlc.User{
		ID:                    row.ID,
		Email:                 row.Email,
		Username:              row.Username,
		FirstName:             row.FirstName,
		LastName:              row.LastName,
		Avatar:                row.Avatar,
		Bio:                   row.Bio,
		Timezone:              row.Timezone,
		Language:              row.Language,
		MobileNumber:          row.MobileNumber,
		Role:                  row.Role,
		Status:                row.Status,
		Password:              row.Password,
		RefreshToken:          row.RefreshToken,
		ResetToken:            row.ResetToken,
		ResetTokenExpiry:      row.ResetTokenExpiry,
		DefaultOrganizationID: row.DefaultOrganizationID,
		LastLoginAt:           row.LastLoginAt,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
		DeletedAt:             row.DeletedAt,
	}, nil
}

func (r *pgRepository) CreateUser(ctx context.Context, params sqlc.CreateUserParams) (*sqlc.User, error) {
	row, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}
	return &sqlc.User{
		ID:                    row.ID,
		Email:                 row.Email,
		Username:              row.Username,
		FirstName:             row.FirstName,
		LastName:              row.LastName,
		Avatar:                row.Avatar,
		Bio:                   row.Bio,
		Timezone:              row.Timezone,
		Language:              row.Language,
		Role:                  row.Role,
		Status:                row.Status,
		Password:              row.Password,
		RefreshToken:          row.RefreshToken,
		ResetToken:            row.ResetToken,
		ResetTokenExpiry:      row.ResetTokenExpiry,
		DefaultOrganizationID: row.DefaultOrganizationID,
		LastLoginAt:           row.LastLoginAt,
		CreatedAt:             row.CreatedAt,
		UpdatedAt:             row.UpdatedAt,
		DeletedAt:             row.DeletedAt,
	}, nil
}

func (r *pgRepository) CreateUserWithProfile(ctx context.Context, params sqlc.CreateUserWithProfileParams) (*sqlc.CreateUserWithProfileRow, error) {
	row, err := r.queries.CreateUserWithProfile(ctx, params)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) CheckUsersShareOrganization(ctx context.Context, u1, u2 uuid.UUID) (bool, error) {
	return r.queries.CheckUsersShareOrganization(ctx, sqlc.CheckUsersShareOrganizationParams{
		UserID:   u1,
		UserID_2: u2,
	})
}

func (r *pgRepository) GetPublicUserByID(ctx context.Context, id uuid.UUID) (*sqlc.GetPublicUserByIDRow, error) {
	row, err := r.queries.GetPublicUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetPublicUserByUsername(ctx context.Context, username string) (*sqlc.GetPublicUserByUsernameRow, error) {
	row, err := r.queries.GetPublicUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) UpdateUserProfile(ctx context.Context, params sqlc.UpdateUserProfileParams) (*sqlc.UpdateUserProfileRow, error) {
	row, err := r.queries.UpdateUserProfile(ctx, params)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) UpdateUserStatus(ctx context.Context, id uuid.UUID, status sqlc.UserStatus) (*sqlc.UpdateUserStatusRow, error) {
	row, err := r.queries.UpdateUserStatus(ctx, sqlc.UpdateUserStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hash pgtype.Text) error {
	return r.queries.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		ID:       id,
		Password: hash,
	})
}

func (r *pgRepository) DeleteUser(ctx context.Context, id, deletedBy uuid.UUID) error {
	return r.queries.DeleteUserSoft(ctx, sqlc.DeleteUserSoftParams{
		ID:        id,
		DeletedBy: pgtype.UUID{Bytes: deletedBy, Valid: true},
	})
}
