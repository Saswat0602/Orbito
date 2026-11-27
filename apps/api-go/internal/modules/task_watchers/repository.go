package task_watchers

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// ErrUniqueViolation marks unique constraint conflict on task_id + user_id.
var ErrUniqueViolation = errors.New("task watcher already exists")

// Repository defines persistence operations for task watchers.
type Repository interface {
	GetTaskScope(ctx context.Context, identifier string) (*sqlc.GetTaskProjectScopeForWatcherRow, error)
	CheckCallerMembership(ctx context.Context, projectID, workspaceID, userID uuid.UUID) (sqlc.CheckCallerTaskMembershipRow, error)
	CheckTargetUserAccess(ctx context.Context, projectID, workspaceID, orgID, userID uuid.UUID) (*sqlc.CheckTargetUserTenantAccessRow, error)
	CheckUserExists(ctx context.Context, userID uuid.UUID) (bool, error)
	CreateWatcher(ctx context.Context, id, taskID, userID uuid.UUID) (uuid.UUID, error)
	GetByID(ctx context.Context, id uuid.UUID) (*sqlc.GetTaskWatcherDetailsByIDRow, error)
	FindByPair(ctx context.Context, taskID, userID uuid.UUID) (*sqlc.FindTaskWatcherByTaskAndUserRow, error)
	DeleteByPair(ctx context.Context, taskID, userID uuid.UUID) (int64, error)
	DeleteByID(ctx context.Context, id uuid.UUID) (int64, error)
	ListFiltered(ctx context.Context, taskID, userID *uuid.UUID) ([]sqlc.ListTaskWatchersFilteredRow, error)
	ListForTask(ctx context.Context, taskID uuid.UUID) ([]sqlc.ListWatchersForTaskRow, error)
	ListWatchedByUser(ctx context.Context, userID uuid.UUID) ([]sqlc.ListWatchedTasksByUserRow, error)
	GetStatsCounts(ctx context.Context, since time.Time, taskID, userID *uuid.UUID) (sqlc.CountTaskWatchersTotalAndRecentRow, error)
	GroupMostWatched(ctx context.Context, userID *uuid.UUID) ([]sqlc.GroupMostWatchedTasksRow, error)
}

type pgRepository struct {
	queries *sqlc.Queries
}

// NewRepository constructs a PostgreSQL task_watchers repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{queries: db.Queries()}
}

func (r *pgRepository) GetTaskScope(ctx context.Context, identifier string) (*sqlc.GetTaskProjectScopeForWatcherRow, error) {
	var arg sqlc.GetTaskProjectScopeForWatcherParams
	if parsed, err := uuid.Parse(identifier); err == nil {
		arg.IsUuid = true
		arg.TaskUuid = pgtype.UUID{Bytes: parsed, Valid: true}
	} else {
		arg.IsUuid = false
		arg.TaskSlug = identifier
	}
	row, err := r.queries.GetTaskProjectScopeForWatcher(ctx, arg)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) CheckCallerMembership(ctx context.Context, projectID, workspaceID, userID uuid.UUID) (sqlc.CheckCallerTaskMembershipRow, error) {
	return r.queries.CheckCallerTaskMembership(ctx, sqlc.CheckCallerTaskMembershipParams{
		ProjectID:   projectID,
		WorkspaceID: workspaceID,
		UserID:      userID,
	})
}

func (r *pgRepository) CheckTargetUserAccess(ctx context.Context, projectID, workspaceID, orgID, userID uuid.UUID) (*sqlc.CheckTargetUserTenantAccessRow, error) {
	row, err := r.queries.CheckTargetUserTenantAccess(ctx, sqlc.CheckTargetUserTenantAccessParams{
		ProjectID:   projectID,
		WorkspaceID: workspaceID,
		OrgID:       orgID,
		UserID:      userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) CheckUserExists(ctx context.Context, userID uuid.UUID) (bool, error) {
	return r.queries.CheckUserExistsByIDSimple(ctx, userID)
}

func (r *pgRepository) CreateWatcher(ctx context.Context, id, taskID, userID uuid.UUID) (uuid.UUID, error) {
	createdID, err := r.queries.CreateTaskWatcherRecord(ctx, sqlc.CreateTaskWatcherRecordParams{
		ID:     id,
		TaskID: taskID,
		UserID: userID,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return uuid.Nil, ErrUniqueViolation
		}
		return uuid.Nil, err
	}
	return createdID, nil
}

func (r *pgRepository) GetByID(ctx context.Context, id uuid.UUID) (*sqlc.GetTaskWatcherDetailsByIDRow, error) {
	row, err := r.queries.GetTaskWatcherDetailsByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) FindByPair(ctx context.Context, taskID, userID uuid.UUID) (*sqlc.FindTaskWatcherByTaskAndUserRow, error) {
	row, err := r.queries.FindTaskWatcherByTaskAndUser(ctx, sqlc.FindTaskWatcherByTaskAndUserParams{
		TaskID: taskID,
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

func (r *pgRepository) DeleteByPair(ctx context.Context, taskID, userID uuid.UUID) (int64, error) {
	return r.queries.DeleteTaskWatcherByPair(ctx, sqlc.DeleteTaskWatcherByPairParams{
		TaskID: taskID,
		UserID: userID,
	})
}

func (r *pgRepository) DeleteByID(ctx context.Context, id uuid.UUID) (int64, error) {
	return r.queries.DeleteTaskWatcherByID(ctx, id)
}

func optUUID(v *uuid.UUID) pgtype.UUID {
	if v == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *v, Valid: true}
}

func (r *pgRepository) ListFiltered(ctx context.Context, taskID, userID *uuid.UUID) ([]sqlc.ListTaskWatchersFilteredRow, error) {
	return r.queries.ListTaskWatchersFiltered(ctx, sqlc.ListTaskWatchersFilteredParams{
		TaskID: optUUID(taskID),
		UserID: optUUID(userID),
	})
}

func (r *pgRepository) ListForTask(ctx context.Context, taskID uuid.UUID) ([]sqlc.ListWatchersForTaskRow, error) {
	return r.queries.ListWatchersForTask(ctx, taskID)
}

func (r *pgRepository) ListWatchedByUser(ctx context.Context, userID uuid.UUID) ([]sqlc.ListWatchedTasksByUserRow, error) {
	return r.queries.ListWatchedTasksByUser(ctx, userID)
}

func (r *pgRepository) GetStatsCounts(ctx context.Context, since time.Time, taskID, userID *uuid.UUID) (sqlc.CountTaskWatchersTotalAndRecentRow, error) {
	return r.queries.CountTaskWatchersTotalAndRecent(ctx, sqlc.CountTaskWatchersTotalAndRecentParams{
		Since:  pgtype.Timestamp{Time: since, Valid: true},
		TaskID: optUUID(taskID),
		UserID: optUUID(userID),
	})
}

func (r *pgRepository) GroupMostWatched(ctx context.Context, userID *uuid.UUID) ([]sqlc.GroupMostWatchedTasksRow, error) {
	return r.queries.GroupMostWatchedTasks(ctx, optUUID(userID))
}
