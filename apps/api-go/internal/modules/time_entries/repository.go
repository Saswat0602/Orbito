package time_entries

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

// CreateParams holds columns required to insert a time entry.
type CreateParams struct {
	ID          uuid.UUID
	TaskID      uuid.UUID
	UserID      uuid.UUID
	TimeSpent   int32
	Description *string
	StartTime   *time.Time
	EndTime     *time.Time
	Date        time.Time
}

// UpdateParams holds optional update fields for a time entry.
type UpdateParams struct {
	ID          uuid.UUID
	Description *string
	TimeSpent   *int32
	StartTime   *time.Time
	EndTime     *time.Time
	Date        *time.Time
}

// Repository defines persistence operations for time entries.
type Repository interface {
	CheckTaskAccess(ctx context.Context, taskID, userID uuid.UUID) (*sqlc.CheckTimeEntryTaskProjectAccessRow, error)
	Create(ctx context.Context, p CreateParams) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*sqlc.GetTimeEntryDetailByIDRow, error)
	ListFiltered(ctx context.Context, taskID, userID *uuid.UUID, startDate, endDate *time.Time) ([]sqlc.ListTimeEntriesFilteredRow, error)
	GetActiveTimer(ctx context.Context, userID uuid.UUID) (*sqlc.GetActiveTimerForUserRow, error)
	Update(ctx context.Context, p UpdateParams) (uuid.UUID, error)
	StopTimer(ctx context.Context, id uuid.UUID, endTime time.Time, timeSpent int32, description *string) (uuid.UUID, error)
	Delete(ctx context.Context, id uuid.UUID) (int64, error)
}

type pgRepository struct {
	queries *sqlc.Queries
}

// NewRepository constructs a PostgreSQL time_entries repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{queries: db.Queries()}
}

func optText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func optTS(t *time.Time) pgtype.Timestamp {
	if t == nil {
		return pgtype.Timestamp{}
	}
	return pgtype.Timestamp{Time: *t, Valid: true}
}

func optUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func (r *pgRepository) CheckTaskAccess(ctx context.Context, taskID, userID uuid.UUID) (*sqlc.CheckTimeEntryTaskProjectAccessRow, error) {
	row, err := r.queries.CheckTimeEntryTaskProjectAccess(ctx, sqlc.CheckTimeEntryTaskProjectAccessParams{
		ID:     taskID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) Create(ctx context.Context, p CreateParams) (uuid.UUID, error) {
	return r.queries.CreateTimeEntryRecord(ctx, sqlc.CreateTimeEntryRecordParams{
		ID:          p.ID,
		TaskID:      p.TaskID,
		UserID:      p.UserID,
		TimeSpent:   p.TimeSpent,
		Description: optText(p.Description),
		StartTime:   optTS(p.StartTime),
		EndTime:     optTS(p.EndTime),
		Date:        pgtype.Timestamp{Time: p.Date, Valid: true},
	})
}

func (r *pgRepository) GetByID(ctx context.Context, id uuid.UUID) (*sqlc.GetTimeEntryDetailByIDRow, error) {
	row, err := r.queries.GetTimeEntryDetailByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) ListFiltered(ctx context.Context, taskID, userID *uuid.UUID, startDate, endDate *time.Time) ([]sqlc.ListTimeEntriesFilteredRow, error) {
	return r.queries.ListTimeEntriesFiltered(ctx, sqlc.ListTimeEntriesFilteredParams{
		TaskID:    optUUID(taskID),
		UserID:    optUUID(userID),
		StartDate: optTS(startDate),
		EndDate:   optTS(endDate),
	})
}

func (r *pgRepository) GetActiveTimer(ctx context.Context, userID uuid.UUID) (*sqlc.GetActiveTimerForUserRow, error) {
	row, err := r.queries.GetActiveTimerForUser(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) Update(ctx context.Context, p UpdateParams) (uuid.UUID, error) {
	var ts pgtype.Int4
	if p.TimeSpent != nil {
		ts = pgtype.Int4{Int32: *p.TimeSpent, Valid: true}
	}
	return r.queries.UpdateTimeEntryByID(ctx, sqlc.UpdateTimeEntryByIDParams{
		ID:          p.ID,
		Description: optText(p.Description),
		TimeSpent:   ts,
		StartTime:   optTS(p.StartTime),
		EndTime:     optTS(p.EndTime),
		Date:        optTS(p.Date),
	})
}

func (r *pgRepository) StopTimer(ctx context.Context, id uuid.UUID, endTime time.Time, timeSpent int32, description *string) (uuid.UUID, error) {
	return r.queries.StopActiveTimerByID(ctx, sqlc.StopActiveTimerByIDParams{
		ID:          id,
		EndTime:     pgtype.Timestamp{Time: endTime, Valid: true},
		TimeSpent:   timeSpent,
		Description: optText(description),
	})
}

func (r *pgRepository) Delete(ctx context.Context, id uuid.UUID) (int64, error) {
	return r.queries.DeleteTimeEntryByID(ctx, id)
}

