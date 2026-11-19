package workspaces_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/modules/workspaces"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) Create(ctx context.Context, params sqlc.CreateWorkspaceParams) (*sqlc.Workspace, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlc.Workspace), args.Error(1)
}

func (m *MockRepo) GetByID(ctx context.Context, id uuid.UUID) (*sqlc.GetWorkspaceByIDRow, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlc.GetWorkspaceByIDRow), args.Error(1)
}

func (m *MockRepo) GetByOrgAndSlug(ctx context.Context, orgID uuid.UUID, slug string) (*sqlc.GetWorkspaceByOrgAndSlugRow, error) {
	args := m.Called(ctx, orgID, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlc.GetWorkspaceByOrgAndSlugRow), args.Error(1)
}

func (m *MockRepo) GetBySlugOnly(ctx context.Context, slug string) (*sqlc.GetWorkspaceBySlugOnlyRow, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlc.GetWorkspaceBySlugOnlyRow), args.Error(1)
}

func (m *MockRepo) ListForUser(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID, search string) ([]sqlc.ListWorkspacesForUserRow, error) {
	args := m.Called(ctx, userID, orgID, search)
	return args.Get(0).([]sqlc.ListWorkspacesForUserRow), args.Error(1)
}

func (m *MockRepo) ListArchived(ctx context.Context, orgID uuid.UUID) ([]sqlc.Workspace, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]sqlc.Workspace), args.Error(1)
}

func (m *MockRepo) ListTree(ctx context.Context, orgID uuid.UUID) ([]sqlc.ListWorkspaceTreeRow, error) {
	args := m.Called(ctx, orgID)
	return args.Get(0).([]sqlc.ListWorkspaceTreeRow), args.Error(1)
}

func (m *MockRepo) Update(ctx context.Context, params sqlc.UpdateWorkspaceParams) (*sqlc.Workspace, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlc.Workspace), args.Error(1)
}

func (m *MockRepo) Archive(ctx context.Context, id, userID uuid.UUID) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockRepo) Unarchive(ctx context.Context, id, userID uuid.UUID) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepo) CheckSlugExists(ctx context.Context, orgID uuid.UUID, slug string, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, orgID, slug, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepo) ListMembersDetails(ctx context.Context, workspaceID uuid.UUID) ([]sqlc.ListWorkspaceMembersDetailsRow, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).([]sqlc.ListWorkspaceMembersDetailsRow), args.Error(1)
}

func (m *MockRepo) CreateMember(ctx context.Context, params sqlc.CreateWorkspaceMemberDirectParams) (*sqlc.WorkspaceMember, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlc.WorkspaceMember), args.Error(1)
}

func (m *MockRepo) CountProjects(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRepo) CountProjectsByStatus(ctx context.Context, workspaceID uuid.UUID, status sqlc.ProjectStatus) (int64, error) {
	args := m.Called(ctx, workspaceID, status)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRepo) CountTasks(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRepo) CountCompletedTasks(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRepo) CountOverdueTasks(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRepo) ProjectStatusStats(ctx context.Context, workspaceID uuid.UUID) ([]sqlc.WorkspaceProjectStatusStatsRow, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).([]sqlc.WorkspaceProjectStatusStatsRow), args.Error(1)
}

func (m *MockRepo) TaskPriorityStats(ctx context.Context, workspaceID uuid.UUID) ([]sqlc.WorkspaceTaskPriorityStatsRow, error) {
	args := m.Called(ctx, workspaceID)
	return args.Get(0).([]sqlc.WorkspaceTaskPriorityStatsRow), args.Error(1)
}

func (m *MockRepo) WithTx(ctx context.Context, fn func(repo workspaces.Repository) error) error {
	return fn(m)
}

func TestWorkspacesService_Create(t *testing.T) {
	mockRepo := new(MockRepo)
	chartsSvc := workspaces.NewChartsService(mockRepo)
	svc := workspaces.NewService(mockRepo, chartsSvc)

	userID := uuid.New()
	orgID := uuid.New()
	req := workspaces.CreateWorkspaceRequest{
		Name:           "Core Engineering",
		Slug:           "core-engineering",
		OrganizationID: orgID,
	}

	t.Run("Creates workspace and assigns creator as OWNER", func(t *testing.T) {
		mockRepo.On("CheckSlugExists", mock.Anything, orgID, "core-engineering", (*uuid.UUID)(nil)).Return(false, nil).Once()

		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(p sqlc.CreateWorkspaceParams) bool {
			return p.Name == "Core Engineering" && p.Slug == "core-engineering" && p.OrganizationID == orgID
		})).Return(&sqlc.Workspace{
			ID:             uuid.New(),
			Name:           "Core Engineering",
			Slug:           "core-engineering",
			OrganizationID: orgID,
			Archive:        false,
			CreatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
			UpdatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
		}, nil).Once()

		mockRepo.On("CreateMember", mock.Anything, mock.MatchedBy(func(p sqlc.CreateWorkspaceMemberDirectParams) bool {
			return p.Role == sqlc.RoleOWNER && p.UserID == userID
		})).Return(&sqlc.WorkspaceMember{
			ID:   uuid.New(),
			Role: sqlc.RoleOWNER,
		}, nil).Once()

		res, err := svc.Create(context.Background(), req, userID)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "Core Engineering", res.Name)
		assert.Equal(t, "core-engineering", res.Slug)
		assert.False(t, res.Archive)
	})
}

func TestWorkspacesService_ChartsAndMetrics(t *testing.T) {
	mockRepo := new(MockRepo)
	chartsSvc := workspaces.NewChartsService(mockRepo)
	svc := workspaces.NewService(mockRepo, chartsSvc)

	wsID := uuid.New()
	userID := uuid.New()

	t.Run("Computes aggregated task percentage and stats across workspace", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, wsID).Return(&sqlc.GetWorkspaceByIDRow{
			ID:   wsID,
			Name: "Core Engineering",
		}, nil).Once()

		mockRepo.On("CountProjects", mock.Anything, wsID).Return(int64(10), nil).Once()
		mockRepo.On("CountProjectsByStatus", mock.Anything, wsID, sqlc.ProjectStatusACTIVE).Return(int64(7), nil).Once()
		mockRepo.On("CountProjectsByStatus", mock.Anything, wsID, sqlc.ProjectStatusCOMPLETED).Return(int64(3), nil).Once()
		mockRepo.On("CountTasks", mock.Anything, wsID).Return(int64(50), nil).Once()
		mockRepo.On("CountCompletedTasks", mock.Anything, wsID).Return(int64(25), nil).Once()
		mockRepo.On("CountOverdueTasks", mock.Anything, wsID).Return(int64(5), nil).Once()

		mockRepo.On("ProjectStatusStats", mock.Anything, wsID).Return([]sqlc.WorkspaceProjectStatusStatsRow{
			{Status: sqlc.ProjectStatusACTIVE, Count: 7},
			{Status: sqlc.ProjectStatusCOMPLETED, Count: 3},
		}, nil).Once()

		mockRepo.On("TaskPriorityStats", mock.Anything, wsID).Return([]sqlc.WorkspaceTaskPriorityStatsRow{
			{Priority: sqlc.TaskPriorityHIGH, Count: 15},
			{Priority: sqlc.TaskPriorityMEDIUM, Count: 35},
		}, nil).Once()

		res, err := svc.GetCharts(context.Background(), wsID, userID)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotNil(t, res.KPIMetrics)
		assert.Equal(t, int64(10), res.KPIMetrics.TotalProjects)
		assert.Equal(t, int64(50), res.KPIMetrics.TotalTasks)
		assert.Equal(t, int64(25), res.KPIMetrics.CompletedTasks)
		assert.Equal(t, 50.0, res.KPIMetrics.CompletionRate)
		assert.Len(t, res.ProjectStatus, 2)
		assert.Len(t, res.TaskPriority, 2)
	})
}

func TestWorkspacesService_ActiveAndArchived(t *testing.T) {
	mockRepo := new(MockRepo)
	chartsSvc := workspaces.NewChartsService(mockRepo)
	svc := workspaces.NewService(mockRepo, chartsSvc)

	orgID := uuid.New()
	userID := uuid.New()

	t.Run("Active directory queries exclude archived workspaces", func(t *testing.T) {
		mockRepo.On("ListForUser", mock.Anything, userID, &orgID, "").Return([]sqlc.ListWorkspacesForUserRow{
			{
				ID:             uuid.New(),
				Name:           "Active Workspace",
				OrganizationID: orgID,
				Archive:        false,
				CreatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
				UpdatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
			},
		}, nil).Once()

		res, err := svc.FindAll(context.Background(), userID, &orgID, "")
		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.False(t, res[0].Archive)
	})
}

func TestWorkspacesHandler_HTTP(t *testing.T) {
	mockRepo := new(MockRepo)
	chartsSvc := workspaces.NewChartsService(mockRepo)
	svc := workspaces.NewService(mockRepo, chartsSvc)
	handler := workspaces.NewHandler(svc)

	userID := uuid.New()
	orgID := uuid.New()
	authMw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := ctxhelper.WithUserID(r.Context(), userID.String())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	router := workspaces.Routes(handler, authMw, nil)

	t.Run("POST / creates workspace and returns 201", func(t *testing.T) {
		mockRepo.On("CheckSlugExists", mock.Anything, orgID, "product-dev", (*uuid.UUID)(nil)).Return(false, nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(&sqlc.Workspace{
			ID:             uuid.New(),
			Name:           "Product Dev",
			Slug:           "product-dev",
			OrganizationID: orgID,
			Archive:        false,
			CreatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
			UpdatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
		}, nil).Once()
		mockRepo.On("CreateMember", mock.Anything, mock.Anything).Return(&sqlc.WorkspaceMember{}, nil).Once()

		body := `{"name": "Product Dev", "slug": "product-dev", "organizationId": "` + orgID.String() + `"}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), `"slug":"product-dev"`)
	})
}
