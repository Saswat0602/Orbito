package task_dependencies

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository defines database operations for task dependencies.
type Repository interface {
	FindTaskID(ctx context.Context, identifier string) (uuid.UUID, bool, error)
	CheckExists(ctx context.Context, dependentTaskID, blockingTaskID uuid.UUID) (bool, error)
	ListBlockingIDs(ctx context.Context, dependentTaskID uuid.UUID) ([]uuid.UUID, error)
	Create(ctx context.Context, id uuid.UUID, depType sqlc.DependencyType, dependentTaskID, blockingTaskID uuid.UUID, createdBy *uuid.UUID) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*sqlc.GetTaskDependencyDetailsByIDRow, error)
	ListAll(ctx context.Context, projectID *uuid.UUID) ([]sqlc.ListTaskDependenciesRow, error)
	ListDependsOn(ctx context.Context, dependentTaskID uuid.UUID) ([]sqlc.ListDependenciesTaskDependsOnRow, error)
	ListBlocks(ctx context.Context, blockingTaskID uuid.UUID) ([]sqlc.ListDependenciesTaskBlocksRow, error)
	Update(ctx context.Context, id uuid.UUID, depType *sqlc.DependencyType, dependentTaskID, blockingTaskID *uuid.UUID) (uuid.UUID, error)
	DeleteByID(ctx context.Context, id uuid.UUID) (int64, error)
	DeleteByPair(ctx context.Context, dependentTaskID, blockingTaskID uuid.UUID) (int64, error)
}

type pgRepository struct {
	queries *sqlc.Queries
}

// NewRepository constructs a PostgreSQL task_dependencies repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{queries: db.Queries()}
}

func (r *pgRepository) FindTaskID(ctx context.Context, identifier string) (uuid.UUID, bool, error) {
	var arg sqlc.FindTaskIDForDependencyParams
	if parsed, err := uuid.Parse(identifier); err == nil {
		arg.IsUuid = true
		arg.TaskUuid = pgtype.UUID{Bytes: parsed, Valid: true}
	} else {
		arg.IsUuid = false
		arg.TaskSlug = identifier
	}
	id, err := r.queries.FindTaskIDForDependency(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, false, nil
		}
		return uuid.Nil, false, err
	}
	return id, true, nil
}

func (r *pgRepository) CheckExists(ctx context.Context, dependentTaskID, blockingTaskID uuid.UUID) (bool, error) {
	return r.queries.CheckDependencyExists(ctx, sqlc.CheckDependencyExistsParams{
		DependentTaskID: dependentTaskID,
		BlockingTaskID:  blockingTaskID,
	})
}

func (r *pgRepository) ListBlockingIDs(ctx context.Context, dependentTaskID uuid.UUID) ([]uuid.UUID, error) {
	return r.queries.ListBlockingIDsForTask(ctx, dependentTaskID)
}

func (r *pgRepository) Create(ctx context.Context, id uuid.UUID, depType sqlc.DependencyType, dependentTaskID, blockingTaskID uuid.UUID, createdBy *uuid.UUID) (uuid.UUID, error) {
	var cb pgtype.UUID
	if createdBy != nil {
		cb = pgtype.UUID{Bytes: *createdBy, Valid: true}
	}
	return r.queries.CreateTaskDependencyRecord(ctx, sqlc.CreateTaskDependencyRecordParams{
		ID:              id,
		Type:            depType,
		DependentTaskID: dependentTaskID,
		BlockingTaskID:  blockingTaskID,
		CreatedByID:     cb,
	})
}

func (r *pgRepository) GetByID(ctx context.Context, id uuid.UUID) (*sqlc.GetTaskDependencyDetailsByIDRow, error) {
	row, err := r.queries.GetTaskDependencyDetailsByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) ListAll(ctx context.Context, projectID *uuid.UUID) ([]sqlc.ListTaskDependenciesRow, error) {
	var pid pgtype.UUID
	if projectID != nil {
		pid = pgtype.UUID{Bytes: *projectID, Valid: true}
	}
	return r.queries.ListTaskDependencies(ctx, pid)
}

func (r *pgRepository) ListDependsOn(ctx context.Context, dependentTaskID uuid.UUID) ([]sqlc.ListDependenciesTaskDependsOnRow, error) {
	return r.queries.ListDependenciesTaskDependsOn(ctx, dependentTaskID)
}

func (r *pgRepository) ListBlocks(ctx context.Context, blockingTaskID uuid.UUID) ([]sqlc.ListDependenciesTaskBlocksRow, error) {
	return r.queries.ListDependenciesTaskBlocks(ctx, blockingTaskID)
}

func (r *pgRepository) Update(ctx context.Context, id uuid.UUID, depType *sqlc.DependencyType, dependentTaskID, blockingTaskID *uuid.UUID) (uuid.UUID, error) {
	var arg sqlc.UpdateTaskDependencyByIDParams
	arg.ID = id
	if depType != nil {
		arg.Type = sqlc.NullDependencyType{DependencyType: *depType, Valid: true}
	}
	if dependentTaskID != nil {
		arg.DependentTaskID = pgtype.UUID{Bytes: *dependentTaskID, Valid: true}
	}
	if blockingTaskID != nil {
		arg.BlockingTaskID = pgtype.UUID{Bytes: *blockingTaskID, Valid: true}
	}
	return r.queries.UpdateTaskDependencyByID(ctx, arg)
}

func (r *pgRepository) DeleteByID(ctx context.Context, id uuid.UUID) (int64, error) {
	return r.queries.DeleteTaskDependencyByID(ctx, id)
}

func (r *pgRepository) DeleteByPair(ctx context.Context, dependentTaskID, blockingTaskID uuid.UUID) (int64, error) {
	return r.queries.DeleteTaskDependencyByPair(ctx, sqlc.DeleteTaskDependencyByPairParams{
		DependentTaskID: dependentTaskID,
		BlockingTaskID:  blockingTaskID,
	})
}
