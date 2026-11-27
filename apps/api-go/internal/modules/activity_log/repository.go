package activity_log

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository defines database persistence for activity logs.
type Repository interface {
	Create(ctx context.Context, arg sqlc.CreateActivityLogRecordParams) (sqlc.ActivityLog, error)
	ResolveTaskByIDOrSlug(ctx context.Context, identifier string) (*sqlc.ResolveActivityTaskByIDOrSlugRow, error)
	ResolveProjectByIDOrSlug(ctx context.Context, identifier string) (*sqlc.ResolveActivityProjectByIDOrSlugRow, error)
	CheckWorkspaceExists(ctx context.Context, workspaceID uuid.UUID) (*sqlc.CheckWorkspaceExistsForActivityRow, error)
	GetOrganizationIDForTask(ctx context.Context, identifier string) (*uuid.UUID, error)
	GetOrganizationIDForProject(ctx context.Context, projectID uuid.UUID) (*uuid.UUID, error)
	GetOrganizationIDForWorkspace(ctx context.Context, workspaceID uuid.UUID) (*uuid.UUID, error)
	GetOrganizationIDForSprint(ctx context.Context, sprintID uuid.UUID) (*uuid.UUID, error)

	CountTaskActivities(ctx context.Context, taskID uuid.UUID) (int64, error)
	ListTaskActivities(ctx context.Context, taskID uuid.UUID, limit, offset int32) ([]sqlc.ListTaskActivitiesRow, error)

	CountWorkspaceActivityOptimized(ctx context.Context, workspaceID uuid.UUID) (int64, error)
	ListWorkspaceActivityOptimized(ctx context.Context, workspaceID uuid.UUID, limit, offset int32) ([]sqlc.ListWorkspaceActivityOptimizedRow, error)

	CountProjectActivities(ctx context.Context, projectID uuid.UUID) (int64, error)
	ListProjectActivities(ctx context.Context, projectID uuid.UUID, limit, offset int32) ([]sqlc.ListProjectActivitiesRow, error)

	CountOrganizationActivityComprehensive(ctx context.Context, arg sqlc.CountOrganizationActivityComprehensiveParams) (int64, error)
	ListOrganizationActivityComprehensive(ctx context.Context, arg sqlc.ListOrganizationActivityComprehensiveParams) ([]sqlc.ListOrganizationActivityComprehensiveRow, error)

	ListOrganizationActivitiesForStats(ctx context.Context, orgID uuid.UUID, since pgtype.Timestamp) ([]sqlc.ListOrganizationActivitiesForStatsRow, error)

	CountUserRecentActivities(ctx context.Context, arg sqlc.CountUserRecentActivitiesParams) (int64, error)
	ListUserRecentActivities(ctx context.Context, arg sqlc.ListUserRecentActivitiesParams) ([]sqlc.ListUserRecentActivitiesRow, error)
}

type pgRepository struct {
	queries *sqlc.Queries
}

// NewRepository constructs a PostgreSQL-backed activity log repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{queries: sqlc.New(db.Pool())}
}

func (r *pgRepository) Create(ctx context.Context, arg sqlc.CreateActivityLogRecordParams) (sqlc.ActivityLog, error) {
	return r.queries.CreateActivityLogRecord(ctx, arg)
}

func (r *pgRepository) ResolveTaskByIDOrSlug(ctx context.Context, identifier string) (*sqlc.ResolveActivityTaskByIDOrSlugRow, error) {
	parsed, parseErr := uuid.Parse(identifier)
	isUUID := parseErr == nil
	var taskUUID pgtype.UUID
	if isUUID {
		taskUUID = pgtype.UUID{Bytes: parsed, Valid: true}
	}
	row, err := r.queries.ResolveActivityTaskByIDOrSlug(ctx, sqlc.ResolveActivityTaskByIDOrSlugParams{
		IsUuid:   isUUID,
		TaskUuid: taskUUID,
		TaskSlug: identifier,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) ResolveProjectByIDOrSlug(ctx context.Context, identifier string) (*sqlc.ResolveActivityProjectByIDOrSlugRow, error) {
	parsed, parseErr := uuid.Parse(identifier)
	isUUID := parseErr == nil
	var projUUID pgtype.UUID
	if isUUID {
		projUUID = pgtype.UUID{Bytes: parsed, Valid: true}
	}
	row, err := r.queries.ResolveActivityProjectByIDOrSlug(ctx, sqlc.ResolveActivityProjectByIDOrSlugParams{
		IsUuid:   isUUID,
		ProjUuid: projUUID,
		ProjSlug: identifier,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) CheckWorkspaceExists(ctx context.Context, workspaceID uuid.UUID) (*sqlc.CheckWorkspaceExistsForActivityRow, error) {
	row, err := r.queries.CheckWorkspaceExistsForActivity(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetOrganizationIDForTask(ctx context.Context, identifier string) (*uuid.UUID, error) {
	parsed, parseErr := uuid.Parse(identifier)
	isUUID := parseErr == nil
	var taskUUID pgtype.UUID
	if isUUID {
		taskUUID = pgtype.UUID{Bytes: parsed, Valid: true}
	}
	id, err := r.queries.GetOrganizationIDForTask(ctx, sqlc.GetOrganizationIDForTaskParams{
		IsUuid:   isUUID,
		TaskUuid: taskUUID,
		TaskSlug: identifier,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}

func (r *pgRepository) GetOrganizationIDForProject(ctx context.Context, projectID uuid.UUID) (*uuid.UUID, error) {
	id, err := r.queries.GetOrganizationIDForProject(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}

func (r *pgRepository) GetOrganizationIDForWorkspace(ctx context.Context, workspaceID uuid.UUID) (*uuid.UUID, error) {
	id, err := r.queries.GetOrganizationIDForWorkspace(ctx, workspaceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}

func (r *pgRepository) GetOrganizationIDForSprint(ctx context.Context, sprintID uuid.UUID) (*uuid.UUID, error) {
	id, err := r.queries.GetOrganizationIDForSprint(ctx, sprintID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &id, nil
}

func (r *pgRepository) CountTaskActivities(ctx context.Context, taskID uuid.UUID) (int64, error) {
	return r.queries.CountTaskActivities(ctx, taskID)
}

func (r *pgRepository) ListTaskActivities(ctx context.Context, taskID uuid.UUID, limit, offset int32) ([]sqlc.ListTaskActivitiesRow, error) {
	return r.queries.ListTaskActivities(ctx, sqlc.ListTaskActivitiesParams{
		EntityID: taskID,
		Limit:    limit,
		Offset:   offset,
	})
}

func (r *pgRepository) CountWorkspaceActivityOptimized(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	return r.queries.CountWorkspaceActivityOptimized(ctx, workspaceID)
}

func (r *pgRepository) ListWorkspaceActivityOptimized(ctx context.Context, workspaceID uuid.UUID, limit, offset int32) ([]sqlc.ListWorkspaceActivityOptimizedRow, error) {
	return r.queries.ListWorkspaceActivityOptimized(ctx, sqlc.ListWorkspaceActivityOptimizedParams{
		WorkspaceID: workspaceID,
		LimitVal:    limit,
		OffsetVal:   offset,
	})
}

func (r *pgRepository) CountProjectActivities(ctx context.Context, projectID uuid.UUID) (int64, error) {
	return r.queries.CountProjectActivities(ctx, projectID)
}

func (r *pgRepository) ListProjectActivities(ctx context.Context, projectID uuid.UUID, limit, offset int32) ([]sqlc.ListProjectActivitiesRow, error) {
	return r.queries.ListProjectActivities(ctx, sqlc.ListProjectActivitiesParams{
		ProjectID: projectID,
		LimitVal:  limit,
		OffsetVal: offset,
	})
}

func (r *pgRepository) CountOrganizationActivityComprehensive(ctx context.Context, arg sqlc.CountOrganizationActivityComprehensiveParams) (int64, error) {
	return r.queries.CountOrganizationActivityComprehensive(ctx, arg)
}

func (r *pgRepository) ListOrganizationActivityComprehensive(ctx context.Context, arg sqlc.ListOrganizationActivityComprehensiveParams) ([]sqlc.ListOrganizationActivityComprehensiveRow, error) {
	return r.queries.ListOrganizationActivityComprehensive(ctx, arg)
}

func (r *pgRepository) ListOrganizationActivitiesForStats(ctx context.Context, orgID uuid.UUID, since pgtype.Timestamp) ([]sqlc.ListOrganizationActivitiesForStatsRow, error) {
	return r.queries.ListOrganizationActivitiesForStats(ctx, sqlc.ListOrganizationActivitiesForStatsParams{
		OrganizationID: pgtype.UUID{Bytes: orgID, Valid: true},
		CreatedAt:      since,
	})
}

func (r *pgRepository) CountUserRecentActivities(ctx context.Context, arg sqlc.CountUserRecentActivitiesParams) (int64, error) {
	return r.queries.CountUserRecentActivities(ctx, arg)
}

func (r *pgRepository) ListUserRecentActivities(ctx context.Context, arg sqlc.ListUserRecentActivitiesParams) ([]sqlc.ListUserRecentActivitiesRow, error) {
	return r.queries.ListUserRecentActivities(ctx, arg)
}

