package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/modules/email"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubRepo struct {
	dueTasks            []sqlc.ListTasksDueWithinWindowForCronRow
	overdueRows         []sqlc.ListOverdueTasksWithAssigneesForCronRow
	weeklyUsers         []sqlc.ListActiveUsersForWeeklySummaryForCronRow
	recurringRows       []sqlc.ListActiveRecurringTasksForCronRow
	deactivatedIDs      []uuid.UUID
	createdOccurrences  int
	advancedSchedules   int
	deletedNotifsCalled bool
	sprintUpdated       bool
	orgIDs              []uuid.UUID
}

func (s *stubRepo) ListTasksDueWithinWindow(_ context.Context, _, _ time.Time) ([]sqlc.ListTasksDueWithinWindowForCronRow, error) {
	return s.dueTasks, nil
}
func (s *stubRepo) ListOverdueTasksWithAssignees(_ context.Context, _ time.Time) ([]sqlc.ListOverdueTasksWithAssigneesForCronRow, error) {
	return s.overdueRows, nil
}
func (s *stubRepo) ListActiveUsersForWeeklySummary(_ context.Context, _ time.Time) ([]sqlc.ListActiveUsersForWeeklySummaryForCronRow, error) {
	return s.weeklyUsers, nil
}
func (s *stubRepo) DeleteOldNotifications(_ context.Context, _ time.Time) (int64, error) {
	s.deletedNotifsCalled = true
	return 4, nil
}
func (s *stubRepo) DeleteOldActivityLogs(_ context.Context, _ time.Time) (int64, error) {
	return 2, nil
}
func (s *stubRepo) DeleteOldRuleExecutions(_ context.Context, _ time.Time) (int64, error) {
	return 1, nil
}
func (s *stubRepo) ActivatePlanningSprints(_ context.Context, _ time.Time) (int64, error) {
	s.sprintUpdated = true
	return 1, nil
}
func (s *stubRepo) CompleteActiveSprints(_ context.Context, _ time.Time) (int64, error) {
	return 1, nil
}
func (s *stubRepo) ListAllOrganizationIDs(_ context.Context) ([]uuid.UUID, error) {
	return s.orgIDs, nil
}
func (s *stubRepo) GetWeeklyAnalyticsForOrg(_ context.Context, _ uuid.UUID, _ time.Time) (sqlc.GetWeeklyAnalyticsForOrgCronRow, error) {
	return sqlc.GetWeeklyAnalyticsForOrgCronRow{TasksCreated: 5, TasksCompleted: 3, TotalTimeSpent: 120}, nil
}
func (s *stubRepo) ListActiveRecurringTasks(_ context.Context) ([]sqlc.ListActiveRecurringTasksForCronRow, error) {
	return s.recurringRows, nil
}
func (s *stubRepo) DeactivateRecurringTask(_ context.Context, recurringID, _ uuid.UUID) error {
	s.deactivatedIDs = append(s.deactivatedIDs, recurringID)
	return nil
}
func (s *stubRepo) FindExistingOccurrenceTask(_ context.Context, _ uuid.UUID, _ string, _ time.Time) (bool, error) {
	return false, nil
}
func (s *stubRepo) ListTaskAssigneeUserIDs(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (s *stubRepo) ListTaskReporterUserIDs(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}
func (s *stubRepo) CreateRecurringOccurrenceTx(_ context.Context, _ sqlc.ListActiveRecurringTasksForCronRow, _, _ []uuid.UUID) (uuid.UUID, error) {
	s.createdOccurrences++
	return uuid.New(), nil
}
func (s *stubRepo) AdvanceRecurringTaskSchedule(_ context.Context, _ uuid.UUID, _ time.Time) error {
	s.advancedSchedules++
	return nil
}

type stubNotifier struct{ taskIDs []uuid.UUID }

func (n *stubNotifier) NotifyTaskDueSoon(_ context.Context, taskID uuid.UUID) error {
	n.taskIDs = append(n.taskIDs, taskID)
	return nil
}

type stubMailer struct{ sent []email.SendEmailDTO }

func (m *stubMailer) SendEmail(_ context.Context, dto email.SendEmailDTO) error {
	m.sent = append(m.sent, dto)
	return nil
}

func TestSchedulerJobsAndLifecycle(t *testing.T) {
	now := time.Date(2026, 11, 27, 9, 0, 0, 0, time.UTC)
	taskID := uuid.New()
	recActiveID := uuid.New()
	recEndedID := uuid.New()
	userID := uuid.New()
	orgID := uuid.New()

	repo := &stubRepo{
		dueTasks: []sqlc.ListTasksDueWithinWindowForCronRow{{ID: taskID, Title: "Due task"}},
		overdueRows: []sqlc.ListOverdueTasksWithAssigneesForCronRow{
			{TaskID: taskID, UserID: userID, UserEmail: "dev@orbito.io", UserFirstName: "Dev", UserLastName: "One"},
			{TaskID: uuid.New(), UserID: userID, UserEmail: "dev@orbito.io", UserFirstName: "Dev", UserLastName: "One"},
		},
		weeklyUsers: []sqlc.ListActiveUsersForWeeklySummaryForCronRow{
			{ID: userID, Email: "dev@orbito.io", FirstName: "Dev", LastName: "One"},
		},
		orgIDs: []uuid.UUID{orgID},
		recurringRows: []sqlc.ListActiveRecurringTasksForCronRow{
			{
				ID:             recActiveID,
				TaskID:         taskID,
				Title:          "Daily standup",
				RecurrenceType: sqlc.RecurrenceTypeDAILY,
				Interval:       1,
				EndType:        sqlc.RecurrenceEndTypeNEVER,
				NextOccurrence: pgtype.Timestamp{Time: now.Add(2 * time.Hour), Valid: true},
				IsActive:       true,
			},
			{
				ID:                recEndedID,
				TaskID:            uuid.New(),
				Title:             "Completed series",
				RecurrenceType:    sqlc.RecurrenceTypeDAILY,
				Interval:          1,
				EndType:           sqlc.RecurrenceEndTypeAFTEROCCURRENCES,
				OccurrenceCount:   pgtype.Int4{Int32: 2, Valid: true},
				CurrentOccurrence: 2,
				NextOccurrence:    pgtype.Timestamp{Time: now.Add(2 * time.Hour), Valid: true},
				IsActive:          true,
			},
		},
	}

	notifier := &stubNotifier{}
	mailer := &stubMailer{}
	s := New(repo, notifier, mailer, nil)
	s.nowFunc = func() time.Time { return now }

	ctx := context.Background()
	s.CheckDueDateReminders(ctx)
	assert.Equal(t, []uuid.UUID{taskID}, notifier.taskIDs)

	s.CheckOverdueTasks(ctx)
	s.SendWeeklySummaries(ctx)
	require.Len(t, mailer.sent, 2)
	assert.Equal(t, "task-overdue", mailer.sent[0].Template)
	assert.Equal(t, "2", mailer.sent[0].Data["taskCount"])
	assert.Equal(t, "weekly-summary", mailer.sent[1].Template)

	s.CleanupOldData(ctx)
	assert.True(t, repo.deletedNotifsCalled)
	s.UpdateSprintStatuses(ctx)
	assert.True(t, repo.sprintUpdated)
	s.GenerateAnalytics(ctx)

	s.GenerateUpcomingRecurringTasks(ctx)
	assert.Equal(t, 1, repo.createdOccurrences)
	assert.Equal(t, 1, repo.advancedSchedules)
	assert.Equal(t, []uuid.UUID{recEndedID}, repo.deactivatedIDs)

	require.NoError(t, s.Start())
	s.Stop()
}
