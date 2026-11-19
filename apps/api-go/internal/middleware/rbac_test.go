package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRBACResolver struct {
	mock.Mock
}

func (m *MockRBACResolver) GetOrgMemberRole(ctx context.Context, userID, orgID uuid.UUID) (middleware.Role, error) {
	args := m.Called(ctx, userID, orgID)
	return args.Get(0).(middleware.Role), args.Error(1)
}

func (m *MockRBACResolver) GetOrgIDBySlug(ctx context.Context, slug string) (uuid.UUID, error) {
	args := m.Called(ctx, slug)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockRBACResolver) GetWorkspaceMemberRole(ctx context.Context, userID, workspaceID uuid.UUID) (middleware.Role, error) {
	args := m.Called(ctx, userID, workspaceID)
	return args.Get(0).(middleware.Role), args.Error(1)
}

func (m *MockRBACResolver) GetWorkspaceIDBySlug(ctx context.Context, slug string) (uuid.UUID, error) {
	args := m.Called(ctx, slug)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockRBACResolver) GetProjectMemberRole(ctx context.Context, userID, projectID uuid.UUID) (middleware.Role, error) {
	args := m.Called(ctx, userID, projectID)
	return args.Get(0).(middleware.Role), args.Error(1)
}

func (m *MockRBACResolver) GetProjectScopeInfo(ctx context.Context, projectID uuid.UUID) (*middleware.ProjectScopeInfo, error) {
	args := m.Called(ctx, projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*middleware.ProjectScopeInfo), args.Error(1)
}

func (m *MockRBACResolver) GetProjectScopeInfoBySlug(ctx context.Context, slug string) (*middleware.ProjectScopeInfo, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*middleware.ProjectScopeInfo), args.Error(1)
}

func TestRequireGlobalRole(t *testing.T) {
	handler := middleware.RequireGlobalRole(middleware.RoleManager)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("SuperAdmin always allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		ctx := ctxhelper.WithUserID(req.Context(), uuid.New().String())
		ctx = ctxhelper.WithUserRole(ctx, string(middleware.RoleSuperAdmin))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req.WithContext(ctx))
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Owner meets Manager role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		ctx := ctxhelper.WithUserID(req.Context(), uuid.New().String())
		ctx = ctxhelper.WithUserRole(ctx, string(middleware.RoleOwner))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req.WithContext(ctx))
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Member rejected for Manager role", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		ctx := ctxhelper.WithUserID(req.Context(), uuid.New().String())
		ctx = ctxhelper.WithUserRole(ctx, string(middleware.RoleMember))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req.WithContext(ctx))
		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "Insufficient privileges")
	})

	t.Run("Unauthenticated rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "Unauthenticated")
	})
}

func TestRequireScopeRole_Organization(t *testing.T) {
	mockRes := new(MockRBACResolver)
	userUUID := uuid.New()
	orgUUID := uuid.New()

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := ctxhelper.WithUserID(req.Context(), userUUID.String())
			ctx = ctxhelper.WithUserRole(ctx, string(middleware.RoleMember))
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})
	r.With(middleware.RequireOrgRole(mockRes, middleware.RoleManager, "id")).Get("/orgs/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("Sufficient role in org succeeds", func(t *testing.T) {
		mockRes.On("GetOrgMemberRole", mock.Anything, userUUID, orgUUID).Return(middleware.RoleOwner, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/orgs/"+orgUUID.String(), nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Insufficient role in org fails", func(t *testing.T) {
		mockRes.On("GetOrgMemberRole", mock.Anything, userUUID, orgUUID).Return(middleware.RoleMember, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/orgs/"+orgUUID.String(), nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "Insufficient role")
	})

	t.Run("Non-member fails", func(t *testing.T) {
		mockRes.On("GetOrgMemberRole", mock.Anything, userUUID, orgUUID).Return(middleware.Role(""), apperror.NotFound("not found")).Once()

		req := httptest.NewRequest(http.MethodGet, "/orgs/"+orgUUID.String(), nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Contains(t, rec.Body.String(), "Not a member of this scope")
	})
}

func TestRequireScopeRole_WorkspaceSlug(t *testing.T) {
	mockRes := new(MockRBACResolver)
	userUUID := uuid.New()
	wsUUID := uuid.New()
	wsSlug := "my-workspace"

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := ctxhelper.WithUserID(req.Context(), userUUID.String())
			ctx = ctxhelper.WithUserRole(ctx, string(middleware.RoleMember))
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})
	r.With(middleware.RequireWorkspaceRole(mockRes, middleware.RoleMember, "slug")).Get("/workspaces/{slug}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("Resolves slug to UUID and authorizes member", func(t *testing.T) {
		mockRes.On("GetWorkspaceIDBySlug", mock.Anything, wsSlug).Return(wsUUID, nil).Once()
		mockRes.On("GetWorkspaceMemberRole", mock.Anything, userUUID, wsUUID).Return(middleware.RoleMember, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/workspaces/"+wsSlug, nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Unknown slug returns 404", func(t *testing.T) {
		mockRes.On("GetWorkspaceIDBySlug", mock.Anything, "non-existent").Return(uuid.Nil, apperror.NotFound("not found")).Once()

		req := httptest.NewRequest(http.MethodGet, "/workspaces/non-existent", nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
		assert.Contains(t, rec.Body.String(), "Workspace not found")
	})
}

func TestRequireScopeRole_ProjectPublicAndInternal(t *testing.T) {
	mockRes := new(MockRBACResolver)
	userUUID := uuid.New()
	projUUID := uuid.New()
	wsUUID := uuid.New()

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := ctxhelper.WithUserID(req.Context(), userUUID.String())
			ctx = ctxhelper.WithUserRole(ctx, string(middleware.RoleViewer))
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	})
	projMW := middleware.RequireProjectRole(mockRes, middleware.RoleViewer, "id")
	r.With(projMW).Get("/projects/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.With(projMW).Post("/projects/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	t.Run("Public project permits GET to viewer", func(t *testing.T) {
		mockRes.On("GetProjectScopeInfo", mock.Anything, projUUID).Return(&middleware.ProjectScopeInfo{
			ID:          projUUID,
			Visibility:  "PUBLIC",
			WorkspaceID: wsUUID,
		}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/projects/"+projUUID.String(), nil)
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("Public project denies POST without membership", func(t *testing.T) {
		mockRes.On("GetProjectScopeInfo", mock.Anything, projUUID).Return(&middleware.ProjectScopeInfo{
			ID:          projUUID,
			Visibility:  "PUBLIC",
			WorkspaceID: wsUUID,
		}, nil).Once()
		mockRes.On("GetProjectMemberRole", mock.Anything, userUUID, projUUID).Return(middleware.Role(""), apperror.NotFound("not found")).Once()

		req := httptest.NewRequest(http.MethodPost, "/projects/"+projUUID.String(), strings.NewReader(`{}`))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("Internal project inherits role from workspace", func(t *testing.T) {
		mockRes.On("GetProjectScopeInfo", mock.Anything, projUUID).Return(&middleware.ProjectScopeInfo{
			ID:          projUUID,
			Visibility:  "INTERNAL",
			WorkspaceID: wsUUID,
		}, nil).Once()
		mockRes.On("GetProjectMemberRole", mock.Anything, userUUID, projUUID).Return(middleware.Role(""), apperror.NotFound("not found")).Once()
		mockRes.On("GetWorkspaceMemberRole", mock.Anything, userUUID, wsUUID).Return(middleware.RoleMember, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/projects/"+projUUID.String(), strings.NewReader(`{}`))
		rec := httptest.NewRecorder()

		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)
	})
}
