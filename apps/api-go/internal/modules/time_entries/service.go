package time_entries

import (
	"context"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// CallerContext describes the authenticated principal invoking time_entries.
type CallerContext struct {
	UserID       uuid.UUID
	IsSuperAdmin bool
}

// Service defines business operations for time tracking and reporting.
type Service interface {
	Create(ctx context.Context, req CreateRequest, caller CallerContext) (*EntryResponse, error)
	FindAll(ctx context.Context, caller CallerContext, taskID, userID *uuid.UUID, startDate, endDate *time.Time) ([]EntryResponse, error)
	GetSummary(ctx context.Context, caller CallerContext, userID, taskID *uuid.UUID, startDate, endDate *time.Time) (*SummaryResponse, error)
	FindOne(ctx context.Context, id uuid.UUID, caller CallerContext) (*EntryResponse, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateRequest, callerID uuid.UUID) (*EntryResponse, error)
	Remove(ctx context.Context, id, callerID uuid.UUID) error
	StartTimer(ctx context.Context, req StartTimerRequest, caller CallerContext) (*StartTimerResponse, error)
	StopTimer(ctx context.Context, req StopTimerRequest, caller CallerContext) (*EntryResponse, error)
	GetActiveTimer(ctx context.Context, targetUserID uuid.UUID, caller CallerContext) (*EntryResponse, error)
}

type service struct {
	repo Repository
}

// NewService constructs a time_entries service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func parseTimeISO(raw *string) (*time.Time, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, *raw); err == nil {
		return &t, nil
	}
	t, err := time.Parse("2006-01-02", *raw)
	if err != nil {
		return nil, apperror.BadRequest("Invalid date format")
	}
	return &t, nil
}

func clampPositiveInt32(m int) int32 {
	if m <= 0 {
		return 1
	}
	if m > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(m)
}

func clampNonNegativeInt32(m int) int32 {
	if m < 0 {
		return 0
	}
	if m > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(m)
}

func safeMinutes(minutes float64) int32 {
	if math.IsNaN(minutes) || minutes <= 0 {
		return 1
	}
	return clampPositiveInt32(int(math.Round(minutes)))
}

func (s *service) checkTaskAccess(ctx context.Context, taskID uuid.UUID, caller CallerContext) (bool, *sqlc.CheckTimeEntryTaskProjectAccessRow, error) {
	row, err := s.repo.CheckTaskAccess(ctx, taskID, caller.UserID)
	if err != nil {
		return false, nil, apperror.Internal("Failed to verify task access")
	}
	if row == nil {
		return false, nil, nil
	}
	if caller.IsSuperAdmin || row.IsPm || row.Visibility == sqlc.ProjectVisibilityPUBLIC {
		return true, row, nil
	}
	if row.Visibility == sqlc.ProjectVisibilityINTERNAL && row.IsWm {
		return true, row, nil
	}
	return false, row, nil
}

func (s *service) Create(ctx context.Context, req CreateRequest, caller CallerContext) (*EntryResponse, error) {
	taskID, err := uuid.Parse(req.TaskID)
	if err != nil {
		return nil, apperror.BadRequest("taskId must be a UUID")
	}

	allowed, taskRow, err := s.checkTaskAccess(ctx, taskID, caller)
	if err != nil {
		return nil, err
	}
	if taskRow == nil {
		return nil, apperror.NotFound("Task not found")
	}
	if !allowed {
		return nil, apperror.Forbidden("You do not have access to this task")
	}

	startPtr, err := parseTimeISO(req.StartTime)
	if err != nil {
		return nil, err
	}
	endPtr, err := parseTimeISO(req.EndTime)
	if err != nil {
		return nil, err
	}
	datePtr, err := parseTimeISO(req.Date)
	if err != nil {
		return nil, err
	}
	entryDate := time.Now().UTC()
	if datePtr != nil {
		entryDate = *datePtr
	}

	var timeSpent int32 = 1
	if startPtr != nil && endPtr != nil {
		if !endPtr.After(*startPtr) {
			return nil, apperror.BadRequest("End time must be after start time")
		}
		timeSpent = safeMinutes(endPtr.Sub(*startPtr).Minutes())
	} else if req.TimeSpent != nil && *req.TimeSpent >= 1 {
		timeSpent = clampPositiveInt32(*req.TimeSpent)
	} else if req.Duration != nil {
		timeSpent = safeMinutes(*req.Duration / 60.0)
	}

	newID, err := s.repo.Create(ctx, CreateParams{
		ID:          uuid.New(),
		TaskID:      taskID,
		UserID:      caller.UserID,
		TimeSpent:   timeSpent,
		Description: req.Description,
		StartTime:   startPtr,
		EndTime:     endPtr,
		Date:        entryDate,
	})
	if err != nil {
		return nil, apperror.Internal("Failed to create time entry")
	}

	detail, err := s.repo.GetByID(ctx, newID)
	if err != nil || detail == nil {
		return nil, apperror.Internal("Failed to load created time entry")
	}
	resp := mapDetailRow(detail, false, true)
	return &resp, nil
}

func (s *service) resolveScopeFilters(ctx context.Context, caller CallerContext, taskID, userID *uuid.UUID) (*uuid.UUID, *uuid.UUID, error) {
	if taskID != nil {
		allowed, taskRow, err := s.checkTaskAccess(ctx, *taskID, caller)
		if err != nil {
			return nil, nil, err
		}
		if taskRow == nil || !allowed {
			return nil, nil, apperror.Forbidden("You do not have access to this task")
		}
		return taskID, userID, nil
	}
	targetUser := caller.UserID
	if caller.IsSuperAdmin && userID != nil {
		targetUser = *userID
	}
	return nil, &targetUser, nil
}

func (s *service) FindAll(ctx context.Context, caller CallerContext, taskID, userID *uuid.UUID, startDate, endDate *time.Time) ([]EntryResponse, error) {
	scopedTask, scopedUser, err := s.resolveScopeFilters(ctx, caller, taskID, userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.ListFiltered(ctx, scopedTask, scopedUser, startDate, endDate)
	if err != nil {
		return nil, apperror.Internal("Failed to list time entries")
	}
	out := make([]EntryResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapFilteredRow(r))
	}
	return out, nil
}

func (s *service) GetSummary(ctx context.Context, caller CallerContext, userID, taskID *uuid.UUID, startDate, endDate *time.Time) (*SummaryResponse, error) {
	scopedTask, scopedUser, err := s.resolveScopeFilters(ctx, caller, taskID, userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.ListFiltered(ctx, scopedTask, scopedUser, startDate, endDate)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch time entry summary")
	}

	var totalMinutes int32
	taskMap := make(map[string]*TaskSummaryItem)
	userMap := make(map[string]*UserSummaryItem)
	taskOrder := make([]string, 0)
	userOrder := make([]string, 0)
	entries := make([]EntryResponse, 0, len(rows))

	for _, r := range rows {
		mapped := mapFilteredRow(r)
		entries = append(entries, mapped)
		totalMinutes += r.TimeSpent

		if _, ok := taskMap[r.TaskSlug]; !ok {
			taskOrder = append(taskOrder, r.TaskSlug)
			taskMap[r.TaskSlug] = &TaskSummaryItem{
				Task: TaskBrief{
					ID:    r.TaskID.String(),
					Title: r.TaskTitle,
					Slug:  r.TaskSlug,
					Project: &ProjectBrief{
						ID:   r.ProjectID.String(),
						Name: r.ProjectName,
						Slug: r.ProjectSlug,
					},
				},
			}
		}
		taskMap[r.TaskSlug].TotalTime += r.TimeSpent
		taskMap[r.TaskSlug].Entries++

		uKey := r.UserID.String()
		if _, ok := userMap[uKey]; !ok {
			userOrder = append(userOrder, uKey)
			userMap[uKey] = &UserSummaryItem{
				User: UserBrief{
					ID:        uKey,
					FirstName: r.UserFirstName,
					LastName:  r.UserLastName,
				},
			}
		}
		userMap[uKey].TotalTime += r.TimeSpent
		userMap[uKey].Entries++
	}

	taskSummary := make([]TaskSummaryItem, 0, len(taskOrder))
	for _, k := range taskOrder {
		taskSummary = append(taskSummary, *taskMap[k])
	}
	userSummary := make([]UserSummaryItem, 0, len(userOrder))
	for _, k := range userOrder {
		userSummary = append(userSummary, *userMap[k])
	}

	return &SummaryResponse{
		TotalTimeSpent:      totalMinutes,
		TotalTimeSpentHours: math.Round((float64(totalMinutes)/60.0)*100) / 100,
		TotalEntries:        len(entries),
		TaskSummary:         taskSummary,
		UserSummary:         userSummary,
		Entries:             entries,
	}, nil
}

func (s *service) FindOne(ctx context.Context, id uuid.UUID, caller CallerContext) (*EntryResponse, error) {
	detail, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch time entry")
	}
	if detail == nil {
		return nil, apperror.NotFound("Time entry not found")
	}

	if detail.UserID != caller.UserID {
		allowed, _, accErr := s.checkTaskAccess(ctx, detail.TaskID, caller)
		if accErr != nil || !allowed {
			return nil, apperror.NotFound("Time entry not found")
		}
	}

	resp := mapDetailRow(detail, true, true)
	return &resp, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req UpdateRequest, callerID uuid.UUID) (*EntryResponse, error) {
	detail, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch time entry")
	}
	if detail == nil {
		return nil, apperror.NotFound("Time entry not found")
	}
	if detail.UserID != callerID {
		return nil, apperror.Forbidden("You can only edit your own time entries")
	}

	startPtr, err := parseTimeISO(req.StartTime)
	if err != nil {
		return nil, err
	}
	endPtr, err := parseTimeISO(req.EndTime)
	if err != nil {
		return nil, err
	}
	datePtr, err := parseTimeISO(req.Date)
	if err != nil {
		return nil, err
	}

	var timeSpentPtr *int32
	if startPtr != nil && endPtr != nil {
		if !endPtr.After(*startPtr) {
			return nil, apperror.BadRequest("End time must be after start time")
		}
		m := safeMinutes(endPtr.Sub(*startPtr).Minutes())
		timeSpentPtr = &m
	} else if req.TimeSpent != nil {
		m := clampPositiveInt32(*req.TimeSpent)
		timeSpentPtr = &m
	}

	if _, updErr := s.repo.Update(ctx, UpdateParams{
		ID:          id,
		Description: req.Description,
		TimeSpent:   timeSpentPtr,
		StartTime:   startPtr,
		EndTime:     endPtr,
		Date:        datePtr,
	}); updErr != nil {
		return nil, apperror.Internal("Failed to update time entry")
	}

	updated, err := s.repo.GetByID(ctx, id)
	if err != nil || updated == nil {
		return nil, apperror.Internal("Failed to load updated time entry")
	}
	resp := mapDetailRow(updated, false, false)
	return &resp, nil
}

func (s *service) Remove(ctx context.Context, id, callerID uuid.UUID) error {
	detail, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return apperror.Internal("Failed to fetch time entry")
	}
	if detail == nil {
		return apperror.NotFound("Time entry not found")
	}
	if detail.UserID != callerID {
		return apperror.Forbidden("You can only delete your own time entries")
	}
	if _, delErr := s.repo.Delete(ctx, id); delErr != nil {
		return apperror.Internal("Failed to delete time entry")
	}
	return nil
}

func (s *service) StartTimer(ctx context.Context, req StartTimerRequest, caller CallerContext) (*StartTimerResponse, error) {
	taskID, err := uuid.Parse(req.TaskID)
	if err != nil {
		return nil, apperror.BadRequest("taskId must be a UUID")
	}

	active, err := s.repo.GetActiveTimer(ctx, caller.UserID)
	if err != nil {
		return nil, apperror.Internal("Failed to check active timer")
	}
	if active != nil {
		return nil, apperror.Conflict("You already have an active timer running. Stop it first.")
	}

	allowed, taskRow, err := s.checkTaskAccess(ctx, taskID, caller)
	if err != nil {
		return nil, err
	}
	if taskRow == nil {
		return nil, apperror.NotFound("Task not found")
	}
	if !allowed {
		return nil, apperror.Forbidden("You do not have access to this task")
	}

	now := time.Now().UTC()
	newID, err := s.repo.Create(ctx, CreateParams{
		ID:          uuid.New(),
		TaskID:      taskID,
		UserID:      caller.UserID,
		TimeSpent:   0,
		Description: req.Description,
		StartTime:   &now,
		Date:        now,
	})
	if err != nil {
		return nil, apperror.Internal("Failed to start timer")
	}

	detail, err := s.repo.GetByID(ctx, newID)
	if err != nil || detail == nil {
		return nil, apperror.Internal("Failed to load started timer")
	}
	resp := mapDetailRow(detail, false, false)
	resp.Task.Project = nil
	resp.User.Avatar = nil
	return &StartTimerResponse{
		Message:     "Timer started successfully",
		ActiveTimer: &resp,
	}, nil
}

func (s *service) StopTimer(ctx context.Context, req StopTimerRequest, caller CallerContext) (*EntryResponse, error) {
	active, err := s.repo.GetActiveTimer(ctx, caller.UserID)
	if err != nil {
		return nil, apperror.Internal("Failed to check active timer")
	}
	if active == nil {
		return nil, apperror.NotFound("No active timer found for this user")
	}

	endTime := time.Now().UTC()
	minutes := int(math.Round(endTime.Sub(active.StartTime.Time).Minutes()))
	if _, stopErr := s.repo.StopTimer(ctx, active.ID, endTime, clampNonNegativeInt32(minutes), req.Description); stopErr != nil {
		return nil, apperror.Internal("Failed to stop timer")
	}

	detail, err := s.repo.GetByID(ctx, active.ID)
	if err != nil || detail == nil {
		return nil, apperror.Internal("Failed to load stopped timer")
	}
	resp := mapDetailRow(detail, false, false)
	return &resp, nil
}

func (s *service) GetActiveTimer(ctx context.Context, targetUserID uuid.UUID, caller CallerContext) (*EntryResponse, error) {
	if !caller.IsSuperAdmin && caller.UserID != targetUserID {
		return nil, apperror.Forbidden("You can only view your own active timer")
	}
	active, err := s.repo.GetActiveTimer(ctx, targetUserID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch active timer")
	}
	if active == nil {
		return nil, nil
	}
	mapped := mapFilteredRow(sqlc.ListTimeEntriesFilteredRow(*active))
	return &mapped, nil
}
