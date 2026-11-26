package labels

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// Service defines business operations for labels.
type Service interface {
	Create(ctx context.Context, req CreateLabelRequest, userID uuid.UUID) (*LabelResponse, error)
	FindAll(ctx context.Context, projectID *uuid.UUID, userID uuid.UUID) ([]LabelResponse, error)
	FindOne(ctx context.Context, id, userID uuid.UUID) (*LabelResponse, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateLabelRequest, userID uuid.UUID) (*LabelResponse, error)
	Remove(ctx context.Context, id, userID uuid.UUID) error
	AssignLabelToTask(ctx context.Context, req AssignLabelRequest, userID uuid.UUID) error
	RemoveLabelFromTask(ctx context.Context, taskIDOrSlug string, labelID, userID uuid.UUID) error
	AssignMultipleLabelsToTask(ctx context.Context, req AssignMultipleLabelsRequest, userID uuid.UUID) error
	GetTaskLabels(ctx context.Context, taskIDOrSlug string, userID uuid.UUID) ([]LabelResponse, error)
}

type service struct {
	repo Repository
}

// NewService constructs a labels service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (s *service) Create(ctx context.Context, req CreateLabelRequest, userID uuid.UUID) (*LabelResponse, error) {
	projectID := uuid.MustParse(req.ProjectID)
	proj, err := s.repo.GetProjectWorkspace(ctx, projectID, userID)
	if err != nil {
		return nil, apperror.NotFound("Project not found")
	}
	if !proj.IsWsMember {
		return nil, apperror.Forbidden("You do not have permission to create labels in this project")
	}

	var desc pgtype.Text
	if req.Description != nil {
		desc = pgtype.Text{String: *req.Description, Valid: true}
	}

	labelID := uuid.New()
	_, err = s.repo.CreateWithWorkspaceTx(ctx, sqlc.CreateLabelRecordParams{
		ID:          labelID,
		Name:        req.Name,
		Color:       req.Color,
		Description: desc,
		ProjectID:   projectID,
		CreatedByID: pgtype.UUID{Bytes: userID, Valid: true},
		UpdatedByID: pgtype.UUID{Bytes: userID, Valid: true},
	}, proj.WorkspaceID)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, apperror.Conflict("Label with this name already exists in this project")
		}
		return nil, fmt.Errorf("create label: %w", err)
	}

	details, err := s.repo.GetDetailsByID(ctx, labelID, userID)
	if err != nil {
		return nil, apperror.NotFound("Label not found")
	}
	return mapDetailsRow(details, true, nil), nil
}

func (s *service) FindAll(ctx context.Context, projectID *uuid.UUID, userID uuid.UUID) ([]LabelResponse, error) {
	if projectID != nil {
		proj, err := s.repo.GetProjectWorkspace(ctx, *projectID, userID)
		if err != nil {
			return nil, apperror.NotFound("Project not found")
		}
		if !proj.IsWsMember {
			return nil, apperror.Forbidden("You do not have permission to view labels in this project")
		}
		rows, err := s.repo.ListByWorkspaceID(ctx, proj.WorkspaceID)
		if err != nil {
			return nil, fmt.Errorf("list workspace labels: %w", err)
		}
		out := make([]LabelResponse, 0, len(rows))
		for _, r := range rows {
			out = append(out, LabelResponse{
				ID:          r.ID.String(),
				Name:        r.Name,
				Color:       r.Color,
				Description: pgTextPtr(r.Description),
				ProjectID:   r.ProjectID.String(),
				CreatedByID: pgUUIDPtr(r.CreatedByID),
				UpdatedByID: pgUUIDPtr(r.UpdatedByID),
				CreatedAt:   formatTS(r.CreatedAt),
				UpdatedAt:   formatTS(r.UpdatedAt),
				Project: &ProjectBrief{
					ID:   r.ProjectID.String(),
					Name: r.ProjectName,
					Slug: r.ProjectSlug,
					Workspace: &WorkspaceBrief{
						ID:   r.WorkspaceID.String(),
						Name: r.WorkspaceName,
						Slug: r.WorkspaceSlug,
					},
				},
				Count: &LabelCount{TaskLabels: r.TaskLabelCount},
			})
		}
		return out, nil
	}

	rows, err := s.repo.ListAccessibleForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list accessible labels: %w", err)
	}
	out := make([]LabelResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, LabelResponse{
			ID:          r.ID.String(),
			Name:        r.Name,
			Color:       r.Color,
			Description: pgTextPtr(r.Description),
			ProjectID:   r.ProjectID.String(),
			CreatedByID: pgUUIDPtr(r.CreatedByID),
			UpdatedByID: pgUUIDPtr(r.UpdatedByID),
			CreatedAt:   formatTS(r.CreatedAt),
			UpdatedAt:   formatTS(r.UpdatedAt),
			Project: &ProjectBrief{
				ID:   r.ProjectID.String(),
				Name: r.ProjectName,
				Slug: r.ProjectSlug,
				Workspace: &WorkspaceBrief{
					ID:   r.WorkspaceID.String(),
					Name: r.WorkspaceName,
					Slug: r.WorkspaceSlug,
				},
			},
			Count: &LabelCount{TaskLabels: r.TaskLabelCount},
		})
	}
	return out, nil
}

func (s *service) FindOne(ctx context.Context, id, userID uuid.UUID) (*LabelResponse, error) {
	details, err := s.repo.GetDetailsByID(ctx, id, userID)
	if err != nil {
		return nil, apperror.NotFound("Label not found")
	}
	if !details.IsWsMember {
		return nil, apperror.Forbidden("You do not have permission to view this label")
	}

	entries, err := s.repo.ListTaskLabelEntries(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list label tasks: %w", err)
	}

	taskIDs := make([]uuid.UUID, 0, len(entries))
	for _, e := range entries {
		taskIDs = append(taskIDs, e.TaskID)
	}
	assigneeRows, err := s.repo.ListTaskAssignees(ctx, taskIDs)
	if err != nil {
		return nil, fmt.Errorf("list task assignees: %w", err)
	}
	assigneeMap := make(map[uuid.UUID][]AssigneeBrief)
	for _, a := range assigneeRows {
		first := a.FirstName
		last := a.LastName
		assigneeMap[a.TaskID] = append(assigneeMap[a.TaskID], AssigneeBrief{
			ID:        a.ID.String(),
			FirstName: &first,
			LastName:  &last,
			Avatar:    pgTextPtr(a.Avatar),
		})
	}

	tlRes := make([]TaskLabelEntry, 0, len(entries))
	for _, e := range entries {
		assignees := assigneeMap[e.TaskID]
		if assignees == nil {
			assignees = []AssigneeBrief{}
		}
		tlRes = append(tlRes, TaskLabelEntry{
			TaskID:      e.TaskID.String(),
			LabelID:     e.LabelID.String(),
			CreatedByID: pgUUIDPtr(e.CreatedByID),
			UpdatedByID: pgUUIDPtr(e.UpdatedByID),
			CreatedAt:   formatTS(e.CreatedAt),
			UpdatedAt:   formatTS(e.UpdatedAt),
			Task: &TaskSummary{
				ID:       e.TaskID.String(),
				Title:    e.TaskTitle,
				Slug:     e.TaskSlug,
				Type:     string(e.TaskType),
				Priority: string(e.TaskPriority),
				Status: &TaskStatusBrief{
					ID:       e.StatusID.String(),
					Name:     e.StatusName,
					Color:    e.StatusColor,
					Category: string(e.StatusCategory),
				},
				Assignees: assignees,
			},
		})
	}

	return mapDetailsRow(details, false, tlRes), nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req UpdateLabelRequest, userID uuid.UUID) (*LabelResponse, error) {
	details, err := s.repo.GetDetailsByID(ctx, id, userID)
	if err != nil {
		return nil, apperror.NotFound("Label not found")
	}
	if !details.IsWsMember {
		return nil, apperror.Forbidden("You do not have permission to update labels in this project")
	}

	params := sqlc.UpdateLabelByIDParams{
		ID:          id,
		UpdatedByID: pgtype.UUID{Bytes: userID, Valid: true},
	}
	if req.Name != nil {
		params.Name = pgtype.Text{String: *req.Name, Valid: true}
	}
	if req.Color != nil {
		params.Color = pgtype.Text{String: *req.Color, Valid: true}
	}
	if req.Description != nil {
		params.Description = pgtype.Text{String: *req.Description, Valid: true}
	}

	updatedID, err := s.repo.UpdateByID(ctx, params)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, apperror.Conflict("Label with this name already exists in this project")
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("Label not found")
		}
		return nil, fmt.Errorf("update label: %w", err)
	}

	updated, err := s.repo.GetDetailsByID(ctx, updatedID, userID)
	if err != nil {
		return nil, apperror.NotFound("Label not found")
	}
	return mapDetailsRow(updated, false, nil), nil
}

func (s *service) Remove(ctx context.Context, id, userID uuid.UUID) error {
	details, err := s.repo.GetDetailsByID(ctx, id, userID)
	if err != nil {
		return apperror.NotFound("Label not found")
	}
	if !details.IsWsMember {
		return apperror.Forbidden("You do not have permission to delete labels in this project")
	}

	affected, err := s.repo.DeleteByID(ctx, id)
	if err != nil {
		return fmt.Errorf("delete label: %w", err)
	}
	if affected == 0 {
		return apperror.NotFound("Label not found")
	}
	return nil
}

func (s *service) resolveTaskAccess(ctx context.Context, taskIDOrSlug string, userID uuid.UUID, forbidMsg string) (*sqlc.FindTaskWorkspaceAccessByIDOrSlugRow, error) {
	task, err := s.repo.FindTaskWorkspaceAccess(ctx, taskIDOrSlug, userID)
	if err != nil {
		return nil, apperror.NotFound("Task not found")
	}
	if !task.IsWsMember {
		return nil, apperror.Forbidden(forbidMsg)
	}
	return task, nil
}

func (s *service) AssignLabelToTask(ctx context.Context, req AssignLabelRequest, userID uuid.UUID) error {
	task, err := s.resolveTaskAccess(ctx, req.TaskID, userID, "You do not have permission to modify labels in this project")
	if err != nil {
		return err
	}

	labelID := uuid.MustParse(req.LabelID)
	lbl, err := s.repo.GetLabelWorkspaceCheck(ctx, labelID, task.WorkspaceID)
	if err != nil {
		return apperror.NotFound("Label not found")
	}
	if task.WorkspaceID != lbl.ProjectWorkspaceID && !lbl.InTargetWorkspace {
		return apperror.Conflict("Task and label must belong to the same workspace or be inherited")
	}

	if err := s.repo.AssignSingleTx(ctx, task.TaskID, labelID, task.WorkspaceID); err != nil {
		if isUniqueViolation(err) {
			return apperror.Conflict("Label is already assigned to this task")
		}
		return fmt.Errorf("assign label to task: %w", err)
	}
	return nil
}

func (s *service) RemoveLabelFromTask(ctx context.Context, taskIDOrSlug string, labelID, userID uuid.UUID) error {
	task, err := s.resolveTaskAccess(ctx, taskIDOrSlug, userID, "You do not have permission to modify labels in this project")
	if err != nil {
		return err
	}
	affected, err := s.repo.RemoveTaskLabel(ctx, task.TaskID, labelID)
	if err != nil {
		return fmt.Errorf("remove label from task: %w", err)
	}
	if affected == 0 {
		return apperror.NotFound("Label assignment not found")
	}
	return nil
}

func (s *service) AssignMultipleLabelsToTask(ctx context.Context, req AssignMultipleLabelsRequest, userID uuid.UUID) error {
	if len(req.LabelIDs) == 0 {
		return apperror.BadRequest("labelIds array cannot be empty")
	}

	task, err := s.resolveTaskAccess(ctx, req.TaskID, userID, "You do not have permission to modify labels in this project")
	if err != nil {
		return err
	}

	uniqueIDs := make([]uuid.UUID, 0, len(req.LabelIDs))
	seen := make(map[uuid.UUID]struct{}, len(req.LabelIDs))
	for _, raw := range req.LabelIDs {
		id := uuid.MustParse(raw)
		if _, exists := seen[id]; !exists {
			seen[id] = struct{}{}
			uniqueIDs = append(uniqueIDs, id)
		}
	}

	validIDs, err := s.repo.ListWorkspaceValidLabelIDs(ctx, uniqueIDs, task.WorkspaceID)
	if err != nil {
		return fmt.Errorf("verify label workspace membership: %w", err)
	}
	if len(validIDs) != len(uniqueIDs) {
		return apperror.NotFound("One or more labels not found or do not belong to the same workspace")
	}

	if err := s.repo.ReplaceTaskLabelsTx(ctx, task.TaskID, task.WorkspaceID, uniqueIDs); err != nil {
		return fmt.Errorf("replace task labels: %w", err)
	}
	return nil
}

func (s *service) GetTaskLabels(ctx context.Context, taskIDOrSlug string, userID uuid.UUID) ([]LabelResponse, error) {
	task, err := s.resolveTaskAccess(ctx, taskIDOrSlug, userID, "You do not have permission to view labels in this project")
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.ListLabelsByTaskID(ctx, task.TaskID)
	if err != nil {
		return nil, fmt.Errorf("list task labels: %w", err)
	}
	out := make([]LabelResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, LabelResponse{
			ID:          r.ID.String(),
			Name:        r.Name,
			Color:       r.Color,
			Description: pgTextPtr(r.Description),
			ProjectID:   r.ProjectID.String(),
			CreatedByID: pgUUIDPtr(r.CreatedByID),
			UpdatedByID: pgUUIDPtr(r.UpdatedByID),
			CreatedAt:   formatTS(r.CreatedAt),
			UpdatedAt:   formatTS(r.UpdatedAt),
			Project: &ProjectBrief{
				ID:   r.ProjectID.String(),
				Name: r.ProjectName,
				Slug: r.ProjectSlug,
			},
		})
	}
	return out, nil
}
