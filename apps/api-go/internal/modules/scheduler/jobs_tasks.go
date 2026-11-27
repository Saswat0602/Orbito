package scheduler

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/modules/email"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// CheckDueDateReminders notifies assignees of tasks due within the next 24 hours.
func (s *Scheduler) CheckDueDateReminders(ctx context.Context) {
	now := s.nowFunc().UTC()
	nextDay := now.Add(24 * time.Hour)

	tasks, err := s.repo.ListTasksDueWithinWindow(ctx, now, nextDay)
	if err != nil {
		s.log.Error("CheckDueDateReminders query failed", slog.String("error", err.Error()))
		return
	}
	if s.notifier == nil {
		return
	}
	for _, t := range tasks {
		if nErr := s.notifier.NotifyTaskDueSoon(ctx, t.ID); nErr != nil {
			s.log.Warn("Failed to send due-soon notification", slog.String("taskId", t.ID.String()), slog.String("error", nErr.Error()))
		}
	}
}

// CheckOverdueTasks sends a morning overdue digest email to assignees with overdue open tasks.
func (s *Scheduler) CheckOverdueTasks(ctx context.Context) {
	now := s.nowFunc().UTC()
	rows, err := s.repo.ListOverdueTasksWithAssignees(ctx, now)
	if err != nil {
		s.log.Error("CheckOverdueTasks query failed", slog.String("error", err.Error()))
		return
	}
	if len(rows) == 0 || s.emailSvc == nil {
		return
	}

	type digestGroup struct {
		email string
		name  string
		tasks []sqlc.ListOverdueTasksWithAssigneesForCronRow
	}
	byUser := make(map[uuid.UUID]*digestGroup)
	for _, row := range rows {
		grp, exists := byUser[row.UserID]
		if !exists {
			fullName := strings.TrimSpace(row.UserFirstName + " " + row.UserLastName)
			grp = &digestGroup{email: row.UserEmail, name: fullName}
			byUser[row.UserID] = grp
		}
		grp.tasks = append(grp.tasks, row)
	}

	for _, grp := range byUser {
		dto := email.SendEmailDTO{
			To:       grp.email,
			Subject:  "Overdue Tasks Reminder",
			Template: "task-overdue",
			Data: map[string]any{
				"assigneeName": grp.name,
				"taskCount":    strconv.Itoa(len(grp.tasks)),
			},
		}
		if sendErr := s.emailSvc.SendEmail(ctx, dto); sendErr != nil {
			s.log.Warn("Failed to queue overdue digest email", slog.String("to", grp.email), slog.String("error", sendErr.Error()))
		}
	}
}

// SendWeeklySummaries delivers Monday morning activity digests to active assignees.
func (s *Scheduler) SendWeeklySummaries(ctx context.Context) {
	now := s.nowFunc().UTC()
	oneWeekAgo := now.AddDate(0, 0, -7)

	users, err := s.repo.ListActiveUsersForWeeklySummary(ctx, oneWeekAgo)
	if err != nil {
		s.log.Error("SendWeeklySummaries query failed", slog.String("error", err.Error()))
		return
	}
	if s.emailSvc == nil {
		return
	}

	weekStart := oneWeekAgo.Format("2006-01-02")
	for _, u := range users {
		dto := email.SendEmailDTO{
			To:       u.Email,
			Subject:  "Weekly Task Summary",
			Template: "weekly-summary",
			Data: map[string]any{
				"userName":  strings.TrimSpace(u.FirstName + " " + u.LastName),
				"weekStart": weekStart,
			},
		}
		if sendErr := s.emailSvc.SendEmail(ctx, dto); sendErr != nil {
			s.log.Warn("Failed to queue weekly summary email", slog.String("to", u.Email), slog.String("error", sendErr.Error()))
		}
	}
}
