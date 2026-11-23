package project_members

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository defines data access operations for project members.
type Repository interface {
	ListMembers(ctx context.Context, projectID uuid.UUID, search *string) ([]sqlc.ListProjectMembersRow, error)
	GetByID(ctx context.Context, id uuid.UUID) (*sqlc.GetProjectMemberByIDRow, error)
	GetByUserAndProject(ctx context.Context, userID, projectID uuid.UUID) (*sqlc.GetProjectMemberByUserAndProjectRow, error)
	Create(ctx context.Context, params sqlc.CreateProjectMemberRecordParams) (*sqlc.ProjectMember, error)
	UpdateRole(ctx context.Context, id uuid.UUID, role sqlc.Role, updatedByID uuid.UUID) (*sqlc.ProjectMember, error)
	Delete(ctx context.Context, id uuid.UUID) error
	CountOwners(ctx context.Context, projectID uuid.UUID) (int64, error)
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]sqlc.ListProjectMembersByWorkspaceRow, error)
	ListUserProjects(ctx context.Context, userID uuid.UUID) ([]sqlc.ListUserProjectsRow, error)
	RoleStats(ctx context.Context, projectID uuid.UUID) ([]sqlc.ProjectMembersRoleStatsRow, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*sqlc.GetUserByIDRow, error)
	GetProjectByID(ctx context.Context, id uuid.UUID) (*sqlc.GetProjectByIDRowRow, error)
	WithTx(ctx context.Context, fn func(repo Repository) error) error
}

type pgRepository struct {
	db      *database.DB
	queries *sqlc.Queries
}

// NewRepository constructs a PostgreSQL project members repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{
		db:      db,
		queries: db.Queries(),
	}
}

func (r *pgRepository) ListMembers(ctx context.Context, projectID uuid.UUID, search *string) ([]sqlc.ListProjectMembersRow, error) {
	var sVal pgtype.Text
	if search != nil && *search != "" {
		sVal = pgtype.Text{String: *search, Valid: true}
	}
	return r.queries.ListProjectMembers(ctx, sqlc.ListProjectMembersParams{
		ProjectID: projectID,
		Search:    sVal,
	})
}

func (r *pgRepository) GetByID(ctx context.Context, id uuid.UUID) (*sqlc.GetProjectMemberByIDRow, error) {
	row, err := r.queries.GetProjectMemberByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetByUserAndProject(ctx context.Context, userID, projectID uuid.UUID) (*sqlc.GetProjectMemberByUserAndProjectRow, error) {
	row, err := r.queries.GetProjectMemberByUserAndProject(ctx, sqlc.GetProjectMemberByUserAndProjectParams{
		UserID:    userID,
		ProjectID: projectID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) Create(ctx context.Context, params sqlc.CreateProjectMemberRecordParams) (*sqlc.ProjectMember, error) {
	pm, err := r.queries.CreateProjectMemberRecord(ctx, params)
	if err != nil {
		return nil, err
	}
	return &pm, nil
}

func (r *pgRepository) UpdateRole(ctx context.Context, id uuid.UUID, role sqlc.Role, updatedByID uuid.UUID) (*sqlc.ProjectMember, error) {
	pm, err := r.queries.UpdateProjectMemberRoleByID(ctx, sqlc.UpdateProjectMemberRoleByIDParams{
		ID:          id,
		Role:        role,
		UpdatedByID: pgtype.UUID{Bytes: updatedByID, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &pm, nil
}

func (r *pgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteProjectMemberByID(ctx, id)
}

func (r *pgRepository) CountOwners(ctx context.Context, projectID uuid.UUID) (int64, error) {
	return r.queries.CountProjectOwnersInProject(ctx, projectID)
}

func (r *pgRepository) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]sqlc.ListProjectMembersByWorkspaceRow, error) {
	return r.queries.ListProjectMembersByWorkspace(ctx, workspaceID)
}

func (r *pgRepository) ListUserProjects(ctx context.Context, userID uuid.UUID) ([]sqlc.ListUserProjectsRow, error) {
	return r.queries.ListUserProjects(ctx, userID)
}

func (r *pgRepository) RoleStats(ctx context.Context, projectID uuid.UUID) ([]sqlc.ProjectMembersRoleStatsRow, error) {
	return r.queries.ProjectMembersRoleStats(ctx, projectID)
}

func (r *pgRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*sqlc.GetUserByIDRow, error) {
	u, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *pgRepository) GetProjectByID(ctx context.Context, id uuid.UUID) (*sqlc.GetProjectByIDRowRow, error) {
	p, err := r.queries.GetProjectByIDRow(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
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
