package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/sanitize"
)

func (s *service) BulkCreate(ctx context.Context, req BulkCreateTasksRequest, userID uuid.UUID, isSuperAdmin bool) (*BulkCreateResponse, error) {
	projectID := uuid.MustParse(req.ProjectID)
	statusID := uuid.MustParse(req.StatusID)

	meta, err := s.repo.GetProjectMeta(ctx, projectID)
	if err != nil {
		return nil, apperror.NotFound("Project not found")
	}
	canChange, err := s.repo.GetProjectAccessForCreate(ctx, projectID, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}
	if !canChange {
		return nil, apperror.Forbidden("Insufficient permissions to create task in this project")
	}

	var sprintUUID pgtype.UUID
	if req.SprintID != nil && *req.SprintID != "" {
		sid := uuid.MustParse(*req.SprintID)
		sprintUUID = pgtype.UUID{Bytes: sid, Valid: true}
	} else {
		sprintUUID, _ = s.repo.GetDefaultSprintID(ctx, projectID)
	}

	items := make([]CreateTxInput, 0, len(req.Tasks))
	for _, t := range req.Tasks {
		taskType := sqlc.TaskTypeTASK
		if t.Type != nil && *t.Type != "" {
			taskType = sqlc.TaskType(*t.Type)
		}
		priority := sqlc.TaskPriorityMEDIUM
		if t.Priority != nil && *t.Priority != "" {
			priority = sqlc.TaskPriority(*t.Priority)
		}
		var desc pgtype.Text
		if t.Description != nil {
			desc = pgtype.Text{String: sanitize.HTML(*t.Description), Valid: true}
		}
		var dueTS pgtype.Timestamp
		if t.DueDate != nil && *t.DueDate != "" {
			parsed, ok := parseISO8601(*t.DueDate)
			if !ok {
				return nil, apperror.BadRequest("dueDate must be a valid ISO 8601 date string")
			}
			dueTS = pgtype.Timestamp{Time: parsed, Valid: true}
		}
		items = append(items, CreateTxInput{
			TaskID:         uuid.New(),
			Title:          sanitize.Text(t.Title),
			Description:    desc,
			Type:           taskType,
			Priority:       priority,
			DueDate:        dueTS,
			ProjectID:      projectID,
			WorkspaceID:    meta.WorkspaceID,
			OrganizationID: meta.OrganizationID,
			StatusID:       &statusID,
			SprintID:       sprintUUID,
			UserID:         userID,
		})
	}

	createdIDs, err := s.repo.BulkCreateTasksTx(ctx, projectID, statusID, sprintUUID, items)
	if err != nil {
		return nil, fmt.Errorf("bulk create tasks: %w", err)
	}
	hydrated, _ := s.hydrateTaskListByIDs(ctx, createdIDs)
	return &BulkCreateResponse{
		Message: "Tasks created successfully",
		Count:   len(hydrated),
		Tasks:   hydrated,
	}, nil
}

func parseUUIDSlice(raw []string) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, 0, len(raw))
	for _, s := range raw {
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, apperror.BadRequest("Invalid UUID: " + s)
		}
		out = append(out, id)
	}
	return out, nil
}

func parseOptUUID(raw *string) (*uuid.UUID, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(*raw)
	if err != nil {
		return nil, apperror.BadRequest("Invalid UUID: " + *raw)
	}
	return &id, nil
}

func (s *service) BulkDeleteTasks(ctx context.Context, req BulkDeleteTasksRequest, userID uuid.UUID, isSuperAdmin bool) (*BulkDeleteResponse, error) {
	if len(req.TaskIDs) == 0 && !req.All {
		return nil, apperror.BadRequest(`No task IDs provided and "all" flag not set`)
	}
	explicitIDs, err := parseUUIDSlice(req.TaskIDs)
	if err != nil {
		return nil, err
	}
	excludedIDs, err := parseUUIDSlice(req.ExcludedIDs)
	if err != nil {
		return nil, err
	}
	projID, err := parseOptUUID(req.ProjectID)
	if err != nil {
		return nil, err
	}

	targetIDs, err := s.repo.ResolveBulkTaskIDs(ctx, req.All, explicitIDs, excludedIDs, projID, nil, nil, "", nil, nil, nil, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("resolve bulk delete targets: %w", err)
	}

	deletable := make([]uuid.UUID, 0, len(targetIDs))
	failed := make([]FailedTaskItem, 0)
	foundSet := make(map[uuid.UUID]struct{}, len(targetIDs))

	for _, tid := range targetIDs {
		accessRow, acc, aErr := s.requireTaskAccess(ctx, tid.String(), userID, isSuperAdmin)
		if aErr != nil {
			failed = append(failed, FailedTaskItem{ID: tid.String(), Reason: aErr.Error()})
			continue
		}
		foundSet[accessRow.ID] = struct{}{}
		if acc.isElevated {
			deletable = append(deletable, accessRow.ID)
		} else {
			failed = append(failed, FailedTaskItem{ID: tid.String(), Reason: "Insufficient permissions"})
		}
	}

	if !req.All {
		for _, id := range explicitIDs {
			if _, ok := foundSet[id]; !ok {
				alreadyFailed := false
				for _, f := range failed {
					if f.ID == id.String() {
						alreadyFailed = true
						break
					}
				}
				if !alreadyFailed {
					failed = append(failed, FailedTaskItem{ID: id.String(), Reason: "Task not found"})
				}
			}
		}
	}

	var deletedCount int64
	if len(deletable) > 0 {
		deletedCount, err = s.repo.BulkDeleteByIDs(ctx, deletable)
		if err != nil {
			return nil, fmt.Errorf("bulk delete tasks: %w", err)
		}
	}

	return &BulkDeleteResponse{
		DeletedCount: deletedCount,
		FailedTasks:  failed,
	}, nil
}

func (s *service) BulkUpdateTasksStatus(ctx context.Context, req BulkUpdateTasksStatusRequest, userID uuid.UUID, isSuperAdmin bool) (*BulkUpdateStatusResponse, error) {
	if len(req.TaskIDs) == 0 && !req.All {
		return nil, apperror.BadRequest(`No task IDs provided and "all" flag not set`)
	}
	explicitIDs, err := parseUUIDSlice(req.TaskIDs)
	if err != nil {
		return nil, err
	}
	excludedIDs, err := parseUUIDSlice(req.ExcludedIDs)
	if err != nil {
		return nil, err
	}
	projID, err := parseOptUUID(req.ProjectID)
	if err != nil {
		return nil, err
	}
	wsID, err := parseOptUUID(req.WorkspaceID)
	if err != nil {
		return nil, err
	}
	sprintID, err := parseOptUUID(req.SprintID)
	if err != nil {
		return nil, err
	}
	targetStatusID, err := parseOptUUID(req.StatusID)
	if err != nil {
		return nil, err
	}

	var globalTargetName string
	if targetStatusID != nil {
		st, stErr := s.repo.GetStatusByID(ctx, *targetStatusID)
		if stErr != nil {
			return nil, apperror.NotFound("Target status not found")
		}
		globalTargetName = st.Name
	}

	searchStr := ""
	if req.Search != nil {
		searchStr = *req.Search
	}

	targetIDs, err := s.repo.ResolveBulkTaskIDs(
		ctx, req.All, explicitIDs, excludedIDs, projID, wsID, sprintID,
		searchStr, parseCSV(req.Statuses), parseCSV(req.Priorities), parseCSV(req.Types), parseCSV(req.Assignees), parseCSV(req.Reporters),
	)
	if err != nil {
		return nil, fmt.Errorf("resolve bulk update targets: %w", err)
	}

	updatedIDs := make([]uuid.UUID, 0, len(targetIDs))
	failed := make([]FailedTaskItem, 0)

	for _, tid := range targetIDs {
		taskRow, acc, aErr := s.requireTaskAccess(ctx, tid.String(), userID, isSuperAdmin)
		if aErr != nil {
			failed = append(failed, FailedTaskItem{ID: tid.String(), Reason: aErr.Error()})
			continue
		}
		if !acc.canChange {
			failed = append(failed, FailedTaskItem{ID: tid.String(), Reason: "Insufficient permissions to update this task"})
			continue
		}

		appliedStatus, sErr := s.repo.FindStatusInWorkflowByIDOrName(ctx, taskRow.ProjectID, targetStatusID)
		if sErr != nil {
			reason := `No "Done" status found for this project workflow`
			if targetStatusID != nil {
				name := globalTargetName
				if name == "" {
					name = "Unknown"
				}
				reason = fmt.Sprintf(`Status %q is not available for this project's workflow`, name)
			}
			failed = append(failed, FailedTaskItem{ID: tid.String(), Reason: reason})
			continue
		}

		params := sqlc.UpdateTaskCoreByIDParams{
			ID:          tid,
			StatusID:    pgtype.UUID{Bytes: appliedStatus.ID, Valid: true},
			UpdatedByID: pgtype.UUID{Bytes: userID, Valid: true},
		}
		if appliedStatus.Category == sqlc.StatusCategoryDONE {
			params.CompletedAt = pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
		} else {
			params.ClearCompletedAt = true
		}
		if uErr := s.repo.UpdateTaskTx(ctx, params, nil, nil); uErr != nil {
			failed = append(failed, FailedTaskItem{ID: tid.String(), Reason: uErr.Error()})
			continue
		}
		updatedIDs = append(updatedIDs, tid)
	}

	updatedTasks, _ := s.hydrateTaskListByIDs(ctx, updatedIDs)
	return &BulkUpdateStatusResponse{
		UpdatedCount: len(updatedTasks),
		UpdatedTasks: updatedTasks,
		FailedTasks:  failed,
	}, nil
}

func (s *service) BulkAssignTasks(ctx context.Context, req BulkAssignTasksRequest, userID uuid.UUID, isSuperAdmin bool) (*BulkAssignResponse, error) {
	if len(req.TaskIDs) == 0 && !req.All {
		return nil, apperror.BadRequest(`No task IDs provided and "all" flag not set`)
	}
	if req.AssigneeIDs == nil {
		return nil, apperror.BadRequest("assigneeIds must be provided (can be empty array to clear assignments)")
	}

	assigneeUUIDs, err := parseUUIDSlice(req.AssigneeIDs)
	if err != nil {
		return nil, err
	}
	if len(assigneeUUIDs) > 0 {
		if _, vErr := s.validateUserIDs(ctx, req.AssigneeIDs); vErr != nil {
			return nil, vErr
		}
	}

	explicitIDs, err := parseUUIDSlice(req.TaskIDs)
	if err != nil {
		return nil, err
	}
	excludedIDs, err := parseUUIDSlice(req.ExcludedIDs)
	if err != nil {
		return nil, err
	}
	projID, err := parseOptUUID(req.ProjectID)
	if err != nil {
		return nil, err
	}
	wsID, err := parseOptUUID(req.WorkspaceID)
	if err != nil {
		return nil, err
	}
	sprintID, err := parseOptUUID(req.SprintID)
	if err != nil {
		return nil, err
	}

	searchStr := ""
	if req.Search != nil {
		searchStr = *req.Search
	}
	targetIDs, err := s.repo.ResolveBulkTaskIDs(
		ctx, req.All, explicitIDs, excludedIDs, projID, wsID, sprintID,
		searchStr, parseCSV(req.Statuses), parseCSV(req.Priorities), parseCSV(req.Types), parseCSV(req.Assignees), parseCSV(req.Reporters),
	)
	if err != nil {
		return nil, fmt.Errorf("resolve bulk assign targets: %w", err)
	}

	allowed := make([]uuid.UUID, 0, len(targetIDs))
	failed := make([]FailedTaskItem, 0)
	for _, tid := range targetIDs {
		_, acc, aErr := s.requireTaskAccess(ctx, tid.String(), userID, isSuperAdmin)
		if aErr != nil {
			failed = append(failed, FailedTaskItem{ID: tid.String(), Reason: aErr.Error()})
			continue
		}
		if !acc.canChange {
			failed = append(failed, FailedTaskItem{ID: tid.String(), Reason: "Insufficient permissions to update this task"})
			continue
		}
		allowed = append(allowed, tid)
	}

	if len(allowed) > 0 {
		if bErr := s.repo.BulkReplaceAssignees(ctx, allowed, assigneeUUIDs); bErr != nil {
			return nil, fmt.Errorf("bulk replace assignees: %w", bErr)
		}
	}

	updatedTasks, _ := s.hydrateTaskListByIDs(ctx, allowed)
	return &BulkAssignResponse{
		AssignedCount: len(updatedTasks),
		UpdatedTasks:  updatedTasks,
		FailedTasks:   failed,
	}, nil
}
