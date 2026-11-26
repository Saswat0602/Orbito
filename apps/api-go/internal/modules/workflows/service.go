package workflows

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// Service defines business operations for workflows.
type Service interface {
	Create(ctx context.Context, req CreateWorkflowRequest, currentUserID uuid.UUID, isSuperAdmin bool) (*WorkflowResponse, error)
	FindAll(ctx context.Context, orgID *uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) ([]WorkflowResponse, error)
	FindByOrganizationSlug(ctx context.Context, slug string, currentUserID uuid.UUID, isSuperAdmin bool) ([]WorkflowResponse, error)
	FindOne(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) (*WorkflowResponse, error)
	GetDefaultWorkflow(ctx context.Context, orgID uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) (*WorkflowResponse, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateWorkflowRequest, currentUserID uuid.UUID, isSuperAdmin bool) (*WorkflowResponse, error)
	MakeDefault(ctx context.Context, id uuid.UUID, orgID uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) (*WorkflowResponse, error)
	Delete(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) error
}

type service struct {
	repo Repository
}

// NewService constructs a new workflows service instance.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) assertOrgMember(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, isSuperAdmin bool, minRole middleware.Role) error {
	if isSuperAdmin {
		return nil
	}

	row, err := s.repo.GetOrgOwnerAndUserRole(ctx, orgID, userID)
	if err != nil {
		return apperror.NotFound("Organization not found")
	}

	if row.OwnerID == userID {
		return nil
	}

	if !row.MemberRole.Valid {
		return apperror.Forbidden("You are not a member of this organization")
	}

	if minRole != "" {
		memberRole := middleware.Role(row.MemberRole.Role)
		if middleware.RoleRank[memberRole] < middleware.RoleRank[minRole] {
			return apperror.Forbidden(fmt.Sprintf("This action requires the %s role or higher in this organization", minRole))
		}
	}

	return nil
}

func (s *service) assertOrgMemberForWorkflow(ctx context.Context, workflowID uuid.UUID, userID uuid.UUID, isSuperAdmin bool, minRole middleware.Role) (*sqlc.GetWorkflowOrgOwnerAndUserRoleRow, error) {
	row, err := s.repo.GetWorkflowOrgOwnerAndUserRole(ctx, workflowID, userID)
	if err != nil {
		return nil, apperror.NotFound("Workflow not found")
	}

	if isSuperAdmin || row.OwnerID == userID {
		return row, nil
	}

	if !row.MemberRole.Valid {
		return nil, apperror.Forbidden("You are not a member of this organization")
	}

	if minRole != "" {
		memberRole := middleware.Role(row.MemberRole.Role)
		if middleware.RoleRank[memberRole] < middleware.RoleRank[minRole] {
			return nil, apperror.Forbidden(fmt.Sprintf("This action requires the %s role or higher in this organization", minRole))
		}
	}

	return row, nil
}

func (s *service) Create(ctx context.Context, req CreateWorkflowRequest, currentUserID uuid.UUID, isSuperAdmin bool) (*WorkflowResponse, error) {
	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		return nil, apperror.BadRequest("Invalid organization ID format")
	}

	if err := s.assertOrgMember(ctx, orgID, currentUserID, isSuperAdmin, ""); err != nil {
		return nil, err
	}

	wfID := uuid.New()
	isDefault := false
	if req.IsDefault != nil && *req.IsDefault {
		isDefault = true
	}

	desc := pgtype.Text{}
	if req.Description != nil {
		desc = pgtype.Text{String: *req.Description, Valid: true}
	}

	wfParams := sqlc.CreateWorkflowParams{
		ID:             wfID,
		Name:           req.Name,
		Description:    desc,
		IsDefault:      isDefault,
		OrganizationID: orgID,
		CreatedByID:    pgtype.UUID{Bytes: currentUserID, Valid: true},
		UpdatedByID:    pgtype.UUID{Bytes: currentUserID, Valid: true},
	}

	statusParams := make([]sqlc.CreateTaskStatusParams, 0, len(DefaultTaskStatuses))
	for _, status := range DefaultTaskStatuses {
		statusParams = append(statusParams, sqlc.CreateTaskStatusParams{
			ID:          uuid.New(),
			Name:        status.Name,
			Color:       status.Color,
			Category:    status.Category,
			Position:    status.Position,
			IsDefault:   status.IsDefault,
			WorkflowID:  wfID,
			CreatedByID: pgtype.UUID{Bytes: currentUserID, Valid: true},
			UpdatedByID: pgtype.UUID{Bytes: currentUserID, Valid: true},
		})
	}

	if _, err := s.repo.Create(ctx, wfParams, statusParams); err != nil {
		return nil, apperror.Internal("Failed to create workflow: " + err.Error())
	}

	return s.FindOne(ctx, wfID, currentUserID, isSuperAdmin)
}

func (s *service) FindAll(ctx context.Context, orgID *uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) ([]WorkflowResponse, error) {
	if orgID != nil {
		if err := s.assertOrgMember(ctx, *orgID, currentUserID, isSuperAdmin, ""); err != nil {
			return nil, err
		}
		rows, err := s.repo.ListByOrgID(ctx, *orgID)
		if err != nil {
			return nil, apperror.Internal("Failed to list workflows: " + err.Error())
		}
		return s.mapOrgWorkflowRows(ctx, rows)
	}

	if isSuperAdmin {
		rows, err := s.repo.ListAll(ctx)
		if err != nil {
			return nil, apperror.Internal("Failed to list workflows: " + err.Error())
		}
		return s.mapAllWorkflowRows(ctx, rows)
	}

	accessibleOrgs, err := s.repo.GetUserAccessibleOrgIDs(ctx, currentUserID)
	if err != nil {
		return nil, apperror.Internal("Failed to get accessible organizations: " + err.Error())
	}
	if len(accessibleOrgs) == 0 {
		return []WorkflowResponse{}, nil
	}

	rows, err := s.repo.ListByOrgIDs(ctx, accessibleOrgs)
	if err != nil {
		return nil, apperror.Internal("Failed to list workflows: " + err.Error())
	}
	return s.mapOrgIDsWorkflowRows(ctx, rows)
}

func (s *service) FindByOrganizationSlug(ctx context.Context, slug string, currentUserID uuid.UUID, isSuperAdmin bool) ([]WorkflowResponse, error) {
	trimmed := strings.TrimSpace(slug)
	if trimmed == "" {
		return nil, apperror.BadRequest("Organization slug must be provided")
	}

	orgRow, err := s.repo.GetOrgBySlug(ctx, trimmed, currentUserID)
	if err != nil {
		return nil, apperror.NotFound(fmt.Sprintf("Organization with slug \"%s\" not found", trimmed))
	}

	if !isSuperAdmin && orgRow.OwnerID != currentUserID && !orgRow.MemberRole.Valid {
		return nil, apperror.Forbidden("You are not a member of this organization")
	}

	rows, err := s.repo.ListByOrgID(ctx, orgRow.ID)
	if err != nil {
		return nil, apperror.Internal("Failed to list workflows: " + err.Error())
	}

	return s.mapOrgWorkflowRows(ctx, rows)
}

func (s *service) FindOne(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) (*WorkflowResponse, error) {
	if _, err := s.assertOrgMemberForWorkflow(ctx, id, currentUserID, isSuperAdmin, ""); err != nil {
		return nil, err
	}

	wfRow, err := s.repo.FindDetailsByID(ctx, id)
	if err != nil {
		return nil, apperror.NotFound("Workflow not found")
	}

	statuses, err := s.repo.FindStatusesByWorkflowID(ctx, id)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch workflow statuses: " + err.Error())
	}

	transitions, err := s.repo.FindTransitionsByWorkflowID(ctx, id)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch workflow transitions: " + err.Error())
	}

	activeCount, _ := s.repo.CountActiveStatuses(ctx, id)
	transitionCount, _ := s.repo.CountTransitions(ctx, id)

	statusResponses := make([]TaskStatusResponse, 0, len(statuses))
	for _, st := range statuses {
		statusResponses = append(statusResponses, TaskStatusResponse{
			ID:          st.ID.String(),
			Name:        st.Name,
			Color:       st.Color,
			Category:    string(st.Category),
			Position:    st.Position,
			IsDefault:   st.IsDefault,
			WorkflowID:  st.WorkflowID.String(),
			CreatedAt:   formatTimestamp(st.CreatedAt),
			UpdatedAt:   formatTimestamp(st.UpdatedAt),
			CreatedByID: pgUUIDToPtr(st.CreatedByID),
			UpdatedByID: pgUUIDToPtr(st.UpdatedByID),
			DeletedAt:   pgTimestampToPtr(st.DeletedAt),
			DeletedBy:   pgUUIDToPtr(st.DeletedBy),
		})
	}

	transitionResponses := make([]StatusTransitionResponse, 0, len(transitions))
	for _, tr := range transitions {
		transitionResponses = append(transitionResponses, StatusTransitionResponse{
			ID:           tr.ID.String(),
			Name:         pgTextToString(tr.Name),
			Description:  pgTextToPtr(tr.Description),
			WorkflowID:   tr.WorkflowID.String(),
			FromStatusID: tr.FromStatusID.String(),
			ToStatusID:   tr.ToStatusID.String(),
			CreatedByID:  pgUUIDToPtr(tr.CreatedByID),
			UpdatedByID:  pgUUIDToPtr(tr.UpdatedByID),
			CreatedAt:    formatTimestamp(tr.CreatedAt),
			UpdatedAt:    formatTimestamp(tr.UpdatedAt),
			FromStatus: &TransitionStatusBrief{
				ID:       tr.FromStatusID_2.String(),
				Name:     tr.FromStatusName,
				Color:    tr.FromStatusColor,
				Category: string(tr.FromStatusCategory),
			},
			ToStatus: &TransitionStatusBrief{
				ID:       tr.ToStatusID_2.String(),
				Name:     tr.ToStatusName,
				Color:    tr.ToStatusColor,
				Category: string(tr.ToStatusCategory),
			},
		})
	}

	res := &WorkflowResponse{
		ID:             wfRow.ID.String(),
		Name:           wfRow.Name,
		Description:    pgTextToPtr(wfRow.Description),
		IsDefault:      wfRow.IsDefault,
		OrganizationID: wfRow.OrganizationID.String(),
		CreatedByID:    pgUUIDToPtr(wfRow.CreatedByID),
		UpdatedByID:    pgUUIDToPtr(wfRow.UpdatedByID),
		CreatedAt:      formatTimestamp(wfRow.CreatedAt),
		UpdatedAt:      formatTimestamp(wfRow.UpdatedAt),
		Organization: &OrgBrief{
			ID:   wfRow.OrgID.String(),
			Name: wfRow.OrgName,
			Slug: wfRow.OrgSlug,
		},
		Statuses:    statusResponses,
		Transitions: transitionResponses,
		Count: &WorkflowCounts{
			Statuses:    activeCount,
			Transitions: transitionCount,
		},
	}

	if wfRow.CreatedByUserID.Valid {
		res.CreatedByUser = &UserBrief{
			ID:        uuid.UUID(wfRow.CreatedByUserID.Bytes).String(),
			Email:     wfRow.CreatedByEmail.String,
			FirstName: pgTextToPtr(wfRow.CreatedByFirstName),
			LastName:  pgTextToPtr(wfRow.CreatedByLastName),
		}
	}
	if wfRow.UpdatedByUserID.Valid {
		res.UpdatedByUser = &UserBrief{
			ID:        uuid.UUID(wfRow.UpdatedByUserID.Bytes).String(),
			Email:     wfRow.UpdatedByEmail.String,
			FirstName: pgTextToPtr(wfRow.UpdatedByFirstName),
			LastName:  pgTextToPtr(wfRow.UpdatedByLastName),
		}
	}

	return res, nil
}

func (s *service) GetDefaultWorkflow(ctx context.Context, orgID uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) (*WorkflowResponse, error) {
	if err := s.assertOrgMember(ctx, orgID, currentUserID, isSuperAdmin, ""); err != nil {
		return nil, err
	}

	wfRow, err := s.repo.FindDefaultByOrgID(ctx, orgID)
	if err != nil {
		return nil, apperror.NotFound("No default workflow found")
	}

	statuses, err := s.repo.FindStatusesByWorkflowID(ctx, wfRow.ID)
	if err != nil {
		return nil, apperror.Internal("Failed to fetch workflow statuses: " + err.Error())
	}

	statusResponses := make([]TaskStatusResponse, 0, len(statuses))
	for _, st := range statuses {
		statusResponses = append(statusResponses, TaskStatusResponse{
			ID:          st.ID.String(),
			Name:        st.Name,
			Color:       st.Color,
			Category:    string(st.Category),
			Position:    st.Position,
			IsDefault:   st.IsDefault,
			WorkflowID:  st.WorkflowID.String(),
			CreatedAt:   formatTimestamp(st.CreatedAt),
			UpdatedAt:   formatTimestamp(st.UpdatedAt),
			CreatedByID: pgUUIDToPtr(st.CreatedByID),
			UpdatedByID: pgUUIDToPtr(st.UpdatedByID),
			DeletedAt:   pgTimestampToPtr(st.DeletedAt),
			DeletedBy:   pgUUIDToPtr(st.DeletedBy),
		})
	}

	return &WorkflowResponse{
		ID:             wfRow.ID.String(),
		Name:           wfRow.Name,
		Description:    pgTextToPtr(wfRow.Description),
		IsDefault:      wfRow.IsDefault,
		OrganizationID: wfRow.OrganizationID.String(),
		CreatedByID:    pgUUIDToPtr(wfRow.CreatedByID),
		UpdatedByID:    pgUUIDToPtr(wfRow.UpdatedByID),
		CreatedAt:      formatTimestamp(wfRow.CreatedAt),
		UpdatedAt:      formatTimestamp(wfRow.UpdatedAt),
		Organization: &OrgBrief{
			ID:   wfRow.OrgID.String(),
			Name: wfRow.OrgName,
			Slug: wfRow.OrgSlug,
		},
		Statuses: statusResponses,
	}, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req UpdateWorkflowRequest, currentUserID uuid.UUID, isSuperAdmin bool) (*WorkflowResponse, error) {
	wfOrg, err := s.assertOrgMemberForWorkflow(ctx, id, currentUserID, isSuperAdmin, middleware.RoleManager)
	if err != nil {
		return nil, err
	}

	if _, err := s.repo.Update(ctx, id, wfOrg.OrganizationID, req.Name, req.Description, req.IsDefault, currentUserID); err != nil {
		return nil, apperror.Internal("Failed to update workflow: " + err.Error())
	}

	return s.FindOne(ctx, id, currentUserID, isSuperAdmin)
}

func (s *service) MakeDefault(ctx context.Context, id uuid.UUID, orgID uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) (*WorkflowResponse, error) {
	if err := s.assertOrgMember(ctx, orgID, currentUserID, isSuperAdmin, middleware.RoleManager); err != nil {
		return nil, err
	}

	wfOrg, err := s.assertOrgMemberForWorkflow(ctx, id, currentUserID, isSuperAdmin, middleware.RoleManager)
	if err != nil {
		return nil, err
	}

	if wfOrg.OrganizationID != orgID {
		return nil, apperror.BadRequest("Workflow does not belong to this organization")
	}

	if _, err := s.repo.SetDefault(ctx, id, orgID, currentUserID); err != nil {
		return nil, apperror.Internal("Failed to set workflow as default: " + err.Error())
	}

	return s.FindOne(ctx, id, currentUserID, isSuperAdmin)
}

func (s *service) Delete(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) error {
	if _, err := s.assertOrgMemberForWorkflow(ctx, id, currentUserID, isSuperAdmin, middleware.RoleManager); err != nil {
		return err
	}

	projectCount, err := s.repo.CountProjectsUsingWorkflow(ctx, id)
	if err != nil {
		return apperror.Internal("Failed to check project dependencies: " + err.Error())
	}

	taskCount, err := s.repo.CountTasksUsingWorkflow(ctx, id)
	if err != nil {
		return apperror.Internal("Failed to check task dependencies: " + err.Error())
	}

	if projectCount > 0 || taskCount > 0 {
		var blockers []string
		if taskCount > 0 {
			plural := "tasks"
			if taskCount == 1 {
				plural = "task"
			}
			blockers = append(blockers, fmt.Sprintf("%d %s", taskCount, plural))
		}
		if projectCount > 0 {
			plural := "projects"
			if projectCount == 1 {
				plural = "project"
			}
			blockers = append(blockers, fmt.Sprintf("%d %s", projectCount, plural))
		}
		return apperror.Conflict(fmt.Sprintf(
			"This workflow is still in use by %s. Move them to another workflow before deleting it, because deleting this workflow would delete those tasks permanently.",
			strings.Join(blockers, " and "),
		))
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return apperror.Internal("Failed to delete workflow: " + err.Error())
	}

	return nil
}

func (s *service) mapOrgWorkflowRows(ctx context.Context, rows []sqlc.ListWorkflowsByOrganizationIDRow) ([]WorkflowResponse, error) {
	res := make([]WorkflowResponse, 0, len(rows))
	for _, r := range rows {
		statuses, _ := s.repo.FindStatusesByWorkflowID(ctx, r.ID)
		activeCount, _ := s.repo.CountActiveStatuses(ctx, r.ID)
		transitionCount, _ := s.repo.CountTransitions(ctx, r.ID)

		statusResponses := make([]TaskStatusResponse, 0, len(statuses))
		for _, st := range statuses {
			statusResponses = append(statusResponses, TaskStatusResponse{
				ID:          st.ID.String(),
				Name:        st.Name,
				Color:       st.Color,
				Category:    string(st.Category),
				Position:    st.Position,
				IsDefault:   st.IsDefault,
				WorkflowID:  st.WorkflowID.String(),
				CreatedAt:   formatTimestamp(st.CreatedAt),
				UpdatedAt:   formatTimestamp(st.UpdatedAt),
				CreatedByID: pgUUIDToPtr(st.CreatedByID),
				UpdatedByID: pgUUIDToPtr(st.UpdatedByID),
				DeletedAt:   pgTimestampToPtr(st.DeletedAt),
				DeletedBy:   pgUUIDToPtr(st.DeletedBy),
			})
		}

		wfRes := WorkflowResponse{
			ID:             r.ID.String(),
			Name:           r.Name,
			Description:    pgTextToPtr(r.Description),
			IsDefault:      r.IsDefault,
			OrganizationID: r.OrganizationID.String(),
			CreatedByID:    pgUUIDToPtr(r.CreatedByID),
			UpdatedByID:    pgUUIDToPtr(r.UpdatedByID),
			CreatedAt:      formatTimestamp(r.CreatedAt),
			UpdatedAt:      formatTimestamp(r.UpdatedAt),
			Organization: &OrgBrief{
				ID:   r.OrgID.String(),
				Name: r.OrgName,
				Slug: r.OrgSlug,
			},
			Statuses: statusResponses,
			Count: &WorkflowCounts{
				Statuses:    activeCount,
				Transitions: transitionCount,
			},
		}

		if r.CreatedByUserID.Valid {
			wfRes.CreatedByUser = &UserBrief{
				ID:        uuid.UUID(r.CreatedByUserID.Bytes).String(),
				Email:     r.CreatedByEmail.String,
				FirstName: pgTextToPtr(r.CreatedByFirstName),
				LastName:  pgTextToPtr(r.CreatedByLastName),
			}
		}
		if r.UpdatedByUserID.Valid {
			wfRes.UpdatedByUser = &UserBrief{
				ID:        uuid.UUID(r.UpdatedByUserID.Bytes).String(),
				Email:     r.UpdatedByEmail.String,
				FirstName: pgTextToPtr(r.UpdatedByFirstName),
				LastName:  pgTextToPtr(r.UpdatedByLastName),
			}
		}

		res = append(res, wfRes)
	}
	return res, nil
}

func (s *service) mapOrgIDsWorkflowRows(ctx context.Context, rows []sqlc.ListWorkflowsByOrganizationIDsRow) ([]WorkflowResponse, error) {
	res := make([]WorkflowResponse, 0, len(rows))
	for _, r := range rows {
		statuses, _ := s.repo.FindStatusesByWorkflowID(ctx, r.ID)
		activeCount, _ := s.repo.CountActiveStatuses(ctx, r.ID)
		transitionCount, _ := s.repo.CountTransitions(ctx, r.ID)

		statusResponses := make([]TaskStatusResponse, 0, len(statuses))
		for _, st := range statuses {
			statusResponses = append(statusResponses, TaskStatusResponse{
				ID:          st.ID.String(),
				Name:        st.Name,
				Color:       st.Color,
				Category:    string(st.Category),
				Position:    st.Position,
				IsDefault:   st.IsDefault,
				WorkflowID:  st.WorkflowID.String(),
				CreatedAt:   formatTimestamp(st.CreatedAt),
				UpdatedAt:   formatTimestamp(st.UpdatedAt),
				CreatedByID: pgUUIDToPtr(st.CreatedByID),
				UpdatedByID: pgUUIDToPtr(st.UpdatedByID),
				DeletedAt:   pgTimestampToPtr(st.DeletedAt),
				DeletedBy:   pgUUIDToPtr(st.DeletedBy),
			})
		}

		wfRes := WorkflowResponse{
			ID:             r.ID.String(),
			Name:           r.Name,
			Description:    pgTextToPtr(r.Description),
			IsDefault:      r.IsDefault,
			OrganizationID: r.OrganizationID.String(),
			CreatedByID:    pgUUIDToPtr(r.CreatedByID),
			UpdatedByID:    pgUUIDToPtr(r.UpdatedByID),
			CreatedAt:      formatTimestamp(r.CreatedAt),
			UpdatedAt:      formatTimestamp(r.UpdatedAt),
			Organization: &OrgBrief{
				ID:   r.OrgID.String(),
				Name: r.OrgName,
				Slug: r.OrgSlug,
			},
			Statuses: statusResponses,
			Count: &WorkflowCounts{
				Statuses:    activeCount,
				Transitions: transitionCount,
			},
		}

		if r.CreatedByUserID.Valid {
			wfRes.CreatedByUser = &UserBrief{
				ID:        uuid.UUID(r.CreatedByUserID.Bytes).String(),
				Email:     r.CreatedByEmail.String,
				FirstName: pgTextToPtr(r.CreatedByFirstName),
				LastName:  pgTextToPtr(r.CreatedByLastName),
			}
		}
		if r.UpdatedByUserID.Valid {
			wfRes.UpdatedByUser = &UserBrief{
				ID:        uuid.UUID(r.UpdatedByUserID.Bytes).String(),
				Email:     r.UpdatedByEmail.String,
				FirstName: pgTextToPtr(r.UpdatedByFirstName),
				LastName:  pgTextToPtr(r.UpdatedByLastName),
			}
		}

		res = append(res, wfRes)
	}
	return res, nil
}

func (s *service) mapAllWorkflowRows(ctx context.Context, rows []sqlc.ListAllWorkflowsRow) ([]WorkflowResponse, error) {
	res := make([]WorkflowResponse, 0, len(rows))
	for _, r := range rows {
		statuses, _ := s.repo.FindStatusesByWorkflowID(ctx, r.ID)
		activeCount, _ := s.repo.CountActiveStatuses(ctx, r.ID)
		transitionCount, _ := s.repo.CountTransitions(ctx, r.ID)

		statusResponses := make([]TaskStatusResponse, 0, len(statuses))
		for _, st := range statuses {
			statusResponses = append(statusResponses, TaskStatusResponse{
				ID:          st.ID.String(),
				Name:        st.Name,
				Color:       st.Color,
				Category:    string(st.Category),
				Position:    st.Position,
				IsDefault:   st.IsDefault,
				WorkflowID:  st.WorkflowID.String(),
				CreatedAt:   formatTimestamp(st.CreatedAt),
				UpdatedAt:   formatTimestamp(st.UpdatedAt),
				CreatedByID: pgUUIDToPtr(st.CreatedByID),
				UpdatedByID: pgUUIDToPtr(st.UpdatedByID),
				DeletedAt:   pgTimestampToPtr(st.DeletedAt),
				DeletedBy:   pgUUIDToPtr(st.DeletedBy),
			})
		}

		wfRes := WorkflowResponse{
			ID:             r.ID.String(),
			Name:           r.Name,
			Description:    pgTextToPtr(r.Description),
			IsDefault:      r.IsDefault,
			OrganizationID: r.OrganizationID.String(),
			CreatedByID:    pgUUIDToPtr(r.CreatedByID),
			UpdatedByID:    pgUUIDToPtr(r.UpdatedByID),
			CreatedAt:      formatTimestamp(r.CreatedAt),
			UpdatedAt:      formatTimestamp(r.UpdatedAt),
			Organization: &OrgBrief{
				ID:   r.OrgID.String(),
				Name: r.OrgName,
				Slug: r.OrgSlug,
			},
			Statuses: statusResponses,
			Count: &WorkflowCounts{
				Statuses:    activeCount,
				Transitions: transitionCount,
			},
		}

		if r.CreatedByUserID.Valid {
			wfRes.CreatedByUser = &UserBrief{
				ID:        uuid.UUID(r.CreatedByUserID.Bytes).String(),
				Email:     r.CreatedByEmail.String,
				FirstName: pgTextToPtr(r.CreatedByFirstName),
				LastName:  pgTextToPtr(r.CreatedByLastName),
			}
		}
		if r.UpdatedByUserID.Valid {
			wfRes.UpdatedByUser = &UserBrief{
				ID:        uuid.UUID(r.UpdatedByUserID.Bytes).String(),
				Email:     r.UpdatedByEmail.String,
				FirstName: pgTextToPtr(r.UpdatedByFirstName),
				LastName:  pgTextToPtr(r.UpdatedByLastName),
			}
		}

		res = append(res, wfRes)
	}
	return res, nil
}
