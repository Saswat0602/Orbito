package workspace_members_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/modules/workspace_members"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) List(ctx context.Context, wsID *uuid.UUID, search string, limit, offset int32) ([]sqlc.ListWorkspaceMembersPaginatedRow, error) {
	args := m.Called(ctx, wsID, search, limit, offset)
	return args.Get(0).([]sqlc.ListWorkspaceMembersPaginatedRow), args.Error(1)
}

func (m *MockRepo) Count(ctx context.Context, wsID *uuid.UUID, search string) (int64, error) {
	args := m.Called(ctx, wsID, search)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRepo) GetByID(ctx context.Context, id uuid.UUID) (*sqlc.GetWorkspaceMemberDetailsByIDRow, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlc.GetWorkspaceMemberDetailsByIDRow), args.Error(1)
}

func (m *MockRepo) GetByUserAndWorkspace(ctx context.Context, userID, wsID uuid.UUID) (*sqlc.GetWorkspaceMemberByUserAndWorkspaceRow, error) {
	args := m.Called(ctx, userID, wsID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlc.GetWorkspaceMemberByUserAndWorkspaceRow), args.Error(1)
}

func (m *MockRepo) Create(ctx context.Context, params sqlc.CreateWorkspaceMemberDirectParams) (*sqlc.WorkspaceMember, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlc.WorkspaceMember), args.Error(1)
}

func (m *MockRepo) UpdateRole(ctx context.Context, id uuid.UUID, role sqlc.Role, updatedByID uuid.UUID) (*sqlc.WorkspaceMember, error) {
	args := m.Called(ctx, id, role, updatedByID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlc.WorkspaceMember), args.Error(1)
}

func (m *MockRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepo) GetUserGlobalRole(ctx context.Context, userID uuid.UUID) (sqlc.Role, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(sqlc.Role), args.Error(1)
}

func (m *MockRepo) WithTx(ctx context.Context, fn func(repo workspace_members.Repository) error) error {
	return fn(m)
}

func TestWorkspaceMembers_List(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := workspace_members.NewService(mockRepo)

	wsID := uuid.New()
	userID := uuid.New()

	t.Run("Returns members list with role and joinedAt timestamps", func(t *testing.T) {
		mockRepo.On("Count", mock.Anything, &wsID, "").Return(int64(1), nil).Once()
		mockRepo.On("List", mock.Anything, &wsID, "", int32(20), int32(0)).Return([]sqlc.ListWorkspaceMembersPaginatedRow{
			{
				ID:            uuid.New(),
				Role:          sqlc.RoleMEMBER,
				JoinedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
				UserID:        userID,
				WorkspaceID:   wsID,
				UserEmail:     "developer@example.com",
				UserFirstName: "Dev",
				UserLastName:  "User",
				UserStatus:    sqlc.UserStatusACTIVE,
				CreatedAt:     pgtype.Timestamp{Time: time.Now(), Valid: true},
				UpdatedAt:     pgtype.Timestamp{Time: time.Now(), Valid: true},
			},
		}, nil).Once()

		res, err := svc.FindAll(context.Background(), &wsID, "", 1, 20, userID)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, int64(1), res.TotalItems)
		assert.Len(t, res.Items, 1)
		assert.Equal(t, "MEMBER", res.Items[0].Role)
		assert.False(t, res.Items[0].JoinedAt.IsZero())
	})
}

func TestWorkspaceMembers_AddMember(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := workspace_members.NewService(mockRepo)

	wsID := uuid.New()
	requesterID := uuid.New()
	targetUserID := uuid.New()
	memberID := uuid.New()

	t.Run("Adds user to workspace with specified role", func(t *testing.T) {
		mockRepo.On("GetByUserAndWorkspace", mock.Anything, targetUserID, wsID).Return((*sqlc.GetWorkspaceMemberByUserAndWorkspaceRow)(nil), nil).Once()
		mockRepo.On("GetUserGlobalRole", mock.Anything, requesterID).Return(sqlc.RoleMEMBER, nil).Once()
		mockRepo.On("GetByUserAndWorkspace", mock.Anything, requesterID, wsID).Return(&sqlc.GetWorkspaceMemberByUserAndWorkspaceRow{
			Role: sqlc.RoleOWNER,
		}, nil).Once()

		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(p sqlc.CreateWorkspaceMemberDirectParams) bool {
			return p.Role == sqlc.RoleMEMBER && p.UserID == targetUserID && p.WorkspaceID == wsID
		})).Return(&sqlc.WorkspaceMember{
			ID:   memberID,
			Role: sqlc.RoleMEMBER,
		}, nil).Once()

		mockRepo.On("GetByID", mock.Anything, memberID).Return(&sqlc.GetWorkspaceMemberDetailsByIDRow{
			ID:          memberID,
			Role:        sqlc.RoleMEMBER,
			UserID:      targetUserID,
			WorkspaceID: wsID,
			UserEmail:   "target@example.com",
		}, nil).Once()

		res, err := svc.AddMember(context.Background(), workspace_members.AddWorkspaceMemberRequest{
			UserID:      targetUserID,
			WorkspaceID: wsID,
			Role:        sqlc.RoleMEMBER,
		}, requesterID)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "MEMBER", res.Role)
	})
}

func TestWorkspaceMembers_Remove(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := workspace_members.NewService(mockRepo)

	wsID := uuid.New()
	requesterID := uuid.New()
	memberID := uuid.New()
	targetUserID := uuid.New()

	t.Run("Removes user from workspace", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, memberID).Return(&sqlc.GetWorkspaceMemberDetailsByIDRow{
			ID:          memberID,
			UserID:      targetUserID,
			WorkspaceID: wsID,
			Role:        sqlc.RoleMEMBER,
		}, nil).Once()

		mockRepo.On("GetUserGlobalRole", mock.Anything, requesterID).Return(sqlc.RoleMEMBER, nil).Once()
		mockRepo.On("GetByUserAndWorkspace", mock.Anything, requesterID, wsID).Return(&sqlc.GetWorkspaceMemberByUserAndWorkspaceRow{
			Role: sqlc.RoleOWNER,
		}, nil).Once()

		mockRepo.On("Delete", mock.Anything, memberID).Return(nil).Once()

		err := svc.Remove(context.Background(), memberID, requesterID)
		assert.NoError(t, err)
	})
}

func TestWorkspaceMembers_HTTP(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := workspace_members.NewService(mockRepo)
	handler := workspace_members.NewHandler(svc)

	userID := uuid.New()
	wsID := uuid.New()
	authMw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := ctxhelper.WithUserID(r.Context(), userID.String())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	router := workspace_members.Routes(handler, authMw)

	t.Run("GET /workspace/:id returns 200 with members", func(t *testing.T) {
		mockRepo.On("Count", mock.Anything, &wsID, "").Return(int64(1), nil).Once()
		mockRepo.On("List", mock.Anything, &wsID, "", int32(20), int32(0)).Return([]sqlc.ListWorkspaceMembersPaginatedRow{
			{
				ID:          uuid.New(),
				Role:        sqlc.RoleMEMBER,
				UserID:      userID,
				WorkspaceID: wsID,
				UserEmail:   "test@example.com",
			},
		}, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/workspace/"+wsID.String(), nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"email":"test@example.com"`)
	})

	t.Run("DELETE /:id returns 204", func(t *testing.T) {
		memberID := uuid.New()
		mockRepo.On("GetByID", mock.Anything, memberID).Return(&sqlc.GetWorkspaceMemberDetailsByIDRow{
			ID:          memberID,
			UserID:      userID,
			WorkspaceID: wsID,
		}, nil).Once()
		mockRepo.On("GetUserGlobalRole", mock.Anything, userID).Return(sqlc.RoleMEMBER, nil).Once()
		mockRepo.On("GetByUserAndWorkspace", mock.Anything, userID, wsID).Return(&sqlc.GetWorkspaceMemberByUserAndWorkspaceRow{
			Role: sqlc.RoleOWNER,
		}, nil).Once()
		mockRepo.On("Delete", mock.Anything, memberID).Return(nil).Once()

		req := httptest.NewRequest(http.MethodDelete, "/"+memberID.String(), nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})
}
