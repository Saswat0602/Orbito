package public

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository abstracts read-only queries required by public routes.
type Repository interface {
	GetSharedTaskByToken(ctx context.Context, token string) (*sqlc.GetPublicSharedTaskByTokenRow, error)
	GetShareAttachmentByID(ctx context.Context, token string, attachmentID uuid.UUID) (*sqlc.GetPublicShareAttachmentByIDRow, error)
	ListWorkspaceProjects(ctx context.Context, workspaceSlug string) ([]sqlc.ListWorkspacePublicProjectsRow, error)
	GetProjectBySlugs(ctx context.Context, workspaceSlug, projectSlug string) (*sqlc.GetPublicProjectBySlugsRow, error)
	GetProjectBySlugOnly(ctx context.Context, projectSlug string) (*sqlc.GetPublicProjectBySlugOnlyRow, error)
	ListStatusesByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]sqlc.ListPublicProjectStatusesByWorkflowRow, error)
	GetProjectKPIs(ctx context.Context, projectID uuid.UUID) (*sqlc.GetPublicProjectKPIsRow, error)
	ListTaskCountsByStatus(ctx context.Context, projectID, workflowID uuid.UUID) ([]sqlc.ListPublicProjectTaskCountsByStatusRow, error)
	ListTaskCountsByType(ctx context.Context, projectID uuid.UUID) ([]sqlc.ListPublicProjectTaskCountsByTypeRow, error)
	ListTaskCountsByPriority(ctx context.Context, projectID uuid.UUID) ([]sqlc.ListPublicProjectTaskCountsByPriorityRow, error)
	ListSprintVelocity(ctx context.Context, projectID uuid.UUID) ([]sqlc.ListPublicProjectSprintVelocityRow, error)
	ListCalendarTasks(ctx context.Context, arg sqlc.ListPublicCalendarTasksParams) ([]sqlc.ListPublicCalendarTasksRow, error)
	ListSprints(ctx context.Context, projectID uuid.UUID) ([]sqlc.ListPublicProjectSprintsRow, error)
	GetSprintByID(ctx context.Context, sprintID uuid.UUID, workspaceSlug, projectSlug string) (*sqlc.GetPublicSprintByIDRow, error)
	ListSprintTasks(ctx context.Context, sprintID uuid.UUID, workspaceSlug, projectSlug string) ([]sqlc.ListPublicSprintTasksRow, error)
	CountProjectTasks(ctx context.Context, arg sqlc.CountPublicProjectTasksParams) (int64, error)
	ListProjectTasks(ctx context.Context, arg sqlc.ListPublicProjectTasksParams) ([]sqlc.ListPublicProjectTasksRow, error)
	ListSubtasksBatch(ctx context.Context, parentTaskIDs []uuid.UUID) ([]sqlc.ListPublicSubtasksBatchRow, error)
	ListTaskLabelsBatch(ctx context.Context, taskIDs []uuid.UUID) ([]sqlc.ListPublicTaskLabelsBatchRow, error)
	ListTaskAssigneesBatch(ctx context.Context, taskIDs []uuid.UUID) ([]sqlc.ListPublicTaskAssigneesBatchRow, error)
	GetTaskByIDOrSlug(ctx context.Context, arg sqlc.GetPublicTaskByIDOrSlugParams) (*sqlc.GetPublicTaskByIDOrSlugRow, error)
	CountTaskActivities(ctx context.Context, taskID uuid.UUID) (int64, error)
	ListTaskActivities(ctx context.Context, taskID uuid.UUID, limit, offset int32) ([]sqlc.ListPublicTaskActivitiesRow, error)
	ListTaskComments(ctx context.Context, taskID uuid.UUID) ([]sqlc.ListPublicTaskCommentsRow, error)
	ListTaskAttachments(ctx context.Context, taskID uuid.UUID) ([]sqlc.ListPublicTaskAttachmentsRow, error)
}

type pgRepository struct {
	q *sqlc.Queries
}

// NewRepository creates a PostgreSQL-backed public read repository.
func NewRepository(db *database.DB) Repository {
	if db == nil {
		return nil
	}
	return &pgRepository{q: sqlc.New(db.Pool())}
}

func wrapSingle[T any](v T, err error) (*T, error) {
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *pgRepository) GetSharedTaskByToken(ctx context.Context, token string) (*sqlc.GetPublicSharedTaskByTokenRow, error) {
	row, err := r.q.GetPublicSharedTaskByToken(ctx, token)
	return wrapSingle(row, err)
}

func (r *pgRepository) GetShareAttachmentByID(ctx context.Context, token string, attachmentID uuid.UUID) (*sqlc.GetPublicShareAttachmentByIDRow, error) {
	row, err := r.q.GetPublicShareAttachmentByID(ctx, sqlc.GetPublicShareAttachmentByIDParams{Token: token, ID: attachmentID})
	return wrapSingle(row, err)
}

func (r *pgRepository) ListWorkspaceProjects(ctx context.Context, workspaceSlug string) ([]sqlc.ListWorkspacePublicProjectsRow, error) {
	return r.q.ListWorkspacePublicProjects(ctx, workspaceSlug)
}

func (r *pgRepository) GetProjectBySlugs(ctx context.Context, workspaceSlug, projectSlug string) (*sqlc.GetPublicProjectBySlugsRow, error) {
	row, err := r.q.GetPublicProjectBySlugs(ctx, sqlc.GetPublicProjectBySlugsParams{Slug: workspaceSlug, Slug_2: projectSlug})
	return wrapSingle(row, err)
}

func (r *pgRepository) GetProjectBySlugOnly(ctx context.Context, projectSlug string) (*sqlc.GetPublicProjectBySlugOnlyRow, error) {
	row, err := r.q.GetPublicProjectBySlugOnly(ctx, projectSlug)
	return wrapSingle(row, err)
}

func (r *pgRepository) ListStatusesByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]sqlc.ListPublicProjectStatusesByWorkflowRow, error) {
	return r.q.ListPublicProjectStatusesByWorkflow(ctx, workflowID)
}

func (r *pgRepository) GetProjectKPIs(ctx context.Context, projectID uuid.UUID) (*sqlc.GetPublicProjectKPIsRow, error) {
	row, err := r.q.GetPublicProjectKPIs(ctx, projectID)
	return wrapSingle(row, err)
}

func (r *pgRepository) ListTaskCountsByStatus(ctx context.Context, projectID, workflowID uuid.UUID) ([]sqlc.ListPublicProjectTaskCountsByStatusRow, error) {
	return r.q.ListPublicProjectTaskCountsByStatus(ctx, sqlc.ListPublicProjectTaskCountsByStatusParams{ProjectID: projectID, WorkflowID: workflowID})
}

func (r *pgRepository) ListTaskCountsByType(ctx context.Context, projectID uuid.UUID) ([]sqlc.ListPublicProjectTaskCountsByTypeRow, error) {
	return r.q.ListPublicProjectTaskCountsByType(ctx, projectID)
}

func (r *pgRepository) ListTaskCountsByPriority(ctx context.Context, projectID uuid.UUID) ([]sqlc.ListPublicProjectTaskCountsByPriorityRow, error) {
	return r.q.ListPublicProjectTaskCountsByPriority(ctx, projectID)
}

func (r *pgRepository) ListSprintVelocity(ctx context.Context, projectID uuid.UUID) ([]sqlc.ListPublicProjectSprintVelocityRow, error) {
	return r.q.ListPublicProjectSprintVelocity(ctx, projectID)
}

func (r *pgRepository) ListCalendarTasks(ctx context.Context, arg sqlc.ListPublicCalendarTasksParams) ([]sqlc.ListPublicCalendarTasksRow, error) {
	return r.q.ListPublicCalendarTasks(ctx, arg)
}

func (r *pgRepository) ListSprints(ctx context.Context, projectID uuid.UUID) ([]sqlc.ListPublicProjectSprintsRow, error) {
	return r.q.ListPublicProjectSprints(ctx, projectID)
}

func (r *pgRepository) GetSprintByID(ctx context.Context, sprintID uuid.UUID, workspaceSlug, projectSlug string) (*sqlc.GetPublicSprintByIDRow, error) {
	row, err := r.q.GetPublicSprintByID(ctx, sqlc.GetPublicSprintByIDParams{ID: sprintID, Slug: workspaceSlug, Slug_2: projectSlug})
	return wrapSingle(row, err)
}

func (r *pgRepository) ListSprintTasks(ctx context.Context, sprintID uuid.UUID, workspaceSlug, projectSlug string) ([]sqlc.ListPublicSprintTasksRow, error) {
	return r.q.ListPublicSprintTasks(ctx, sqlc.ListPublicSprintTasksParams{SprintID: pgtype.UUID{Bytes: sprintID, Valid: true}, Slug: workspaceSlug, Slug_2: projectSlug})
}

func (r *pgRepository) CountProjectTasks(ctx context.Context, arg sqlc.CountPublicProjectTasksParams) (int64, error) {
	return r.q.CountPublicProjectTasks(ctx, arg)
}

func (r *pgRepository) ListProjectTasks(ctx context.Context, arg sqlc.ListPublicProjectTasksParams) ([]sqlc.ListPublicProjectTasksRow, error) {
	return r.q.ListPublicProjectTasks(ctx, arg)
}

func (r *pgRepository) ListSubtasksBatch(ctx context.Context, parentTaskIDs []uuid.UUID) ([]sqlc.ListPublicSubtasksBatchRow, error) {
	return r.q.ListPublicSubtasksBatch(ctx, parentTaskIDs)
}

func (r *pgRepository) ListTaskLabelsBatch(ctx context.Context, taskIDs []uuid.UUID) ([]sqlc.ListPublicTaskLabelsBatchRow, error) {
	return r.q.ListPublicTaskLabelsBatch(ctx, taskIDs)
}

func (r *pgRepository) ListTaskAssigneesBatch(ctx context.Context, taskIDs []uuid.UUID) ([]sqlc.ListPublicTaskAssigneesBatchRow, error) {
	return r.q.ListPublicTaskAssigneesBatch(ctx, taskIDs)
}

func (r *pgRepository) GetTaskByIDOrSlug(ctx context.Context, arg sqlc.GetPublicTaskByIDOrSlugParams) (*sqlc.GetPublicTaskByIDOrSlugRow, error) {
	row, err := r.q.GetPublicTaskByIDOrSlug(ctx, arg)
	return wrapSingle(row, err)
}

func (r *pgRepository) CountTaskActivities(ctx context.Context, taskID uuid.UUID) (int64, error) {
	return r.q.CountPublicTaskActivities(ctx, taskID)
}

func (r *pgRepository) ListTaskActivities(ctx context.Context, taskID uuid.UUID, limit, offset int32) ([]sqlc.ListPublicTaskActivitiesRow, error) {
	return r.q.ListPublicTaskActivities(ctx, sqlc.ListPublicTaskActivitiesParams{ID: taskID, Limit: limit, Offset: offset})
}

func (r *pgRepository) ListTaskComments(ctx context.Context, taskID uuid.UUID) ([]sqlc.ListPublicTaskCommentsRow, error) {
	return r.q.ListPublicTaskComments(ctx, taskID)
}

func (r *pgRepository) ListTaskAttachments(ctx context.Context, taskID uuid.UUID) ([]sqlc.ListPublicTaskAttachmentsRow, error) {
	return r.q.ListPublicTaskAttachments(ctx, taskID)
}
