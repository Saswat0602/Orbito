package tasks

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// CalculateNextOccurrence computes the next recurring task occurrence timestamp matching NestJS RecurrenceService.
func CalculateNextOccurrence(current time.Time, cfg RecurrenceConfigRequest) time.Time {
	interval := int32(1)
	if cfg.Interval != nil && *cfg.Interval > 0 {
		interval = *cfg.Interval
	}

	cur := current.UTC()
	switch cfg.RecurrenceType {
	case string(sqlc.RecurrenceTypeDAILY), string(sqlc.RecurrenceTypeCUSTOM):
		return cur.AddDate(0, 0, int(interval))
	case string(sqlc.RecurrenceTypeWEEKLY):
		if len(cfg.DaysOfWeek) > 0 {
			sorted := append([]int32(nil), cfg.DaysOfWeek...)
			sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
			weekday := int(cur.Weekday())
			if weekday < 0 || weekday > 6 {
				weekday = 0
			}
			curDay := int32(weekday) // #nosec G115 -- bounded [0,6]
			for _, d := range sorted {
				if d > curDay {
					return cur.AddDate(0, 0, int(d-curDay))
				}
			}
			daysUntilEnd := 7 - curDay
			firstNext := sorted[0]
			return cur.AddDate(0, 0, int(daysUntilEnd+7*(interval-1)+firstNext))
		}
		return cur.AddDate(0, 0, 7*int(interval))
	case string(sqlc.RecurrenceTypeMONTHLY):
		firstOfTarget := time.Date(cur.Year(), cur.Month()+time.Month(interval), 1, cur.Hour(), cur.Minute(), cur.Second(), cur.Nanosecond(), time.UTC)
		targetDay := cur.Day()
		if cfg.DayOfMonth != nil && *cfg.DayOfMonth > 0 {
			targetDay = int(*cfg.DayOfMonth)
		}
		maxDay := daysInMonth(firstOfTarget.Year(), firstOfTarget.Month())
		if targetDay > maxDay {
			targetDay = maxDay
		}
		return time.Date(firstOfTarget.Year(), firstOfTarget.Month(), targetDay, cur.Hour(), cur.Minute(), cur.Second(), cur.Nanosecond(), time.UTC)
	case string(sqlc.RecurrenceTypeQUARTERLY):
		firstOfTarget := time.Date(cur.Year(), cur.Month()+time.Month(3*interval), 1, cur.Hour(), cur.Minute(), cur.Second(), cur.Nanosecond(), time.UTC)
		targetDay := cur.Day()
		if cfg.DayOfMonth != nil && *cfg.DayOfMonth > 0 {
			targetDay = int(*cfg.DayOfMonth)
		}
		maxDay := daysInMonth(firstOfTarget.Year(), firstOfTarget.Month())
		if targetDay > maxDay {
			targetDay = maxDay
		}
		return time.Date(firstOfTarget.Year(), firstOfTarget.Month(), targetDay, cur.Hour(), cur.Minute(), cur.Second(), cur.Nanosecond(), time.UTC)
	case string(sqlc.RecurrenceTypeYEARLY):
		targetMonth := cur.Month()
		if cfg.MonthOfYear != nil && *cfg.MonthOfYear >= 1 && *cfg.MonthOfYear <= 12 {
			targetMonth = time.Month(*cfg.MonthOfYear)
		}
		targetYear := cur.Year() + int(interval)
		targetDay := cur.Day()
		if cfg.DayOfMonth != nil && *cfg.DayOfMonth > 0 {
			targetDay = int(*cfg.DayOfMonth)
		}
		maxDay := daysInMonth(targetYear, targetMonth)
		if targetDay > maxDay {
			targetDay = maxDay
		}
		return time.Date(targetYear, targetMonth, targetDay, cur.Hour(), cur.Minute(), cur.Second(), cur.Nanosecond(), time.UTC)
	default:
		return cur.AddDate(0, 0, int(interval))
	}
}

// IsRecurrenceComplete determines whether a recurring schedule has reached its termination condition.
func IsRecurrenceComplete(rt *sqlc.RecurringTask, now time.Time) bool {
	if rt == nil {
		return true
	}
	switch rt.EndType {
	case sqlc.RecurrenceEndTypeNEVER:
		return false
	case sqlc.RecurrenceEndTypeONDATE:
		if !rt.EndDate.Valid {
			return false
		}
		return !now.UTC().Before(rt.EndDate.Time.UTC())
	case sqlc.RecurrenceEndTypeAFTEROCCURRENCES:
		if !rt.OccurrenceCount.Valid {
			return false
		}
		return rt.CurrentOccurrence >= rt.OccurrenceCount.Int32
	default:
		return false
	}
}

func buildRecurrenceParams(taskID uuid.UUID, req RecurrenceConfigRequest, baseTime time.Time) (sqlc.CreateRecurringTaskConfigParams, error) {
	interval := int32(1)
	if req.Interval != nil && *req.Interval > 0 {
		interval = *req.Interval
	}
	nextOcc := CalculateNextOccurrence(baseTime, req)

	var endDate pgtype.Timestamp
	if req.EndDate != nil && *req.EndDate != "" {
		parsed, ok := parseISO8601(*req.EndDate)
		if !ok {
			return sqlc.CreateRecurringTaskConfigParams{}, apperror.BadRequest("endDate must be a valid ISO 8601 date string")
		}
		endDate = pgtype.Timestamp{Time: parsed, Valid: true}
	}
	var dayOfMonth, monthOfYear, occCount pgtype.Int4
	if req.DayOfMonth != nil {
		dayOfMonth = pgtype.Int4{Int32: *req.DayOfMonth, Valid: true}
	}
	if req.MonthOfYear != nil {
		monthOfYear = pgtype.Int4{Int32: *req.MonthOfYear, Valid: true}
	}
	if req.OccurrenceCount != nil {
		occCount = pgtype.Int4{Int32: *req.OccurrenceCount, Valid: true}
	}
	days := req.DaysOfWeek
	if days == nil {
		days = []int32{}
	}

	return sqlc.CreateRecurringTaskConfigParams{
		ID:                uuid.New(),
		TaskID:            taskID,
		RecurrenceType:    sqlc.RecurrenceType(req.RecurrenceType),
		Interval:          interval,
		DaysOfWeek:        days,
		DayOfMonth:        dayOfMonth,
		MonthOfYear:       monthOfYear,
		EndType:           sqlc.RecurrenceEndType(req.EndType),
		EndDate:           endDate,
		OccurrenceCount:   occCount,
		CurrentOccurrence: 1,
		NextOccurrence:    pgtype.Timestamp{Time: nextOcc, Valid: true},
	}, nil
}

func (s *service) AddRecurrence(ctx context.Context, taskIDOrSlug string, req RecurrenceConfigRequest, userID uuid.UUID, isSuperAdmin bool) (*RecurrenceResponse, error) {
	taskRow, _, err := s.requireTaskAccess(ctx, taskIDOrSlug, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}
	existing, err := s.repo.GetRecurringConfig(ctx, taskRow.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("check existing recurrence: %w", err)
	}
	if existing != nil {
		return nil, apperror.BadRequest("This task is already a recurring task")
	}

	baseDate := time.Now().UTC()
	if taskRow.DueDate.Valid {
		baseDate = taskRow.DueDate.Time.UTC()
	}
	params, err := buildRecurrenceParams(taskRow.ID, req, baseDate)
	if err != nil {
		return nil, err
	}

	created, err := s.repo.CreateRecurrenceConfig(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create recurrence config: %w", err)
	}
	if _, err := s.repo.SetTaskIsRecurring(ctx, taskRow.ID, true); err != nil {
		return nil, fmt.Errorf("mark task recurring: %w", err)
	}
	return mapRecurrenceModel(created), nil
}

func (s *service) UpdateRecurrenceConfig(ctx context.Context, taskIDOrSlug string, req RecurrenceConfigRequest, userID uuid.UUID, isSuperAdmin bool) (*RecurrenceResponse, error) {
	taskRow, _, err := s.requireTaskAccess(ctx, taskIDOrSlug, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}
	existing, err := s.repo.GetRecurringConfig(ctx, taskRow.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("fetch recurrence config: %w", err)
	}
	if existing == nil {
		return nil, apperror.BadRequest("This task is not a recurring task")
	}

	baseDate := time.Now().UTC()
	if taskRow.DueDate.Valid {
		baseDate = taskRow.DueDate.Time.UTC()
	}
	built, err := buildRecurrenceParams(taskRow.ID, req, baseDate)
	if err != nil {
		return nil, err
	}
	updated, err := s.repo.UpdateRecurrenceConfig(ctx, sqlc.UpdateRecurringTaskConfigByIDParams{
		ID:              existing.ID,
		RecurrenceType:  built.RecurrenceType,
		Interval:        built.Interval,
		DaysOfWeek:      built.DaysOfWeek,
		DayOfMonth:      built.DayOfMonth,
		MonthOfYear:     built.MonthOfYear,
		EndType:         built.EndType,
		EndDate:         built.EndDate,
		OccurrenceCount: built.OccurrenceCount,
		NextOccurrence:  built.NextOccurrence,
	})
	if err != nil {
		return nil, fmt.Errorf("update recurrence config: %w", err)
	}
	return mapRecurrenceModel(updated), nil
}

func (s *service) StopRecurrence(ctx context.Context, taskIDOrSlug string, userID uuid.UUID, isSuperAdmin bool) (*TaskResponse, error) {
	taskRow, _, err := s.requireTaskAccess(ctx, taskIDOrSlug, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}
	existing, err := s.repo.GetRecurringConfig(ctx, taskRow.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("fetch recurrence config: %w", err)
	}
	if existing == nil {
		return nil, apperror.BadRequest("This task is not a recurring task")
	}
	if delErr := s.repo.DeleteRecurrenceConfig(ctx, existing.ID); delErr != nil {
		return nil, fmt.Errorf("delete recurrence config: %w", delErr)
	}
	if _, setErr := s.repo.SetTaskIsRecurring(ctx, taskRow.ID, false); setErr != nil {
		return nil, fmt.Errorf("unset recurring flag: %w", setErr)
	}
	return s.FindOne(ctx, taskRow.ID.String(), userID, isSuperAdmin)
}

func (s *service) CompleteOccurrenceAndGenerateNext(ctx context.Context, taskIDOrSlug string, userID uuid.UUID, isSuperAdmin bool) (*CompleteOccurrenceResponse, error) {
	taskRow, _, err := s.requireTaskAccess(ctx, taskIDOrSlug, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}
	existing, err := s.repo.GetRecurringConfig(ctx, taskRow.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("fetch recurrence config: %w", err)
	}
	if existing == nil {
		return nil, apperror.BadRequest("This task is not a recurring task")
	}

	nowStr := time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	completed, err := s.Update(ctx, taskRow.ID.String(), UpdateTaskRequest{CompletedAt: &nowStr}, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}

	if IsRecurrenceComplete(existing, time.Now().UTC()) {
		return &CompleteOccurrenceResponse{CompletedTask: completed, NextTask: nil}, nil
	}

	baseDate := time.Now().UTC()
	if taskRow.DueDate.Valid {
		baseDate = taskRow.DueDate.Time.UTC()
	}
	interval := existing.Interval
	cfg := RecurrenceConfigRequest{
		RecurrenceType:  string(existing.RecurrenceType),
		Interval:        &interval,
		DaysOfWeek:      existing.DaysOfWeek,
		DayOfMonth:      pgInt4Ptr(existing.DayOfMonth),
		MonthOfYear:     pgInt4Ptr(existing.MonthOfYear),
		EndType:         string(existing.EndType),
		OccurrenceCount: pgInt4Ptr(existing.OccurrenceCount),
	}
	nextOcc := CalculateNextOccurrence(baseDate, cfg)
	nextDueStr := nextOcc.Format("2006-01-02T15:04:05.000Z")

	assigneeIDs := make([]string, 0, len(completed.Assignees))
	for _, a := range completed.Assignees {
		assigneeIDs = append(assigneeIDs, a.ID)
	}
	reporterIDs := make([]string, 0, len(completed.Reporters))
	for _, rep := range completed.Reporters {
		reporterIDs = append(reporterIDs, rep.ID)
	}
	isFalse := false
	nextTask, err := s.Create(ctx, CreateTaskRequest{
		Title:       completed.Title,
		Description: completed.Description,
		Type:        &completed.Type,
		Priority:    &completed.Priority,
		ProjectID:   completed.ProjectID,
		StatusID:    &completed.StatusID,
		SprintID:    completed.SprintID,
		DueDate:     &nextDueStr,
		AssigneeIDs: assigneeIDs,
		ReporterIDs: reporterIDs,
		IsRecurring: &isFalse,
	}, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}

	if incErr := s.repo.IncrementRecurrenceOccurrence(ctx, existing.ID, existing.CurrentOccurrence+1, nextOcc); incErr != nil {
		return nil, fmt.Errorf("increment occurrence: %w", incErr)
	}

	return &CompleteOccurrenceResponse{
		CompletedTask: completed,
		NextTask:      nextTask,
	}, nil
}

func (s *service) GetRecurringTasks(ctx context.Context, projectID, userID uuid.UUID, isSuperAdmin bool) ([]TaskResponse, error) {
	meta, err := s.repo.GetProjectMeta(ctx, projectID)
	if err != nil {
		return nil, apperror.NotFound("Project not found")
	}
	if _, authErr := s.checkOrgAccess(ctx, meta.OrganizationID, userID, isSuperAdmin); authErr != nil {
		return nil, authErr
	}

	rows, err := s.repo.ListRecurringByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list recurring tasks: %w", err)
	}
	taskIDs := make([]uuid.UUID, 0, len(rows))
	for _, r := range rows {
		taskIDs = append(taskIDs, r.ID)
	}
	assigneeMap, _, _ := s.loadAssigneesAndReporters(ctx, taskIDs)

	out := make([]TaskResponse, 0, len(rows))
	for _, r := range rows {
		var recCfg *RecurrenceResponse
		if r.RecID.Valid {
			days := r.DaysOfWeek
			if days == nil {
				days = []int32{}
			}
			recCfg = &RecurrenceResponse{
				ID:                uuid.UUID(r.RecID.Bytes).String(),
				TaskID:            r.ID.String(),
				RecurrenceType:    string(r.RecurrenceType.RecurrenceType),
				Interval:          r.Interval.Int32,
				DaysOfWeek:        days,
				DayOfMonth:        pgInt4Ptr(r.DayOfMonth),
				MonthOfYear:       pgInt4Ptr(r.MonthOfYear),
				EndType:           string(r.EndType.RecurrenceEndType),
				EndDate:           pgTSPtr(r.EndDate),
				OccurrenceCount:   pgInt4Ptr(r.OccurrenceCount),
				CurrentOccurrence: r.CurrentOccurrence.Int32,
				NextOccurrence:    formatTS(r.NextOccurrence),
				IsActive:          r.RecIsActive.Bool,
			}
		}
		assignees := assigneeMap[r.ID]
		if assignees == nil {
			assignees = []UserBrief{}
		}
		out = append(out, TaskResponse{
			ID:                r.ID.String(),
			Title:             r.Title,
			Description:       pgTextPtr(r.Description),
			Type:              string(r.Type),
			Priority:          string(r.Priority),
			TaskNumber:        r.TaskNumber,
			Slug:              r.Slug,
			StartDate:         pgTSPtr(r.StartDate),
			DueDate:           pgTSPtr(r.DueDate),
			CompletedAt:       pgTSPtr(r.CompletedAt),
			StoryPoints:       pgInt4Ptr(r.StoryPoints),
			OriginalEstimate:  pgInt4Ptr(r.OriginalEstimate),
			RemainingEstimate: pgInt4Ptr(r.RemainingEstimate),
			ProjectID:         r.ProjectID.String(),
			StatusID:          r.StatusID.String(),
			SprintID:          pgUUIDPtr(r.SprintID),
			ParentTaskID:      pgUUIDPtr(r.ParentTaskID),
			CreatedBy:         pgUUIDPtr(r.CreatedByID),
			CreatedAt:         formatTS(r.CreatedAt),
			UpdatedAt:         formatTS(r.UpdatedAt),
			IsArchived:        r.IsArchived,
			AllowEmailReplies: r.AllowEmailReplies,
			IsRecurring:       r.IsRecurring,
			Status: &StatusBrief{
				ID:       r.StatusID.String(),
				Name:     r.StatusName,
				Color:    r.StatusColor,
				Category: string(r.StatusCategory),
			},
			Assignees:       assignees,
			Reporters:       []UserBrief{},
			RecurringConfig: recCfg,
		})
	}
	return out, nil
}
