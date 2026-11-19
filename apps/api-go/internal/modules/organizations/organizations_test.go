package organizations_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/modules/organizations"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) Create(ctx context.Context, params sqlc.CreateOrganizationParams) (*sqlc.Organization, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlc.Organization), args.Error(1)
}

func (m *MockRepo) GetDetailsByID(ctx context.Context, id uuid.UUID) (*sqlc.GetOrganizationDetailsByIDRow, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlc.GetOrganizationDetailsByIDRow), args.Error(1)
}

func (m *MockRepo) GetBySlug(ctx context.Context, slug string) (*sqlc.GetOrganizationBySlugRow, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlc.GetOrganizationBySlugRow), args.Error(1)
}

func (m *MockRepo) ListForUser(ctx context.Context, userID uuid.UUID) ([]sqlc.ListOrganizationsForUserRow, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]sqlc.ListOrganizationsForUserRow), args.Error(1)
}

func (m *MockRepo) Update(ctx context.Context, params sqlc.UpdateOrganizationParams) (*sqlc.Organization, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlc.Organization), args.Error(1)
}

func (m *MockRepo) Archive(ctx context.Context, id, userID uuid.UUID) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepo) TransferOwner(ctx context.Context, id, newOwnerID, updatedByID uuid.UUID) error {
	args := m.Called(ctx, id, newOwnerID, updatedByID)
	return args.Error(0)
}

func (m *MockRepo) CreateMember(ctx context.Context, params sqlc.CreateOrganizationMemberParams) (*sqlc.OrganizationMember, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlc.OrganizationMember), args.Error(1)
}

func (m *MockRepo) GetMember(ctx context.Context, userID, orgID uuid.UUID) (*sqlc.OrganizationMember, error) {
	args := m.Called(ctx, userID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlc.OrganizationMember), args.Error(1)
}

func (m *MockRepo) UpdateMemberRole(ctx context.Context, orgID, userID uuid.UUID, role sqlc.Role, updatedByID uuid.UUID) error {
	args := m.Called(ctx, orgID, userID, role, updatedByID)
	return args.Error(0)
}

func (m *MockRepo) CheckSlugExists(ctx context.Context, slug string, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, slug, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepo) SetUserDefaultOrg(ctx context.Context, userID, orgID uuid.UUID) error {
	args := m.Called(ctx, userID, orgID)
	return args.Error(0)
}

func (m *MockRepo) ClearUserDefaultOrg(ctx context.Context, orgID uuid.UUID) error {
	args := m.Called(ctx, orgID)
	return args.Error(0)
}

func (m *MockRepo) CreateDefaultWorkflow(ctx context.Context, orgID, userID uuid.UUID) (*sqlc.Workflow, error) {
	args := m.Called(ctx, orgID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlc.Workflow), args.Error(1)
}

func (m *MockRepo) GetStats(ctx context.Context, orgID uuid.UUID) (*organizations.OrganizationStatsResponse, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*organizations.OrganizationStatsResponse), args.Error(1)
}

func (m *MockRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*sqlc.GetUserByIDRow, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlc.GetUserByIDRow), args.Error(1)
}

func (m *MockRepo) WithTx(ctx context.Context, fn func(repo organizations.Repository) error) error {
	return fn(m)
}

func TestOrganizationsService_Create(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := organizations.NewService(mockRepo)

	userID := uuid.New()
	req := organizations.CreateOrganizationRequest{
		Name: "Acme Corp",
	}

	t.Run("Creates organization with default workflow and owner member", func(t *testing.T) {
		mockRepo.On("CheckSlugExists", mock.Anything, "acme-corp", (*uuid.UUID)(nil)).Return(false, nil).Once()
		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(p sqlc.CreateOrganizationParams) bool {
			return p.Name == "Acme Corp" && p.Slug == "acme-corp" && p.OwnerID == userID
		})).Return(&sqlc.Organization{
			ID:        uuid.New(),
			Name:      "Acme Corp",
			Slug:      "acme-corp",
			OwnerID:   userID,
			CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
			UpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		}, nil).Once()

		mockRepo.On("CreateMember", mock.Anything, mock.MatchedBy(func(p sqlc.CreateOrganizationMemberParams) bool {
			return p.Role == sqlc.RoleOWNER && p.UserID == userID
		})).Return(&sqlc.OrganizationMember{
			ID:   uuid.New(),
			Role: sqlc.RoleOWNER,
		}, nil).Once()

		mockRepo.On("SetUserDefaultOrg", mock.Anything, userID, mock.Anything).Return(nil).Once()
		mockRepo.On("CreateDefaultWorkflow", mock.Anything, mock.Anything, userID).Return(&sqlc.Workflow{}, nil).Once()

		res, err := svc.Create(context.Background(), req, userID)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "Acme Corp", res.Name)
		assert.Equal(t, "acme-corp", res.Slug)
		assert.Equal(t, userID, res.OwnerID)
	})

	t.Run("Appends suffix when slug already exists", func(t *testing.T) {
		mockRepo.On("CheckSlugExists", mock.Anything, "acme-corp", (*uuid.UUID)(nil)).Return(true, nil).Once()
		mockRepo.On("CheckSlugExists", mock.Anything, "acme-corp-1", (*uuid.UUID)(nil)).Return(false, nil).Once()

		mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(p sqlc.CreateOrganizationParams) bool {
			return p.Slug == "acme-corp-1"
		})).Return(&sqlc.Organization{
			ID:        uuid.New(),
			Name:      "Acme Corp",
			Slug:      "acme-corp-1",
			OwnerID:   userID,
			CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
			UpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		}, nil).Once()

		mockRepo.On("CreateMember", mock.Anything, mock.Anything).Return(&sqlc.OrganizationMember{}, nil).Once()
		mockRepo.On("SetUserDefaultOrg", mock.Anything, userID, mock.Anything).Return(nil).Once()
		mockRepo.On("CreateDefaultWorkflow", mock.Anything, mock.Anything, userID).Return(&sqlc.Workflow{}, nil).Once()

		res, err := svc.Create(context.Background(), req, userID)
		assert.NoError(t, err)
		assert.Equal(t, "acme-corp-1", res.Slug)
	})
}

func TestOrganizationsService_TransferOwnership(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := organizations.NewService(mockRepo)

	currentOwnerID := uuid.New()
	newOwnerID := uuid.New()
	orgID := uuid.New()

	t.Run("Transfers ownership atomically within transaction", func(t *testing.T) {
		mockRepo.On("GetDetailsByID", mock.Anything, orgID).Return(&sqlc.GetOrganizationDetailsByIDRow{
			ID:      orgID,
			Name:    "Acme Corp",
			OwnerID: currentOwnerID,
		}, nil).Once()

		mockRepo.On("GetUserByID", mock.Anything, newOwnerID).Return(&sqlc.GetUserByIDRow{
			ID:        newOwnerID,
			Status:    sqlc.UserStatusACTIVE,
			DeletedAt: pgtype.Timestamp{Valid: false},
		}, nil).Once()

		// New owner is existing member with MEMBER role
		mockRepo.On("GetMember", mock.Anything, newOwnerID, orgID).Return(&sqlc.OrganizationMember{
			ID:   uuid.New(),
			Role: sqlc.RoleMEMBER,
		}, nil).Once()

		mockRepo.On("UpdateMemberRole", mock.Anything, orgID, newOwnerID, sqlc.RoleOWNER, currentOwnerID).Return(nil).Once()

		// Previous owner demoted to MANAGER
		mockRepo.On("GetMember", mock.Anything, currentOwnerID, orgID).Return(&sqlc.OrganizationMember{
			ID:   uuid.New(),
			Role: sqlc.RoleOWNER,
		}, nil).Once()

		mockRepo.On("UpdateMemberRole", mock.Anything, orgID, currentOwnerID, sqlc.RoleMANAGER, currentOwnerID).Return(nil).Once()
		mockRepo.On("TransferOwner", mock.Anything, orgID, newOwnerID, currentOwnerID).Return(nil).Once()

		// Final reload
		mockRepo.On("GetDetailsByID", mock.Anything, orgID).Return(&sqlc.GetOrganizationDetailsByIDRow{
			ID:          orgID,
			Name:        "Acme Corp",
			OwnerID:     newOwnerID,
			OwnerUserID: newOwnerID,
		}, nil).Once()

		res, err := svc.TransferOwnership(context.Background(), orgID, newOwnerID, currentOwnerID)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, newOwnerID, res.OwnerID)
	})
}

func TestOrganizationsHandler_HTTP(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := organizations.NewService(mockRepo)
	handler := organizations.NewHandler(svc)

	userID := uuid.New()
	authMw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := ctxhelper.WithUserID(r.Context(), userID.String())
			ctx = ctxhelper.WithUserRole(ctx, string(middleware.RoleOwner))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	router := organizations.Routes(handler, authMw, nil)

	t.Run("POST / creates organization", func(t *testing.T) {
		mockRepo.On("CheckSlugExists", mock.Anything, "my-org", (*uuid.UUID)(nil)).Return(false, nil).Once()
		mockRepo.On("Create", mock.Anything, mock.Anything).Return(&sqlc.Organization{
			ID:        uuid.New(),
			Name:      "My Org",
			Slug:      "my-org",
			OwnerID:   userID,
			CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
			UpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		}, nil).Once()
		mockRepo.On("CreateMember", mock.Anything, mock.Anything).Return(&sqlc.OrganizationMember{}, nil).Once()
		mockRepo.On("SetUserDefaultOrg", mock.Anything, userID, mock.Anything).Return(nil).Once()
		mockRepo.On("CreateDefaultWorkflow", mock.Anything, mock.Anything, userID).Return(&sqlc.Workflow{}, nil).Once()

		body := `{"name": "My Org"}`
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Contains(t, rec.Body.String(), `"slug":"my-org"`)
	})
}
