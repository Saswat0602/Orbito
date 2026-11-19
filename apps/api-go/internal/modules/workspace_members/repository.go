package workspace_members

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository manages persistence for workspace members.
type Repository interface {
	List(ctx context.Context, wsID *uuid.UUID, search string, limit, offset int32) ([]sqlc.ListWorkspaceMembersPaginatedRow, error)
	Count(ctx context.Context, wsID *uuid.UUID, search string) (int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*sqlc.GetWorkspaceMemberDetailsByIDRow, error)
	GetByUserAndWorkspace(ctx context.Context, userID, wsID uuid.UUID) (*sqlc.GetWorkspaceMemberByUserAndWorkspaceRow, error)
	Create(ctx context.Context, params sqlc.CreateWorkspaceMemberDirectParams) (*sqlc.WorkspaceMember, error)
	UpdateRole(ctx context.Context, id uuid.UUID, role sqlc.Role, updatedByID uuid.UUID) (*sqlc.WorkspaceMember, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetUserGlobalRole(ctx context.Context, userID uuid.UUID) (sqlc.Role, error)
	WithTx(ctx context.Context, fn func(repo Repository) error) error
}

type pgRepository struct {
	db      *database.DB
	queries *sqlc.Queries
}

// NewRepository constructs a PostgreSQL workspace members repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{
		db:      db,
		queries: db.Queries(),
	}
}

func (r *pgRepository) List(ctx context.Context, wsID *uuid.UUID, search string, limit, offset int32) ([]sqlc.ListWorkspaceMembersPaginatedRow, error) {
	var wsUUID uuid.UUID
	if wsID != nil {
		wsUUID = *wsID
	}
	return r.queries.ListWorkspaceMembersPaginated(ctx, sqlc.ListWorkspaceMembersPaginatedParams{
		Column1: wsUUID,
		Column2: search,
		Limit:   limit,
		Offset:  offset,
	})
}

func (r *pgRepository) Count(ctx context.Context, wsID *uuid.UUID, search string) (int64, error) {
	var wsUUID uuid.UUID
	if wsID != nil {
		wsUUID = *wsID
	}
	return r.queries.CountWorkspaceMembersFiltered(ctx, sqlc.CountWorkspaceMembersFilteredParams{
		Column1: wsUUID,
		Column2: search,
	})
}

func (r *pgRepository) GetByID(ctx context.Context, id uuid.UUID) (*sqlc.GetWorkspaceMemberDetailsByIDRow, error) {
	row, err := r.queries.GetWorkspaceMemberDetailsByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetByUserAndWorkspace(ctx context.Context, userID, wsID uuid.UUID) (*sqlc.GetWorkspaceMemberByUserAndWorkspaceRow, error) {
	row, err := r.queries.GetWorkspaceMemberByUserAndWorkspace(ctx, sqlc.GetWorkspaceMemberByUserAndWorkspaceParams{
		UserID:      userID,
		WorkspaceID: wsID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) Create(ctx context.Context, params sqlc.CreateWorkspaceMemberDirectParams) (*sqlc.WorkspaceMember, error) {
	m, err := r.queries.CreateWorkspaceMemberDirect(ctx, params)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *pgRepository) UpdateRole(ctx context.Context, id uuid.UUID, role sqlc.Role, updatedByID uuid.UUID) (*sqlc.WorkspaceMember, error) {
	m, err := r.queries.UpdateWorkspaceMemberRoleByID(ctx, sqlc.UpdateWorkspaceMemberRoleByIDParams{
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
	return r.queries.DeleteWorkspaceMemberByID(ctx, id)
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
