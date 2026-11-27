package task_attachments

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// CreateParams holds columns needed to persist an attachment row.
type CreateParams struct {
	ID         uuid.UUID
	FileName   string
	FilePath   *string
	FileSize   int32
	MimeType   string
	URL        *string
	StorageKey *string
	TaskID     uuid.UUID
	CreatedBy  uuid.UUID
}

// Repository defines persistence operations for task attachments.
type Repository interface {
	CheckAccess(ctx context.Context, userID uuid.UUID, taskIdentifier string) (*sqlc.CheckAttachmentAccessTaskAndUserRow, error)
	Create(ctx context.Context, p CreateParams) (uuid.UUID, error)
	GetDetailByID(ctx context.Context, id uuid.UUID) (*sqlc.GetTaskAttachmentDetailByIDRow, error)
	ListByTaskID(ctx context.Context, taskID uuid.UUID) ([]sqlc.ListTaskAttachmentDetailsByTaskIDRow, error)
	GroupStatsByMime(ctx context.Context, taskID uuid.UUID) ([]sqlc.GroupAttachmentStatsByMimeRow, error)
	CountRecent(ctx context.Context, taskID uuid.UUID, since time.Time) (int64, error)
	DeleteByID(ctx context.Context, id uuid.UUID) (int64, error)
}

type pgRepository struct {
	queries *sqlc.Queries
}

// NewRepository constructs a PostgreSQL task_attachments repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{queries: db.Queries()}
}

func (r *pgRepository) CheckAccess(ctx context.Context, userID uuid.UUID, taskIdentifier string) (*sqlc.CheckAttachmentAccessTaskAndUserRow, error) {
	arg := sqlc.CheckAttachmentAccessTaskAndUserParams{UserID: userID}
	if parsed, err := uuid.Parse(taskIdentifier); err == nil {
		arg.IsUuid = true
		arg.TaskUuid = pgtype.UUID{Bytes: parsed, Valid: true}
	} else {
		arg.IsUuid = false
		arg.TaskSlug = taskIdentifier
	}
	row, err := r.queries.CheckAttachmentAccessTaskAndUser(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func optText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func (r *pgRepository) Create(ctx context.Context, p CreateParams) (uuid.UUID, error) {
	return r.queries.CreateTaskAttachmentRecord(ctx, sqlc.CreateTaskAttachmentRecordParams{
		ID:          p.ID,
		FileName:    p.FileName,
		FilePath:    optText(p.FilePath),
		FileSize:    p.FileSize,
		MimeType:    p.MimeType,
		Url:         optText(p.URL),
		StorageKey:  optText(p.StorageKey),
		TaskID:      p.TaskID,
		CreatedByID: pgtype.UUID{Bytes: p.CreatedBy, Valid: true},
	})
}

func (r *pgRepository) GetDetailByID(ctx context.Context, id uuid.UUID) (*sqlc.GetTaskAttachmentDetailByIDRow, error) {
	row, err := r.queries.GetTaskAttachmentDetailByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) ListByTaskID(ctx context.Context, taskID uuid.UUID) ([]sqlc.ListTaskAttachmentDetailsByTaskIDRow, error) {
	return r.queries.ListTaskAttachmentDetailsByTaskID(ctx, taskID)
}

func (r *pgRepository) GroupStatsByMime(ctx context.Context, taskID uuid.UUID) ([]sqlc.GroupAttachmentStatsByMimeRow, error) {
	return r.queries.GroupAttachmentStatsByMime(ctx, taskID)
}

func (r *pgRepository) CountRecent(ctx context.Context, taskID uuid.UUID, since time.Time) (int64, error) {
	return r.queries.CountRecentTaskAttachments(ctx, sqlc.CountRecentTaskAttachmentsParams{
		TaskID:  taskID,
		Column2: pgtype.Timestamp{Time: since, Valid: true},
	})
}

func (r *pgRepository) DeleteByID(ctx context.Context, id uuid.UUID) (int64, error) {
	return r.queries.DeleteTaskAttachmentByID(ctx, id)
}
