package scheduler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/modules/task_ranks"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository encapsulates all SQL queries used by background cron jobs.
type Repository interface {
	ListTasksDueWithinWindow(ctx context.Context, start, end time.Time) ([]sqlc.ListTasksDueWithinWindowForCronRow, error)
	ListOverdueTasksWithAssignees(ctx context.Context, cutoff time.Time) ([]sqlc.ListOverdueTasksWithAssigneesForCronRow, error)
	ListActiveUsersForWeeklySummary(ctx context.Context, since time.Time) ([]sqlc.ListActiveUsersForWeeklySummaryForCronRow, error)
	DeleteOldNotifications(ctx context.Context, cutoff time.Time) (int64, error)
	DeleteOldActivityLogs(ctx context.Context, cutoff time.Time) (int64, error)
	DeleteOldRuleExecutions(ctx context.Context, cutoff time.Time) (int64, error)
	ActivatePlanningSprints(ctx context.Context, now time.Time) (int64, error)
	CompleteActiveSprints(ctx context.Context, now time.Time) (int64, error)
	ListAllOrganizationIDs(ctx context.Context) ([]uuid.UUID, error)
	GetWeeklyAnalyticsForOrg(ctx context.Context, orgID uuid.UUID, since time.Time) (sqlc.GetWeeklyAnalyticsForOrgCronRow, error)

	ListActiveRecurringTasks(ctx context.Context) ([]sqlc.ListActiveRecurringTasksForCronRow, error)
	DeactivateRecurringTask(ctx context.Context, recurringID, taskID uuid.UUID) error
	FindExistingOccurrenceTask(ctx context.Context, projectID uuid.UUID, title string, dueDate time.Time) (bool, error)
	ListTaskAssigneeUserIDs(ctx context.Context, taskID uuid.UUID) ([]uuid.UUID, error)
	ListTaskReporterUserIDs(ctx context.Context, taskID uuid.UUID) ([]uuid.UUID, error)
	CreateRecurringOccurrenceTx(ctx context.Context, row sqlc.ListActiveRecurringTasksForCronRow, assignees, reporters []uuid.UUID) (uuid.UUID, error)
	AdvanceRecurringTaskSchedule(ctx context.Context, recurringID uuid.UUID, nextOcc time.Time) error
}

type pgRepository struct {
	db      *database.DB
	queries *sqlc.Queries
}

// NewRepository creates a PostgreSQL repository for the scheduler module.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{
		db:      db,
		queries: sqlc.New(db.Pool()),
	}
}

func (r *pgRepository) ListTasksDueWithinWindow(ctx context.Context, start, end time.Time) ([]sqlc.ListTasksDueWithinWindowForCronRow, error) {
	return r.queries.ListTasksDueWithinWindowForCron(ctx, sqlc.ListTasksDueWithinWindowForCronParams{
		DueDate:   pgtype.Timestamp{Time: start.UTC(), Valid: true},
		DueDate_2: pgtype.Timestamp{Time: end.UTC(), Valid: true},
	})
}

func (r *pgRepository) ListOverdueTasksWithAssignees(ctx context.Context, cutoff time.Time) ([]sqlc.ListOverdueTasksWithAssigneesForCronRow, error) {
	return r.queries.ListOverdueTasksWithAssigneesForCron(ctx, pgtype.Timestamp{Time: cutoff.UTC(), Valid: true})
}

func (r *pgRepository) ListActiveUsersForWeeklySummary(ctx context.Context, since time.Time) ([]sqlc.ListActiveUsersForWeeklySummaryForCronRow, error) {
	return r.queries.ListActiveUsersForWeeklySummaryForCron(ctx, pgtype.Timestamp{Time: since.UTC(), Valid: true})
}

func (r *pgRepository) DeleteOldNotifications(ctx context.Context, cutoff time.Time) (int64, error) {
	return r.queries.DeleteOldNotificationsForCron(ctx, pgtype.Timestamp{Time: cutoff.UTC(), Valid: true})
}

func (r *pgRepository) DeleteOldActivityLogs(ctx context.Context, cutoff time.Time) (int64, error) {
	return r.queries.DeleteOldActivityLogsForCron(ctx, pgtype.Timestamp{Time: cutoff.UTC(), Valid: true})
}

func (r *pgRepository) DeleteOldRuleExecutions(ctx context.Context, cutoff time.Time) (int64, error) {
	return r.queries.DeleteOldRuleExecutionsForCron(ctx, pgtype.Timestamp{Time: cutoff.UTC(), Valid: true})
}

func (r *pgRepository) ActivatePlanningSprints(ctx context.Context, now time.Time) (int64, error) {
	return r.queries.ActivatePlanningSprintsForCron(ctx, pgtype.Timestamp{Time: now.UTC(), Valid: true})
}

func (r *pgRepository) CompleteActiveSprints(ctx context.Context, now time.Time) (int64, error) {
	return r.queries.CompleteActiveSprintsForCron(ctx, pgtype.Timestamp{Time: now.UTC(), Valid: true})
}

func (r *pgRepository) ListAllOrganizationIDs(ctx context.Context) ([]uuid.UUID, error) {
	return r.queries.ListAllOrganizationIDsForCron(ctx)
}

func (r *pgRepository) GetWeeklyAnalyticsForOrg(ctx context.Context, orgID uuid.UUID, since time.Time) (sqlc.GetWeeklyAnalyticsForOrgCronRow, error) {
	return r.queries.GetWeeklyAnalyticsForOrgCron(ctx, sqlc.GetWeeklyAnalyticsForOrgCronParams{
		OrgID: orgID,
		Since: pgtype.Timestamp{Time: since.UTC(), Valid: true},
	})
}

func (r *pgRepository) ListActiveRecurringTasks(ctx context.Context) ([]sqlc.ListActiveRecurringTasksForCronRow, error) {
	return r.queries.ListActiveRecurringTasksForCron(ctx)
}

func (r *pgRepository) DeactivateRecurringTask(ctx context.Context, recurringID, taskID uuid.UUID) error {
	return r.db.WithTx(ctx, func(q *sqlc.Queries) error {
		if err := q.DeactivateRecurringTaskForCron(ctx, recurringID); err != nil {
			return err
		}
		return q.ClearTaskIsRecurringForCron(ctx, taskID)
	})
}

func (r *pgRepository) FindExistingOccurrenceTask(ctx context.Context, projectID uuid.UUID, title string, dueDate time.Time) (bool, error) {
	_, err := r.queries.FindExistingOccurrenceTaskForCron(ctx, sqlc.FindExistingOccurrenceTaskForCronParams{
		ProjectID: projectID,
		Title:     title,
		DueDate:   pgtype.Timestamp{Time: dueDate.UTC(), Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *pgRepository) ListTaskAssigneeUserIDs(ctx context.Context, taskID uuid.UUID) ([]uuid.UUID, error) {
	return r.queries.ListTaskAssigneeUserIDsForCron(ctx, taskID)
}

func (r *pgRepository) ListTaskReporterUserIDs(ctx context.Context, taskID uuid.UUID) ([]uuid.UUID, error) {
	return r.queries.ListTaskReporterUserIDsForCron(ctx, taskID)
}

func (r *pgRepository) CreateRecurringOccurrenceTx(
	ctx context.Context,
	row sqlc.ListActiveRecurringTasksForCronRow,
	assignees, reporters []uuid.UUID,
) (uuid.UUID, error) {
	var createdID uuid.UUID
	err := r.db.WithTx(ctx, func(q *sqlc.Queries) error {
		lastNum, numErr := q.GetLastTaskNumberInProjectForCron(ctx, row.ProjectID)
		if numErr != nil {
			return numErr
		}
		nextNum := lastNum + 1
		slug := fmt.Sprintf("%s-%d", row.ProjectSlug, nextNum)

		id, insErr := q.InsertRecurringOccurrenceTaskForCron(ctx, sqlc.InsertRecurringOccurrenceTaskForCronParams{
			ID:          uuid.New(),
			Title:       row.Title,
			Description: row.Description,
			Type:        row.Type,
			Priority:    row.Priority,
			ProjectID:   row.ProjectID,
			StatusID:    row.StatusID,
			SprintID:    row.SprintID,
			DueDate:     row.NextOccurrence,
			TaskNumber:  nextNum,
			Slug:        slug,
			CreatedByID: row.CreatedByID,
		})
		if insErr != nil {
			return insErr
		}
		createdID = id

		for _, uid := range assignees {
			if aErr := q.InsertTaskAssignee(ctx, sqlc.InsertTaskAssigneeParams{TaskID: id, UserID: uid}); aErr != nil {
				return aErr
			}
		}
		for _, uid := range reporters {
			if rErr := q.InsertTaskReporter(ctx, sqlc.InsertTaskReporterParams{TaskID: id, UserID: uid}); rErr != nil {
				return rErr
			}
		}
		return task_ranks.SeedForTasksBatchWithQueries(ctx, q, []uuid.UUID{id}, row.ProjectID, row.WorkspaceID, row.OrganizationID)
	})
	return createdID, err
}

func (r *pgRepository) AdvanceRecurringTaskSchedule(ctx context.Context, recurringID uuid.UUID, nextOcc time.Time) error {
	return r.queries.AdvanceRecurringTaskScheduleForCron(ctx, sqlc.AdvanceRecurringTaskScheduleForCronParams{
		ID:             recurringID,
		NextOccurrence: pgtype.Timestamp{Time: nextOcc.UTC(), Valid: true},
	})
}
