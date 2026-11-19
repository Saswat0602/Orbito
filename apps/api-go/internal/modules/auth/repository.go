package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository provides data access for authentication and user account persistence.
type Repository interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (*sqlc.User, error)
	GetUserByEmail(ctx context.Context, email string) (*sqlc.User, error)
	GetUserByUsername(ctx context.Context, username string) (*sqlc.User, error)
	CountSuperAdmins(ctx context.Context) (int64, error)
	CountTotalUsers(ctx context.Context) (int64, error)
	CreateUser(ctx context.Context, params sqlc.CreateUserParams) (*sqlc.User, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	UpdateRole(ctx context.Context, id uuid.UUID, role sqlc.Role) error
	UpdateRefreshToken(ctx context.Context, id uuid.UUID, token pgtype.Text) error
	SetResetToken(ctx context.Context, id uuid.UUID, token pgtype.Text, expiry pgtype.Timestamp) error
	GetUserByResetToken(ctx context.Context, token string) (*sqlc.User, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, hash pgtype.Text) error
	FindInvitationByToken(ctx context.Context, token string) (*sqlc.FindInvitationByTokenRow, error)
	AddOrganizationMember(ctx context.Context, id, userID, orgID uuid.UUID, role sqlc.Role) error
	GetOrganizationByID(ctx context.Context, id uuid.UUID) (*sqlc.Organization, error)
}

type pgRepository struct {
	queries *sqlc.Queries
}

// NewRepository constructs a new PostgreSQL auth repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{queries: db.Queries()}
}

func (r *pgRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*sqlc.User, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	u := mapUserRow(row)
	return &u, nil
}

func (r *pgRepository) GetUserByEmail(ctx context.Context, email string) (*sqlc.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	u := mapUserEmailRow(row)
	return &u, nil
}

func (r *pgRepository) GetUserByUsername(ctx context.Context, username string) (*sqlc.User, error) {
	row, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	u := mapUserUsernameRow(row)
	return &u, nil
}

func (r *pgRepository) CountSuperAdmins(ctx context.Context) (int64, error) {
	return r.queries.CountSuperAdmins(ctx)
}

func (r *pgRepository) CountTotalUsers(ctx context.Context) (int64, error) {
	return r.queries.CountTotalUsers(ctx)
}

func (r *pgRepository) CreateUser(ctx context.Context, params sqlc.CreateUserParams) (*sqlc.User, error) {
	row, err := r.queries.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}
	u := mapCreateUserRow(row)
	return &u, nil
}

func (r *pgRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	return r.queries.UpdateUserLastLogin(ctx, id)
}

func (r *pgRepository) UpdateRole(ctx context.Context, id uuid.UUID, role sqlc.Role) error {
	return r.queries.UpdateUserRole(ctx, sqlc.UpdateUserRoleParams{
		ID:   id,
		Role: role,
	})
}

func (r *pgRepository) UpdateRefreshToken(ctx context.Context, id uuid.UUID, token pgtype.Text) error {
	return r.queries.UpdateUserRefreshToken(ctx, sqlc.UpdateUserRefreshTokenParams{
		ID:           id,
		RefreshToken: token,
	})
}

func (r *pgRepository) SetResetToken(ctx context.Context, id uuid.UUID, token pgtype.Text, expiry pgtype.Timestamp) error {
	return r.queries.SetUserResetToken(ctx, sqlc.SetUserResetTokenParams{
		ID:               id,
		ResetToken:       token,
		ResetTokenExpiry: expiry,
	})
}

func (r *pgRepository) GetUserByResetToken(ctx context.Context, token string) (*sqlc.User, error) {
	row, err := r.queries.GetUserByResetToken(ctx, pgtype.Text{String: token, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	u := mapUserResetRow(row)
	return &u, nil
}

func (r *pgRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hash pgtype.Text) error {
	return r.queries.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		ID:       id,
		Password: hash,
	})
}

func (r *pgRepository) FindInvitationByToken(ctx context.Context, token string) (*sqlc.FindInvitationByTokenRow, error) {
	row, err := r.queries.FindInvitationByToken(ctx, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) AddOrganizationMember(ctx context.Context, id, userID, orgID uuid.UUID, role sqlc.Role) error {
	return r.queries.AddOrganizationMember(ctx, sqlc.AddOrganizationMemberParams{
		ID:             id,
		UserID:         userID,
		OrganizationID: orgID,
		Role:           role,
	})
}

func (r *pgRepository) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*sqlc.Organization, error) {
	row, err := r.queries.GetOrganizationByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &sqlc.Organization{
		ID:   row.ID,
		Name: row.Name,
		Slug: row.Slug,
	}, nil
}

func mapUserRow(r sqlc.GetUserByIDRow) sqlc.User {
	return sqlc.User{
		ID:                    r.ID,
		Email:                 r.Email,
		Username:              r.Username,
		FirstName:             r.FirstName,
		LastName:              r.LastName,
		Avatar:                r.Avatar,
		Bio:                   r.Bio,
		Timezone:              r.Timezone,
		Language:              r.Language,
		MobileNumber:          r.MobileNumber,
		Role:                  r.Role,
		Status:                r.Status,
		Password:              r.Password,
		RefreshToken:          r.RefreshToken,
		ResetToken:            r.ResetToken,
		ResetTokenExpiry:      r.ResetTokenExpiry,
		DefaultOrganizationID: r.DefaultOrganizationID,
		LastLoginAt:           r.LastLoginAt,
		CreatedAt:             r.CreatedAt,
		UpdatedAt:             r.UpdatedAt,
		DeletedAt:             r.DeletedAt,
	}
}

func mapUserEmailRow(r sqlc.GetUserByEmailRow) sqlc.User {
	return sqlc.User{
		ID:                    r.ID,
		Email:                 r.Email,
		Username:              r.Username,
		FirstName:             r.FirstName,
		LastName:              r.LastName,
		Avatar:                r.Avatar,
		Bio:                   r.Bio,
		Timezone:              r.Timezone,
		Language:              r.Language,
		MobileNumber:          r.MobileNumber,
		Role:                  r.Role,
		Status:                r.Status,
		Password:              r.Password,
		RefreshToken:          r.RefreshToken,
		ResetToken:            r.ResetToken,
		ResetTokenExpiry:      r.ResetTokenExpiry,
		DefaultOrganizationID: r.DefaultOrganizationID,
		LastLoginAt:           r.LastLoginAt,
		CreatedAt:             r.CreatedAt,
		UpdatedAt:             r.UpdatedAt,
		DeletedAt:             r.DeletedAt,
	}
}

func mapUserUsernameRow(r sqlc.GetUserByUsernameRow) sqlc.User {
	return sqlc.User{
		ID:                    r.ID,
		Email:                 r.Email,
		Username:              r.Username,
		FirstName:             r.FirstName,
		LastName:              r.LastName,
		Avatar:                r.Avatar,
		Bio:                   r.Bio,
		Timezone:              r.Timezone,
		Language:              r.Language,
		MobileNumber:          r.MobileNumber,
		Role:                  r.Role,
		Status:                r.Status,
		Password:              r.Password,
		RefreshToken:          r.RefreshToken,
		ResetToken:            r.ResetToken,
		ResetTokenExpiry:      r.ResetTokenExpiry,
		DefaultOrganizationID: r.DefaultOrganizationID,
		LastLoginAt:           r.LastLoginAt,
		CreatedAt:             r.CreatedAt,
		UpdatedAt:             r.UpdatedAt,
		DeletedAt:             r.DeletedAt,
	}
}

func mapCreateUserRow(r sqlc.CreateUserRow) sqlc.User {
	return sqlc.User{
		ID:                    r.ID,
		Email:                 r.Email,
		Username:              r.Username,
		FirstName:             r.FirstName,
		LastName:              r.LastName,
		Avatar:                r.Avatar,
		Bio:                   r.Bio,
		Timezone:              r.Timezone,
		Language:              r.Language,
		Role:                  r.Role,
		Status:                r.Status,
		Password:              r.Password,
		RefreshToken:          r.RefreshToken,
		ResetToken:            r.ResetToken,
		ResetTokenExpiry:      r.ResetTokenExpiry,
		DefaultOrganizationID: r.DefaultOrganizationID,
		LastLoginAt:           r.LastLoginAt,
		CreatedAt:             r.CreatedAt,
		UpdatedAt:             r.UpdatedAt,
		DeletedAt:             r.DeletedAt,
	}
}

func mapUserResetRow(r sqlc.GetUserByResetTokenRow) sqlc.User {
	return sqlc.User{
		ID:                    r.ID,
		Email:                 r.Email,
		Username:              r.Username,
		FirstName:             r.FirstName,
		LastName:              r.LastName,
		Avatar:                r.Avatar,
		Bio:                   r.Bio,
		Timezone:              r.Timezone,
		Language:              r.Language,
		Role:                  r.Role,
		Status:                r.Status,
		Password:              r.Password,
		RefreshToken:          r.RefreshToken,
		ResetToken:            r.ResetToken,
		ResetTokenExpiry:      r.ResetTokenExpiry,
		DefaultOrganizationID: r.DefaultOrganizationID,
		LastLoginAt:           r.LastLoginAt,
		CreatedAt:             r.CreatedAt,
		UpdatedAt:             r.UpdatedAt,
		DeletedAt:             r.DeletedAt,
	}
}
