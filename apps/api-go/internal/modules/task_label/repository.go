package task_label

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository defines persistence operations for task label assignments.
type Repository interface {
	FindTask(ctx context.Context, identifier string, userID uuid.UUID) (*sqlc.FindTaskForTaskLabelRow, error)
	GetLabelWorkspace(ctx context.Context, labelID, workspaceID uuid.UUID) (*sqlc.GetLabelWorkspaceForTaskLabelRow, error)
	FindLabelsInWorkspace(ctx context.Context, labelIDs []uuid.UUID, workspaceID uuid.UUID) ([]uuid.UUID, error)
	CheckAssignmentExists(ctx context.Context, taskID, labelID uuid.UUID) (bool, error)
	AssignSingleTx(ctx context.Context, taskID, labelID, workspaceID, userID uuid.UUID) (*sqlc.GetTaskLabelPairDetailsRow, error)
	ReplaceMultipleTx(ctx context.Context, taskID uuid.UUID, labelIDs []uuid.UUID, workspaceID, userID uuid.UUID) ([]sqlc.GetTaskLabelPairDetailsRow, error)
	ListAccessibleProjectIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	ListByProjectIDs(ctx context.Context, projectIDs []uuid.UUID) ([]sqlc.ListTaskLabelsByProjectIDsRow, error)
	GetPairDetails(ctx context.Context, taskID, labelID uuid.UUID) (*sqlc.GetTaskLabelPairDetailsRow, error)
	DeletePair(ctx context.Context, taskID, labelID uuid.UUID) (int64, error)
}

type pgRepository struct {
	db      *database.DB
	queries *sqlc.Queries
}

// NewRepository constructs a PostgreSQL task_label repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{db: db, queries: db.Queries()}
}

func (r *pgRepository) FindTask(ctx context.Context, identifier string, userID uuid.UUID) (*sqlc.FindTaskForTaskLabelRow, error) {
	arg := sqlc.FindTaskForTaskLabelParams{UserID: userID}
	if parsed, err := uuid.Parse(identifier); err == nil {
		arg.IsUuid = true
		arg.TaskUuid = pgtype.UUID{Bytes: parsed, Valid: true}
	} else {
		arg.IsUuid = false
		arg.TaskSlug = identifier
	}
	row, err := r.queries.FindTaskForTaskLabel(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetLabelWorkspace(ctx context.Context, labelID, workspaceID uuid.UUID) (*sqlc.GetLabelWorkspaceForTaskLabelRow, error) {
	row, err := r.queries.GetLabelWorkspaceForTaskLabel(ctx, sqlc.GetLabelWorkspaceForTaskLabelParams{
		ID:          labelID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) FindLabelsInWorkspace(ctx context.Context, labelIDs []uuid.UUID, workspaceID uuid.UUID) ([]uuid.UUID, error) {
	return r.queries.FindLabelsInWorkspaceByIDs(ctx, sqlc.FindLabelsInWorkspaceByIDsParams{
		Column1: labelIDs,
		Column2: workspaceID,
	})
}

func (r *pgRepository) CheckAssignmentExists(ctx context.Context, taskID, labelID uuid.UUID) (bool, error) {
	return r.queries.CheckTaskLabelAssignmentExists(ctx, sqlc.CheckTaskLabelAssignmentExistsParams{
		TaskID:  taskID,
		LabelID: labelID,
	})
}

func (r *pgRepository) AssignSingleTx(ctx context.Context, taskID, labelID, workspaceID, userID uuid.UUID) (*sqlc.GetTaskLabelPairDetailsRow, error) {
	var out sqlc.GetTaskLabelPairDetailsRow
	err := r.db.WithTx(ctx, func(q *sqlc.Queries) error {
		if insErr := q.InsertTaskLabelWithAudit(ctx, sqlc.InsertTaskLabelWithAuditParams{
			TaskID:      taskID,
			LabelID:     labelID,
			CreatedByID: pgtype.UUID{Bytes: userID, Valid: true},
		}); insErr != nil {
			return insErr
		}
		if wsErr := q.UpsertWorkspaceLabelForTaskLabel(ctx, sqlc.UpsertWorkspaceLabelForTaskLabelParams{
			WorkspaceID: workspaceID,
			LabelID:     labelID,
		}); wsErr != nil {
			return wsErr
		}
		det, getErr := q.GetTaskLabelPairDetails(ctx, sqlc.GetTaskLabelPairDetailsParams{
			TaskID:  taskID,
			LabelID: labelID,
		})
		if getErr != nil {
			return getErr
		}
		out = det
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *pgRepository) ReplaceMultipleTx(ctx context.Context, taskID uuid.UUID, labelIDs []uuid.UUID, workspaceID, userID uuid.UUID) ([]sqlc.GetTaskLabelPairDetailsRow, error) {
	results := make([]sqlc.GetTaskLabelPairDetailsRow, 0, len(labelIDs))
	err := r.db.WithTx(ctx, func(q *sqlc.Queries) error {
		if delErr := q.DeleteTaskLabelsByTaskID(ctx, taskID); delErr != nil {
			return delErr
		}
		actor := pgtype.UUID{Bytes: userID, Valid: true}
		for _, lid := range labelIDs {
			if insErr := q.InsertTaskLabelWithAudit(ctx, sqlc.InsertTaskLabelWithAuditParams{
				TaskID:      taskID,
				LabelID:     lid,
				CreatedByID: actor,
			}); insErr != nil {
				return insErr
			}
			if wsErr := q.UpsertWorkspaceLabelForTaskLabel(ctx, sqlc.UpsertWorkspaceLabelForTaskLabelParams{
				WorkspaceID: workspaceID,
				LabelID:     lid,
			}); wsErr != nil {
				return wsErr
			}
			det, getErr := q.GetTaskLabelPairDetails(ctx, sqlc.GetTaskLabelPairDetailsParams{
				TaskID:  taskID,
				LabelID: lid,
			})
			if getErr != nil {
				return getErr
			}
			results = append(results, det)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (r *pgRepository) ListAccessibleProjectIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return r.queries.ListAccessibleProjectIDsForTaskLabels(ctx, userID)
}

func (r *pgRepository) ListByProjectIDs(ctx context.Context, projectIDs []uuid.UUID) ([]sqlc.ListTaskLabelsByProjectIDsRow, error) {
	return r.queries.ListTaskLabelsByProjectIDs(ctx, projectIDs)
}

func (r *pgRepository) GetPairDetails(ctx context.Context, taskID, labelID uuid.UUID) (*sqlc.GetTaskLabelPairDetailsRow, error) {
	row, err := r.queries.GetTaskLabelPairDetails(ctx, sqlc.GetTaskLabelPairDetailsParams{
		TaskID:  taskID,
		LabelID: labelID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) DeletePair(ctx context.Context, taskID, labelID uuid.UUID) (int64, error) {
	return r.queries.DeleteTaskLabelPair(ctx, sqlc.DeleteTaskLabelPairParams{
		TaskID:  taskID,
		LabelID: labelID,
	})
}
