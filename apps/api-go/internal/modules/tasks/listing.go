package tasks

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/modules/task_ranks"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

func (s *service) hydrateTaskListByIDs(ctx context.Context, taskIDs []uuid.UUID) ([]TaskResponse, error) {
	if len(taskIDs) == 0 {
		return []TaskResponse{}, nil
	}
	assigneeMap, reporterMap, labelMap := s.loadAllListRelations(ctx, taskIDs)
	out := make([]TaskResponse, 0, len(taskIDs))
	for _, tid := range taskIDs {
		full, err := s.repo.GetTaskFullByID(ctx, tid)
		if err != nil {
			continue
		}
		item := mapTaskRowBase(full)
		if a := assigneeMap[tid]; a != nil {
			item.Assignees = a
		}
		if r := reporterMap[tid]; r != nil {
			item.Reporters = r
		}
		if l := labelMap[tid]; l != nil {
			item.Labels = l
		}
		out = append(out, item)
	}
	return out, nil
}

func (s *service) loadAssigneesAndReporters(ctx context.Context, taskIDs []uuid.UUID) (map[uuid.UUID][]UserBrief, map[uuid.UUID][]UserBrief, error) {
	aRows, err := s.repo.ListAssigneesBatch(ctx, taskIDs)
	if err != nil {
		return nil, nil, err
	}
	rRows, err := s.repo.ListReportersBatch(ctx, taskIDs)
	if err != nil {
		return nil, nil, err
	}
	aMap := make(map[uuid.UUID][]UserBrief)
	for _, a := range aRows {
		f, l := a.FirstName, a.LastName
		aMap[a.TaskID] = append(aMap[a.TaskID], UserBrief{
			ID: a.ID.String(), Email: a.Email, FirstName: &f, LastName: &l, Avatar: pgTextPtr(a.Avatar),
		})
	}
	rMap := make(map[uuid.UUID][]UserBrief)
	for _, r := range rRows {
		f, l := r.FirstName, r.LastName
		rMap[r.TaskID] = append(rMap[r.TaskID], UserBrief{
			ID: r.ID.String(), Email: r.Email, FirstName: &f, LastName: &l, Avatar: pgTextPtr(r.Avatar),
		})
	}
	return aMap, rMap, nil
}

func (s *service) loadAllListRelations(ctx context.Context, taskIDs []uuid.UUID) (map[uuid.UUID][]UserBrief, map[uuid.UUID][]UserBrief, map[uuid.UUID][]TaskLabelResponse) {
	aMap, rMap, _ := s.loadAssigneesAndReporters(ctx, taskIDs)
	lRows, _ := s.repo.ListLabelsBatch(ctx, taskIDs)
	lMap := make(map[uuid.UUID][]TaskLabelResponse)
	for _, l := range lRows {
		lMap[l.TaskID] = append(lMap[l.TaskID], TaskLabelResponse{
			TaskID:      l.TaskID.String(),
			LabelID:     l.LabelID.String(),
			Name:        l.Name,
			Color:       l.Color,
			Description: pgTextPtr(l.Description),
		})
	}
	return aMap, rMap, lMap
}

func (s *service) findPaginated(ctx context.Context, q TaskFilterQuery, userID uuid.UUID, isSuperAdmin, excludeDone bool) (*PaginatedTasksResponse, error) {
	orgID, projectIDs, workspaceIDs, err := s.repo.ResolveScopeContext(ctx, q.OrganizationID, q.ProjectIDs, q.WorkspaceIDs, userID)
	if err != nil {
		return nil, err
	}
	acc, err := s.repo.GetOrgAccess(ctx, orgID, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}

	page := max(1, q.Page)
	limit := max(1, q.Limit)
	q.Page = page
	q.Limit = limit

	ids, total, fc, err := s.repo.FindTasksDynamic(ctx, orgID, projectIDs, workspaceIDs, acc, userID, q, excludeDone)
	if err != nil {
		return nil, fmt.Errorf("query filtered tasks: %w", err)
	}

	tasks, err := s.hydrateTaskListByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	return &PaginatedTasksResponse{
		Data:         tasks,
		Total:        total,
		Page:         page,
		Limit:        limit,
		TotalPages:   totalPages,
		FilterCounts: &fc,
	}, nil
}

func (s *service) FindAll(ctx context.Context, q TaskFilterQuery, userID uuid.UUID, isSuperAdmin bool) (*PaginatedTasksResponse, error) {
	return s.findPaginated(ctx, q, userID, isSuperAdmin, true)
}

func (s *service) GetTasks(ctx context.Context, q TaskFilterQuery, userID uuid.UUID, isSuperAdmin bool) (*PaginatedTasksResponse, error) {
	return s.findPaginated(ctx, q, userID, isSuperAdmin, false)
}

func (s *service) GetTasksGrouped(ctx context.Context, q GroupedTasksQuery, userID uuid.UUID, isSuperAdmin bool) (*GroupedTasksResponse, error) {
	orgID, projectIDs, workspaceIDs, err := s.repo.ResolveScopeContext(ctx, q.OrganizationID, q.ProjectIDs, q.WorkspaceIDs, userID)
	if err != nil {
		return nil, err
	}
	acc, err := s.repo.GetOrgAccess(ctx, orgID, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}

	buckets, err := s.repo.QueryGroupedTasks(ctx, orgID, projectIDs, workspaceIDs, acc, userID, q)
	if err != nil {
		return nil, fmt.Errorf("query grouped tasks: %w", err)
	}

	var allIDs []uuid.UUID
	for _, b := range buckets {
		for _, t := range b.Tasks {
			if id, pErr := uuid.Parse(t.ID); pErr == nil {
				allIDs = append(allIDs, id)
			}
		}
	}
	hydrated, _ := s.hydrateTaskListByIDs(ctx, allIDs)
	byID := make(map[string]TaskResponse, len(hydrated))
	for _, t := range hydrated {
		byID[t.ID] = t
	}
	for i := range buckets {
		filled := make([]TaskResponse, 0, len(buckets[i].Tasks))
		for _, placeholder := range buckets[i].Tasks {
			if t, ok := byID[placeholder.ID]; ok {
				filled = append(filled, t)
			}
		}
		buckets[i].Tasks = filled
		if q.GroupKey != "" {
			p := max(1, q.Page)
			buckets[i].Page = &p
		}
	}

	return &GroupedTasksResponse{
		Groups:        buckets,
		GroupBy:       q.GroupBy,
		Page:          max(1, q.Page),
		LimitPerGroup: max(1, q.LimitPerGroup),
	}, nil
}

func (s *service) GetTasksGroupedByStatus(ctx context.Context, q ByStatusQuery, userID uuid.UUID, isSuperAdmin bool) (*ByStatusEnvelope, error) {
	info, err := s.repo.GetProjectWorkflowInfoBySlug(ctx, q.Slug)
	if err != nil {
		return nil, err
	}
	acc, err := s.repo.GetProjectAccessByID(ctx, info.ProjectID, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}

	statuses := info.Statuses
	if q.StatusID != "" {
		filtered := make([]sqlc.TaskStatus, 0, 1)
		for _, st := range statuses {
			if st.ID.String() == q.StatusID {
				filtered = append(filtered, st)
			}
		}
		if len(filtered) == 0 {
			return nil, apperror.NotFound(fmt.Sprintf("Status with ID %s not found in project workflow", q.StatusID))
		}
		statuses = filtered
	}

	var sprintPtr *uuid.UUID
	if q.SprintID != "" {
		if sid, sErr := uuid.Parse(q.SprintID); sErr == nil {
			sprintPtr = &sid
		}
	}

	page := max(1, q.Page)
	limit := min(100, max(1, q.Limit))
	offset := (page - 1) * limit
	restrictedToUser := !acc.IsElevated

	columns := make([]StatusColumnBucket, 0, len(statuses))
	var totalTasks, loadedTasks int64

	for _, st := range statuses {
		taskIDs, colTotal, qErr := s.repo.QueryColumnTasks(ctx, info.ProjectID, st.ID, sprintPtr, q.IncludeSubtasks, restrictedToUser, userID, limit, offset)
		if qErr != nil {
			return nil, fmt.Errorf("query column tasks: %w", qErr)
		}
		tasks, _ := s.hydrateTaskListByIDs(ctx, taskIDs)
		totalPages := int(math.Ceil(float64(colTotal) / float64(limit)))

		columns = append(columns, StatusColumnBucket{
			StatusID:       st.ID.String(),
			StatusName:     st.Name,
			StatusColor:    st.Color,
			StatusCategory: string(st.Category),
			Tasks:          tasks,
			Pagination: ColumnPagination{
				Total:       colTotal,
				Page:        page,
				Limit:       limit,
				TotalPages:  totalPages,
				HasNextPage: page < totalPages,
				HasPrevPage: page > 1,
			},
		})
		totalTasks += colTotal
		loadedTasks += int64(len(tasks))
	}

	return &ByStatusEnvelope{
		Data: columns,
		Meta: ByStatusMeta{
			TotalTasks:    totalTasks,
			LoadedTasks:   loadedTasks,
			TotalStatuses: len(columns),
			FetchedAt:     time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		},
	}, nil
}

func (s *service) FindTodaysTasks(ctx context.Context, orgRaw string, page, limit int, userID uuid.UUID, isSuperAdmin bool) (*TodayTasksResponse, error) {
	orgID, _, _, err := s.repo.ResolveScopeContext(ctx, orgRaw, nil, nil, userID)
	if err != nil {
		return nil, err
	}
	acc, err := s.repo.GetOrgAccess(ctx, orgID, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	dayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.UTC)

	page = max(1, page)
	limit = min(100, max(1, limit))
	offset := (page - 1) * limit

	ids, total, err := s.repo.FindTodayTaskIDs(ctx, orgID, acc, userID, dayStart, dayEnd, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("find today tasks: %w", err)
	}
	tasks, err := s.hydrateTaskListByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	return &TodayTasksResponse{
		Tasks: tasks,
		Pagination: StandardPagination{
			CurrentPage: page,
			TotalPages:  totalPages,
			TotalCount:  total,
			HasNextPage: page < totalPages,
			HasPrevPage: page > 1,
		},
	}, nil
}

func (s *service) FindByOrganization(ctx context.Context, orgID uuid.UUID, priority, search string, page, limit int, userID uuid.UUID, isSuperAdmin bool) (*TodayTasksResponse, error) {
	if _, err := s.repo.GetOrgAccess(ctx, orgID, userID, isSuperAdmin); err != nil {
		return nil, err
	}
	page = max(1, page)
	limit = min(100, max(1, limit))
	offset := (page - 1) * limit

	ids, total, err := s.repo.FindOrgTaskIDs(ctx, orgID, userID, priority, search, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("find tasks by organization: %w", err)
	}
	tasks, err := s.hydrateTaskListByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	return &TodayTasksResponse{
		Tasks: tasks,
		Pagination: StandardPagination{
			CurrentPage: page,
			TotalPages:  totalPages,
			TotalCount:  total,
			HasNextPage: page < totalPages,
			HasPrevPage: page > 1,
		},
	}, nil
}

func (s *service) ReorderTask(ctx context.Context, taskIDOrSlug string, req task_ranks.ReorderRequest, userID uuid.UUID, isSuperAdmin bool) error {
	if s.rankSvc == nil {
		return nil
	}
	taskRow, _, err := s.requireTaskAccess(ctx, taskIDOrSlug, userID, isSuperAdmin)
	if err != nil {
		return err
	}
	if req.AfterTaskID != nil && *req.AfterTaskID != "" {
		if _, parseErr := uuid.Parse(*req.AfterTaskID); parseErr != nil {
			if resolved, rErr := s.repo.ResolveTaskIDBySlug(ctx, *req.AfterTaskID); rErr == nil {
				sID := resolved.String()
				req.AfterTaskID = &sID
			}
		}
	}
	if req.BeforeTaskID != nil && *req.BeforeTaskID != "" {
		if _, parseErr := uuid.Parse(*req.BeforeTaskID); parseErr != nil {
			if resolved, rErr := s.repo.ResolveTaskIDBySlug(ctx, *req.BeforeTaskID); rErr == nil {
				sID := resolved.String()
				req.BeforeTaskID = &sID
			}
		}
	}
	return s.rankSvc.Reorder(ctx, taskRow.ID, req)
}

func (s *service) checkOrgAccess(ctx context.Context, orgID, userID uuid.UUID, isSuperAdmin bool) (AccessContext, error) {
	return s.repo.GetOrgAccess(ctx, orgID, userID, isSuperAdmin)
}

func (s *service) requireTaskAccess(ctx context.Context, idOrSlug string, userID uuid.UUID, isSuperAdmin bool) (*sqlc.GetTaskAccessDetailsRow, effectiveAccess, error) {
	row, err := s.repo.GetTaskAccessDetails(ctx, idOrSlug, userID)
	if err != nil {
		return nil, effectiveAccess{}, apperror.NotFound("Task not found")
	}
	acc, err := evaluateTaskAccess(row, userID, isSuperAdmin)
	if err != nil {
		return nil, effectiveAccess{}, err
	}
	return row, acc, nil
}

func memberRestrictedViolation(req UpdateTaskRequest) string {
	checks := []struct {
		name string
		set  bool
	}{
		{"title", req.Title != nil},
		{"description", req.Description != nil},
		{"priority", req.Priority != nil},
		{"startDate", req.StartDate != nil},
		{"dueDate", req.DueDate != nil},
		{"type", req.Type != nil},
		{"storyPoints", req.StoryPoints != nil},
		{"originalEstimate", req.OriginalEstimate != nil},
		{"remainingEstimate", req.RemainingEstimate != nil},
		{"projectId", req.ProjectID != nil},
		{"assigneeIds", req.AssigneeIDs != nil},
		{"reporterIds", req.ReporterIDs != nil},
		{"sprintId", req.SprintID != nil || req.ClearSprintID},
		{"parentTaskId", req.ParentTaskID != nil || req.ClearParentTaskID},
		{"isRecurring", req.IsRecurring != nil},
		{"recurrenceConfig", req.RecurrenceConfig != nil},
		{"allowEmailReplies", req.AllowEmailReplies != nil},
		{"stopRecurrence", req.StopRecurrence != nil},
	}
	for _, c := range checks {
		if c.set {
			return c.name
		}
	}
	return ""
}

func validateRoleForMember(role middleware.Role, req UpdateTaskRequest) error {
	if role != middleware.RoleMember {
		return nil
	}
	if violation := memberRestrictedViolation(req); violation != "" {
		return apperror.Forbidden(fmt.Sprintf("Members are not allowed to update the %s field", violation))
	}
	return nil
}

func parseCSV(raw *string) []string {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil
	}
	parts := strings.Split(*raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}
