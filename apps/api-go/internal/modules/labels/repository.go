package labels

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository defines persistence operations for labels.
type Repository interface {
	GetProjectWorkspace(ctx context.Context, projectID, userID uuid.UUID) (*sqlc.GetProjectWorkspaceForLabelRow, error)
	CreateWithWorkspaceTx(ctx context.Context, arg sqlc.CreateLabelRecordParams, workspaceID uuid.UUID) (uuid.UUID, error)
	GetDetailsByID(ctx context.Context, labelID, userID uuid.UUID) (*sqlc.GetLabelDetailsByIDRow, error)
	ListByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]sqlc.ListLabelsByWorkspaceIDRow, error)
	ListAccessibleForUser(ctx context.Context, userID uuid.UUID) ([]sqlc.ListAccessibleLabelsForUserRow, error)
	ListTaskLabelEntries(ctx context.Context, labelID uuid.UUID) ([]sqlc.ListTaskLabelEntriesByLabelIDRow, error)
	ListTaskAssignees(ctx context.Context, taskIDs []uuid.UUID) ([]sqlc.ListTaskAssigneesForTaskIDsRow, error)
	UpdateByID(ctx context.Context, arg sqlc.UpdateLabelByIDParams) (uuid.UUID, error)
	DeleteByID(ctx context.Context, id uuid.UUID) (int64, error)
	FindTaskWorkspaceAccess(ctx context.Context, taskIDOrSlug string, userID uuid.UUID) (*sqlc.FindTaskWorkspaceAccessByIDOrSlugRow, error)
	GetLabelWorkspaceCheck(ctx context.Context, labelID, targetWorkspaceID uuid.UUID) (*sqlc.GetLabelWorkspaceCheckRow, error)
	AssignSingleTx(ctx context.Context, taskID, labelID, workspaceID uuid.UUID) error
	RemoveTaskLabel(ctx context.Context, taskID, labelID uuid.UUID) (int64, error)
	ListWorkspaceValidLabelIDs(ctx context.Context, labelIDs []uuid.UUID, workspaceID uuid.UUID) ([]uuid.UUID, error)
	ReplaceTaskLabelsTx(ctx context.Context, taskID, workspaceID uuid.UUID, labelIDs []uuid.UUID) error
	ListLabelsByTaskID(ctx context.Context, taskID uuid.UUID) ([]sqlc.ListLabelsByTaskIDRow, error)
}

type pgRepository struct {
	db      *database.DB
	queries *sqlc.Queries
}

// NewRepository constructs a PostgreSQL labels repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{db: db, queries: db.Queries()}
}

func (r *pgRepository) GetProjectWorkspace(ctx context.Context, projectID, userID uuid.UUID) (*sqlc.GetProjectWorkspaceForLabelRow, error) {
	row, err := r.queries.GetProjectWorkspaceForLabel(ctx, sqlc.GetProjectWorkspaceForLabelParams{
		ID:     projectID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) CreateWithWorkspaceTx(ctx context.Context, arg sqlc.CreateLabelRecordParams, workspaceID uuid.UUID) (uuid.UUID, error) {
	var createdID uuid.UUID
	err := r.db.WithTx(ctx, func(q *sqlc.Queries) error {
		id, err := q.CreateLabelRecord(ctx, arg)
		if err != nil {
			return err
		}
		createdID = id
		return q.UpsertWorkspaceLabel(ctx, sqlc.UpsertWorkspaceLabelParams{
			WorkspaceID: workspaceID,
			LabelID:     id,
		})
	})
	return createdID, err
}

func (r *pgRepository) GetDetailsByID(ctx context.Context, labelID, userID uuid.UUID) (*sqlc.GetLabelDetailsByIDRow, error) {
	row, err := r.queries.GetLabelDetailsByID(ctx, sqlc.GetLabelDetailsByIDParams{
		ID:     labelID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) ListByWorkspaceID(ctx context.Context, workspaceID uuid.UUID) ([]sqlc.ListLabelsByWorkspaceIDRow, error) {
	return r.queries.ListLabelsByWorkspaceID(ctx, workspaceID)
}

func (r *pgRepository) ListAccessibleForUser(ctx context.Context, userID uuid.UUID) ([]sqlc.ListAccessibleLabelsForUserRow, error) {
	return r.queries.ListAccessibleLabelsForUser(ctx, userID)
}

func (r *pgRepository) ListTaskLabelEntries(ctx context.Context, labelID uuid.UUID) ([]sqlc.ListTaskLabelEntriesByLabelIDRow, error) {
	return r.queries.ListTaskLabelEntriesByLabelID(ctx, labelID)
}

func (r *pgRepository) ListTaskAssignees(ctx context.Context, taskIDs []uuid.UUID) ([]sqlc.ListTaskAssigneesForTaskIDsRow, error) {
	if len(taskIDs) == 0 {
		return nil, nil
	}
	return r.queries.ListTaskAssigneesForTaskIDs(ctx, taskIDs)
}

func (r *pgRepository) UpdateByID(ctx context.Context, arg sqlc.UpdateLabelByIDParams) (uuid.UUID, error) {
	return r.queries.UpdateLabelByID(ctx, arg)
}

func (r *pgRepository) DeleteByID(ctx context.Context, id uuid.UUID) (int64, error) {
	return r.queries.DeleteLabelByID(ctx, id)
}

func (r *pgRepository) FindTaskWorkspaceAccess(ctx context.Context, taskIDOrSlug string, userID uuid.UUID) (*sqlc.FindTaskWorkspaceAccessByIDOrSlugRow, error) {
	params := sqlc.FindTaskWorkspaceAccessByIDOrSlugParams{
		UserID: userID,
	}
	if parsed, err := uuid.Parse(taskIDOrSlug); err == nil {
		params.IsUuid = true
		params.TaskUuid = pgtype.UUID{Bytes: parsed, Valid: true}
	} else {
		params.IsUuid = false
		params.TaskSlug = taskIDOrSlug
	}
	row, err := r.queries.FindTaskWorkspaceAccessByIDOrSlug(ctx, params)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetLabelWorkspaceCheck(ctx context.Context, labelID, targetWorkspaceID uuid.UUID) (*sqlc.GetLabelWorkspaceCheckRow, error) {
	row, err := r.queries.GetLabelWorkspaceCheck(ctx, sqlc.GetLabelWorkspaceCheckParams{
		ID:          labelID,
		WorkspaceID: targetWorkspaceID,
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) AssignSingleTx(ctx context.Context, taskID, labelID, workspaceID uuid.UUID) error {
	return r.db.WithTx(ctx, func(q *sqlc.Queries) error {
		if err := q.AssignTaskLabel(ctx, sqlc.AssignTaskLabelParams{
			TaskID:  taskID,
			LabelID: labelID,
		}); err != nil {
			return err
		}
		return q.UpsertWorkspaceLabel(ctx, sqlc.UpsertWorkspaceLabelParams{
			WorkspaceID: workspaceID,
			LabelID:     labelID,
		})
	})
}

func (r *pgRepository) RemoveTaskLabel(ctx context.Context, taskID, labelID uuid.UUID) (int64, error) {
	return r.queries.RemoveTaskLabel(ctx, sqlc.RemoveTaskLabelParams{
		TaskID:  taskID,
		LabelID: labelID,
	})
}

func (r *pgRepository) ListWorkspaceValidLabelIDs(ctx context.Context, labelIDs []uuid.UUID, workspaceID uuid.UUID) ([]uuid.UUID, error) {
	return r.queries.ListWorkspaceValidLabelIDs(ctx, sqlc.ListWorkspaceValidLabelIDsParams{
		Column1:     labelIDs,
		WorkspaceID: workspaceID,
	})
}

func (r *pgRepository) ReplaceTaskLabelsTx(ctx context.Context, taskID, workspaceID uuid.UUID, labelIDs []uuid.UUID) error {
	return r.db.WithTx(ctx, func(q *sqlc.Queries) error {
		if err := q.DeleteAllTaskLabelsByTaskID(ctx, taskID); err != nil {
			return err
		}
		for _, labelID := range labelIDs {
			if err := q.AssignTaskLabel(ctx, sqlc.AssignTaskLabelParams{
				TaskID:  taskID,
				LabelID: labelID,
			}); err != nil {
				return err
			}
			if err := q.UpsertWorkspaceLabel(ctx, sqlc.UpsertWorkspaceLabelParams{
				WorkspaceID: workspaceID,
				LabelID:     labelID,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *pgRepository) ListLabelsByTaskID(ctx context.Context, taskID uuid.UUID) ([]sqlc.ListLabelsByTaskIDRow, error) {
	return r.queries.ListLabelsByTaskID(ctx, taskID)
}
