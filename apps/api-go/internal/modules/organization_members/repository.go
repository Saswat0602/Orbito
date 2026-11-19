package organization_members

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository defines data access for organization membership operations.
type Repository interface {
	List(ctx context.Context, orgID uuid.UUID, search string) ([]sqlc.ListOrganizationMembersRow, error)
	ListBySlug(ctx context.Context, slug, search string, limit, offset int32) ([]sqlc.ListOrganizationMembersBySlugRow, error)
	CountBySlug(ctx context.Context, slug, search string) (int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*sqlc.GetOrganizationMemberByIDRow, error)
	GetByUserAndOrg(ctx context.Context, userID, orgID uuid.UUID) (*sqlc.GetOrganizationMemberByUserAndOrgRow, error)
	Create(ctx context.Context, params sqlc.CreateOrganizationMemberParams) (*sqlc.OrganizationMember, error)
	UpdateRole(ctx context.Context, id uuid.UUID, role sqlc.Role, updatedByID uuid.UUID) (*sqlc.OrganizationMember, error)
	Delete(ctx context.Context, id uuid.UUID) error
	SetDefaultOrg(ctx context.Context, userID, orgID uuid.UUID) error
	ClearDefaultOrg(ctx context.Context, userID, orgID uuid.UUID) error
	GetUserGlobalRole(ctx context.Context, userID uuid.UUID) (sqlc.Role, error)
	WithTx(ctx context.Context, fn func(repo Repository) error) error
}

type pgRepository struct {
	db      *database.DB
	queries *sqlc.Queries
}

// NewRepository creates a PostgreSQL-backed repository for organization members.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{
		db:      db,
		queries: db.Queries(),
	}
}

func (r *pgRepository) List(ctx context.Context, orgID uuid.UUID, search string) ([]sqlc.ListOrganizationMembersRow, error) {
	return r.queries.ListOrganizationMembers(ctx, sqlc.ListOrganizationMembersParams{
		OrganizationID: orgID,
		Column2:        search,
	})
}

func (r *pgRepository) ListBySlug(ctx context.Context, slug, search string, limit, offset int32) ([]sqlc.ListOrganizationMembersBySlugRow, error) {
	return r.queries.ListOrganizationMembersBySlug(ctx, sqlc.ListOrganizationMembersBySlugParams{
		Slug:    slug,
		Column2: search,
		Limit:   limit,
		Offset:  offset,
	})
}

func (r *pgRepository) CountBySlug(ctx context.Context, slug, search string) (int64, error) {
	return r.queries.CountOrganizationMembersBySlug(ctx, sqlc.CountOrganizationMembersBySlugParams{
		Slug:    slug,
		Column2: search,
	})
}

func (r *pgRepository) GetByID(ctx context.Context, id uuid.UUID) (*sqlc.GetOrganizationMemberByIDRow, error) {
	row, err := r.queries.GetOrganizationMemberByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetByUserAndOrg(ctx context.Context, userID, orgID uuid.UUID) (*sqlc.GetOrganizationMemberByUserAndOrgRow, error) {
	row, err := r.queries.GetOrganizationMemberByUserAndOrg(ctx, sqlc.GetOrganizationMemberByUserAndOrgParams{
		UserID:         userID,
		OrganizationID: orgID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) Create(ctx context.Context, params sqlc.CreateOrganizationMemberParams) (*sqlc.OrganizationMember, error) {
	m, err := r.queries.CreateOrganizationMember(ctx, params)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *pgRepository) UpdateRole(ctx context.Context, id uuid.UUID, role sqlc.Role, updatedByID uuid.UUID) (*sqlc.OrganizationMember, error) {
	m, err := r.queries.UpdateOrganizationMemberRoleByID(ctx, sqlc.UpdateOrganizationMemberRoleByIDParams{
		ID:          id,
		Role:        role,
		UpdatedByID: pgtype.UUID{Bytes: updatedByID, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *pgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteOrganizationMemberByID(ctx, id)
}

func (r *pgRepository) SetDefaultOrg(ctx context.Context, userID, orgID uuid.UUID) error {
	return r.queries.SetUserDefaultOrgForUser(ctx, sqlc.SetUserDefaultOrgForUserParams{
		ID:                    userID,
		DefaultOrganizationID: pgtype.UUID{Bytes: orgID, Valid: true},
	})
}

func (r *pgRepository) ClearDefaultOrg(ctx context.Context, userID, orgID uuid.UUID) error {
	return r.queries.ClearUserDefaultOrgIfMatches(ctx, sqlc.ClearUserDefaultOrgIfMatchesParams{
		ID:                    userID,
		DefaultOrganizationID: pgtype.UUID{Bytes: orgID, Valid: true},
	})
}

func (r *pgRepository) GetUserGlobalRole(ctx context.Context, userID uuid.UUID) (sqlc.Role, error) {
	u, err := r.queries.GetUserByID(ctx, userID)
	if err != nil {
		return "", err
	}
	return u.Role, nil
}

func (r *pgRepository) WithTx(ctx context.Context, fn func(repo Repository) error) error {
	return r.db.WithTx(ctx, func(q *sqlc.Queries) error {
		txRepo := &pgRepository{
			db:      r.db,
			queries: q,
		}
		return fn(txRepo)
	})
}
