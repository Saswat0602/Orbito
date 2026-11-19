package workspaces

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository defines data access methods for workspaces.
type Repository interface {
	Create(ctx context.Context, params sqlc.CreateWorkspaceParams) (*sqlc.Workspace, error)
	GetByID(ctx context.Context, id uuid.UUID) (*sqlc.GetWorkspaceByIDRow, error)
	GetByOrgAndSlug(ctx context.Context, orgID uuid.UUID, slug string) (*sqlc.GetWorkspaceByOrgAndSlugRow, error)
	GetBySlugOnly(ctx context.Context, slug string) (*sqlc.GetWorkspaceBySlugOnlyRow, error)
	ListForUser(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID, search string) ([]sqlc.ListWorkspacesForUserRow, error)
	ListArchived(ctx context.Context, orgID uuid.UUID) ([]sqlc.Workspace, error)
	ListTree(ctx context.Context, orgID uuid.UUID) ([]sqlc.ListWorkspaceTreeRow, error)
	Update(ctx context.Context, params sqlc.UpdateWorkspaceParams) (*sqlc.Workspace, error)
	Archive(ctx context.Context, id, userID uuid.UUID) error
	Unarchive(ctx context.Context, id, userID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	CheckSlugExists(ctx context.Context, orgID uuid.UUID, slug string, excludeID *uuid.UUID) (bool, error)
	ListMembersDetails(ctx context.Context, workspaceID uuid.UUID) ([]sqlc.ListWorkspaceMembersDetailsRow, error)
	CreateMember(ctx context.Context, params sqlc.CreateWorkspaceMemberDirectParams) (*sqlc.WorkspaceMember, error)
	CountProjects(ctx context.Context, workspaceID uuid.UUID) (int64, error)
	CountProjectsByStatus(ctx context.Context, workspaceID uuid.UUID, status sqlc.ProjectStatus) (int64, error)
	CountTasks(ctx context.Context, workspaceID uuid.UUID) (int64, error)
	CountCompletedTasks(ctx context.Context, workspaceID uuid.UUID) (int64, error)
	CountOverdueTasks(ctx context.Context, workspaceID uuid.UUID) (int64, error)
	ProjectStatusStats(ctx context.Context, workspaceID uuid.UUID) ([]sqlc.WorkspaceProjectStatusStatsRow, error)
	TaskPriorityStats(ctx context.Context, workspaceID uuid.UUID) ([]sqlc.WorkspaceTaskPriorityStatsRow, error)
	WithTx(ctx context.Context, fn func(repo Repository) error) error
}

type pgRepository struct {
	db      *database.DB
	queries *sqlc.Queries
}

// NewRepository constructs a PostgreSQL workspace repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{
		db:      db,
		queries: db.Queries(),
	}
}

func (r *pgRepository) Create(ctx context.Context, params sqlc.CreateWorkspaceParams) (*sqlc.Workspace, error) {
	ws, err := r.queries.CreateWorkspace(ctx, params)
	if err != nil {
		return nil, err
	}
	return &ws, nil
}

func (r *pgRepository) GetByID(ctx context.Context, id uuid.UUID) (*sqlc.GetWorkspaceByIDRow, error) {
	row, err := r.queries.GetWorkspaceByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetByOrgAndSlug(ctx context.Context, orgID uuid.UUID, slug string) (*sqlc.GetWorkspaceByOrgAndSlugRow, error) {
	row, err := r.queries.GetWorkspaceByOrgAndSlug(ctx, sqlc.GetWorkspaceByOrgAndSlugParams{
		OrganizationID: orgID,
		Slug:           slug,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetBySlugOnly(ctx context.Context, slug string) (*sqlc.GetWorkspaceBySlugOnlyRow, error) {
	row, err := r.queries.GetWorkspaceBySlugOnly(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) ListForUser(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID, search string) ([]sqlc.ListWorkspacesForUserRow, error) {
	var orgUUID uuid.UUID
	if orgID != nil {
		orgUUID = *orgID
	}
	return r.queries.ListWorkspacesForUser(ctx, sqlc.ListWorkspacesForUserParams{
		UserID:  userID,
		Column2: orgUUID,
		Column3: search,
	})
}

func (r *pgRepository) ListArchived(ctx context.Context, orgID uuid.UUID) ([]sqlc.Workspace, error) {
	return r.queries.ListArchivedWorkspaces(ctx, orgID)
}

func (r *pgRepository) ListTree(ctx context.Context, orgID uuid.UUID) ([]sqlc.ListWorkspaceTreeRow, error) {
	return r.queries.ListWorkspaceTree(ctx, orgID)
}

func (r *pgRepository) Update(ctx context.Context, params sqlc.UpdateWorkspaceParams) (*sqlc.Workspace, error) {
	ws, err := r.queries.UpdateWorkspace(ctx, params)
	if err != nil {
		return nil, err
	}
	return &ws, nil
}

func (r *pgRepository) Archive(ctx context.Context, id, userID uuid.UUID) error {
	return r.queries.ArchiveWorkspace(ctx, sqlc.ArchiveWorkspaceParams{
		ID:          id,
		UpdatedByID: pgtype.UUID{Bytes: userID, Valid: true},
	})
}

func (r *pgRepository) Unarchive(ctx context.Context, id, userID uuid.UUID) error {
	return r.queries.UnarchiveWorkspace(ctx, sqlc.UnarchiveWorkspaceParams{
		ID:          id,
		UpdatedByID: pgtype.UUID{Bytes: userID, Valid: true},
	})
}

func (r *pgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteWorkspace(ctx, id)
}

func (r *pgRepository) CheckSlugExists(ctx context.Context, orgID uuid.UUID, slug string, excludeID *uuid.UUID) (bool, error) {
	var exc uuid.UUID
	if excludeID != nil {
		exc = *excludeID
	}
	return r.queries.CheckWorkspaceSlugExists(ctx, sqlc.CheckWorkspaceSlugExistsParams{
		OrganizationID: orgID,
		Slug:           slug,
		Column3:        exc,
	})
}

func (r *pgRepository) ListMembersDetails(ctx context.Context, workspaceID uuid.UUID) ([]sqlc.ListWorkspaceMembersDetailsRow, error) {
	return r.queries.ListWorkspaceMembersDetails(ctx, workspaceID)
}

func (r *pgRepository) CreateMember(ctx context.Context, params sqlc.CreateWorkspaceMemberDirectParams) (*sqlc.WorkspaceMember, error) {
	m, err := r.queries.CreateWorkspaceMemberDirect(ctx, params)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *pgRepository) CountProjects(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	return r.queries.CountWorkspaceProjects(ctx, workspaceID)
}

func (r *pgRepository) CountProjectsByStatus(ctx context.Context, workspaceID uuid.UUID, status sqlc.ProjectStatus) (int64, error) {
	return r.queries.CountWorkspaceProjectsByStatus(ctx, sqlc.CountWorkspaceProjectsByStatusParams{
		WorkspaceID: workspaceID,
		Status:      status,
	})
}

func (r *pgRepository) CountTasks(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	return r.queries.CountWorkspaceTasks(ctx, workspaceID)
}

func (r *pgRepository) CountCompletedTasks(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	return r.queries.CountWorkspaceCompletedTasks(ctx, workspaceID)
}

func (r *pgRepository) CountOverdueTasks(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	return r.queries.CountWorkspaceOverdueTasks(ctx, workspaceID)
}

func (r *pgRepository) ProjectStatusStats(ctx context.Context, workspaceID uuid.UUID) ([]sqlc.WorkspaceProjectStatusStatsRow, error) {
	return r.queries.WorkspaceProjectStatusStats(ctx, workspaceID)
}

func (r *pgRepository) TaskPriorityStats(ctx context.Context, workspaceID uuid.UUID) ([]sqlc.WorkspaceTaskPriorityStatsRow, error) {
	return r.queries.WorkspaceTaskPriorityStats(ctx, workspaceID)
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
