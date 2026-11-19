package projects

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository defines data access operations for projects.
type Repository interface {
	Create(ctx context.Context, params sqlc.CreateProjectParams) (*sqlc.Project, error)
	GetByID(ctx context.Context, id uuid.UUID) (*sqlc.GetProjectByIDRowRow, error)
	GetBySlug(ctx context.Context, slug string) (*sqlc.GetProjectBySlugRowRow, error)
	GetByKey(ctx context.Context, workspaceID uuid.UUID, key string) (*sqlc.GetProjectByKeyRowRow, error)
	ListForUser(ctx context.Context, params sqlc.ListProjectsForUserParams) ([]sqlc.ListProjectsForUserRow, error)
	CountForUser(ctx context.Context, params sqlc.CountProjectsForUserParams) (int64, error)
	ListByOrganization(ctx context.Context, params sqlc.ListProjectsByOrganizationParams) ([]sqlc.ListProjectsByOrganizationRow, error)
	CountByOrganization(ctx context.Context, params sqlc.CountProjectsByOrganizationParams) (int64, error)
	ListArchivedByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]sqlc.ListArchivedProjectsByWorkspaceRow, error)
	ListArchivedByOrganization(ctx context.Context, organizationID uuid.UUID) ([]sqlc.ListArchivedProjectsByOrganizationRow, error)
	Update(ctx context.Context, params sqlc.UpdateProjectParams) (*sqlc.Project, error)
	Archive(ctx context.Context, id, userID uuid.UUID) error
	Unarchive(ctx context.Context, id, userID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	CheckSlugExists(ctx context.Context, slug string, excludeID *uuid.UUID) (bool, error)
	ListSlugsStartingWith(ctx context.Context, prefix string) ([]string, error)
	GetDefaultWorkflowForOrg(ctx context.Context, orgID uuid.UUID) (*sqlc.Workflow, error)
	GetWorkflowByID(ctx context.Context, id uuid.UUID) (*sqlc.Workflow, error)
	CreateInitialSprint(ctx context.Context, params sqlc.CreateInitialSprintParams) (*sqlc.Sprint, error)
	CreateProjectMember(ctx context.Context, params sqlc.CreateProjectMemberDirectParams) (*sqlc.ProjectMember, error)
	GetWorkspaceByID(ctx context.Context, id uuid.UUID) (*sqlc.GetWorkspaceByIDRow, error)
	GetOrganizationByID(ctx context.Context, id uuid.UUID) (*sqlc.GetOrganizationDetailsByIDRow, error)
	CountTasks(ctx context.Context, id uuid.UUID) (int64, error)
	CountCompletedTasks(ctx context.Context, id uuid.UUID) (int64, error)
	CountMembers(ctx context.Context, id uuid.UUID) (int64, error)
	TaskStatusStats(ctx context.Context, id uuid.UUID) ([]sqlc.ProjectTaskStatusStatsRow, error)
	TaskPriorityStats(ctx context.Context, id uuid.UUID) ([]sqlc.ProjectTaskPriorityStatsRow, error)
	WithTx(ctx context.Context, fn func(repo Repository) error) error
}

type pgRepository struct {
	db      *database.DB
	queries *sqlc.Queries
}

// NewRepository constructs a PostgreSQL projects repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{
		db:      db,
		queries: db.Queries(),
	}
}

func (r *pgRepository) Create(ctx context.Context, params sqlc.CreateProjectParams) (*sqlc.Project, error) {
	p, err := r.queries.CreateProject(ctx, params)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *pgRepository) GetByID(ctx context.Context, id uuid.UUID) (*sqlc.GetProjectByIDRowRow, error) {
	row, err := r.queries.GetProjectByIDRow(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetBySlug(ctx context.Context, slug string) (*sqlc.GetProjectBySlugRowRow, error) {
	row, err := r.queries.GetProjectBySlugRow(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetByKey(ctx context.Context, workspaceID uuid.UUID, key string) (*sqlc.GetProjectByKeyRowRow, error) {
	row, err := r.queries.GetProjectByKeyRow(ctx, sqlc.GetProjectByKeyRowParams{
		WorkspaceID: workspaceID,
		TaskPrefix:  pgtype.Text{String: key, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) ListForUser(ctx context.Context, params sqlc.ListProjectsForUserParams) ([]sqlc.ListProjectsForUserRow, error) {
	return r.queries.ListProjectsForUser(ctx, params)
}

func (r *pgRepository) CountForUser(ctx context.Context, params sqlc.CountProjectsForUserParams) (int64, error) {
	return r.queries.CountProjectsForUser(ctx, params)
}

func (r *pgRepository) ListByOrganization(ctx context.Context, params sqlc.ListProjectsByOrganizationParams) ([]sqlc.ListProjectsByOrganizationRow, error) {
	return r.queries.ListProjectsByOrganization(ctx, params)
}

func (r *pgRepository) CountByOrganization(ctx context.Context, params sqlc.CountProjectsByOrganizationParams) (int64, error) {
	return r.queries.CountProjectsByOrganization(ctx, params)
}

func (r *pgRepository) ListArchivedByWorkspace(ctx context.Context, workspaceID uuid.UUID) ([]sqlc.ListArchivedProjectsByWorkspaceRow, error) {
	return r.queries.ListArchivedProjectsByWorkspace(ctx, workspaceID)
}

func (r *pgRepository) ListArchivedByOrganization(ctx context.Context, organizationID uuid.UUID) ([]sqlc.ListArchivedProjectsByOrganizationRow, error) {
	return r.queries.ListArchivedProjectsByOrganization(ctx, organizationID)
}

func (r *pgRepository) Update(ctx context.Context, params sqlc.UpdateProjectParams) (*sqlc.Project, error) {
	p, err := r.queries.UpdateProject(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *pgRepository) Archive(ctx context.Context, id, userID uuid.UUID) error {
	return r.queries.ArchiveProject(ctx, sqlc.ArchiveProjectParams{
		ID:          id,
		UpdatedByID: pgtype.UUID{Bytes: userID, Valid: true},
	})
}

func (r *pgRepository) Unarchive(ctx context.Context, id, userID uuid.UUID) error {
	return r.queries.UnarchiveProject(ctx, sqlc.UnarchiveProjectParams{
		ID:          id,
		UpdatedByID: pgtype.UUID{Bytes: userID, Valid: true},
	})
}

func (r *pgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteProject(ctx, id)
}

func (r *pgRepository) CheckSlugExists(ctx context.Context, slug string, excludeID *uuid.UUID) (bool, error) {
	count, err := r.queries.CheckProjectSlugExists(ctx, sqlc.CheckProjectSlugExistsParams{
		Slug:      slug,
		ExcludeID: toPgUUID(excludeID),
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *pgRepository) ListSlugsStartingWith(ctx context.Context, prefix string) ([]string, error) {
	return r.queries.ListProjectSlugsStartingWith(ctx, pgtype.Text{String: prefix, Valid: true})
}

func (r *pgRepository) GetDefaultWorkflowForOrg(ctx context.Context, orgID uuid.UUID) (*sqlc.Workflow, error) {
	wf, err := r.queries.GetDefaultWorkflowForOrg(ctx, orgID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &wf, nil
}

func (r *pgRepository) GetWorkflowByID(ctx context.Context, id uuid.UUID) (*sqlc.Workflow, error) {
	wf, err := r.queries.GetWorkflowByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &wf, nil
}

func (r *pgRepository) CreateInitialSprint(ctx context.Context, params sqlc.CreateInitialSprintParams) (*sqlc.Sprint, error) {
	s, err := r.queries.CreateInitialSprint(ctx, params)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *pgRepository) CreateProjectMember(ctx context.Context, params sqlc.CreateProjectMemberDirectParams) (*sqlc.ProjectMember, error) {
	m, err := r.queries.CreateProjectMemberDirect(ctx, params)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *pgRepository) GetWorkspaceByID(ctx context.Context, id uuid.UUID) (*sqlc.GetWorkspaceByIDRow, error) {
	ws, err := r.queries.GetWorkspaceByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &ws, nil
}

func (r *pgRepository) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*sqlc.GetOrganizationDetailsByIDRow, error) {
	org, err := r.queries.GetOrganizationDetailsByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &org, nil
}

func (r *pgRepository) CountTasks(ctx context.Context, id uuid.UUID) (int64, error) {
	return r.queries.CountProjectTasks(ctx, id)
}

func (r *pgRepository) CountCompletedTasks(ctx context.Context, id uuid.UUID) (int64, error) {
	return r.queries.CountProjectCompletedTasks(ctx, id)
}

func (r *pgRepository) CountMembers(ctx context.Context, id uuid.UUID) (int64, error) {
	return r.queries.CountProjectMembers(ctx, id)
}

func (r *pgRepository) TaskStatusStats(ctx context.Context, id uuid.UUID) ([]sqlc.ProjectTaskStatusStatsRow, error) {
	return r.queries.ProjectTaskStatusStats(ctx, id)
}

func (r *pgRepository) TaskPriorityStats(ctx context.Context, id uuid.UUID) ([]sqlc.ProjectTaskPriorityStatsRow, error) {
	return r.queries.ProjectTaskPriorityStats(ctx, id)
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

func toPgUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}
