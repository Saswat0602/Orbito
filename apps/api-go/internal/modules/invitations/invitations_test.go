package invitations

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/modules/email"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEmailSender struct {
	sent []email.SendEmailDTO
}

func (m *mockEmailSender) SendEmail(_ context.Context, dto email.SendEmailDTO) error {
	m.sent = append(m.sent, dto)
	return nil
}

type mockRBACResolver struct {
	orgRole       middleware.Role
	workspaceRole middleware.Role
	projectRole   middleware.Role
}

func (m *mockRBACResolver) GetOrgMemberRole(_ context.Context, _, _ uuid.UUID) (middleware.Role, error) {
	return m.orgRole, nil
}

func (m *mockRBACResolver) GetWorkspaceMemberRole(_ context.Context, _, _ uuid.UUID) (middleware.Role, error) {
	return m.workspaceRole, nil
}

func (m *mockRBACResolver) GetProjectMemberRole(_ context.Context, _, _ uuid.UUID) (middleware.Role, error) {
	return m.projectRole, nil
}

func (m *mockRBACResolver) GetOrgIDBySlug(_ context.Context, _ string) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (m *mockRBACResolver) GetWorkspaceIDBySlug(_ context.Context, _ string) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func (m *mockRBACResolver) GetProjectScopeInfo(_ context.Context, _ uuid.UUID) (*middleware.ProjectScopeInfo, error) {
	return nil, nil
}

func (m *mockRBACResolver) GetProjectScopeInfoBySlug(_ context.Context, _ string) (*middleware.ProjectScopeInfo, error) {
	return nil, nil
}

type mockRepo struct {
	invitations      map[uuid.UUID]*sqlc.Invitation
	invitationsByTok map[string]*sqlc.Invitation
	orgs             map[uuid.UUID]*sqlc.GetOrganizationDetailsByIDRow
	workspaces       map[uuid.UUID]*sqlc.GetWorkspaceByIDRow
	projects         map[uuid.UUID]*sqlc.GetProjectWithWorkspaceAndOrgRow
	users            map[string]*sqlc.GetUserByEmailRow
	usersByID        map[uuid.UUID]*sqlc.GetUserByIDRow
	orgMembers       map[string]bool
	workspaceMembers map[string]bool
	projectMembers   map[string]bool
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		invitations:      make(map[uuid.UUID]*sqlc.Invitation),
		invitationsByTok: make(map[string]*sqlc.Invitation),
		orgs:             make(map[uuid.UUID]*sqlc.GetOrganizationDetailsByIDRow),
		workspaces:       make(map[uuid.UUID]*sqlc.GetWorkspaceByIDRow),
		projects:         make(map[uuid.UUID]*sqlc.GetProjectWithWorkspaceAndOrgRow),
		users:            make(map[string]*sqlc.GetUserByEmailRow),
		usersByID:        make(map[uuid.UUID]*sqlc.GetUserByIDRow),
		orgMembers:       make(map[string]bool),
		workspaceMembers: make(map[string]bool),
		projectMembers:   make(map[string]bool),
	}
}

func (m *mockRepo) Create(_ context.Context, params sqlc.CreateInvitationParams) (*sqlc.Invitation, error) {
	inv := &sqlc.Invitation{
		ID:             params.ID,
		InviterID:      params.InviterID,
		InviteeEmail:   params.InviteeEmail,
		OrganizationID: params.OrganizationID,
		WorkspaceID:    params.WorkspaceID,
		ProjectID:      params.ProjectID,
		Role:           params.Role,
		Status:         params.Status,
		Token:          params.Token,
		ExpiresAt:      params.ExpiresAt,
		CreatedAt:      toPgTimestamp(time.Now()),
		UpdatedAt:      toPgTimestamp(time.Now()),
	}
	m.invitations[inv.ID] = inv
	m.invitationsByTok[inv.Token] = inv
	return inv, nil
}

func (m *mockRepo) GetByID(_ context.Context, id uuid.UUID) (*sqlc.GetInvitationByIDRow, error) {
	inv, ok := m.invitations[id]
	if !ok {
		return nil, nil
	}
	return &sqlc.GetInvitationByIDRow{
		ID:               inv.ID,
		InviterID:        inv.InviterID,
		InviteeEmail:     inv.InviteeEmail,
		OrganizationID:   inv.OrganizationID,
		WorkspaceID:      inv.WorkspaceID,
		ProjectID:        inv.ProjectID,
		Role:             inv.Role,
		Status:           inv.Status,
		Token:            inv.Token,
		ExpiresAt:        inv.ExpiresAt,
		CreatedAt:        inv.CreatedAt,
		UpdatedAt:        inv.UpdatedAt,
		InviterUserID:    inv.InviterID,
		InviterEmail:     "inviter@example.com",
		InviterFirstName: "Jane",
		InviterLastName:  "Doe",
		OrgName:          pgtype.Text{String: "Acme Org", Valid: true},
		OrgSlug:          pgtype.Text{String: "acme-org", Valid: true},
	}, nil
}

func (m *mockRepo) GetByToken(_ context.Context, token string) (*sqlc.GetInvitationByTokenRow, error) {
	inv, ok := m.invitationsByTok[token]
	if !ok {
		return nil, nil
	}
	return &sqlc.GetInvitationByTokenRow{
		ID:               inv.ID,
		InviterID:        inv.InviterID,
		InviteeEmail:     inv.InviteeEmail,
		OrganizationID:   inv.OrganizationID,
		WorkspaceID:      inv.WorkspaceID,
		ProjectID:        inv.ProjectID,
		Role:             inv.Role,
		Status:           inv.Status,
		Token:            inv.Token,
		ExpiresAt:        inv.ExpiresAt,
		CreatedAt:        inv.CreatedAt,
		UpdatedAt:        inv.UpdatedAt,
		InviterUserID:    inv.InviterID,
		InviterEmail:     "inviter@example.com",
		InviterFirstName: "Jane",
		InviterLastName:  "Doe",
		OrgName:          pgtype.Text{String: "Acme Org", Valid: true},
		OrgSlug:          pgtype.Text{String: "acme-org", Valid: true},
	}, nil
}

func (m *mockRepo) FindPending(_ context.Context, params sqlc.FindPendingInvitationParams) (*sqlc.Invitation, error) {
	for _, inv := range m.invitations {
		if inv.InviteeEmail == params.InviteeEmail && inv.Status == sqlc.InvitationStatusPENDING {
			return inv, nil
		}
	}
	return nil, nil
}

func (m *mockRepo) GetUserInvitations(_ context.Context, email string) ([]sqlc.GetUserInvitationsRow, error) {
	var res []sqlc.GetUserInvitationsRow
	for _, inv := range m.invitations {
		if inv.InviteeEmail == email && inv.Status == sqlc.InvitationStatusPENDING {
			res = append(res, sqlc.GetUserInvitationsRow{
				ID:               inv.ID,
				InviterID:        inv.InviterID,
				InviteeEmail:     inv.InviteeEmail,
				OrganizationID:   inv.OrganizationID,
				WorkspaceID:      inv.WorkspaceID,
				ProjectID:        inv.ProjectID,
				Role:             inv.Role,
				Status:           inv.Status,
				Token:            inv.Token,
				ExpiresAt:        inv.ExpiresAt,
				CreatedAt:        inv.CreatedAt,
				UpdatedAt:        inv.UpdatedAt,
				InviterUserID:    inv.InviterID,
				InviterEmail:     "inviter@example.com",
				InviterFirstName: "Jane",
				InviterLastName:  "Doe",
			})
		}
	}
	return res, nil
}

func (m *mockRepo) GetEntityInvitations(_ context.Context, _ string, _ uuid.UUID) ([]sqlc.GetEntityInvitationsRow, error) {
	var res []sqlc.GetEntityInvitationsRow
	for _, inv := range m.invitations {
		res = append(res, sqlc.GetEntityInvitationsRow{
			ID:               inv.ID,
			InviterID:        inv.InviterID,
			InviteeEmail:     inv.InviteeEmail,
			OrganizationID:   inv.OrganizationID,
			WorkspaceID:      inv.WorkspaceID,
			ProjectID:        inv.ProjectID,
			Role:             inv.Role,
			Status:           inv.Status,
			Token:            inv.Token,
			ExpiresAt:        inv.ExpiresAt,
			CreatedAt:        inv.CreatedAt,
			UpdatedAt:        inv.UpdatedAt,
			InviterUserID:    inv.InviterID,
			InviterEmail:     "inviter@example.com",
			InviterFirstName: "Jane",
			InviterLastName:  "Doe",
		})
	}
	return res, nil
}

func (m *mockRepo) ExpirePendingEntityInvitations(_ context.Context, _ string, _ uuid.UUID) error {
	for _, inv := range m.invitations {
		if inv.Status == sqlc.InvitationStatusPENDING && inv.ExpiresAt.Time.Before(time.Now()) {
			inv.Status = sqlc.InvitationStatusEXPIRED
		}
	}
	return nil
}

func (m *mockRepo) UpdateStatus(_ context.Context, id uuid.UUID, status sqlc.InvitationStatus) error {
	if inv, ok := m.invitations[id]; ok {
		inv.Status = status
	}
	return nil
}

func (m *mockRepo) Resend(_ context.Context, params sqlc.ResendInvitationParams) (*sqlc.Invitation, error) {
	inv, ok := m.invitations[params.ID]
	if !ok {
		return nil, nil
	}
	delete(m.invitationsByTok, inv.Token)
	inv.Token = params.Token
	inv.ExpiresAt = params.ExpiresAt
	inv.Status = sqlc.InvitationStatusPENDING
	inv.InviterID = params.InviterID
	m.invitationsByTok[inv.Token] = inv
	return inv, nil
}

func (m *mockRepo) Delete(_ context.Context, id uuid.UUID) error {
	if inv, ok := m.invitations[id]; ok {
		delete(m.invitationsByTok, inv.Token)
		delete(m.invitations, id)
	}
	return nil
}

func (m *mockRepo) ListRelatedProjectInvitations(_ context.Context, _ string, _ uuid.UUID) ([]sqlc.Invitation, error) {
	return nil, nil
}

func (m *mockRepo) GetProjectWithWorkspaceAndOrg(_ context.Context, id uuid.UUID) (*sqlc.GetProjectWithWorkspaceAndOrgRow, error) {
	return m.projects[id], nil
}

func (m *mockRepo) CheckProjectBelongsToWorkspace(_ context.Context, projectID, workspaceID uuid.UUID) (bool, error) {
	p, ok := m.projects[projectID]
	if !ok {
		return false, nil
	}
	return p.WorkspaceID == workspaceID, nil
}

func (m *mockRepo) CheckOrgMemberExists(_ context.Context, userID, orgID uuid.UUID) (bool, error) {
	return m.orgMembers[userID.String()+":"+orgID.String()], nil
}

func (m *mockRepo) CheckWorkspaceMemberExists(_ context.Context, userID, wsID uuid.UUID) (bool, error) {
	return m.workspaceMembers[userID.String()+":"+wsID.String()], nil
}

func (m *mockRepo) CheckProjectMemberExists(_ context.Context, userID, projID uuid.UUID) (bool, error) {
	return m.projectMembers[userID.String()+":"+projID.String()], nil
}

func (m *mockRepo) GetUserByEmail(_ context.Context, email string) (*sqlc.GetUserByEmailRow, error) {
	return m.users[email], nil
}

func (m *mockRepo) GetUserByID(_ context.Context, id uuid.UUID) (*sqlc.GetUserByIDRow, error) {
	return m.usersByID[id], nil
}

func (m *mockRepo) GetWorkspaceByID(_ context.Context, id uuid.UUID) (*sqlc.GetWorkspaceByIDRow, error) {
	return m.workspaces[id], nil
}

func (m *mockRepo) GetOrganizationByID(_ context.Context, id uuid.UUID) (*sqlc.GetOrganizationDetailsByIDRow, error) {
	return m.orgs[id], nil
}

func (m *mockRepo) CreateOrganizationMember(_ context.Context, params sqlc.CreateOrganizationMemberParams) (*sqlc.OrganizationMember, error) {
	m.orgMembers[params.UserID.String()+":"+params.OrganizationID.String()] = true
	return &sqlc.OrganizationMember{
		ID:             params.ID,
		UserID:         params.UserID,
		OrganizationID: params.OrganizationID,
		Role:           params.Role,
	}, nil
}

func (m *mockRepo) CreateWorkspaceMember(_ context.Context, params sqlc.CreateWorkspaceMemberDirectParams) (*sqlc.WorkspaceMember, error) {
	m.workspaceMembers[params.UserID.String()+":"+params.WorkspaceID.String()] = true
	return &sqlc.WorkspaceMember{
		ID:          params.ID,
		UserID:      params.UserID,
		WorkspaceID: params.WorkspaceID,
		Role:        params.Role,
	}, nil
}

func (m *mockRepo) CreateProjectMember(_ context.Context, params sqlc.CreateProjectMemberDirectParams) (*sqlc.ProjectMember, error) {
	m.projectMembers[params.UserID.String()+":"+params.ProjectID.String()] = true
	return &sqlc.ProjectMember{
		ID:        params.ID,
		UserID:    params.UserID,
		ProjectID: params.ProjectID,
		Role:      params.Role,
	}, nil
}

func (m *mockRepo) WithTx(_ context.Context, fn func(repo Repository) error) error {
	return fn(m)
}

func TestInvitationsService(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New()
	wsID := uuid.New()
	inviterID := uuid.New()
	inviteeID := uuid.New()

	repo := newMockRepo()
	repo.orgs[orgID] = &sqlc.GetOrganizationDetailsByIDRow{
		ID:   orgID,
		Name: "Acme Org",
		Slug: "acme-org",
	}
	repo.workspaces[wsID] = &sqlc.GetWorkspaceByIDRow{
		ID:             wsID,
		Name:           "Core Engineering",
		Slug:           "core-engineering",
		OrganizationID: orgID,
	}

	rbac := &mockRBACResolver{
		orgRole:       middleware.RoleOwner,
		workspaceRole: middleware.RoleOwner,
	}
	mockEmail := &mockEmailSender{}

	svc := NewService(repo, rbac, mockEmail, "http://localhost:3001")

	t.Run("Create invitation - valid parameters", func(t *testing.T) {
		res, err := svc.CreateInvitation(ctx, CreateInvitationRequest{
			InviteeEmail:   "newuser@example.com",
			OrganizationID: &orgID,
			Role:           "MEMBER",
		}, inviterID, false)

		require.NoError(t, err)
		invResp, ok := res.(*InvitationResponse)
		require.True(t, ok)
		assert.Equal(t, "newuser@example.com", invResp.InviteeEmail)
		assert.Equal(t, "PENDING", invResp.Status)
		assert.Equal(t, "MEMBER", invResp.Role)
		assert.NotEmpty(t, invResp.Token)
		assert.True(t, invResp.ExpiresAt.After(time.Now()))
		assert.True(t, invResp.EmailSent)
		assert.Len(t, mockEmail.sent, 1)
	})

	t.Run("Create invitation - missing email returns 400", func(t *testing.T) {
		_, err := svc.CreateInvitation(ctx, CreateInvitationRequest{
			OrganizationID: &orgID,
			Role:           "MEMBER",
		}, inviterID, false)

		require.Error(t, err)
		var appErr *apperror.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, 400, appErr.StatusCode)
	})

	t.Run("Create invitation - invalid role returns 400", func(t *testing.T) {
		_, err := svc.CreateInvitation(ctx, CreateInvitationRequest{
			InviteeEmail:   "test@example.com",
			OrganizationID: &orgID,
			Role:           "SUPER_ADMIN",
		}, inviterID, false)

		require.Error(t, err)
		var appErr *apperror.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, 400, appErr.StatusCode)
	})

	t.Run("Create invitation - cannot grant role higher than own returns 403", func(t *testing.T) {
		managerRBAC := &mockRBACResolver{
			orgRole: middleware.RoleManager,
		}
		managerSvc := NewService(repo, managerRBAC, mockEmail, "http://localhost:3001")

		_, err := managerSvc.CreateInvitation(ctx, CreateInvitationRequest{
			InviteeEmail:   "candidate@example.com",
			OrganizationID: &orgID,
			Role:           "OWNER",
		}, inviterID, false)

		require.Error(t, err)
		var appErr *apperror.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, 403, appErr.StatusCode)
	})

	t.Run("Create invitation - already member of org returns 400", func(t *testing.T) {
		repo.users["existing@example.com"] = &sqlc.GetUserByEmailRow{
			ID:    inviteeID,
			Email: "existing@example.com",
		}
		repo.orgMembers[inviteeID.String()+":"+orgID.String()] = true

		_, err := svc.CreateInvitation(ctx, CreateInvitationRequest{
			InviteeEmail:   "existing@example.com",
			OrganizationID: &orgID,
			Role:           "MEMBER",
		}, inviterID, false)

		require.Error(t, err)
		var appErr *apperror.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, 400, appErr.StatusCode)
	})

	t.Run("Create invitation - direct add for existing org member to workspace", func(t *testing.T) {
		res, err := svc.CreateInvitation(ctx, CreateInvitationRequest{
			InviteeEmail: "existing@example.com",
			WorkspaceID:  &wsID,
			Role:         "MEMBER",
		}, inviterID, false)

		require.NoError(t, err)
		directAdd, ok := res.(*DirectAddResponse)
		require.True(t, ok)
		assert.Equal(t, "direct_add", directAdd.Type)
		assert.True(t, repo.workspaceMembers[inviteeID.String()+":"+wsID.String()])
	})

	t.Run("Accept invitation - valid token adds member and marks ACCEPTED", func(t *testing.T) {
		userEmail := "acceptor@example.com"
		acceptorID := uuid.New()
		repo.usersByID[acceptorID] = &sqlc.GetUserByIDRow{
			ID:    acceptorID,
			Email: userEmail,
		}

		cRes, err := svc.CreateInvitation(ctx, CreateInvitationRequest{
			InviteeEmail:   userEmail,
			OrganizationID: &orgID,
			Role:           "MEMBER",
		}, inviterID, false)
		require.NoError(t, err)
		invResp := cRes.(*InvitationResponse)

		accRes, err := svc.AcceptInvitation(ctx, invResp.Token, acceptorID)
		require.NoError(t, err)
		assert.Equal(t, "Invitation accepted successfully", accRes.Message)

		// Member must be added to organization
		assert.True(t, repo.orgMembers[acceptorID.String()+":"+orgID.String()])
		// Invitation must be ACCEPTED
		stored := repo.invitationsByTok[invResp.Token]
		assert.Equal(t, sqlc.InvitationStatusACCEPTED, stored.Status)
	})

	t.Run("Accept invitation - expired token returns 400", func(t *testing.T) {
		userEmail := "expired@example.com"
		expiredUserID := uuid.New()
		repo.usersByID[expiredUserID] = &sqlc.GetUserByIDRow{
			ID:    expiredUserID,
			Email: userEmail,
		}

		cRes, err := svc.CreateInvitation(ctx, CreateInvitationRequest{
			InviteeEmail:   userEmail,
			OrganizationID: &orgID,
			Role:           "MEMBER",
		}, inviterID, false)
		require.NoError(t, err)
		invResp := cRes.(*InvitationResponse)

		// Set expired in repository
		repo.invitationsByTok[invResp.Token].ExpiresAt = toPgTimestamp(time.Now().Add(-1 * time.Hour))

		_, err = svc.AcceptInvitation(ctx, invResp.Token, expiredUserID)
		require.Error(t, err)
		var appErr *apperror.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, 400, appErr.StatusCode)
		assert.Equal(t, "Invitation has expired", appErr.Message)
	})

	t.Run("Verify invitation - token status checking", func(t *testing.T) {
		vEmail := "verify@example.com"
		cRes, err := svc.CreateInvitation(ctx, CreateInvitationRequest{
			InviteeEmail:   vEmail,
			OrganizationID: &orgID,
			Role:           "VIEWER",
		}, inviterID, false)
		require.NoError(t, err)
		invResp := cRes.(*InvitationResponse)

		vRes, err := svc.VerifyInvitation(ctx, invResp.Token)
		require.NoError(t, err)
		assert.True(t, vRes.Success)
		assert.Contains(t, vRes.Message, "Valid invitation to join")
	})

	t.Run("Decline invitation - marks DECLINED", func(t *testing.T) {
		dEmail := "decline@example.com"
		cRes, err := svc.CreateInvitation(ctx, CreateInvitationRequest{
			InviteeEmail:   dEmail,
			OrganizationID: &orgID,
			Role:           "MEMBER",
		}, inviterID, false)
		require.NoError(t, err)
		invResp := cRes.(*InvitationResponse)

		dRes, err := svc.DeclineInvitation(ctx, invResp.Token)
		require.NoError(t, err)
		assert.Equal(t, "Invitation declined successfully", dRes.Message)
		assert.Equal(t, sqlc.InvitationStatusDECLINED, repo.invitationsByTok[invResp.Token].Status)
	})

	t.Run("Resend invitation - updates token and resets expiry", func(t *testing.T) {
		rEmail := "resend@example.com"
		cRes, err := svc.CreateInvitation(ctx, CreateInvitationRequest{
			InviteeEmail:   rEmail,
			OrganizationID: &orgID,
			Role:           "MEMBER",
		}, inviterID, false)
		require.NoError(t, err)
		invResp := cRes.(*InvitationResponse)
		oldToken := invResp.Token

		rRes, err := svc.ResendInvitation(ctx, invResp.ID, inviterID)
		require.NoError(t, err)
		assert.NotEqual(t, oldToken, rRes.Invitation.Token)
		assert.True(t, rRes.EmailSent)
	})

	t.Run("Delete invitation - removes record", func(t *testing.T) {
		delEmail := "delete@example.com"
		cRes, err := svc.CreateInvitation(ctx, CreateInvitationRequest{
			InviteeEmail:   delEmail,
			OrganizationID: &orgID,
			Role:           "MEMBER",
		}, inviterID, false)
		require.NoError(t, err)
		invResp := cRes.(*InvitationResponse)

		delRes, err := svc.DeleteInvitation(ctx, invResp.ID)
		require.NoError(t, err)
		assert.Equal(t, "Invitation deleted successfully", delRes.Message)
		assert.Nil(t, repo.invitations[invResp.ID])
	})
}
