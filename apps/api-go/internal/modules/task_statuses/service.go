package task_statuses

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// Service defines business operations for task statuses.
type Service interface {
	Create(ctx context.Context, req CreateTaskStatusRequest, userID uuid.UUID, isSuperAdmin bool) (*TaskStatusResponse, error)
	CreateFromProject(ctx context.Context, req CreateTaskStatusFromProjectRequest, userID uuid.UUID, isSuperAdmin bool) (*TaskStatusResponse, error)
	FindAll(ctx context.Context, workflowID, orgID *uuid.UUID) ([]TaskStatusResponse, error)
	FindByProject(ctx context.Context, projectID uuid.UUID) ([]TaskStatusResponse, error)
	FindOne(ctx context.Context, id uuid.UUID) (*TaskStatusResponse, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateTaskStatusRequest, userID uuid.UUID, isSuperAdmin bool) (*TaskStatusResponse, error)
	UpdatePositions(ctx context.Context, req UpdatePositionsRequest, userID uuid.UUID, isSuperAdmin bool) ([]TaskStatusResponse, error)
	Remove(ctx context.Context, id, userID uuid.UUID, isSuperAdmin bool) error
	FindDeleted(ctx context.Context, workflowID *uuid.UUID) ([]TaskStatusResponse, error)
	Restore(ctx context.Context, id, userID uuid.UUID, isSuperAdmin bool) (*TaskStatusResponse, error)
}

type service struct {
	repo Repository
}

// NewService constructs a task_statuses service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func checkOrgAuthorization(ownerID uuid.UUID, memberRole sqlc.NullRole, userID uuid.UUID, isSuperAdmin bool, minRole middleware.Role) error {
	if isSuperAdmin || ownerID == userID {
		return nil
	}
	if !memberRole.Valid {
		return apperror.Forbidden("You are not a member of this organization")
	}
	if minRole != "" && middleware.RoleRank[middleware.Role(memberRole.Role)] < middleware.RoleRank[minRole] {
		return apperror.Forbidden(fmt.Sprintf("This action requires the %s role or higher in this organization", minRole))
	}
	return nil
}

func (s *service) assertOrgMemberForWorkflow(ctx context.Context, workflowID, userID uuid.UUID, isSuperAdmin bool, minRole middleware.Role) error {
	if isSuperAdmin {
		return nil
	}
	row, err := s.repo.GetWorkflowAuth(ctx, workflowID, userID)
	if err != nil {
		return apperror.NotFound("Workflow not found")
	}
	return checkOrgAuthorization(row.OwnerID, row.MemberRole, userID, isSuperAdmin, minRole)
}

func (s *service) assertOrgMemberForStatus(ctx context.Context, statusID, userID uuid.UUID, isSuperAdmin bool, minRole middleware.Role) (*sqlc.GetTaskStatusAuthByIDRow, error) {
	row, err := s.repo.GetStatusAuth(ctx, statusID, userID)
	if err != nil {
		return nil, apperror.NotFound("Task status not found")
	}
	if err := checkOrgAuthorization(row.OwnerID, row.MemberRole, userID, isSuperAdmin, minRole); err != nil {
		return nil, err
	}
	return row, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func safeCountInt32(n int64) int32 {
	if n < 0 {
		return 0
	}
	if n > 2147483647 {
		return 2147483647
	}
	return int32(n)
}

func (s *service) Create(ctx context.Context, req CreateTaskStatusRequest, userID uuid.UUID, isSuperAdmin bool) (*TaskStatusResponse, error) {
	workflowID, err := uuid.Parse(req.WorkflowID)
	if err != nil {
		return nil, apperror.NotFound("Workflow not found")
	}
	if authErr := s.assertOrgMemberForWorkflow(ctx, workflowID, userID, isSuperAdmin, ""); authErr != nil {
		return nil, authErr
	}

	existingDeleted, err := s.repo.FindSoftDeleted(ctx, workflowID, req.Name)
	if err != nil {
		return nil, fmt.Errorf("lookup soft-deleted status: %w", err)
	}
	if existingDeleted != nil {
		restored, rErr := s.repo.RestoreOnCreate(ctx, existingDeleted.ID, userID)
		if rErr != nil {
			if isUniqueViolation(rErr) {
				return nil, apperror.Conflict("Status with this name already exists in this workflow")
			}
			return nil, fmt.Errorf("restore soft-deleted status: %w", rErr)
		}
		return mapBasicStatus(restored), nil
	}

	count, err := s.repo.CountWorkflowStatuses(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("count workflow statuses: %w", err)
	}

	created, err := s.repo.CreateNode(ctx, sqlc.CreateTaskStatusNodeParams{
		ID:          uuid.New(),
		Name:        req.Name,
		Color:       req.Color,
		Category:    sqlc.StatusCategory(req.Category),
		Position:    safeCountInt32(count + 1),
		IsDefault:   false,
		WorkflowID:  workflowID,
		CreatedByID: pgtype.UUID{Bytes: userID, Valid: true},
		UpdatedByID: pgtype.UUID{Bytes: userID, Valid: true},
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, apperror.Conflict("Status with this name already exists in this workflow")
		}
		return nil, fmt.Errorf("create task status: %w", err)
	}

	full, err := s.repo.GetFullByID(ctx, created.ID)
	if err != nil {
		return mapBasicStatus(created), nil
	}
	return mapFullStatus(full), nil
}

func (s *service) CreateFromProject(ctx context.Context, req CreateTaskStatusFromProjectRequest, userID uuid.UUID, isSuperAdmin bool) (*TaskStatusResponse, error) {
	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return nil, apperror.NotFound("Project or associated workflow not found")
	}
	proj, err := s.repo.GetProjectWorkflow(ctx, projectID)
	if err != nil {
		return nil, apperror.NotFound("Project or associated workflow not found")
	}

	if authErr := s.assertOrgMemberForWorkflow(ctx, proj.WorkflowID, userID, isSuperAdmin, ""); authErr != nil {
		return nil, authErr
	}

	count, err := s.repo.CountWorkflowStatuses(ctx, proj.WorkflowID)
	if err != nil {
		return nil, fmt.Errorf("count workflow statuses: %w", err)
	}

	created, err := s.repo.CreateNode(ctx, sqlc.CreateTaskStatusNodeParams{
		ID:          uuid.New(),
		Name:        req.Name,
		Color:       DefaultFromProjectColor,
		Category:    DefaultFromProjectCategory,
		Position:    safeCountInt32(count),
		IsDefault:   count == 0,
		WorkflowID:  proj.WorkflowID,
		CreatedByID: pgtype.UUID{Bytes: userID, Valid: true},
		UpdatedByID: pgtype.UUID{Bytes: userID, Valid: true},
	})
	if err != nil {
		if isUniqueViolation(err) {
			return nil, apperror.Conflict("Status with this name already exists in this workflow")
		}
		return nil, fmt.Errorf("create status from project: %w", err)
	}

	full, err := s.repo.GetFullByID(ctx, created.ID)
	if err != nil {
		return mapBasicStatus(created), nil
	}
	return mapFullStatus(full), nil
}

func (s *service) FindAll(ctx context.Context, workflowID, orgID *uuid.UUID) ([]TaskStatusResponse, error) {
	var wfParam, orgParam pgtype.UUID
	if workflowID != nil {
		wfParam = pgtype.UUID{Bytes: *workflowID, Valid: true}
	} else if orgID != nil {
		orgParam = pgtype.UUID{Bytes: *orgID, Valid: true}
	}

	rows, err := s.repo.ListFiltered(ctx, wfParam, orgParam)
	if err != nil {
		return nil, fmt.Errorf("list task statuses: %w", err)
	}
	res := make([]TaskStatusResponse, 0, len(rows))
	for _, r := range rows {
		res = append(res, TaskStatusResponse{
			ID:          r.ID.String(),
			Name:        r.Name,
			Color:       r.Color,
			Category:    string(r.Category),
			Position:    r.Position,
			IsDefault:   r.IsDefault,
			WorkflowID:  r.WorkflowID.String(),
			CreatedAt:   formatTS(r.CreatedAt),
			UpdatedAt:   formatTS(r.UpdatedAt),
			CreatedByID: pgUUIDPtr(r.CreatedByID),
			UpdatedByID: pgUUIDPtr(r.UpdatedByID),
			DeletedAt:   pgTSPtr(r.DeletedAt),
			DeletedBy:   pgUUIDPtr(r.DeletedBy),
			Workflow: &WorkflowBrief{
				ID:          r.WorkflowID.String(),
				Name:        r.WorkflowName,
				Description: pgTextPtr(r.WorkflowDescription),
				Organization: &OrgBrief{
					ID:   r.OrgID.String(),
					Name: r.OrgName,
					Slug: r.OrgSlug,
				},
			},
			Count: &TaskCount{Tasks: r.TaskCount},
		})
	}
	return res, nil
}

func (s *service) FindByProject(ctx context.Context, projectID uuid.UUID) ([]TaskStatusResponse, error) {
	rows, err := s.repo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("list project statuses: %w", err)
	}
	res := make([]TaskStatusResponse, 0, len(rows))
	for _, r := range rows {
		res = append(res, TaskStatusResponse{
			ID:          r.ID.String(),
			Name:        r.Name,
			Color:       r.Color,
			Category:    string(r.Category),
			Position:    r.Position,
			IsDefault:   r.IsDefault,
			WorkflowID:  r.WorkflowID.String(),
			CreatedAt:   formatTS(r.CreatedAt),
			UpdatedAt:   formatTS(r.UpdatedAt),
			CreatedByID: pgUUIDPtr(r.CreatedByID),
			UpdatedByID: pgUUIDPtr(r.UpdatedByID),
			DeletedAt:   pgTSPtr(r.DeletedAt),
			DeletedBy:   pgUUIDPtr(r.DeletedBy),
			Workflow: &WorkflowBrief{
				ID:          r.WorkflowID.String(),
				Name:        r.WorkflowName,
				Description: pgTextPtr(r.WorkflowDescription),
				Organization: &OrgBrief{
					ID:   r.OrgID.String(),
					Name: r.OrgName,
					Slug: r.OrgSlug,
				},
			},
			Count: &TaskCount{Tasks: r.TaskCount},
		})
	}
	return res, nil
}

func (s *service) FindOne(ctx context.Context, id uuid.UUID) (*TaskStatusResponse, error) {
	row, err := s.repo.GetActiveFullByID(ctx, id)
	if err != nil {
		return nil, apperror.NotFound("Task status not found")
	}
	fromRows, err := s.repo.ListFromTransitions(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list outgoing transitions: %w", err)
	}
	toRows, err := s.repo.ListToTransitions(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("list incoming transitions: %w", err)
	}

	fromList := make([]StatusTransitionItem, 0, len(fromRows))
	for _, t := range fromRows {
		fromList = append(fromList, StatusTransitionItem{
			ID:           t.ID.String(),
			Name:         pgTextPtr(t.Name),
			Description:  pgTextPtr(t.Description),
			WorkflowID:   t.WorkflowID.String(),
			FromStatusID: t.FromStatusID.String(),
			ToStatusID:   t.ToStatusID.String(),
			CreatedByID:  pgUUIDPtr(t.CreatedByID),
			UpdatedByID:  pgUUIDPtr(t.UpdatedByID),
			CreatedAt:    formatTS(t.CreatedAt),
			UpdatedAt:    formatTS(t.UpdatedAt),
			ToStatus: &TransitionStatusBrief{
				ID:       t.StatusID.String(),
				Name:     t.StatusName,
				Color:    t.StatusColor,
				Category: string(t.StatusCategory),
			},
		})
	}

	toList := make([]StatusTransitionItem, 0, len(toRows))
	for _, t := range toRows {
		toList = append(toList, StatusTransitionItem{
			ID:           t.ID.String(),
			Name:         pgTextPtr(t.Name),
			Description:  pgTextPtr(t.Description),
			WorkflowID:   t.WorkflowID.String(),
			FromStatusID: t.FromStatusID.String(),
			ToStatusID:   t.ToStatusID.String(),
			CreatedByID:  pgUUIDPtr(t.CreatedByID),
			UpdatedByID:  pgUUIDPtr(t.UpdatedByID),
			CreatedAt:    formatTS(t.CreatedAt),
			UpdatedAt:    formatTS(t.UpdatedAt),
			FromStatus: &TransitionStatusBrief{
				ID:       t.StatusID.String(),
				Name:     t.StatusName,
				Color:    t.StatusColor,
				Category: string(t.StatusCategory),
			},
		})
	}

	return &TaskStatusResponse{
		ID:          row.ID.String(),
		Name:        row.Name,
		Color:       row.Color,
		Category:    string(row.Category),
		Position:    row.Position,
		IsDefault:   row.IsDefault,
		WorkflowID:  row.WorkflowID.String(),
		CreatedAt:   formatTS(row.CreatedAt),
		UpdatedAt:   formatTS(row.UpdatedAt),
		CreatedByID: pgUUIDPtr(row.CreatedByID),
		UpdatedByID: pgUUIDPtr(row.UpdatedByID),
		DeletedAt:   pgTSPtr(row.DeletedAt),
		DeletedBy:   pgUUIDPtr(row.DeletedBy),
		Workflow: &WorkflowBrief{
			ID:          row.WorkflowID.String(),
			Name:        row.WorkflowName,
			Description: pgTextPtr(row.WorkflowDescription),
			Organization: &OrgBrief{
				ID:   row.OrgID.String(),
				Name: row.OrgName,
				Slug: row.OrgSlug,
			},
		},
		FromTransitions: fromList,
		ToTransitions:   toList,
		Count:           &TaskCount{Tasks: row.TaskCount},
	}, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req UpdateTaskStatusRequest, userID uuid.UUID, isSuperAdmin bool) (*TaskStatusResponse, error) {
	if _, err := s.assertOrgMemberForStatus(ctx, id, userID, isSuperAdmin, middleware.RoleManager); err != nil {
		return nil, err
	}

	params := sqlc.UpdateTaskStatusByIDParams{
		ID:          id,
		UpdatedByID: pgtype.UUID{Bytes: userID, Valid: true},
	}
	if req.Name != nil {
		params.Name = pgtype.Text{String: *req.Name, Valid: true}
	}
	if req.Color != nil {
		params.Color = pgtype.Text{String: *req.Color, Valid: true}
	}
	if req.Category != nil {
		params.Category = sqlc.NullStatusCategory{StatusCategory: sqlc.StatusCategory(*req.Category), Valid: true}
	}
	if req.Position != nil {
		params.Position = pgtype.Int4{Int32: *req.Position, Valid: true}
	}

	updatedID, err := s.repo.UpdateByID(ctx, params)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, apperror.Conflict("Status with this name already exists in this workflow")
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("Task status not found")
		}
		return nil, fmt.Errorf("update task status: %w", err)
	}

	full, err := s.repo.GetFullByID(ctx, updatedID)
	if err != nil {
		return nil, apperror.NotFound("Task status not found")
	}
	return mapFullStatus(full), nil
}

func (s *service) UpdatePositions(ctx context.Context, req UpdatePositionsRequest, userID uuid.UUID, isSuperAdmin bool) ([]TaskStatusResponse, error) {
	updates := make([]PositionUpdate, 0, len(req.StatusUpdates))
	for _, item := range req.StatusUpdates {
		statusID, err := uuid.Parse(item.ID)
		if err != nil {
			return nil, apperror.NotFound("Task status not found")
		}
		if _, err := s.assertOrgMemberForStatus(ctx, statusID, userID, isSuperAdmin, ""); err != nil {
			return nil, err
		}
		updates = append(updates, PositionUpdate{ID: statusID, Position: item.Position})
	}

	rows, err := s.repo.UpdatePositionsTx(ctx, updates, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("One or more task statuses not found")
		}
		return nil, fmt.Errorf("update task status positions: %w", err)
	}

	out := make([]TaskStatusResponse, 0, len(rows))
	for i := range rows {
		out = append(out, *mapFullStatus(&rows[i]))
	}
	return out, nil
}

func (s *service) Remove(ctx context.Context, id, userID uuid.UUID, isSuperAdmin bool) error {
	status, err := s.assertOrgMemberForStatus(ctx, id, userID, isSuperAdmin, middleware.RoleManager)
	if err != nil {
		return err
	}
	if status.DeletedAt.Valid {
		return apperror.Conflict("This status has already been deleted")
	}

	if err := s.repo.SoftDeleteTx(ctx, id, status.WorkflowID, status.Position, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperror.NotFound("Task status not found")
		}
		return fmt.Errorf("soft-delete task status: %w", err)
	}
	return nil
}

func (s *service) FindDeleted(ctx context.Context, workflowID *uuid.UUID) ([]TaskStatusResponse, error) {
	var wfParam pgtype.UUID
	if workflowID != nil {
		wfParam = pgtype.UUID{Bytes: *workflowID, Valid: true}
	}
	rows, err := s.repo.ListDeleted(ctx, wfParam)
	if err != nil {
		return nil, fmt.Errorf("list deleted task statuses: %w", err)
	}

	res := make([]TaskStatusResponse, 0, len(rows))
	for _, r := range rows {
		var delUser *UserBrief
		if r.DeletedByUserID.Valid {
			delUser = &UserBrief{
				ID:        uuid.UUID(r.DeletedByUserID.Bytes).String(),
				Email:     r.DeletedByEmail.String,
				FirstName: pgTextPtr(r.DeletedByFirstName),
				LastName:  pgTextPtr(r.DeletedByLastName),
			}
		}
		res = append(res, TaskStatusResponse{
			ID:          r.ID.String(),
			Name:        r.Name,
			Color:       r.Color,
			Category:    string(r.Category),
			Position:    r.Position,
			IsDefault:   r.IsDefault,
			WorkflowID:  r.WorkflowID.String(),
			CreatedAt:   formatTS(r.CreatedAt),
			UpdatedAt:   formatTS(r.UpdatedAt),
			CreatedByID: pgUUIDPtr(r.CreatedByID),
			UpdatedByID: pgUUIDPtr(r.UpdatedByID),
			DeletedAt:   pgTSPtr(r.DeletedAt),
			DeletedBy:   pgUUIDPtr(r.DeletedBy),
			Workflow: &WorkflowBrief{
				ID:          r.WorkflowID.String(),
				Name:        r.WorkflowName,
				Description: pgTextPtr(r.WorkflowDescription),
				Organization: &OrgBrief{
					ID:   r.OrgID.String(),
					Name: r.OrgName,
					Slug: r.OrgSlug,
				},
			},
			DeletedByUser: delUser,
		})
	}
	return res, nil
}

func (s *service) Restore(ctx context.Context, id, userID uuid.UUID, isSuperAdmin bool) (*TaskStatusResponse, error) {
	status, err := s.assertOrgMemberForStatus(ctx, id, userID, isSuperAdmin, middleware.RoleManager)
	if err != nil {
		return nil, err
	}
	if !status.DeletedAt.Valid {
		return nil, apperror.Conflict("This status is not deleted and cannot be restored")
	}

	conflict, err := s.repo.CheckActiveNameConflict(ctx, status.WorkflowID, status.Name, id)
	if err != nil {
		return nil, fmt.Errorf("check active status conflict: %w", err)
	}
	if conflict {
		return nil, apperror.Conflict("A status with this name already exists in the workflow")
	}

	restoredID, err := s.repo.RestoreByID(ctx, id, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("Task status not found")
		}
		return nil, fmt.Errorf("restore task status: %w", err)
	}
	full, err := s.repo.GetFullByID(ctx, restoredID)
	if err != nil {
		return nil, apperror.NotFound("Task status not found")
	}
	return mapFullStatus(full), nil
}

func mapBasicStatus(s *sqlc.TaskStatus) *TaskStatusResponse {
	return &TaskStatusResponse{
		ID:          s.ID.String(),
		Name:        s.Name,
		Color:       s.Color,
		Category:    string(s.Category),
		Position:    s.Position,
		IsDefault:   s.IsDefault,
		WorkflowID:  s.WorkflowID.String(),
		CreatedAt:   formatTS(s.CreatedAt),
		UpdatedAt:   formatTS(s.UpdatedAt),
		CreatedByID: pgUUIDPtr(s.CreatedByID),
		UpdatedByID: pgUUIDPtr(s.UpdatedByID),
		DeletedAt:   pgTSPtr(s.DeletedAt),
		DeletedBy:   pgUUIDPtr(s.DeletedBy),
	}
}

func mapFullStatus(r *sqlc.GetTaskStatusFullByIDRow) *TaskStatusResponse {
	var createdUser, updatedUser *UserBrief
	if r.CreatedByUserID.Valid {
		createdUser = &UserBrief{
			ID:        uuid.UUID(r.CreatedByUserID.Bytes).String(),
			Email:     r.CreatedByEmail.String,
			FirstName: pgTextPtr(r.CreatedByFirstName),
			LastName:  pgTextPtr(r.CreatedByLastName),
		}
	}
	if r.UpdatedByUserID.Valid {
		updatedUser = &UserBrief{
			ID:        uuid.UUID(r.UpdatedByUserID.Bytes).String(),
			Email:     r.UpdatedByEmail.String,
			FirstName: pgTextPtr(r.UpdatedByFirstName),
			LastName:  pgTextPtr(r.UpdatedByLastName),
		}
	}
	return &TaskStatusResponse{
		ID:          r.ID.String(),
		Name:        r.Name,
		Color:       r.Color,
		Category:    string(r.Category),
		Position:    r.Position,
		IsDefault:   r.IsDefault,
		WorkflowID:  r.WorkflowID.String(),
		CreatedAt:   formatTS(r.CreatedAt),
		UpdatedAt:   formatTS(r.UpdatedAt),
		CreatedByID: pgUUIDPtr(r.CreatedByID),
		UpdatedByID: pgUUIDPtr(r.UpdatedByID),
		DeletedAt:   pgTSPtr(r.DeletedAt),
		DeletedBy:   pgUUIDPtr(r.DeletedBy),
		Workflow: &WorkflowBrief{
			ID:          r.WorkflowID.String(),
			Name:        r.WorkflowName,
			Description: pgTextPtr(r.WorkflowDescription),
			Organization: &OrgBrief{
				ID:   r.OrgID.String(),
				Name: r.OrgName,
				Slug: r.OrgSlug,
			},
		},
		CreatedByUser: createdUser,
		UpdatedByUser: updatedUser,
		Count:         &TaskCount{Tasks: r.TaskCount},
	}
}
