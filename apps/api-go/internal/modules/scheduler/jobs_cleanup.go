package scheduler

import (
	"context"
	"log/slog"
)

// CleanupOldData purges expired notifications (>30d), activity logs (>90d), and rule executions (>30d).
func (s *Scheduler) CleanupOldData(ctx context.Context) {
	now := s.nowFunc().UTC()
	thirtyDaysAgo := now.AddDate(0, 0, -30)
	ninetyDaysAgo := now.AddDate(0, 0, -90)

	notifDeleted, err := s.repo.DeleteOldNotifications(ctx, thirtyDaysAgo)
	if err != nil {
		s.log.Warn("Failed deleting old notifications", slog.String("error", err.Error()))
	}
	logsDeleted, err := s.repo.DeleteOldActivityLogs(ctx, ninetyDaysAgo)
	if err != nil {
		s.log.Warn("Failed deleting old activity logs", slog.String("error", err.Error()))
	}
	execsDeleted, err := s.repo.DeleteOldRuleExecutions(ctx, thirtyDaysAgo)
	if err != nil {
		s.log.Warn("Failed deleting old rule executions", slog.String("error", err.Error()))
	}
	s.log.Info("CleanupOldData completed",
		slog.Int64("notifications", notifDeleted),
		slog.Int64("activityLogs", logsDeleted),
		slog.Int64("ruleExecutions", execsDeleted),
	)
}

// UpdateSprintStatuses transitions planning sprints whose startDate elapsed and active sprints whose endDate passed.
func (s *Scheduler) UpdateSprintStatuses(ctx context.Context) {
	now := s.nowFunc().UTC()
	activated, err := s.repo.ActivatePlanningSprints(ctx, now)
	if err != nil {
		s.log.Warn("ActivatePlanningSprints failed", slog.String("error", err.Error()))
	}
	completed, err := s.repo.CompleteActiveSprints(ctx, now)
	if err != nil {
		s.log.Warn("CompleteActiveSprints failed", slog.String("error", err.Error()))
	}
	s.log.Info("UpdateSprintStatuses completed",
		slog.Int64("activated", activated),
		slog.Int64("completed", completed),
	)
}

// GenerateAnalytics computes weekly task and time-entry statistics across active organizations.
func (s *Scheduler) GenerateAnalytics(ctx context.Context) {
	now := s.nowFunc().UTC()
	sevenDaysAgo := now.AddDate(0, 0, -7)

	orgIDs, err := s.repo.ListAllOrganizationIDs(ctx)
	if err != nil {
		s.log.Error("GenerateAnalytics failed listing organizations", slog.String("error", err.Error()))
		return
	}
	for _, orgID := range orgIDs {
		metrics, mErr := s.repo.GetWeeklyAnalyticsForOrg(ctx, orgID, sevenDaysAgo)
		if mErr != nil {
			s.log.Warn("Failed computing org analytics", slog.String("orgId", orgID.String()), slog.String("error", mErr.Error()))
			continue
		}
		s.log.Info("Weekly org analytics generated",
			slog.String("organizationId", orgID.String()),
			slog.Int64("tasksCreated", metrics.TasksCreated),
			slog.Int64("tasksCompleted", metrics.TasksCompleted),
			slog.Int64("totalTimeSpent", metrics.TotalTimeSpent),
		)
	}
}
