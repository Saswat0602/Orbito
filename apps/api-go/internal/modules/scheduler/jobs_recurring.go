package scheduler

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/orbito/api-go/internal/modules/tasks"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// GenerateUpcomingRecurringTasks runs daily to materialize tasks due within the next 24h.
func (s *Scheduler) GenerateUpcomingRecurringTasks(ctx context.Context) {
	items, err := s.repo.ListActiveRecurringTasks(ctx)
	if err != nil {
		s.log.Error("ListActiveRecurringTasks failed", slog.String("error", err.Error()))
		return
	}
	now := s.nowFunc().UTC()

	for i := range items {
		s.processRecurringItem(ctx, &items[i], now)
	}
}

func (s *Scheduler) processRecurringItem(ctx context.Context, row *sqlc.ListActiveRecurringTasksForCronRow, now time.Time) {
	model := &sqlc.RecurringTask{
		ID:                row.ID,
		TaskID:            row.TaskID,
		RecurrenceType:    row.RecurrenceType,
		Interval:          row.Interval,
		DaysOfWeek:        row.DaysOfWeek,
		DayOfMonth:        row.DayOfMonth,
		MonthOfYear:       row.MonthOfYear,
		EndType:           row.EndType,
		EndDate:           row.EndDate,
		OccurrenceCount:   row.OccurrenceCount,
		CurrentOccurrence: row.CurrentOccurrence,
		IsActive:          row.IsActive,
		NextOccurrence:    row.NextOccurrence,
	}
	if tasks.IsRecurrenceComplete(model, now) {
		if dErr := s.repo.DeactivateRecurringTask(ctx, row.ID, row.TaskID); dErr != nil {
			s.log.Warn("Failed deactivating completed recurrence", slog.String("recurrenceId", row.ID.String()), slog.String("error", dErr.Error()))
		}
		return
	}

	if !row.NextOccurrence.Valid {
		return
	}
	nextDue := row.NextOccurrence.Time.UTC()
	if nextDue.After(now.Add(24 * time.Hour)) {
		return
	}

	exists, err := s.repo.FindExistingOccurrenceTask(ctx, row.ProjectID, row.Title, nextDue)
	if err != nil || exists {
		return
	}

	assignees, err := s.repo.ListTaskAssigneeUserIDs(ctx, row.TaskID)
	if err != nil {
		s.log.Warn("Failed loading assignees for recurring task", slog.String("taskId", row.TaskID.String()), slog.String("error", err.Error()))
		return
	}
	reporters, err := s.repo.ListTaskReporterUserIDs(ctx, row.TaskID)
	if err != nil {
		s.log.Warn("Failed loading reporters for recurring task", slog.String("taskId", row.TaskID.String()), slog.String("error", err.Error()))
		return
	}

	created := false
	for attempt := 0; attempt < 3; attempt++ {
		if _, cErr := s.repo.CreateRecurringOccurrenceTx(ctx, *row, assignees, reporters); cErr != nil {
			var pgErr *pgconn.PgError
			if errors.As(cErr, &pgErr) && pgErr.Code == "23505" && attempt < 2 {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			s.log.Warn("Failed inserting recurring task occurrence", slog.String("recurrenceId", row.ID.String()), slog.String("error", cErr.Error()))
			return
		}
		created = true
		break
	}
	if !created {
		return
	}

	nextSchedule := computeNextSchedule(row, nextDue)
	if aErr := s.repo.AdvanceRecurringTaskSchedule(ctx, row.ID, nextSchedule); aErr != nil {
		s.log.Warn("Failed advancing recurring schedule", slog.String("recurrenceId", row.ID.String()), slog.String("error", aErr.Error()))
	}
}

func computeNextSchedule(row *sqlc.ListActiveRecurringTasksForCronRow, base time.Time) time.Time {
	interval := row.Interval
	cfg := tasks.RecurrenceConfigRequest{
		RecurrenceType: string(row.RecurrenceType),
		Interval:       &interval,
		DaysOfWeek:     row.DaysOfWeek,
	}
	if row.DayOfMonth.Valid {
		v := row.DayOfMonth.Int32
		cfg.DayOfMonth = &v
	}
	if row.MonthOfYear.Valid {
		v := row.MonthOfYear.Int32
		cfg.MonthOfYear = &v
	}
	return tasks.CalculateNextOccurrence(base, cfg)
}
