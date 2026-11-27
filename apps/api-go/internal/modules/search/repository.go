package search

import (
	"context"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// AccessibleScopes holds a user's permitted organization, workspace, and project IDs.
type AccessibleScopes struct {
	OrganizationIDs []uuid.UUID
	WorkspaceIDs    []uuid.UUID
	ProjectIDs      []uuid.UUID
}

// IsEmpty reports whether the user belongs to no tenant scopes at all.
func (s AccessibleScopes) IsEmpty() bool {
	return len(s.OrganizationIDs) == 0 && len(s.WorkspaceIDs) == 0 && len(s.ProjectIDs) == 0
}

// Repository defines scoped read queries powering global, advanced, and suggestion search.
type Repository interface {
	GetUserScopes(ctx context.Context, userID uuid.UUID) (AccessibleScopes, error)
	SearchTasksGlobal(ctx context.Context, arg sqlc.SearchTasksGlobalQueryParams) ([]sqlc.SearchTasksGlobalQueryRow, error)
	SearchProjectsGlobal(ctx context.Context, arg sqlc.SearchProjectsGlobalQueryParams) ([]sqlc.SearchProjectsGlobalQueryRow, error)
	SearchUsersGlobal(ctx context.Context, arg sqlc.SearchUsersGlobalQueryParams) ([]sqlc.SearchUsersGlobalQueryRow, error)
	SearchCommentsGlobal(ctx context.Context, arg sqlc.SearchCommentsGlobalQueryParams) ([]sqlc.SearchCommentsGlobalQueryRow, error)
	SearchAttachmentsGlobal(ctx context.Context, arg sqlc.SearchAttachmentsGlobalQueryParams) ([]sqlc.SearchAttachmentsGlobalQueryRow, error)
	SearchSprintsGlobal(ctx context.Context, arg sqlc.SearchSprintsGlobalQueryParams) ([]sqlc.SearchSprintsGlobalQueryRow, error)
	CountTasksAdvanced(ctx context.Context, arg sqlc.CountTasksAdvancedQueryParams) (int64, error)
	SearchTasksAdvanced(ctx context.Context, arg sqlc.SearchTasksAdvancedQueryParams) ([]sqlc.SearchTasksAdvancedQueryRow, error)
	ListAssigneesBatch(ctx context.Context, taskIDs []uuid.UUID) ([]sqlc.ListTaskAssigneesBatchRow, error)
	ListReportersBatch(ctx context.Context, taskIDs []uuid.UUID) ([]sqlc.ListTaskReportersBatchRow, error)
	SearchTaskTitleSuggestions(ctx context.Context, arg sqlc.SearchTaskTitleSuggestionsParams) ([]string, error)
	SearchProjectNameSuggestions(ctx context.Context, arg sqlc.SearchProjectNameSuggestionsParams) ([]string, error)
}

type pgRepository struct {
	queries *sqlc.Queries
}

// NewRepository constructs a PostgreSQL-backed search repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{
		queries: sqlc.New(db.Pool()),
	}
}

func (r *pgRepository) GetUserScopes(ctx context.Context, userID uuid.UUID) (AccessibleScopes, error) {
	orgs, err := r.queries.GetUserAccessibleOrganizationIDsForSearch(ctx, userID)
	if err != nil {
		return AccessibleScopes{}, err
	}
	wss, err := r.queries.GetUserAccessibleWorkspaceIDsForSearch(ctx, userID)
	if err != nil {
		return AccessibleScopes{}, err
	}
	projs, err := r.queries.GetUserAccessibleProjectIDsForSearch(ctx, userID)
	if err != nil {
		return AccessibleScopes{}, err
	}
	return AccessibleScopes{
		OrganizationIDs: orgs,
		WorkspaceIDs:    wss,
		ProjectIDs:      projs,
	}, nil
}

func (r *pgRepository) SearchTasksGlobal(ctx context.Context, arg sqlc.SearchTasksGlobalQueryParams) ([]sqlc.SearchTasksGlobalQueryRow, error) {
	return r.queries.SearchTasksGlobalQuery(ctx, arg)
}

func (r *pgRepository) SearchProjectsGlobal(ctx context.Context, arg sqlc.SearchProjectsGlobalQueryParams) ([]sqlc.SearchProjectsGlobalQueryRow, error) {
	return r.queries.SearchProjectsGlobalQuery(ctx, arg)
}

func (r *pgRepository) SearchUsersGlobal(ctx context.Context, arg sqlc.SearchUsersGlobalQueryParams) ([]sqlc.SearchUsersGlobalQueryRow, error) {
	return r.queries.SearchUsersGlobalQuery(ctx, arg)
}

func (r *pgRepository) SearchCommentsGlobal(ctx context.Context, arg sqlc.SearchCommentsGlobalQueryParams) ([]sqlc.SearchCommentsGlobalQueryRow, error) {
	return r.queries.SearchCommentsGlobalQuery(ctx, arg)
}

func (r *pgRepository) SearchAttachmentsGlobal(ctx context.Context, arg sqlc.SearchAttachmentsGlobalQueryParams) ([]sqlc.SearchAttachmentsGlobalQueryRow, error) {
	return r.queries.SearchAttachmentsGlobalQuery(ctx, arg)
}

func (r *pgRepository) SearchSprintsGlobal(ctx context.Context, arg sqlc.SearchSprintsGlobalQueryParams) ([]sqlc.SearchSprintsGlobalQueryRow, error) {
	return r.queries.SearchSprintsGlobalQuery(ctx, arg)
}

func (r *pgRepository) CountTasksAdvanced(ctx context.Context, arg sqlc.CountTasksAdvancedQueryParams) (int64, error) {
	return r.queries.CountTasksAdvancedQuery(ctx, arg)
}

func (r *pgRepository) SearchTasksAdvanced(ctx context.Context, arg sqlc.SearchTasksAdvancedQueryParams) ([]sqlc.SearchTasksAdvancedQueryRow, error) {
	return r.queries.SearchTasksAdvancedQuery(ctx, arg)
}

func (r *pgRepository) ListAssigneesBatch(ctx context.Context, taskIDs []uuid.UUID) ([]sqlc.ListTaskAssigneesBatchRow, error) {
	if len(taskIDs) == 0 {
		return nil, nil
	}
	return r.queries.ListTaskAssigneesBatch(ctx, taskIDs)
}

func (r *pgRepository) ListReportersBatch(ctx context.Context, taskIDs []uuid.UUID) ([]sqlc.ListTaskReportersBatchRow, error) {
	if len(taskIDs) == 0 {
		return nil, nil
	}
	return r.queries.ListTaskReportersBatch(ctx, taskIDs)
}

func (r *pgRepository) SearchTaskTitleSuggestions(ctx context.Context, arg sqlc.SearchTaskTitleSuggestionsParams) ([]string, error) {
	return r.queries.SearchTaskTitleSuggestions(ctx, arg)
}

func (r *pgRepository) SearchProjectNameSuggestions(ctx context.Context, arg sqlc.SearchProjectNameSuggestionsParams) ([]string, error) {
	return r.queries.SearchProjectNameSuggestions(ctx, arg)
}
