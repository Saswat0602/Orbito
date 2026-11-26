package task_statuses

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository defines persistence operations for task statuses.
type Repository interface {
	GetWorkflowAuth(ctx context.Context, workflowID, userID uuid.UUID) (*sqlc.GetWorkflowOrgOwnerAndUserRoleRow, error)
	GetStatusAuth(ctx context.Context, statusID, userID uuid.UUID) (*sqlc.GetTaskStatusAuthByIDRow, error)
	GetProjectWorkflow(ctx context.Context, projectID uuid.UUID) (*sqlc.GetProjectWorkflowForStatusCreateRow, error)
	FindSoftDeleted(ctx context.Context, workflowID uuid.UUID, name string) (*sqlc.TaskStatus, error)
	RestoreOnCreate(ctx context.Context, id, updatedByID uuid.UUID) (*sqlc.TaskStatus, error)
	CountWorkflowStatuses(ctx context.Context, workflowID uuid.UUID) (int64, error)
	CreateNode(ctx context.Context, arg sqlc.CreateTaskStatusNodeParams) (*sqlc.TaskStatus, error)
	GetFullByID(ctx context.Context, id uuid.UUID) (*sqlc.GetTaskStatusFullByIDRow, error)
	GetActiveFullByID(ctx context.Context, id uuid.UUID) (*sqlc.GetActiveTaskStatusFullByIDRow, error)
	ListFiltered(ctx context.Context, workflowID pgtype.UUID, orgID pgtype.UUID) ([]sqlc.ListTaskStatusesFilteredRow, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]sqlc.ListTaskStatusesByProjectIDRow, error)
	ListFromTransitions(ctx context.Context, statusID uuid.UUID) ([]sqlc.ListFromTransitionsForStatusRow, error)
	ListToTransitions(ctx context.Context, statusID uuid.UUID) ([]sqlc.ListToTransitionsForStatusRow, error)
	ListDeleted(ctx context.Context, workflowID pgtype.UUID) ([]sqlc.ListDeletedTaskStatusesRow, error)
	UpdateByID(ctx context.Context, arg sqlc.UpdateTaskStatusByIDParams) (uuid.UUID, error)
	UpdatePositionsTx(ctx context.Context, updates []PositionUpdate, updatedByID uuid.UUID) ([]sqlc.GetTaskStatusFullByIDRow, error)
	SoftDeleteTx(ctx context.Context, id, workflowID uuid.UUID, position int32, userID uuid.UUID) error
	CheckActiveNameConflict(ctx context.Context, workflowID uuid.UUID, name string, excludeID uuid.UUID) (bool, error)
	RestoreByID(ctx context.Context, id, updatedByID uuid.UUID) (uuid.UUID, error)
}

// PositionUpdate captures a single status position change.
type PositionUpdate struct {
	ID       uuid.UUID
	Position int32
}

type pgRepository struct {
	db      *database.DB
	queries *sqlc.Queries
}

// NewRepository constructs a PostgreSQL task_statuses repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{db: db, queries: db.Queries()}
}

func (r *pgRepository) GetWorkflowAuth(ctx context.Context, workflowID, userID uuid.UUID) (*sqlc.GetWorkflowOrgOwnerAndUserRoleRow, error) {
	row, err := r.queries.GetWorkflowOrgOwnerAndUserRole(ctx, sqlc.GetWorkflowOrgOwnerAndUserRoleParams{
		ID:     workflowID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetStatusAuth(ctx context.Context, statusID, userID uuid.UUID) (*sqlc.GetTaskStatusAuthByIDRow, error) {
	row, err := r.queries.GetTaskStatusAuthByID(ctx, sqlc.GetTaskStatusAuthByIDParams{
		ID:     statusID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetProjectWorkflow(ctx context.Context, projectID uuid.UUID) (*sqlc.GetProjectWorkflowForStatusCreateRow, error) {
	row, err := r.queries.GetProjectWorkflowForStatusCreate(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) FindSoftDeleted(ctx context.Context, workflowID uuid.UUID, name string) (*sqlc.TaskStatus, error) {
	row, err := r.queries.FindSoftDeletedTaskStatus(ctx, sqlc.FindSoftDeletedTaskStatusParams{
		WorkflowID: workflowID,
		Name:       name,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) RestoreOnCreate(ctx context.Context, id, updatedByID uuid.UUID) (*sqlc.TaskStatus, error) {
	row, err := r.queries.RestoreSoftDeletedTaskStatusOnCreate(ctx, sqlc.RestoreSoftDeletedTaskStatusOnCreateParams{
		ID:          id,
		UpdatedByID: pgtype.UUID{Bytes: updatedByID, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) CountWorkflowStatuses(ctx context.Context, workflowID uuid.UUID) (int64, error) {
	return r.queries.CountAllWorkflowStatuses(ctx, workflowID)
}

func (r *pgRepository) CreateNode(ctx context.Context, arg sqlc.CreateTaskStatusNodeParams) (*sqlc.TaskStatus, error) {
	row, err := r.queries.CreateTaskStatusNode(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetFullByID(ctx context.Context, id uuid.UUID) (*sqlc.GetTaskStatusFullByIDRow, error) {
	row, err := r.queries.GetTaskStatusFullByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetActiveFullByID(ctx context.Context, id uuid.UUID) (*sqlc.GetActiveTaskStatusFullByIDRow, error) {
	row, err := r.queries.GetActiveTaskStatusFullByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) ListFiltered(ctx context.Context, workflowID pgtype.UUID, orgID pgtype.UUID) ([]sqlc.ListTaskStatusesFilteredRow, error) {
	return r.queries.ListTaskStatusesFiltered(ctx, sqlc.ListTaskStatusesFilteredParams{
		WorkflowID:     workflowID,
		OrganizationID: orgID,
	})
}

func (r *pgRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]sqlc.ListTaskStatusesByProjectIDRow, error) {
	return r.queries.ListTaskStatusesByProjectID(ctx, projectID)
}

func (r *pgRepository) ListFromTransitions(ctx context.Context, statusID uuid.UUID) ([]sqlc.ListFromTransitionsForStatusRow, error) {
	return r.queries.ListFromTransitionsForStatus(ctx, statusID)
}

func (r *pgRepository) ListToTransitions(ctx context.Context, statusID uuid.UUID) ([]sqlc.ListToTransitionsForStatusRow, error) {
	return r.queries.ListToTransitionsForStatus(ctx, statusID)
}

func (r *pgRepository) ListDeleted(ctx context.Context, workflowID pgtype.UUID) ([]sqlc.ListDeletedTaskStatusesRow, error) {
	return r.queries.ListDeletedTaskStatuses(ctx, workflowID)
}

func (r *pgRepository) UpdateByID(ctx context.Context, arg sqlc.UpdateTaskStatusByIDParams) (uuid.UUID, error) {
	return r.queries.UpdateTaskStatusByID(ctx, arg)
}

func (r *pgRepository) UpdatePositionsTx(ctx context.Context, updates []PositionUpdate, updatedByID uuid.UUID) ([]sqlc.GetTaskStatusFullByIDRow, error) {
	out := make([]sqlc.GetTaskStatusFullByIDRow, 0, len(updates))
	err := r.db.WithTx(ctx, func(q *sqlc.Queries) error {
		for _, item := range updates {
			id, err := q.UpdateTaskStatusPositionByID(ctx, sqlc.UpdateTaskStatusPositionByIDParams{
				ID:          item.ID,
				Position:    item.Position,
				UpdatedByID: pgtype.UUID{Bytes: updatedByID, Valid: true},
			})
			if err != nil {
				return err
			}
			full, err := q.GetTaskStatusFullByID(ctx, id)
			if err != nil {
				return err
			}
			out = append(out, full)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *pgRepository) SoftDeleteTx(ctx context.Context, id, workflowID uuid.UUID, position int32, userID uuid.UUID) error {
	return r.db.WithTx(ctx, func(q *sqlc.Queries) error {
		if err := q.DeleteTransitionsByStatusID(ctx, id); err != nil {
			return err
		}
		if err := q.DecrementStatusPositionsAfterDelete(ctx, sqlc.DecrementStatusPositionsAfterDeleteParams{
			WorkflowID:  workflowID,
			Position:    position,
			UpdatedByID: pgtype.UUID{Bytes: userID, Valid: true},
		}); err != nil {
			return err
		}
		return q.SoftDeleteTaskStatusByID(ctx, sqlc.SoftDeleteTaskStatusByIDParams{
			ID:        id,
			DeletedBy: pgtype.UUID{Bytes: userID, Valid: true},
		})
	})
}

func (r *pgRepository) CheckActiveNameConflict(ctx context.Context, workflowID uuid.UUID, name string, excludeID uuid.UUID) (bool, error) {
	return r.queries.CheckActiveTaskStatusNameConflict(ctx, sqlc.CheckActiveTaskStatusNameConflictParams{
		WorkflowID: workflowID,
		Name:       name,
		ID:         excludeID,
	})
}

func (r *pgRepository) RestoreByID(ctx context.Context, id, updatedByID uuid.UUID) (uuid.UUID, error) {
	return r.queries.RestoreDeletedTaskStatusByID(ctx, sqlc.RestoreDeletedTaskStatusByIDParams{
		ID:          id,
		UpdatedByID: pgtype.UUID{Bytes: updatedByID, Valid: true},
	})
}
