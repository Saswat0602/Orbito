package invitations

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/modules/email"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// EmailSender defines an optional interface for queuing invitation emails.
type EmailSender interface {
	SendEmail(ctx context.Context, dto email.SendEmailDTO) error
}

// Service defines the business logic contract for invitations.
type Service interface {
	CreateInvitation(ctx context.Context, req CreateInvitationRequest, inviterID uuid.UUID, isSuperAdmin bool) (any, error)
	AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) (*AcceptInvitationResponse, error)
	DeclineInvitation(ctx context.Context, token string) (*MessageResponse, error)
	VerifyInvitation(ctx context.Context, token string) (*VerifyInvitationResponse, error)
	GetUserInvitations(ctx context.Context, userEmail string) ([]InvitationResponse, error)
	GetEntityInvitations(ctx context.Context, entityType string, entityID uuid.UUID) ([]InvitationResponse, error)
	ResendInvitation(ctx context.Context, invitationID, currentUserID uuid.UUID) (*ResendInvitationResponse, error)
	DeleteInvitation(ctx context.Context, invitationID uuid.UUID) (*MessageResponse, error)
}

type service struct {
	repo         Repository
	rbacResolver middleware.RBACResolver
	emailSender  EmailSender
	frontendURL  string
}

// NewService instantiates an invitations service.
func NewService(repo Repository, rbacResolver middleware.RBACResolver, emailSender EmailSender, frontendURL string) Service {
	if frontendURL == "" {
		frontendURL = "http://localhost:3001"
	}
	return &service{
		repo:         repo,
		rbacResolver: rbacResolver,
		emailSender:  emailSender,
		frontendURL:  frontendURL,
	}
}

func parseInvitableRole(role string) (middleware.Role, error) {
	r := middleware.Role(strings.ToUpper(strings.TrimSpace(role)))
	switch r {
	case middleware.RoleOwner, middleware.RoleManager, middleware.RoleMember, middleware.RoleViewer:
		return r, nil
	default:
		return "", apperror.BadRequest(fmt.Sprintf("Invalid invitation role: %s", role))
	}
}

func (s *service) resolveInviterAuthority(ctx context.Context, req CreateInvitationRequest, inviterID, owningOrgID uuid.UUID) (middleware.Role, error) {
	var candidates []middleware.Role

	if req.ProjectID != nil {
		pRole, err := s.rbacResolver.GetProjectMemberRole(ctx, inviterID, *req.ProjectID)
		if err == nil && (pRole == middleware.RoleOwner || pRole == middleware.RoleManager) {
			candidates = append(candidates, pRole)
		}
	}
	if req.WorkspaceID != nil {
		wRole, err := s.rbacResolver.GetWorkspaceMemberRole(ctx, inviterID, *req.WorkspaceID)
		if err == nil && (wRole == middleware.RoleOwner || wRole == middleware.RoleManager) {
			candidates = append(candidates, wRole)
		}
	}
	oRole, err := s.rbacResolver.GetOrgMemberRole(ctx, inviterID, owningOrgID)
	if err == nil && (oRole == middleware.RoleOwner || oRole == middleware.RoleManager) {
		candidates = append(candidates, oRole)
	}

	if len(candidates) == 0 {
		return "", apperror.Forbidden("You do not have permission to invite members to this scope")
	}

	best := candidates[0]
	for _, c := range candidates[1:] {
		if middleware.RoleRank[c] > middleware.RoleRank[best] {
			best = c
		}
	}
	return best, nil
}

func (s *service) assertCanInvite(ctx context.Context, req CreateInvitationRequest, inviterID, owningOrgID uuid.UUID, isSuperAdmin bool) (middleware.Role, error) {
	requestedRole, err := parseInvitableRole(req.Role)
	if err != nil {
		return "", err
	}

	if isSuperAdmin {
		return requestedRole, nil
	}

	bestRole, err := s.resolveInviterAuthority(ctx, req, inviterID, owningOrgID)
	if err != nil {
		return "", err
	}

	if middleware.RoleRank[requestedRole] > middleware.RoleRank[bestRole] {
		return "", apperror.Forbidden("You cannot grant a role higher than your own")
	}

	return requestedRole, nil
}

func (s *service) checkExistingMembership(ctx context.Context, req CreateInvitationRequest, userUUID uuid.UUID) error {
	if req.OrganizationID != nil {
		exists, err := s.repo.CheckOrgMemberExists(ctx, userUUID, *req.OrganizationID)
		if err != nil {
			return err
		}
		if exists {
			return apperror.BadRequest("User is already a member of this organization")
		}
	}

	if req.WorkspaceID != nil {
		exists, err := s.repo.CheckWorkspaceMemberExists(ctx, userUUID, *req.WorkspaceID)
		if err != nil {
			return err
		}
		if exists {
			return apperror.BadRequest("User is already a member of this workspace")
		}
	}

	if req.ProjectID != nil {
		exists, err := s.repo.CheckProjectMemberExists(ctx, userUUID, *req.ProjectID)
		if err != nil {
			return err
		}
		if exists {
			return apperror.BadRequest("User is already a member of this project")
		}
	}

	return nil
}

func (s *service) CreateInvitation(ctx context.Context, req CreateInvitationRequest, inviterID uuid.UUID, isSuperAdmin bool) (any, error) {
	emailAddr := strings.ToLower(strings.TrimSpace(req.InviteeEmail))
	if emailAddr == "" {
		emailAddr = strings.ToLower(strings.TrimSpace(req.Email))
	}
	if emailAddr == "" {
		return nil, apperror.BadRequest("Either inviteeEmail or email must be provided")
	}
	req.InviteeEmail = emailAddr

	targetCount := 0
	if req.OrganizationID != nil {
		targetCount++
	}
	if req.WorkspaceID != nil {
		targetCount++
	}
	if req.ProjectID != nil {
		targetCount++
	}
	if targetCount != 1 {
		return nil, apperror.BadRequest("Must specify exactly one of: organizationId, workspaceId, or projectId")
	}

	if len(req.ProjectIDs) > 0 && req.WorkspaceID == nil {
		return nil, apperror.BadRequest("projectIds can only be used with workspaceId")
	}

	var owningOrgID uuid.UUID
	var wsTarget *sqlc.GetWorkspaceByIDRow
	var projTarget *sqlc.GetProjectWithWorkspaceAndOrgRow

	if req.OrganizationID != nil {
		owningOrgID = *req.OrganizationID
		org, err := s.repo.GetOrganizationByID(ctx, owningOrgID)
		if err != nil {
			return nil, err
		}
		if org == nil {
			return nil, apperror.NotFound("Organization not found")
		}
	} else if req.WorkspaceID != nil {
		ws, err := s.repo.GetWorkspaceByID(ctx, *req.WorkspaceID)
		if err != nil {
			return nil, err
		}
		if ws == nil {
			return nil, apperror.NotFound("Workspace not found")
		}
		owningOrgID = ws.OrganizationID
		wsTarget = ws

		if len(req.ProjectIDs) > 0 {
			for _, pid := range req.ProjectIDs {
				belongs, bErr := s.repo.CheckProjectBelongsToWorkspace(ctx, pid, *req.WorkspaceID)
				if bErr != nil {
					return nil, bErr
				}
				if !belongs {
					return nil, apperror.BadRequest(fmt.Sprintf("Some project IDs do not belong to this workspace: %s", pid))
				}
			}
		}
	} else {
		proj, err := s.repo.GetProjectWithWorkspaceAndOrg(ctx, *req.ProjectID)
		if err != nil {
			return nil, err
		}
		if proj == nil {
			return nil, apperror.NotFound("Project not found")
		}
		owningOrgID = proj.OrganizationID
		projTarget = proj
	}

	requestedRole, err := s.assertCanInvite(ctx, req, inviterID, owningOrgID, isSuperAdmin)
	if err != nil {
		return nil, err
	}

	existingUser, err := s.repo.GetUserByEmail(ctx, emailAddr)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		if mErr := s.checkExistingMembership(ctx, req, existingUser.ID); mErr != nil {
			return nil, mErr
		}

		if req.WorkspaceID != nil || req.ProjectID != nil {
			isOrgMember, oErr := s.repo.CheckOrgMemberExists(ctx, existingUser.ID, owningOrgID)
			if oErr != nil {
				return nil, oErr
			}
			if isOrgMember {
				return s.addExistingOrgMemberDirectly(ctx, req, existingUser, inviterID, wsTarget, projTarget, requestedRole)
			}
		}
	}

	pending, err := s.repo.FindPending(ctx, sqlc.FindPendingInvitationParams{
		InviteeEmail:   emailAddr,
		OrganizationID: toPgUUID(req.OrganizationID),
		WorkspaceID:    toPgUUID(req.WorkspaceID),
		ProjectID:      toPgUUID(req.ProjectID),
	})
	if err != nil {
		return nil, err
	}
	if pending != nil {
		return nil, apperror.BadRequest("Invitation already exists for this email")
	}

	tokenBytes := make([]byte, 32)
	if _, rErr := rand.Read(tokenBytes); rErr != nil {
		return nil, apperror.Internal("Failed to generate secure invitation token", rErr)
	}
	token := hex.EncodeToString(tokenBytes)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	invitation, err := s.repo.Create(ctx, sqlc.CreateInvitationParams{
		ID:             uuid.New(),
		InviterID:      inviterID,
		InviteeEmail:   emailAddr,
		OrganizationID: toPgUUID(req.OrganizationID),
		WorkspaceID:    toPgUUID(req.WorkspaceID),
		ProjectID:      toPgUUID(req.ProjectID),
		Role:           string(requestedRole),
		Status:         sqlc.InvitationStatusPENDING,
		Token:          token,
		ExpiresAt:      toPgTimestamp(expiresAt),
	})
	if err != nil {
		return nil, err
	}

	if req.WorkspaceID != nil && len(req.ProjectIDs) > 0 {
		for _, pid := range req.ProjectIDs {
			pTokenBytes := make([]byte, 32)
			_, _ = rand.Read(pTokenBytes)
			pToken := hex.EncodeToString(pTokenBytes)
			_, _ = s.repo.Create(ctx, sqlc.CreateInvitationParams{
				ID:           uuid.New(),
				InviterID:    inviterID,
				InviteeEmail: emailAddr,
				ProjectID:    pgtype.UUID{Bytes: pid, Valid: true},
				Role:         string(requestedRole),
				Status:       sqlc.InvitationStatusPENDING,
				Token:        pToken,
				ExpiresAt:    toPgTimestamp(expiresAt),
			})
		}
	}

	fullInv, err := s.repo.GetByID(ctx, invitation.ID)
	if err != nil || fullInv == nil {
		return mapInvitationToResponse(invitation, nil, true), nil
	}

	emailSent := s.trySendInvitationEmail(ctx, fullInv)

	res := mapInvitationRowToResponse(fullInv)
	res.EmailSent = emailSent
	return res, nil
}

func (s *service) addExistingOrgMemberDirectly(
	ctx context.Context,
	req CreateInvitationRequest,
	targetUser *sqlc.GetUserByEmailRow,
	inviterID uuid.UUID,
	wsTarget *sqlc.GetWorkspaceByIDRow,
	projTarget *sqlc.GetProjectWithWorkspaceAndOrgRow,
	role middleware.Role,
) (*DirectAddResponse, error) {
	if req.WorkspaceID != nil && wsTarget != nil {
		member, err := s.repo.CreateWorkspaceMember(ctx, sqlc.CreateWorkspaceMemberDirectParams{
			ID:          uuid.New(),
			UserID:      targetUser.ID,
			WorkspaceID: wsTarget.ID,
			Role:        sqlc.Role(role),
		})
		if err != nil {
			return nil, err
		}

		for _, pid := range req.ProjectIDs {
			_, _ = s.repo.CreateProjectMember(ctx, sqlc.CreateProjectMemberDirectParams{
				ID:          uuid.New(),
				UserID:      targetUser.ID,
				ProjectID:   pid,
				Role:        sqlc.Role(role),
				CreatedByID: pgtype.UUID{Bytes: inviterID, Valid: true},
			})
		}

		return &DirectAddResponse{
			Type:    "direct_add",
			Message: fmt.Sprintf("User %s was added directly to the workspace as they are already an organization member", targetUser.Email),
			Member:  member,
			Entity: map[string]any{
				"type": "workspace",
				"id":   wsTarget.ID,
				"name": wsTarget.Name,
				"slug": wsTarget.Slug,
			},
			User: map[string]any{
				"id":        targetUser.ID,
				"email":     targetUser.Email,
				"firstName": targetUser.FirstName,
				"lastName":  targetUser.LastName,
			},
			AddedToProjects: req.ProjectIDs,
		}, nil
	}

	if req.ProjectID != nil && projTarget != nil {
		wsMemberExists, wErr := s.repo.CheckWorkspaceMemberExists(ctx, targetUser.ID, projTarget.WorkspaceID)
		if wErr != nil {
			return nil, wErr
		}
		if !wsMemberExists {
			_, _ = s.repo.CreateWorkspaceMember(ctx, sqlc.CreateWorkspaceMemberDirectParams{
				ID:          uuid.New(),
				UserID:      targetUser.ID,
				WorkspaceID: projTarget.WorkspaceID,
				Role:        sqlc.RoleMEMBER,
			})
		}

		member, err := s.repo.CreateProjectMember(ctx, sqlc.CreateProjectMemberDirectParams{
			ID:          uuid.New(),
			UserID:      targetUser.ID,
			ProjectID:   projTarget.ID,
			Role:        sqlc.Role(role),
			CreatedByID: pgtype.UUID{Bytes: inviterID, Valid: true},
		})
		if err != nil {
			return nil, err
		}

		return &DirectAddResponse{
			Type:    "direct_add",
			Message: fmt.Sprintf("User %s was added directly to the project as they are already an organization member", targetUser.Email),
			Member:  member,
			Entity: map[string]any{
				"type": "project",
				"id":   projTarget.ID,
				"name": projTarget.Name,
				"slug": projTarget.Slug,
			},
			User: map[string]any{
				"id":        targetUser.ID,
				"email":     targetUser.Email,
				"firstName": targetUser.FirstName,
				"lastName":  targetUser.LastName,
			},
			AddedToWorkspace: !wsMemberExists,
		}, nil
	}

	return nil, apperror.BadRequest("Invalid direct add parameters")
}

func (s *service) trySendInvitationEmail(ctx context.Context, inv *sqlc.GetInvitationByIDRow) bool {
	if s.emailSender == nil {
		return false
	}

	entityName := "an Orbito workspace"
	if inv.OrgName.Valid {
		entityName = inv.OrgName.String
	} else if inv.WsName.Valid {
		entityName = inv.WsName.String
	} else if inv.ProjName.Valid {
		entityName = inv.ProjName.String
	}

	inviteLink := fmt.Sprintf("%s/invite?token=%s", s.frontendURL, inv.Token)
	err := s.emailSender.SendEmail(ctx, email.SendEmailDTO{
		To:       inv.InviteeEmail,
		Subject:  fmt.Sprintf("You've been invited to %s", entityName),
		Template: "send-invitation",
		Data: map[string]any{
			"workspace": map[string]any{
				"name": entityName,
			},
			"inviteLink": inviteLink,
		},
		Priority: "HIGH",
	})
	if err != nil {
		log.Printf("[Invitations] Failed to send email to %s: %v", inv.InviteeEmail, err)
		return false
	}
	return true
}

func (s *service) AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) (*AcceptInvitationResponse, error) {
	inv, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, apperror.NotFound("Invitation not found")
	}

	if inv.Status != sqlc.InvitationStatusPENDING {
		return nil, apperror.BadRequest("Invitation has already been processed")
	}

	if inv.ExpiresAt.Time.Before(time.Now()) {
		_ = s.repo.UpdateStatus(ctx, inv.ID, sqlc.InvitationStatusEXPIRED)
		return nil, apperror.BadRequest("Invitation has expired")
	}

	u, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, apperror.NotFound("User not found")
	}

	if !strings.EqualFold(u.Email, inv.InviteeEmail) {
		return nil, apperror.BadRequest("Invitation email does not match user email")
	}

	txErr := s.repo.WithTx(ctx, func(txRepo Repository) error {
		if inv.OrganizationID.Valid {
			orgID := uuid.UUID(inv.OrganizationID.Bytes)
			exists, _ := txRepo.CheckOrgMemberExists(ctx, userID, orgID)
			if !exists {
				_, cErr := txRepo.CreateOrganizationMember(ctx, sqlc.CreateOrganizationMemberParams{
					ID:             uuid.New(),
					UserID:         userID,
					OrganizationID: orgID,
					Role:           sqlc.Role(inv.Role),
					IsDefault:      false,
				})
				if cErr != nil {
					return cErr
				}
			}
		} else if inv.WorkspaceID.Valid {
			wsID := uuid.UUID(inv.WorkspaceID.Bytes)
			ws, _ := txRepo.GetWorkspaceByID(ctx, wsID)
			if ws != nil {
				orgExists, _ := txRepo.CheckOrgMemberExists(ctx, userID, ws.OrganizationID)
				if !orgExists {
					_, _ = txRepo.CreateOrganizationMember(ctx, sqlc.CreateOrganizationMemberParams{
						ID:             uuid.New(),
						UserID:         userID,
						OrganizationID: ws.OrganizationID,
						Role:           sqlc.RoleMEMBER,
						IsDefault:      false,
					})
				}
			}

			wsExists, _ := txRepo.CheckWorkspaceMemberExists(ctx, userID, wsID)
			if !wsExists {
				_, wErr := txRepo.CreateWorkspaceMember(ctx, sqlc.CreateWorkspaceMemberDirectParams{
					ID:          uuid.New(),
					UserID:      userID,
					WorkspaceID: wsID,
					Role:        sqlc.Role(inv.Role),
				})
				if wErr != nil {
					return wErr
				}
			}

			related, _ := txRepo.ListRelatedProjectInvitations(ctx, inv.InviteeEmail, wsID)
			for _, r := range related {
				if r.ProjectID.Valid {
					projID := uuid.UUID(r.ProjectID.Bytes)
					pmExists, _ := txRepo.CheckProjectMemberExists(ctx, userID, projID)
					if !pmExists {
						_, _ = txRepo.CreateProjectMember(ctx, sqlc.CreateProjectMemberDirectParams{
							ID:          uuid.New(),
							UserID:      userID,
							ProjectID:   projID,
							Role:        sqlc.Role(r.Role),
							CreatedByID: pgtype.UUID{Bytes: r.InviterID, Valid: true},
						})
					}
					_ = txRepo.UpdateStatus(ctx, r.ID, sqlc.InvitationStatusACCEPTED)
				}
			}
		} else if inv.ProjectID.Valid {
			projID := uuid.UUID(inv.ProjectID.Bytes)
			proj, _ := txRepo.GetProjectWithWorkspaceAndOrg(ctx, projID)
			if proj != nil {
				orgExists, _ := txRepo.CheckOrgMemberExists(ctx, userID, proj.OrganizationID)
				if !orgExists {
					_, _ = txRepo.CreateOrganizationMember(ctx, sqlc.CreateOrganizationMemberParams{
						ID:             uuid.New(),
						UserID:         userID,
						OrganizationID: proj.OrganizationID,
						Role:           sqlc.RoleMEMBER,
						IsDefault:      false,
					})
				}
				wsExists, _ := txRepo.CheckWorkspaceMemberExists(ctx, userID, proj.WorkspaceID)
				if !wsExists {
					_, _ = txRepo.CreateWorkspaceMember(ctx, sqlc.CreateWorkspaceMemberDirectParams{
						ID:          uuid.New(),
						UserID:      userID,
						WorkspaceID: proj.WorkspaceID,
						Role:        sqlc.RoleMEMBER,
					})
				}
			}

			pmExists, _ := txRepo.CheckProjectMemberExists(ctx, userID, projID)
			if !pmExists {
				_, pErr := txRepo.CreateProjectMember(ctx, sqlc.CreateProjectMemberDirectParams{
					ID:          uuid.New(),
					UserID:      userID,
					ProjectID:   projID,
					Role:        sqlc.Role(inv.Role),
					CreatedByID: pgtype.UUID{Bytes: inv.InviterID, Valid: true},
				})
				if pErr != nil {
					return pErr
				}
			}
		}

		return txRepo.UpdateStatus(ctx, inv.ID, sqlc.InvitationStatusACCEPTED)
	})

	if txErr != nil {
		return nil, txErr
	}

	entityType := "project"
	entityName := "Unknown"
	if inv.OrganizationID.Valid {
		entityType = "organization"
		if inv.OrgName.Valid {
			entityName = inv.OrgName.String
		}
	} else if inv.WorkspaceID.Valid {
		entityType = "workspace"
		if inv.WsName.Valid {
			entityName = inv.WsName.String
		}
	} else if inv.ProjectID.Valid {
		entityType = "project"
		if inv.ProjName.Valid {
			entityName = inv.ProjName.String
		}
	}

	return &AcceptInvitationResponse{
		Message: "Invitation accepted successfully",
		Invitation: map[string]string{
			"id":         inv.ID.String(),
			"entityType": entityType,
			"entityName": entityName,
		},
	}, nil
}

func (s *service) DeclineInvitation(ctx context.Context, token string) (*MessageResponse, error) {
	inv, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, apperror.NotFound("Invitation not found")
	}

	if inv.Status != sqlc.InvitationStatusPENDING {
		return nil, apperror.BadRequest("Invitation has already been processed")
	}

	if err := s.repo.UpdateStatus(ctx, inv.ID, sqlc.InvitationStatusDECLINED); err != nil {
		return nil, err
	}

	return &MessageResponse{Message: "Invitation declined successfully"}, nil
}

func (s *service) VerifyInvitation(ctx context.Context, token string) (*VerifyInvitationResponse, error) {
	inv, err := s.repo.GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, apperror.NotFound("Invitation not found")
	}

	now := time.Now()
	isExpired := inv.ExpiresAt.Time.Before(now)
	isValid := inv.Status == sqlc.InvitationStatusPENDING && !isExpired

	u, _ := s.repo.GetUserByEmail(ctx, strings.ToLower(inv.InviteeEmail))
	inviteeExists := u != nil

	entityName := "Unknown"
	entityType := "unknown"
	if inv.OrganizationID.Valid {
		entityType = "organization"
		if inv.OrgName.Valid {
			entityName = inv.OrgName.String
		}
	} else if inv.WorkspaceID.Valid {
		entityType = "workspace"
		if inv.WsName.Valid {
			entityName = inv.WsName.String
		}
	} else if inv.ProjectID.Valid {
		entityType = "project"
		if inv.ProjName.Valid {
			entityName = inv.ProjName.String
		}
	}

	msg := fmt.Sprintf("Valid invitation to join %s", entityName)
	if !isValid {
		if isExpired {
			msg = "Invitation has expired"
		} else {
			msg = "Invitation has already been processed"
		}
	}

	return &VerifyInvitationResponse{
		Success:       isValid,
		Message:       msg,
		InviteeExists: inviteeExists,
		Invitation: map[string]any{
			"id":           inv.ID,
			"inviteeEmail": inv.InviteeEmail,
			"role":         inv.Role,
			"status":       string(inv.Status),
			"expiresAt":    inv.ExpiresAt.Time,
			"entityType":   entityType,
			"entityName":   entityName,
		},
	}, nil
}

func (s *service) GetUserInvitations(ctx context.Context, userEmail string) ([]InvitationResponse, error) {
	rows, err := s.repo.GetUserInvitations(ctx, strings.ToLower(strings.TrimSpace(userEmail)))
	if err != nil {
		return nil, err
	}

	res := make([]InvitationResponse, 0, len(rows))
	for _, r := range rows {
		row := r
		res = append(res, *mapUserInvitationRowToResponse(&row))
	}
	return res, nil
}

func (s *service) GetEntityInvitations(ctx context.Context, entityType string, entityID uuid.UUID) ([]InvitationResponse, error) {
	normType := strings.ToLower(strings.TrimSpace(entityType))
	if normType != "organization" && normType != "workspace" && normType != "project" {
		return nil, apperror.BadRequest("Invalid entity type")
	}

	_ = s.repo.ExpirePendingEntityInvitations(ctx, normType, entityID)

	rows, err := s.repo.GetEntityInvitations(ctx, normType, entityID)
	if err != nil {
		return nil, err
	}

	res := make([]InvitationResponse, 0, len(rows))
	for _, r := range rows {
		row := r
		res = append(res, *mapEntityInvitationRowToResponse(&row))
	}
	return res, nil
}

func (s *service) ResendInvitation(ctx context.Context, invitationID, currentUserID uuid.UUID) (*ResendInvitationResponse, error) {
	inv, err := s.repo.GetByID(ctx, invitationID)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, apperror.NotFound("Invitation not found")
	}

	if inv.Status == sqlc.InvitationStatusDECLINED {
		return nil, apperror.BadRequest(fmt.Sprintf("Cannot resend %s invitation", strings.ToLower(string(inv.Status))))
	}

	newTokenBytes := make([]byte, 32)
	_, _ = rand.Read(newTokenBytes)
	newToken := hex.EncodeToString(newTokenBytes)
	newExpiresAt := time.Now().Add(7 * 24 * time.Hour)

	updated, err := s.repo.Resend(ctx, sqlc.ResendInvitationParams{
		ID:        invitationID,
		Token:     newToken,
		ExpiresAt: toPgTimestamp(newExpiresAt),
		InviterID: currentUserID,
	})
	if err != nil {
		return nil, err
	}

	fullUpdated, _ := s.repo.GetByID(ctx, updated.ID)
	emailSent := false
	if fullUpdated != nil {
		emailSent = s.trySendInvitationEmail(ctx, fullUpdated)
	}

	msg := "Invitation updated successfully, but email delivery failed"
	var emailErr *string
	if emailSent {
		msg = "Invitation resent successfully"
	} else {
		errStr := "email delivery failed"
		emailErr = &errStr
	}

	return &ResendInvitationResponse{
		Message:    msg,
		Invitation: mapInvitationRowToResponse(fullUpdated),
		EmailSent:  emailSent,
		EmailError: emailErr,
	}, nil
}

func (s *service) DeleteInvitation(ctx context.Context, invitationID uuid.UUID) (*MessageResponse, error) {
	inv, err := s.repo.GetByID(ctx, invitationID)
	if err != nil {
		return nil, err
	}
	if inv == nil {
		return nil, apperror.NotFound("Invitation not found")
	}

	if err := s.repo.Delete(ctx, invitationID); err != nil {
		return nil, err
	}

	return &MessageResponse{Message: "Invitation deleted successfully"}, nil
}

// Helpers

func toPgUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func toPgTimestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: t, Valid: true}
}

func mapInvitationToResponse(inv *sqlc.Invitation, inviter *InviterUser, emailSent bool) *InvitationResponse {
	var orgID, wsID, projID *uuid.UUID
	if inv.OrganizationID.Valid {
		id := uuid.UUID(inv.OrganizationID.Bytes)
		orgID = &id
	}
	if inv.WorkspaceID.Valid {
		id := uuid.UUID(inv.WorkspaceID.Bytes)
		wsID = &id
	}
	if inv.ProjectID.Valid {
		id := uuid.UUID(inv.ProjectID.Bytes)
		projID = &id
	}

	return &InvitationResponse{
		ID:             inv.ID,
		InviterID:      inv.InviterID,
		InviteeEmail:   inv.InviteeEmail,
		OrganizationID: orgID,
		WorkspaceID:    wsID,
		ProjectID:      projID,
		Role:           inv.Role,
		Status:         string(inv.Status),
		Token:          inv.Token,
		ExpiresAt:      inv.ExpiresAt.Time,
		CreatedAt:      inv.CreatedAt.Time,
		UpdatedAt:      inv.UpdatedAt.Time,
		Inviter:        inviter,
		EmailSent:      emailSent,
	}
}

func mapInvitationRowToResponse(row *sqlc.GetInvitationByIDRow) *InvitationResponse {
	if row == nil {
		return nil
	}
	var orgID, wsID, projID *uuid.UUID
	var org, ws, proj *EntityTarget

	if row.OrganizationID.Valid {
		id := uuid.UUID(row.OrganizationID.Bytes)
		orgID = &id
		if row.OrgName.Valid {
			org = &EntityTarget{
				ID:   id,
				Name: row.OrgName.String,
				Slug: row.OrgSlug.String,
			}
		}
	}
	if row.WorkspaceID.Valid {
		id := uuid.UUID(row.WorkspaceID.Bytes)
		wsID = &id
		if row.WsName.Valid {
			ws = &EntityTarget{
				ID:   id,
				Name: row.WsName.String,
				Slug: row.WsSlug.String,
			}
		}
	}
	if row.ProjectID.Valid {
		id := uuid.UUID(row.ProjectID.Bytes)
		projID = &id
		if row.ProjName.Valid {
			proj = &EntityTarget{
				ID:   id,
				Name: row.ProjName.String,
				Slug: row.ProjSlug.String,
			}
		}
	}

	var username *string
	if row.InviterUsername.Valid {
		username = &row.InviterUsername.String
	}

	return &InvitationResponse{
		ID:             row.ID,
		InviterID:      row.InviterID,
		InviteeEmail:   row.InviteeEmail,
		OrganizationID: orgID,
		WorkspaceID:    wsID,
		ProjectID:      projID,
		Role:           row.Role,
		Status:         string(row.Status),
		Token:          row.Token,
		ExpiresAt:      row.ExpiresAt.Time,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
		Inviter: &InviterUser{
			ID:        row.InviterUserID,
			Email:     row.InviterEmail,
			FirstName: row.InviterFirstName,
			LastName:  row.InviterLastName,
			Username:  username,
		},
		Organization: org,
		Workspace:    ws,
		Project:      proj,
	}
}

func mapUserInvitationRowToResponse(row *sqlc.GetUserInvitationsRow) *InvitationResponse {
	var orgID, wsID, projID *uuid.UUID
	var org, ws, proj *EntityTarget

	if row.OrganizationID.Valid {
		id := uuid.UUID(row.OrganizationID.Bytes)
		orgID = &id
		if row.OrgName.Valid {
			org = &EntityTarget{
				ID:   id,
				Name: row.OrgName.String,
				Slug: row.OrgSlug.String,
			}
		}
	}
	if row.WorkspaceID.Valid {
		id := uuid.UUID(row.WorkspaceID.Bytes)
		wsID = &id
		if row.WsName.Valid {
			ws = &EntityTarget{
				ID:   id,
				Name: row.WsName.String,
				Slug: row.WsSlug.String,
			}
		}
	}
	if row.ProjectID.Valid {
		id := uuid.UUID(row.ProjectID.Bytes)
		projID = &id
		if row.ProjName.Valid {
			proj = &EntityTarget{
				ID:   id,
				Name: row.ProjName.String,
				Slug: row.ProjSlug.String,
			}
		}
	}

	var username *string
	if row.InviterUsername.Valid {
		username = &row.InviterUsername.String
	}

	return &InvitationResponse{
		ID:             row.ID,
		InviterID:      row.InviterID,
		InviteeEmail:   row.InviteeEmail,
		OrganizationID: orgID,
		WorkspaceID:    wsID,
		ProjectID:      projID,
		Role:           row.Role,
		Status:         string(row.Status),
		Token:          row.Token,
		ExpiresAt:      row.ExpiresAt.Time,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
		Inviter: &InviterUser{
			ID:        row.InviterUserID,
			Email:     row.InviterEmail,
			FirstName: row.InviterFirstName,
			LastName:  row.InviterLastName,
			Username:  username,
		},
		Organization: org,
		Workspace:    ws,
		Project:      proj,
	}
}

func mapEntityInvitationRowToResponse(row *sqlc.GetEntityInvitationsRow) *InvitationResponse {
	var orgID, wsID, projID *uuid.UUID
	var org, ws, proj *EntityTarget

	if row.OrganizationID.Valid {
		id := uuid.UUID(row.OrganizationID.Bytes)
		orgID = &id
		if row.OrgName.Valid {
			org = &EntityTarget{
				ID:   id,
				Name: row.OrgName.String,
				Slug: row.OrgSlug.String,
			}
		}
	}
	if row.WorkspaceID.Valid {
		id := uuid.UUID(row.WorkspaceID.Bytes)
		wsID = &id
		if row.WsName.Valid {
			ws = &EntityTarget{
				ID:   id,
				Name: row.WsName.String,
				Slug: row.WsSlug.String,
			}
		}
	}
	if row.ProjectID.Valid {
		id := uuid.UUID(row.ProjectID.Bytes)
		projID = &id
		if row.ProjName.Valid {
			proj = &EntityTarget{
				ID:   id,
				Name: row.ProjName.String,
				Slug: row.ProjSlug.String,
			}
		}
	}

	var username *string
	if row.InviterUsername.Valid {
		username = &row.InviterUsername.String
	}

	return &InvitationResponse{
		ID:             row.ID,
		InviterID:      row.InviterID,
		InviteeEmail:   row.InviteeEmail,
		OrganizationID: orgID,
		WorkspaceID:    wsID,
		ProjectID:      projID,
		Role:           row.Role,
		Status:         string(row.Status),
		Token:          row.Token,
		ExpiresAt:      row.ExpiresAt.Time,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
		Inviter: &InviterUser{
			ID:        row.InviterUserID,
			Email:     row.InviterEmail,
			FirstName: row.InviterFirstName,
			LastName:  row.InviterLastName,
			Username:  username,
		},
		Organization: org,
		Workspace:    ws,
		Project:      proj,
	}
}
