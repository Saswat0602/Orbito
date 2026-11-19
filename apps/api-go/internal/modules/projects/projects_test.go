package projects

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRBACResolver struct {
	workspaceRole middleware.Role
	projectRole   middleware.Role
}

func (m *mockRBACResolver) GetOrgMemberRole(_ context.Context, _, _ uuid.UUID) (middleware.Role, error) {
	return middleware.RoleOwner, nil
}
func (m *mockRBACResolver) GetOrgIDBySlug(_ context.Context, _ string) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (m *mockRBACResolver) GetWorkspaceMemberRole(_ context.Context, _, _ uuid.UUID) (middleware.Role, error) {
	return m.workspaceRole, nil
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
	projects         map[uuid.UUID]*sqlc.Project
	workspaces       map[uuid.UUID]*sqlc.GetWorkspaceByIDRow
	orgs             map[uuid.UUID]*sqlc.GetOrganizationDetailsByIDRow
	workflows        map[uuid.UUID]*sqlc.Workflow
	defaultWorkflows map[uuid.UUID]*sqlc.Workflow
	sprints          map[uuid.UUID]*sqlc.Sprint
	members          map[string]*sqlc.ProjectMember
	tasksCount       int64
	completedCount   int64
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		projects:         make(map[uuid.UUID]*sqlc.Project),
		workspaces:       make(map[uuid.UUID]*sqlc.GetWorkspaceByIDRow),
		orgs:             make(map[uuid.UUID]*sqlc.GetOrganizationDetailsByIDRow),
		workflows:        make(map[uuid.UUID]*sqlc.Workflow),
		defaultWorkflows: make(map[uuid.UUID]*sqlc.Workflow),
		sprints:          make(map[uuid.UUID]*sqlc.Sprint),
		members:          make(map[string]*sqlc.ProjectMember),
	}
}

func (m *mockRepo) Create(_ context.Context, params sqlc.CreateProjectParams) (*sqlc.Project, error) {
	p := &sqlc.Project{
		ID:          params.ID,
		Name:        params.Name,
		Slug:        params.Slug,
		TaskPrefix:  params.TaskPrefix,
		Description: params.Description,
		Avatar:      params.Avatar,
		Color:       params.Color,
		Status:      params.Status,
		Priority:    params.Priority,
		Visibility:  params.Visibility,
		StartDate:   params.StartDate,
		EndDate:     params.EndDate,
		WorkflowID:  params.WorkflowID,
		WorkspaceID: params.WorkspaceID,
		CreatedByID: params.CreatedByID,
		UpdatedByID: params.UpdatedByID,
		CreatedAt:   toPgTimestamp(time.Now()),
		UpdatedAt:   toPgTimestamp(time.Now()),
		Archive:     false,
	}
	m.projects[p.ID] = p
	return p, nil
}

func (m *mockRepo) GetByID(_ context.Context, id uuid.UUID) (*sqlc.GetProjectByIDRowRow, error) {
	p, ok := m.projects[id]
	if !ok {
		return nil, nil
	}
	return &sqlc.GetProjectByIDRowRow{
		ID:               p.ID,
		Name:             p.Name,
		Slug:             p.Slug,
		TaskPrefix:       p.TaskPrefix,
		Description:      p.Description,
		Avatar:           p.Avatar,
		Color:            p.Color,
		Status:           p.Status,
		Priority:         p.Priority,
		Visibility:       p.Visibility,
		WorkflowID:       p.WorkflowID,
		WorkspaceID:      p.WorkspaceID,
		Archive:          p.Archive,
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
		WorkspaceName:    "Core Workspace",
		WorkspaceSlug:    "core-workspace",
		OrganizationID:   uuid.New(),
		OrganizationName: "Acme Org",
		OrganizationSlug: "acme-org",
	}, nil
}

func (m *mockRepo) GetBySlug(_ context.Context, slug string) (*sqlc.GetProjectBySlugRowRow, error) {
	for _, p := range m.projects {
		if p.Slug == slug {
			return &sqlc.GetProjectBySlugRowRow{
				ID:               p.ID,
				Name:             p.Name,
				Slug:             p.Slug,
				TaskPrefix:       p.TaskPrefix,
				Description:      p.Description,
				Avatar:           p.Avatar,
				Color:            p.Color,
				Status:           p.Status,
				Priority:         p.Priority,
				Visibility:       p.Visibility,
				WorkflowID:       p.WorkflowID,
				WorkspaceID:      p.WorkspaceID,
				Archive:          p.Archive,
				CreatedAt:        p.CreatedAt,
				UpdatedAt:        p.UpdatedAt,
				WorkspaceName:    "Core Workspace",
				WorkspaceSlug:    "core-workspace",
				OrganizationID:   uuid.New(),
				OrganizationName: "Acme Org",
				OrganizationSlug: "acme-org",
			}, nil
		}
	}
	return nil, nil
}

func (m *mockRepo) GetByKey(_ context.Context, workspaceID uuid.UUID, key string) (*sqlc.GetProjectByKeyRowRow, error) {
	for _, p := range m.projects {
		if p.WorkspaceID == workspaceID && (p.TaskPrefix.String == key || p.Slug == key) {
			return &sqlc.GetProjectByKeyRowRow{
				ID:               p.ID,
				Name:             p.Name,
				Slug:             p.Slug,
				TaskPrefix:       p.TaskPrefix,
				Description:      p.Description,
				Avatar:           p.Avatar,
				Color:            p.Color,
				Status:           p.Status,
				Priority:         p.Priority,
				Visibility:       p.Visibility,
				WorkflowID:       p.WorkflowID,
				WorkspaceID:      p.WorkspaceID,
				Archive:          p.Archive,
				CreatedAt:        p.CreatedAt,
				UpdatedAt:        p.UpdatedAt,
				WorkspaceName:    "Core Workspace",
				WorkspaceSlug:    "core-workspace",
				OrganizationID:   uuid.New(),
				OrganizationName: "Acme Org",
				OrganizationSlug: "acme-org",
			}, nil
		}
	}
	return nil, nil
}

func (m *mockRepo) ListForUser(_ context.Context, _ sqlc.ListProjectsForUserParams) ([]sqlc.ListProjectsForUserRow, error) {
	var res []sqlc.ListProjectsForUserRow
	for _, p := range m.projects {
		if !p.Archive {
			res = append(res, sqlc.ListProjectsForUserRow{
				ID:            p.ID,
				Name:          p.Name,
				Slug:          p.Slug,
				TaskPrefix:    p.TaskPrefix,
				Description:   p.Description,
				Avatar:        p.Avatar,
				Color:         p.Color,
				Status:        p.Status,
				Priority:      p.Priority,
				Visibility:    p.Visibility,
				WorkflowID:    p.WorkflowID,
				WorkspaceID:   p.WorkspaceID,
				CreatedAt:     p.CreatedAt,
				UpdatedAt:     p.UpdatedAt,
				Archive:       p.Archive,
				WorkspaceName: "Core Workspace",
				WorkspaceSlug: "core-workspace",
				TaskCount:     m.tasksCount,
				MemberCount:   int64(len(m.members)),
			})
		}
	}
	return res, nil
}

func (m *mockRepo) CountForUser(_ context.Context, _ sqlc.CountProjectsForUserParams) (int64, error) {
	var count int64
	for _, p := range m.projects {
		if !p.Archive {
			count++
		}
	}
	return count, nil
}

func (m *mockRepo) ListByOrganization(_ context.Context, _ sqlc.ListProjectsByOrganizationParams) ([]sqlc.ListProjectsByOrganizationRow, error) {
	return nil, nil
}

func (m *mockRepo) CountByOrganization(_ context.Context, _ sqlc.CountProjectsByOrganizationParams) (int64, error) {
	return 0, nil
}

func (m *mockRepo) ListArchivedByWorkspace(_ context.Context, wsID uuid.UUID) ([]sqlc.ListArchivedProjectsByWorkspaceRow, error) {
	var res []sqlc.ListArchivedProjectsByWorkspaceRow
	for _, p := range m.projects {
		if p.Archive && p.WorkspaceID == wsID {
			res = append(res, sqlc.ListArchivedProjectsByWorkspaceRow{
				ID:          p.ID,
				Name:        p.Name,
				Slug:        p.Slug,
				WorkspaceID: p.WorkspaceID,
				Archive:     true,
			})
		}
	}
	return res, nil
}

func (m *mockRepo) ListArchivedByOrganization(_ context.Context, _ uuid.UUID) ([]sqlc.ListArchivedProjectsByOrganizationRow, error) {
	return nil, nil
}

func (m *mockRepo) Update(_ context.Context, params sqlc.UpdateProjectParams) (*sqlc.Project, error) {
	p, ok := m.projects[params.ID]
	if !ok {
		return nil, nil
	}
	if params.Name.Valid {
		p.Name = params.Name.String
	}
	if params.Color.Valid {
		p.Color = params.Color.String
	}
	p.UpdatedAt = toPgTimestamp(time.Now())
	return p, nil
}

func (m *mockRepo) Archive(_ context.Context, id, _ uuid.UUID) error {
	if p, ok := m.projects[id]; ok {
		p.Archive = true
	}
	return nil
}

func (m *mockRepo) Unarchive(_ context.Context, id, _ uuid.UUID) error {
	if p, ok := m.projects[id]; ok {
		p.Archive = false
	}
	return nil
}

func (m *mockRepo) Delete(_ context.Context, id uuid.UUID) error {
	delete(m.projects, id)
	return nil
}

func (m *mockRepo) CheckSlugExists(_ context.Context, slug string, _ *uuid.UUID) (bool, error) {
	for _, p := range m.projects {
		if p.Slug == slug {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockRepo) ListSlugsStartingWith(_ context.Context, prefix string) ([]string, error) {
	var res []string
	for _, p := range m.projects {
		if len(p.Slug) >= len(prefix) && p.Slug[:len(prefix)] == prefix {
			res = append(res, p.Slug)
		}
	}
	return res, nil
}

func (m *mockRepo) GetDefaultWorkflowForOrg(_ context.Context, orgID uuid.UUID) (*sqlc.Workflow, error) {
	return m.defaultWorkflows[orgID], nil
}

func (m *mockRepo) GetWorkflowByID(_ context.Context, id uuid.UUID) (*sqlc.Workflow, error) {
	return m.workflows[id], nil
}

func (m *mockRepo) CreateInitialSprint(_ context.Context, params sqlc.CreateInitialSprintParams) (*sqlc.Sprint, error) {
	s := &sqlc.Sprint{
		ID:        params.ID,
		Name:      params.Name,
		Goal:      params.Goal,
		Status:    params.Status,
		IsDefault: params.IsDefault,
		ProjectID: params.ProjectID,
	}
	m.sprints[s.ID] = s
	return s, nil
}

func (m *mockRepo) CreateProjectMember(_ context.Context, params sqlc.CreateProjectMemberDirectParams) (*sqlc.ProjectMember, error) {
	pm := &sqlc.ProjectMember{
		ID:        params.ID,
		UserID:    params.UserID,
		ProjectID: params.ProjectID,
		Role:      params.Role,
	}
	m.members[params.UserID.String()+":"+params.ProjectID.String()] = pm
	return pm, nil
}

func (m *mockRepo) GetWorkspaceByID(_ context.Context, id uuid.UUID) (*sqlc.GetWorkspaceByIDRow, error) {
	return m.workspaces[id], nil
}

func (m *mockRepo) GetOrganizationByID(_ context.Context, id uuid.UUID) (*sqlc.GetOrganizationDetailsByIDRow, error) {
	return m.orgs[id], nil
}

func (m *mockRepo) CountTasks(_ context.Context, _ uuid.UUID) (int64, error) {
	return m.tasksCount, nil
}

func (m *mockRepo) CountCompletedTasks(_ context.Context, _ uuid.UUID) (int64, error) {
	return m.completedCount, nil
}

func (m *mockRepo) CountMembers(_ context.Context, _ uuid.UUID) (int64, error) {
	return int64(len(m.members)), nil
}

func (m *mockRepo) TaskStatusStats(_ context.Context, _ uuid.UUID) ([]sqlc.ProjectTaskStatusStatsRow, error) {
	return nil, nil
}

func (m *mockRepo) TaskPriorityStats(_ context.Context, _ uuid.UUID) ([]sqlc.ProjectTaskPriorityStatsRow, error) {
	return nil, nil
}

func (m *mockRepo) WithTx(_ context.Context, fn func(repo Repository) error) error {
	return fn(m)
}

func TestProjectsService(t *testing.T) {
	ctx := context.Background()
	orgID := uuid.New()
	wsID := uuid.New()
	wfID := uuid.New()
	userID := uuid.New()

	repo := newMockRepo()
	repo.orgs[orgID] = &sqlc.GetOrganizationDetailsByIDRow{
		ID:      orgID,
		Name:    "Acme Org",
		Slug:    "acme-org",
		OwnerID: userID,
		Archive: false,
	}
	repo.workspaces[wsID] = &sqlc.GetWorkspaceByIDRow{
		ID:             wsID,
		Name:           "Core Workspace",
		Slug:           "core-workspace",
		OrganizationID: orgID,
		Archive:        false,
	}
	defaultWf := &sqlc.Workflow{
		ID:             wfID,
		Name:           "Default Workflow",
		IsDefault:      true,
		OrganizationID: orgID,
	}
	repo.defaultWorkflows[orgID] = defaultWf
	repo.workflows[wfID] = defaultWf

	rbac := &mockRBACResolver{
		workspaceRole: middleware.RoleOwner,
		projectRole:   middleware.RoleOwner,
	}

	svc := NewService(repo, rbac)

	t.Run("Create project - assigns custom task prefix and default workflow", func(t *testing.T) {
		pPrefix := "ORB"
		res, err := svc.Create(ctx, CreateProjectRequest{
			Name:        "Platform Engineering",
			TaskPrefix:  &pPrefix,
			WorkspaceID: wsID,
			Color:       "#10b981",
		}, userID, false)

		require.NoError(t, err)
		assert.Equal(t, "Platform Engineering", res.Name)
		assert.Equal(t, "platform-engineering", res.Slug)
		assert.NotNil(t, res.TaskPrefix)
		assert.Equal(t, "ORB", *res.TaskPrefix)
		assert.Equal(t, wfID, res.WorkflowID)
		assert.Equal(t, wsID, res.WorkspaceID)

		// Initial sprint must be created
		assert.Len(t, repo.sprints, 1)
		// Creator must be added as member
		assert.Len(t, repo.members, 1)
	})

	t.Run("Create project - auto-generates task prefix if omitted", func(t *testing.T) {
		res, err := svc.Create(ctx, CreateProjectRequest{
			Name:        "Mobile Application",
			WorkspaceID: wsID,
		}, userID, false)

		require.NoError(t, err)
		assert.NotNil(t, res.TaskPrefix)
		assert.Equal(t, "MA", *res.TaskPrefix)
	})

	t.Run("Create project - missing name returns 400", func(t *testing.T) {
		_, err := svc.Create(ctx, CreateProjectRequest{
			WorkspaceID: wsID,
		}, userID, false)

		require.Error(t, err)
		var appErr *apperror.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, 400, appErr.StatusCode)
	})

	t.Run("Create project - non-existent workspace returns 404", func(t *testing.T) {
		_, err := svc.Create(ctx, CreateProjectRequest{
			Name:        "Test Project",
			WorkspaceID: uuid.New(),
		}, userID, false)

		require.Error(t, err)
		var appErr *apperror.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, 404, appErr.StatusCode)
	})

	t.Run("FindAll projects - lists active projects", func(t *testing.T) {
		res, err := svc.FindAll(ctx, &wsID, nil, nil, nil, 1, 10, userID, false)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, res.Total, int64(2))
		assert.Len(t, res.Data, int(res.Total))
	})

	t.Run("GetStats - returns total, completed, and completion rate", func(t *testing.T) {
		projList, err := svc.FindAll(ctx, &wsID, nil, nil, nil, 1, 10, userID, false)
		require.NoError(t, err)
		require.NotEmpty(t, projList.Data)
		targetID := projList.Data[0].ID

		repo.tasksCount = 10
		repo.completedCount = 4

		stats, err := svc.GetStats(ctx, targetID, userID, false)
		require.NoError(t, err)
		assert.Equal(t, int64(10), stats.TotalTasks)
		assert.Equal(t, int64(4), stats.CompletedTasks)
		assert.Equal(t, 40.0, stats.CompletionRate)
	})

	t.Run("Update project", func(t *testing.T) {
		projList, err := svc.FindAll(ctx, &wsID, nil, nil, nil, 1, 10, userID, false)
		require.NoError(t, err)
		targetID := projList.Data[0].ID

		newName := "Updated Platform Engineering"
		res, err := svc.Update(ctx, targetID, UpdateProjectRequest{
			Name: &newName,
		}, userID, false)
		require.NoError(t, err)
		assert.Equal(t, newName, res.Name)
	})

	t.Run("Archive and Unarchive project", func(t *testing.T) {
		projList, err := svc.FindAll(ctx, &wsID, nil, nil, nil, 1, 10, userID, false)
		require.NoError(t, err)
		targetID := projList.Data[0].ID

		err = svc.Archive(ctx, targetID, userID, false)
		require.NoError(t, err)
		assert.True(t, repo.projects[targetID].Archive)

		err = svc.Unarchive(ctx, targetID, userID, false)
		require.NoError(t, err)
		assert.False(t, repo.projects[targetID].Archive)
	})

	t.Run("Delete project", func(t *testing.T) {
		projList, err := svc.FindAll(ctx, &wsID, nil, nil, nil, 1, 10, userID, false)
		require.NoError(t, err)
		targetID := projList.Data[0].ID

		err = svc.Delete(ctx, targetID, userID, false)
		require.NoError(t, err)
		assert.Nil(t, repo.projects[targetID])
	})
}
