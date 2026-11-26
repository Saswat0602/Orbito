package task_watchers

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// CallerContext describes the authenticated user invoking task_watchers.
type CallerContext struct {
	UserID       uuid.UUID
	IsSuperAdmin bool
}

// Service defines business operations for task watchers.
type Service interface {
	Create(ctx context.Context, req CreateRequest) (*WatcherResponse, error)
	Unwatch(ctx context.Context, req CreateRequest) error
	ToggleWatch(ctx context.Context, req CreateRequest) (*ToggleWatchResponse, error)
	FindAll(ctx context.Context, caller CallerContext, taskIdentifier string, userID *uuid.UUID) ([]WatcherResponse, error)
	FindOne(ctx context.Context, id uuid.UUID, caller CallerContext) (*WatcherResponse, error)
	GetTaskWatchers(ctx context.Context, taskIdentifier string) ([]WatcherResponse, error)
	GetUserWatchedTasks(ctx context.Context, userID uuid.UUID) ([]WatcherResponse, error)
	IsUserWatchingTask(ctx context.Context, userID uuid.UUID, taskIdentifier string) (bool, error)
	Remove(ctx context.Context, id uuid.UUID, callerUserID uuid.UUID) error
	GetWatcherStats(ctx context.Context, taskID, userID *uuid.UUID) (*WatcherStatsResponse, error)
}

type service struct {
	repo Repository
}

// NewService constructs a task_watchers service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) canAccessScope(ctx context.Context, scope *sqlc.GetTaskProjectScopeForWatcherRow, caller CallerContext) (bool, error) {
	if caller.IsSuperAdmin {
		return true, nil
	}
	if scope.Visibility == sqlc.ProjectVisibilityPUBLIC {
		return true, nil
	}
	mem, err := s.repo.CheckCallerMembership(ctx, scope.ProjectID, scope.WorkspaceID, caller.UserID)
	if err != nil {
		return false, err
	}
	if mem.IsPm {
		return true, nil
	}
	if scope.Visibility == sqlc.ProjectVisibilityINTERNAL && mem.IsWm {
		return true, nil
	}
	return false, nil
}

func (s *service) Create(ctx context.Context, req CreateRequest) (*WatcherResponse, error) {
	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, apperror.BadRequest("Validation failed (uuid is expected)")
	}

	scope, err := s.repo.GetTaskScope(ctx, req.TaskID)
	if err != nil {
		return nil, apperror.Internal("Failed to lookup task")
	}
	if scope == nil {
		return nil, apperror.NotFound("Task not found")
	}

	userAccess, err := s.repo.CheckTargetUserAccess(ctx, scope.ProjectID, scope.WorkspaceID, scope.OrganizationID, targetUserID)
	if err != nil {
		return nil, apperror.Internal("Failed to lookup user")
	}
	if userAccess == nil {
		return nil, apperror.NotFound("User not found")
	}
	if !userAccess.IsPm && !userAccess.IsWm && !userAccess.IsOm {
		return nil, apperror.Forbidden("User does not have access to this task")
	}

	createdID, err := s.repo.CreateWatcher(ctx, uuid.New(), scope.TaskID, targetUserID)
	if err != nil {
		if errors.Is(err, ErrUniqueViolation) {
			return nil, apperror.Conflict("User is already watching this task")
		}
		return nil, apperror.Internal("Failed to watch task")
	}

	detail, err := s.repo.GetByID(ctx, createdID)
	if err != nil || detail == nil {
		return nil, apperror.Internal("Failed to fetch task watcher")
	}
	resp := mapDetailRow(detail)
	return &resp, nil
}

func (s *service) Unwatch(ctx context.Context, req CreateRequest) error {
	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		return apperror.BadRequest("Validation failed (uuid is expected)")
	}
	scope, err := s.repo.GetTaskScope(ctx, req.TaskID)
	if err != nil {
		return apperror.Internal("Failed to lookup task")
	}
	if scope == nil {
		return apperror.NotFound("Task watcher not found")
	}

	affected, err := s.repo.DeleteByPair(ctx, scope.TaskID, targetUserID)
	if err != nil {
		return apperror.Internal("Failed to unwatch task")
	}
	if affected == 0 {
		return apperror.NotFound("Task watcher not found")
	}
	return nil
}

func (s *service) ToggleWatch(ctx context.Context, req CreateRequest) (*ToggleWatchResponse, error) {
	targetUserID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, apperror.BadRequest("Validation failed (uuid is expected)")
	}
	watching, err := s.IsUserWatchingTask(ctx, targetUserID, req.TaskID)
	if err != nil {
		return nil, err
	}
	if watching {
		if unErr := s.Unwatch(ctx, req); unErr != nil {
			return nil, unErr
		}
		return &ToggleWatchResponse{IsWatching: false}, nil
	}
	created, err := s.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return &ToggleWatchResponse{IsWatching: true, Watcher: created}, nil
}

func (s *service) FindAll(ctx context.Context, caller CallerContext, taskIdentifier string, userID *uuid.UUID) ([]WatcherResponse, error) {
	var taskUUIDPtr, filterUserPtr *uuid.UUID
	if taskIdentifier != "" {
		scope, err := s.repo.GetTaskScope(ctx, taskIdentifier)
		if err != nil {
			return nil, apperror.Internal("Failed to lookup task")
		}
		if scope == nil {
			return []WatcherResponse{}, nil
		}
		allowed, accErr := s.canAccessScope(ctx, scope, caller)
		if accErr != nil {
			return nil, apperror.Internal("Failed to verify task access")
		}
		if !allowed {
			return nil, apperror.Forbidden("You do not have access to this task")
		}
		taskUUIDPtr = &scope.TaskID
		filterUserPtr = userID
	} else {
		if caller.IsSuperAdmin && userID != nil {
			filterUserPtr = userID
		} else {
			uid := caller.UserID
			filterUserPtr = &uid
		}
	}

	rows, err := s.repo.ListFiltered(ctx, taskUUIDPtr, filterUserPtr)
	if err != nil {
		return nil, apperror.Internal("Failed to list task watchers")
	}
	out := make([]WatcherResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapFilteredRow(r))
	}
	return out, nil
}

func (s *service) FindOne(ctx context.Context, id uuid.UUID, caller CallerContext) (*WatcherResponse, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch task watcher")
	}
	if row == nil {
		return nil, apperror.NotFound("Task watcher not found")
	}

	if row.UserID != caller.UserID {
		scope, scopeErr := s.repo.GetTaskScope(ctx, row.TaskID.String())
		if scopeErr != nil || scope == nil {
			return nil, apperror.NotFound("Task watcher not found")
		}
		allowed, accErr := s.canAccessScope(ctx, scope, caller)
		if accErr != nil || !allowed {
			return nil, apperror.NotFound("Task watcher not found")
		}
	}

	resp := mapDetailRow(row)
	return &resp, nil
}

func (s *service) GetTaskWatchers(ctx context.Context, taskIdentifier string) ([]WatcherResponse, error) {
	scope, err := s.repo.GetTaskScope(ctx, taskIdentifier)
	if err != nil {
		return nil, apperror.Internal("Failed to lookup task")
	}
	if scope == nil {
		return nil, apperror.NotFound("Task not found")
	}
	rows, err := s.repo.ListForTask(ctx, scope.TaskID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch watchers")
	}
	out := make([]WatcherResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapTaskWatcherRow(r))
	}
	return out, nil
}

func (s *service) GetUserWatchedTasks(ctx context.Context, userID uuid.UUID) ([]WatcherResponse, error) {
	exists, err := s.repo.CheckUserExists(ctx, userID)
	if err != nil {
		return nil, apperror.Internal("Failed to verify user")
	}
	if !exists {
		return nil, apperror.NotFound("User not found")
	}
	rows, err := s.repo.ListWatchedByUser(ctx, userID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch watched tasks")
	}
	out := make([]WatcherResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapUserWatchedTaskRow(r))
	}
	return out, nil
}

func (s *service) IsUserWatchingTask(ctx context.Context, userID uuid.UUID, taskIdentifier string) (bool, error) {
	scope, err := s.repo.GetTaskScope(ctx, taskIdentifier)
	if err != nil {
		return false, apperror.Internal("Failed to lookup task")
	}
	if scope == nil {
		return false, nil
	}
	pair, err := s.repo.FindByPair(ctx, scope.TaskID, userID)
	if err != nil {
		return false, apperror.Internal("Failed to check watch state")
	}
	return pair != nil, nil
}

func (s *service) Remove(ctx context.Context, id uuid.UUID, callerUserID uuid.UUID) error {
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return apperror.Internal("Failed to fetch task watcher")
	}
	if row == nil {
		return apperror.NotFound("Task watcher not found")
	}
	if row.UserID != callerUserID {
		return apperror.Forbidden("You can only remove your own watch on tasks")
	}
	if _, delErr := s.repo.DeleteByID(ctx, id); delErr != nil {
		return apperror.Internal("Failed to remove task watcher")
	}
	return nil
}

func (s *service) GetWatcherStats(ctx context.Context, taskID, userID *uuid.UUID) (*WatcherStatsResponse, error) {
	since := time.Now().UTC().Add(-7 * 24 * time.Hour)
	counts, err := s.repo.GetStatsCounts(ctx, since, taskID, userID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch watcher statistics")
	}

	var groups []WatcherGroupCount
	if taskID == nil {
		rawGroups, grpErr := s.repo.GroupMostWatched(ctx, userID)
		if grpErr != nil {
			return nil, apperror.Internal("Failed to group watched tasks")
		}
		for _, g := range rawGroups {
			groups = append(groups, WatcherGroupCount{
				TaskID: g.TaskID.String(),
				Count:  map[string]int{"taskId": int(g.WatchCount)},
			})
		}
	}

	return &WatcherStatsResponse{
		TotalWatchers:    counts.TotalWatchers,
		MostWatchedTasks: groups,
		RecentWatches:    counts.RecentWatches,
	}, nil
}
