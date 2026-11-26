package sprints

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository defines persistence operations for sprints.
type Repository interface {
	GetProjectAccessByID(ctx context.Context, projectID, userID uuid.UUID) (*sqlc.GetProjectAccessInfoForSprintRow, error)
	GetProjectAccessBySlug(ctx context.Context, slug string, userID uuid.UUID) (*sqlc.GetProjectAccessInfoForSprintBySlugRow, error)
	CheckSlugExists(ctx context.Context, projectID uuid.UUID, slug string, excludeID *uuid.UUID) (bool, error)
	UpdateSlugByID(ctx context.Context, sprintID uuid.UUID, slug string) error
	HasActiveSprint(ctx context.Context, projectID uuid.UUID, excludeID *uuid.UUID) (bool, error)
	Create(ctx context.Context, arg sqlc.CreateSprintRecordParams) (uuid.UUID, error)
	GetDetailsByID(ctx context.Context, id uuid.UUID) (*sqlc.GetSprintDetailsByIDRow, error)
	GetByProjectAndSlug(ctx context.Context, projectSlug, sprintSlug string) (*sqlc.GetSprintByProjectAndSlugRow, error)
	GetActiveByProjectID(ctx context.Context, projectID uuid.UUID) (*sqlc.GetActiveSprintByProjectIDRow, error)
	ListFiltered(ctx context.Context, projectID pgtype.UUID, projectSlug, status pgtype.Text) ([]sqlc.ListSprintsFilteredRow, error)
	ListSprintTasks(ctx context.Context, sprintID uuid.UUID) ([]sqlc.ListSprintTasksBySprintIDRow, error)
	ListTaskAssignees(ctx context.Context, taskIDs []uuid.UUID) ([]sqlc.ListSprintTaskAssigneesRow, error)
	ListTaskReporters(ctx context.Context, taskIDs []uuid.UUID) ([]sqlc.ListSprintTaskReportersRow, error)
	UpdateByID(ctx context.Context, arg sqlc.UpdateSprintByIDParams) (uuid.UUID, error)
	DeleteByID(ctx context.Context, id uuid.UUID) (int64, error)
	ArchiveByID(ctx context.Context, id uuid.UUID) (int64, error)
	LogActivity(ctx context.Context, arg sqlc.CreateSprintActivityLogParams) error
}

type pgRepository struct {
	queries *sqlc.Queries
}

// NewRepository constructs a PostgreSQL sprints repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{queries: db.Queries()}
}

func (r *pgRepository) GetProjectAccessByID(ctx context.Context, projectID, userID uuid.UUID) (*sqlc.GetProjectAccessInfoForSprintRow, error) {
	row, err := r.queries.GetProjectAccessInfoForSprint(ctx, sqlc.GetProjectAccessInfoForSprintParams{
		ID:     projectID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetProjectAccessBySlug(ctx context.Context, slug string, userID uuid.UUID) (*sqlc.GetProjectAccessInfoForSprintBySlugRow, error) {
	row, err := r.queries.GetProjectAccessInfoForSprintBySlug(ctx, sqlc.GetProjectAccessInfoForSprintBySlugParams{
		Slug:   slug,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) CheckSlugExists(ctx context.Context, projectID uuid.UUID, slug string, excludeID *uuid.UUID) (bool, error) {
	var ex pgtype.UUID
	if excludeID != nil {
		ex = pgtype.UUID{Bytes: *excludeID, Valid: true}
	}
	return r.queries.CheckSprintSlugExists(ctx, sqlc.CheckSprintSlugExistsParams{
		ProjectID: projectID,
		Slug:      pgtype.Text{String: slug, Valid: true},
		ExcludeID: ex,
	})
}

func (r *pgRepository) UpdateSlugByID(ctx context.Context, sprintID uuid.UUID, slug string) error {
	return r.queries.UpdateSprintSlugByID(ctx, sqlc.UpdateSprintSlugByIDParams{
		ID:   sprintID,
		Slug: pgtype.Text{String: slug, Valid: true},
	})
}

func (r *pgRepository) HasActiveSprint(ctx context.Context, projectID uuid.UUID, excludeID *uuid.UUID) (bool, error) {
	var ex pgtype.UUID
	if excludeID != nil {
		ex = pgtype.UUID{Bytes: *excludeID, Valid: true}
	}
	return r.queries.HasActiveSprintInProject(ctx, sqlc.HasActiveSprintInProjectParams{
		ProjectID: projectID,
		ExcludeID: ex,
	})
}

func (r *pgRepository) Create(ctx context.Context, arg sqlc.CreateSprintRecordParams) (uuid.UUID, error) {
	return r.queries.CreateSprintRecord(ctx, arg)
}

func (r *pgRepository) GetDetailsByID(ctx context.Context, id uuid.UUID) (*sqlc.GetSprintDetailsByIDRow, error) {
	row, err := r.queries.GetSprintDetailsByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetByProjectAndSlug(ctx context.Context, projectSlug, sprintSlug string) (*sqlc.GetSprintByProjectAndSlugRow, error) {
	row, err := r.queries.GetSprintByProjectAndSlug(ctx, sqlc.GetSprintByProjectAndSlugParams{
		Slug:   projectSlug,
		Slug_2: pgtype.Text{String: sprintSlug, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetActiveByProjectID(ctx context.Context, projectID uuid.UUID) (*sqlc.GetActiveSprintByProjectIDRow, error) {
	row, err := r.queries.GetActiveSprintByProjectID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) ListFiltered(ctx context.Context, projectID pgtype.UUID, projectSlug, status pgtype.Text) ([]sqlc.ListSprintsFilteredRow, error) {
	return r.queries.ListSprintsFiltered(ctx, sqlc.ListSprintsFilteredParams{
		ProjectID:   projectID,
		ProjectSlug: projectSlug,
		Status:      status,
	})
}

func (r *pgRepository) ListSprintTasks(ctx context.Context, sprintID uuid.UUID) ([]sqlc.ListSprintTasksBySprintIDRow, error) {
	return r.queries.ListSprintTasksBySprintID(ctx, pgtype.UUID{Bytes: sprintID, Valid: true})
}

func (r *pgRepository) ListTaskAssignees(ctx context.Context, taskIDs []uuid.UUID) ([]sqlc.ListSprintTaskAssigneesRow, error) {
	if len(taskIDs) == 0 {
		return nil, nil
	}
	return r.queries.ListSprintTaskAssignees(ctx, taskIDs)
}

func (r *pgRepository) ListTaskReporters(ctx context.Context, taskIDs []uuid.UUID) ([]sqlc.ListSprintTaskReportersRow, error) {
	if len(taskIDs) == 0 {
		return nil, nil
	}
	return r.queries.ListSprintTaskReporters(ctx, taskIDs)
}

func (r *pgRepository) UpdateByID(ctx context.Context, arg sqlc.UpdateSprintByIDParams) (uuid.UUID, error) {
	return r.queries.UpdateSprintByID(ctx, arg)
}

func (r *pgRepository) DeleteByID(ctx context.Context, id uuid.UUID) (int64, error) {
	return r.queries.DeleteSprintByID(ctx, id)
}

func (r *pgRepository) ArchiveByID(ctx context.Context, id uuid.UUID) (int64, error) {
	return r.queries.ArchiveSprintByID(ctx, id)
}

func (r *pgRepository) LogActivity(ctx context.Context, arg sqlc.CreateSprintActivityLogParams) error {
	return r.queries.CreateSprintActivityLog(ctx, arg)
}

