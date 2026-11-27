package task_label

import (
	"context"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// Service defines business operations for task label assignments.
type Service interface {
	AssignLabel(ctx context.Context, req AssignRequest, userID uuid.UUID) (*AssignmentResponse, error)
	AssignMultiple(ctx context.Context, req AssignMultipleRequest, userID uuid.UUID) ([]AssignmentResponse, error)
	FindAll(ctx context.Context, userID uuid.UUID) ([]AssignmentResponse, error)
	Remove(ctx context.Context, taskIdentifier string, labelID, userID uuid.UUID) (*AssignmentResponse, error)
}

type service struct {
	repo Repository
}

// NewService constructs a task_label service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) AssignLabel(ctx context.Context, req AssignRequest, userID uuid.UUID) (*AssignmentResponse, error) {
	labelID, err := uuid.Parse(req.LabelID)
	if err != nil {
		return nil, apperror.BadRequest("labelId must be a UUID")
	}

	task, err := s.repo.FindTask(ctx, req.TaskID, userID)
	if err != nil {
		return nil, apperror.Internal("Failed to lookup task")
	}
	if task == nil {
		return nil, apperror.NotFound("Task not found")
	}
	if !task.IsWsMember {
		return nil, apperror.Forbidden("You do not have permission to modify labels in this project")
	}
	if task.ProjectArchived {
		return nil, apperror.BadRequest("Cannot assign label to task in archived project")
	}

	lbl, err := s.repo.GetLabelWorkspace(ctx, labelID, task.WorkspaceID)
	if err != nil {
		return nil, apperror.Internal("Failed to lookup label")
	}
	if lbl == nil {
		return nil, apperror.NotFound("Label not found")
	}
	if lbl.ProjectWorkspaceID != task.WorkspaceID && !lbl.InTargetWs {
		return nil, apperror.BadRequest("Label does not belong to the same workspace as the task")
	}

	exists, err := s.repo.CheckAssignmentExists(ctx, task.TaskID, labelID)
	if err != nil {
		return nil, apperror.Internal("Failed to check existing label assignment")
	}
	if exists {
		return nil, apperror.BadRequest("Label is already assigned to this task")
	}

	row, err := s.repo.AssignSingleTx(ctx, task.TaskID, labelID, task.WorkspaceID, userID)
	if err != nil || row == nil {
		return nil, apperror.Internal("Failed to assign label to task")
	}
	resp := mapPairRow(*row)
	return &resp, nil
}

func (s *service) AssignMultiple(ctx context.Context, req AssignMultipleRequest, userID uuid.UUID) ([]AssignmentResponse, error) {
	if len(req.LabelIDs) == 0 {
		return nil, apperror.BadRequest("labelIds array cannot be empty")
	}
	labelUUIDs := make([]uuid.UUID, 0, len(req.LabelIDs))
	seen := make(map[uuid.UUID]struct{}, len(req.LabelIDs))
	for _, raw := range req.LabelIDs {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			return nil, apperror.BadRequest("Each labelId must be a UUID")
		}
		if _, dup := seen[parsed]; !dup {
			seen[parsed] = struct{}{}
			labelUUIDs = append(labelUUIDs, parsed)
		}
	}

	task, err := s.repo.FindTask(ctx, req.TaskID, userID)
	if err != nil {
		return nil, apperror.Internal("Failed to lookup task")
	}
	if task == nil {
		return nil, apperror.NotFound("Task not found")
	}
	if !task.IsWsMember {
		return nil, apperror.Forbidden("You do not have permission to modify labels in this project")
	}
	if task.ProjectArchived {
		return nil, apperror.BadRequest("Cannot assign labels to task in archived project")
	}

	validIDs, err := s.repo.FindLabelsInWorkspace(ctx, labelUUIDs, task.WorkspaceID)
	if err != nil {
		return nil, apperror.Internal("Failed to verify workspace labels")
	}
	if len(validIDs) != len(labelUUIDs) {
		return nil, apperror.NotFound("One or more labels not found or do not belong to the same workspace")
	}

	rows, err := s.repo.ReplaceMultipleTx(ctx, task.TaskID, labelUUIDs, task.WorkspaceID, userID)
	if err != nil {
		return nil, apperror.Internal("Failed to assign labels to task")
	}
	out := make([]AssignmentResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapPairRow(r))
	}
	return out, nil
}

func (s *service) FindAll(ctx context.Context, userID uuid.UUID) ([]AssignmentResponse, error) {
	projectIDs, err := s.repo.ListAccessibleProjectIDs(ctx, userID)
	if err != nil {
		return nil, apperror.Internal("Failed to resolve accessible projects")
	}
	if len(projectIDs) == 0 {
		return nil, apperror.Forbidden("You do not have permission to view task labels")
	}

	rows, err := s.repo.ListByProjectIDs(ctx, projectIDs)
	if err != nil {
		return nil, apperror.Internal("Failed to list task labels")
	}
	out := make([]AssignmentResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapPairRow(sqlc.GetTaskLabelPairDetailsRow(r)))
	}
	return out, nil
}

func (s *service) Remove(ctx context.Context, taskIdentifier string, labelID, userID uuid.UUID) (*AssignmentResponse, error) {
	task, err := s.repo.FindTask(ctx, taskIdentifier, userID)
	if err != nil {
		return nil, apperror.Internal("Failed to lookup task")
	}
	if task == nil {
		return nil, apperror.NotFound("Task label assignment not found")
	}

	pair, err := s.repo.GetPairDetails(ctx, task.TaskID, labelID)
	if err != nil {
		return nil, apperror.Internal("Failed to lookup task label assignment")
	}
	if pair == nil {
		return nil, apperror.NotFound("Task label assignment not found")
	}
	if !task.IsWsMember {
		return nil, apperror.Forbidden("You do not have permission to modify labels in this project")
	}
	if task.ProjectArchived {
		return nil, apperror.BadRequest("Cannot remove label from task in archived project")
	}

	affected, err := s.repo.DeletePair(ctx, task.TaskID, labelID)
	if err != nil {
		return nil, apperror.Internal("Failed to remove label from task")
	}
	if affected == 0 {
		return nil, apperror.NotFound("Task label assignment not found")
	}

	resp := mapPairRow(*pair)
	return &resp, nil
}
