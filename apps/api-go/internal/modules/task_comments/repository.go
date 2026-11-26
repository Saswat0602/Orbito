package task_comments

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository defines database operations for task comments.
type Repository interface {
	CheckAccess(ctx context.Context, userID uuid.UUID, taskIdentifier string) (*sqlc.CheckCommentAccessTaskAndUserRow, error)
	GetBasicByID(ctx context.Context, id uuid.UUID) (*sqlc.GetCommentBasicByIDRow, error)
	Create(ctx context.Context, id uuid.UUID, content string, taskID, authorID uuid.UUID, parentCommentID *uuid.UUID) (uuid.UUID, error)
	GetDetailByID(ctx context.Context, id uuid.UUID) (*sqlc.GetTaskCommentDetailByIDRow, error)
	CountTopLevel(ctx context.Context, taskID uuid.UUID) (int64, error)
	ListTopLevelAsc(ctx context.Context, taskID uuid.UUID, limit, offset int32) ([]sqlc.ListTopLevelTaskCommentsAscRow, error)
	ListTopLevelDesc(ctx context.Context, taskID uuid.UUID, limit, offset int32) ([]sqlc.ListTopLevelTaskCommentsDescRow, error)
	ListRepliesByParentIDs(ctx context.Context, parentIDs []uuid.UUID) ([]sqlc.ListRepliesByParentIDsRow, error)
	UpdateContent(ctx context.Context, id uuid.UUID, content string) (uuid.UUID, error)
	DeleteByID(ctx context.Context, id uuid.UUID) (int64, error)
}

type pgRepository struct {
	queries *sqlc.Queries
}

// NewRepository constructs a PostgreSQL task_comments repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{queries: db.Queries()}
}

func (r *pgRepository) CheckAccess(ctx context.Context, userID uuid.UUID, taskIdentifier string) (*sqlc.CheckCommentAccessTaskAndUserRow, error) {
	arg := sqlc.CheckCommentAccessTaskAndUserParams{UserID: userID}
	if parsed, err := uuid.Parse(taskIdentifier); err == nil {
		arg.IsUuid = true
		arg.TaskUuid = pgtype.UUID{Bytes: parsed, Valid: true}
	} else {
		arg.IsUuid = false
		arg.TaskSlug = taskIdentifier
	}
	row, err := r.queries.CheckCommentAccessTaskAndUser(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetBasicByID(ctx context.Context, id uuid.UUID) (*sqlc.GetCommentBasicByIDRow, error) {
	row, err := r.queries.GetCommentBasicByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) Create(ctx context.Context, id uuid.UUID, content string, taskID, authorID uuid.UUID, parentCommentID *uuid.UUID) (uuid.UUID, error) {
	var parent pgtype.UUID
	if parentCommentID != nil {
		parent = pgtype.UUID{Bytes: *parentCommentID, Valid: true}
	}
	return r.queries.CreateTaskCommentRecord(ctx, sqlc.CreateTaskCommentRecordParams{
		ID:              id,
		Content:         content,
		TaskID:          taskID,
		AuthorID:        authorID,
		ParentCommentID: parent,
	})
}

func (r *pgRepository) GetDetailByID(ctx context.Context, id uuid.UUID) (*sqlc.GetTaskCommentDetailByIDRow, error) {
	row, err := r.queries.GetTaskCommentDetailByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) CountTopLevel(ctx context.Context, taskID uuid.UUID) (int64, error) {
	return r.queries.CountTopLevelTaskComments(ctx, taskID)
}

func (r *pgRepository) ListTopLevelAsc(ctx context.Context, taskID uuid.UUID, limit, offset int32) ([]sqlc.ListTopLevelTaskCommentsAscRow, error) {
	return r.queries.ListTopLevelTaskCommentsAsc(ctx, sqlc.ListTopLevelTaskCommentsAscParams{
		TaskID: taskID,
		Limit:  limit,
		Offset: offset,
	})
}

func (r *pgRepository) ListTopLevelDesc(ctx context.Context, taskID uuid.UUID, limit, offset int32) ([]sqlc.ListTopLevelTaskCommentsDescRow, error) {
	return r.queries.ListTopLevelTaskCommentsDesc(ctx, sqlc.ListTopLevelTaskCommentsDescParams{
		TaskID: taskID,
		Limit:  limit,
		Offset: offset,
	})
}

func (r *pgRepository) ListRepliesByParentIDs(ctx context.Context, parentIDs []uuid.UUID) ([]sqlc.ListRepliesByParentIDsRow, error) {
	if len(parentIDs) == 0 {
		return nil, nil
	}
	return r.queries.ListRepliesByParentIDs(ctx, parentIDs)
}

func (r *pgRepository) UpdateContent(ctx context.Context, id uuid.UUID, content string) (uuid.UUID, error) {
	return r.queries.UpdateTaskCommentContentByID(ctx, sqlc.UpdateTaskCommentContentByIDParams{
		ID:      id,
		Content: content,
	})
}

func (r *pgRepository) DeleteByID(ctx context.Context, id uuid.UUID) (int64, error) {
	return r.queries.DeleteTaskCommentByID(ctx, id)
}
