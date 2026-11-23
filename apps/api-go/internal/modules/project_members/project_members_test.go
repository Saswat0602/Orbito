package project_members

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRBACResolver struct {
	projectRole middleware.Role
}

func (m *mockRBACResolver) GetOrgMemberRole(_ context.Context, _, _ uuid.UUID) (middleware.Role, error) {
	return middleware.RoleOwner, nil
}
func (m *mockRBACResolver) GetOrgIDBySlug(_ context.Context, _ string) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (m *mockRBACResolver) GetWorkspaceMemberRole(_ context.Context, _, _ uuid.UUID) (middleware.Role, error) {
	return middleware.RoleOwner, nil
}
func (m *mockRBACResolver) GetWorkspaceIDBySlug(_ context.Context, _ string) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (m *mockRBACResolver) GetProjectMemberRole(_ context.Context, _, _ uuid.UUID) (middleware.Role, error) {
	return m.projectRole, nil
}
func (m *mockRBACResolver) GetProjectScopeInfo(_ context.Context, _ uuid.UUID) (*middleware.ProjectScopeInfo, error) {
	return nil, nil
}
func (m *mockRBACResolver) GetProjectScopeInfoBySlug(_ context.Context, _ string) (*middleware.ProjectScopeInfo, error) {
	return nil, nil
}

type mockRepo struct {
	projects map[uuid.UUID]*sqlc.GetProjectByIDRowRow
	users    map[uuid.UUID]*sqlc.GetUserByIDRow
	members  map[uuid.UUID]*sqlc.ProjectMember
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		projects: make(map[uuid.UUID]*sqlc.GetProjectByIDRowRow),
		users:    make(map[uuid.UUID]*sqlc.GetUserByIDRow),
		members:  make(map[uuid.UUID]*sqlc.ProjectMember),
	}
}

func (m *mockRepo) ListMembers(_ context.Context, projectID uuid.UUID, _ *string) ([]sqlc.ListProjectMembersRow, error) {
	var res []sqlc.ListProjectMembersRow
	for _, pm := range m.members {
		if pm.ProjectID == projectID {
			u := m.users[pm.UserID]
			email, fn, ln := "user@example.com", "Test", "User"
			if u != nil {
				email = u.Email
				fn = u.FirstName
				ln = u.LastName
			}
			res = append(res, sqlc.ListProjectMembersRow{
				ID:             pm.ID,
				Role:           pm.Role,
				JoinedAt:       pm.JoinedAt,
				UserID:         pm.UserID,
				ProjectID:      pm.ProjectID,
				MemberUserID:   pm.UserID,
				UserEmail:      email,
				UserFirstName:  fn,
				UserLastName:   ln,
				ProjectName:    "Orbito Project",
				ProjectSlug:    "orbito-project",
				UserGlobalRole: sqlc.RoleMEMBER,
			})
		}
	}
	return res, nil
}

func (m *mockRepo) GetByID(_ context.Context, id uuid.UUID) (*sqlc.GetProjectMemberByIDRow, error) {
	pm, ok := m.members[id]
	if !ok {
		return nil, nil
	}
	u := m.users[pm.UserID]
	email, fn, ln := "user@example.com", "Test", "User"
	if u != nil {
		email = u.Email
		fn = u.FirstName
		ln = u.LastName
	}
	return &sqlc.GetProjectMemberByIDRow{
		ID:            pm.ID,
		Role:          pm.Role,
		JoinedAt:      pm.JoinedAt,
		UserID:        pm.UserID,
		ProjectID:     pm.ProjectID,
		MemberUserID:  pm.UserID,
		UserEmail:     email,
		UserFirstName: fn,
		UserLastName:  ln,
		ProjectName:   "Orbito Project",
		ProjectSlug:   "orbito-project",
		WorkspaceID:   uuid.New(),
	}, nil
}

func (m *mockRepo) GetByUserAndProject(_ context.Context, userID, projectID uuid.UUID) (*sqlc.GetProjectMemberByUserAndProjectRow, error) {
	for _, pm := range m.members {
		if pm.UserID == userID && pm.ProjectID == projectID {
			u := m.users[pm.UserID]
			email, fn, ln := "user@example.com", "Test", "User"
			if u != nil {
				email = u.Email
				fn = u.FirstName
				ln = u.LastName
			}
			return &sqlc.GetProjectMemberByUserAndProjectRow{
				ID:            pm.ID,
				Role:          pm.Role,
				JoinedAt:      pm.JoinedAt,
				UserID:        pm.UserID,
				ProjectID:     pm.ProjectID,
				MemberUserID:  pm.UserID,
				UserEmail:     email,
				UserFirstName: fn,
				UserLastName:  ln,
			}, nil
		}
	}
	return nil, nil
}

func (m *mockRepo) Create(_ context.Context, params sqlc.CreateProjectMemberRecordParams) (*sqlc.ProjectMember, error) {
	pm := &sqlc.ProjectMember{
		ID:        params.ID,
		UserID:    params.UserID,
		ProjectID: params.ProjectID,
		Role:      params.Role,
		JoinedAt:  pgtype.Timestamp{Time: time.Now(), Valid: true},
		CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
	}
	m.members[pm.ID] = pm
	return pm, nil
}

func (m *mockRepo) UpdateRole(_ context.Context, id uuid.UUID, role sqlc.Role, _ uuid.UUID) (*sqlc.ProjectMember, error) {
	pm, ok := m.members[id]
	if !ok {
		return nil, nil
	}
	pm.Role = role
	return pm, nil
}

func (m *mockRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.members, id)
	return nil
}

func (m *mockRepo) CountOwners(_ context.Context, projectID uuid.UUID) (int64, error) {
	var count int64
	for _, pm := range m.members {
		if pm.ProjectID == projectID && pm.Role == sqlc.RoleOWNER {
			count++
		}
	}
	return count, nil
}

func (m *mockRepo) ListByWorkspace(_ context.Context, _ uuid.UUID) ([]sqlc.ListProjectMembersByWorkspaceRow, error) {
	return nil, nil
}

func (m *mockRepo) ListUserProjects(_ context.Context, _ uuid.UUID) ([]sqlc.ListUserProjectsRow, error) {
	return nil, nil
}

func (m *mockRepo) RoleStats(_ context.Context, projectID uuid.UUID) ([]sqlc.ProjectMembersRoleStatsRow, error) {
	counts := make(map[sqlc.Role]int64)
	for _, pm := range m.members {
		if pm.ProjectID == projectID {
			counts[pm.Role]++
		}
	}
	var res []sqlc.ProjectMembersRoleStatsRow
	for role, cnt := range counts {
		res = append(res, sqlc.ProjectMembersRoleStatsRow{
			Role:  role,
			Count: cnt,
		})
	}
	return res, nil
}

func (m *mockRepo) GetUserByID(_ context.Context, id uuid.UUID) (*sqlc.GetUserByIDRow, error) {
	return m.users[id], nil
}

func (m *mockRepo) GetProjectByID(_ context.Context, id uuid.UUID) (*sqlc.GetProjectByIDRowRow, error) {
	return m.projects[id], nil
}

func (m *mockRepo) WithTx(_ context.Context, fn func(repo Repository) error) error {
	return fn(m)
}

func TestProjectMembersService(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	ownerUserID := uuid.New()
	targetUserID := uuid.New()

	repo := newMockRepo()
	repo.projects[projectID] = &sqlc.GetProjectByIDRowRow{
		ID:          projectID,
		Name:        "Orbito Project",
		Slug:        "orbito-project",
		WorkspaceID: uuid.New(),
	}
	repo.users[ownerUserID] = &sqlc.GetUserByIDRow{
		ID:        ownerUserID,
		Email:     "owner@example.com",
		FirstName: "Project",
		LastName:  "Owner",
	}
	repo.users[targetUserID] = &sqlc.GetUserByIDRow{
		ID:        targetUserID,
		Email:     "target@example.com",
		FirstName: "Target",
		LastName:  "Member",
	}

	// Add initial owner
	ownerMemberID := uuid.New()
	repo.members[ownerMemberID] = &sqlc.ProjectMember{
		ID:        ownerMemberID,
		UserID:    ownerUserID,
		ProjectID: projectID,
		Role:      sqlc.RoleOWNER,
		JoinedAt:  pgtype.Timestamp{Time: time.Now(), Valid: true},
	}

	rbac := &mockRBACResolver{
		projectRole: middleware.RoleOwner,
	}

	svc := NewService(repo, rbac)

	t.Run("Create project member - adds member with designated role", func(t *testing.T) {
		roleStr := "MEMBER"
		res, err := svc.Create(ctx, CreateProjectMemberRequest{
			UserID:    targetUserID,
			ProjectID: projectID,
			Role:      &roleStr,
		}, ownerUserID, false)

		require.NoError(t, err)
		assert.Equal(t, targetUserID, res.UserID)
		assert.Equal(t, projectID, res.ProjectID)
		assert.Equal(t, "MEMBER", res.Role)
		assert.Equal(t, "target@example.com", res.User.Email)
	})

	t.Run("Create project member - already exists returns 409", func(t *testing.T) {
		roleStr := "VIEWER"
		_, err := svc.Create(ctx, CreateProjectMemberRequest{
			UserID:    targetUserID,
			ProjectID: projectID,
			Role:      &roleStr,
		}, ownerUserID, false)

		require.Error(t, err)
		var appErr *apperror.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, 409, appErr.StatusCode)
	})

	t.Run("ListMembers - returns project members with profiles", func(t *testing.T) {
		members, err := svc.ListMembers(ctx, projectID, nil, ownerUserID, false)
		require.NoError(t, err)
		assert.Len(t, members, 2)
	})

	t.Run("UpdateRole - cannot demote sole project owner", func(t *testing.T) {
		_, err := svc.UpdateRole(ctx, ownerMemberID, UpdateProjectMemberRoleRequest{
			Role: "VIEWER",
		}, ownerUserID, false)

		require.Error(t, err)
		var appErr *apperror.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, 400, appErr.StatusCode)
		assert.Equal(t, "Cannot demote the sole project owner", appErr.Message)
	})

	t.Run("Remove - cannot remove sole project owner", func(t *testing.T) {
		err := svc.Remove(ctx, ownerMemberID, ownerUserID, false)
		require.Error(t, err)
		var appErr *apperror.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, 400, appErr.StatusCode)
		assert.Equal(t, "Cannot remove the sole project owner", appErr.Message)
	})

	t.Run("Remove - deletes non-owner member", func(t *testing.T) {
		// Find target member
		targetMember, err := svc.FindByUserAndProject(ctx, targetUserID, projectID, ownerUserID, false)
		require.NoError(t, err)

		err = svc.Remove(ctx, targetMember.ID, ownerUserID, false)
		require.NoError(t, err)
		assert.Nil(t, repo.members[targetMember.ID])
	})

	t.Run("GetStats - returns role breakdown", func(t *testing.T) {
		stats, err := svc.GetStats(ctx, projectID, ownerUserID, false)
		require.NoError(t, err)
		assert.Equal(t, int64(1), stats.TotalMembers)
		assert.Equal(t, int64(1), stats.ByRole["OWNER"])
	})
}
