package gantt

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository defines persistence operations for Gantt charts and resource workload.
type Repository interface {
	GetProjectBasic(ctx context.Context, projectID uuid.UUID) (*sqlc.GetProjectBasicForGanttRow, error)
	GetSprintBasic(ctx context.Context, sprintID uuid.UUID) (*sqlc.GetSprintBasicForGanttRow, error)
	ListSprintsByProject(ctx context.Context, projectID uuid.UUID) ([]sqlc.ListSprintsForGanttByProjectRow, error)
	ListProjectTasks(ctx context.Context, projectID uuid.UUID) ([]sqlc.ListTasksForProjectGanttRow, error)
	ListSprintTasks(ctx context.Context, sprintID uuid.UUID) ([]sqlc.ListTasksForSprintGanttRow, error)
	ListAssigneesByTaskIDs(ctx context.Context, taskIDs []uuid.UUID) ([]sqlc.ListTaskAssigneesForGanttRow, error)
	ListDependenciesByProject(ctx context.Context, projectID uuid.UUID) ([]sqlc.ListTaskDependenciesForGanttRow, error)
	ListResourceTasks(ctx context.Context, projectID uuid.UUID) ([]sqlc.ListResourceTasksForProjectGanttRow, error)
}

type pgRepository struct {
	queries *sqlc.Queries
}

// NewRepository constructs a PostgreSQL gantt repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{queries: db.Queries()}
}

func (r *pgRepository) GetProjectBasic(ctx context.Context, projectID uuid.UUID) (*sqlc.GetProjectBasicForGanttRow, error) {
	row, err := r.queries.GetProjectBasicForGantt(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetSprintBasic(ctx context.Context, sprintID uuid.UUID) (*sqlc.GetSprintBasicForGanttRow, error) {
	row, err := r.queries.GetSprintBasicForGantt(ctx, sprintID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) ListSprintsByProject(ctx context.Context, projectID uuid.UUID) ([]sqlc.ListSprintsForGanttByProjectRow, error) {
	return r.queries.ListSprintsForGanttByProject(ctx, projectID)
}

func (r *pgRepository) ListProjectTasks(ctx context.Context, projectID uuid.UUID) ([]sqlc.ListTasksForProjectGanttRow, error) {
	return r.queries.ListTasksForProjectGantt(ctx, projectID)
}

func (r *pgRepository) ListSprintTasks(ctx context.Context, sprintID uuid.UUID) ([]sqlc.ListTasksForSprintGanttRow, error) {
	return r.queries.ListTasksForSprintGantt(ctx, pgtype.UUID{Bytes: sprintID, Valid: true})
}

func (r *pgRepository) ListAssigneesByTaskIDs(ctx context.Context, taskIDs []uuid.UUID) ([]sqlc.ListTaskAssigneesForGanttRow, error) {
	if len(taskIDs) == 0 {
		return nil, nil
	}
	return r.queries.ListTaskAssigneesForGantt(ctx, taskIDs)
}

func (r *pgRepository) ListDependenciesByProject(ctx context.Context, projectID uuid.UUID) ([]sqlc.ListTaskDependenciesForGanttRow, error) {
	return r.queries.ListTaskDependenciesForGantt(ctx, projectID)
}

func (r *pgRepository) ListResourceTasks(ctx context.Context, projectID uuid.UUID) ([]sqlc.ListResourceTasksForProjectGanttRow, error) {
	return r.queries.ListResourceTasksForProjectGantt(ctx, projectID)
}
