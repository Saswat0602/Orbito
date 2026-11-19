package organization_members_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/modules/organization_members"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) List(ctx context.Context, orgID uuid.UUID, search string) ([]sqlc.ListOrganizationMembersRow, error) {
	args := m.Called(ctx, orgID, search)
	return args.Get(0).([]sqlc.ListOrganizationMembersRow), args.Error(1)
}

func (m *MockRepo) ListBySlug(ctx context.Context, slug, search string, limit, offset int32) ([]sqlc.ListOrganizationMembersBySlugRow, error) {
	args := m.Called(ctx, slug, search, limit, offset)
	return args.Get(0).([]sqlc.ListOrganizationMembersBySlugRow), args.Error(1)
}

func (m *MockRepo) CountBySlug(ctx context.Context, slug, search string) (int64, error) {
	args := m.Called(ctx, slug, search)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRepo) GetByID(ctx context.Context, id uuid.UUID) (*sqlc.GetOrganizationMemberByIDRow, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlc.GetOrganizationMemberByIDRow), args.Error(1)
}

func (m *MockRepo) GetByUserAndOrg(ctx context.Context, userID, orgID uuid.UUID) (*sqlc.GetOrganizationMemberByUserAndOrgRow, error) {
	args := m.Called(ctx, userID, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlc.GetOrganizationMemberByUserAndOrgRow), args.Error(1)
}

func (m *MockRepo) Create(ctx context.Context, params sqlc.CreateOrganizationMemberParams) (*sqlc.OrganizationMember, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlc.OrganizationMember), args.Error(1)
}

func (m *MockRepo) UpdateRole(ctx context.Context, id uuid.UUID, role sqlc.Role, updatedByID uuid.UUID) (*sqlc.OrganizationMember, error) {
	args := m.Called(ctx, id, role, updatedByID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqlc.OrganizationMember), args.Error(1)
}

func (m *MockRepo) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepo) SetDefaultOrg(ctx context.Context, userID, orgID uuid.UUID) error {
	args := m.Called(ctx, userID, orgID)
	return args.Error(0)
}

func (m *MockRepo) ClearDefaultOrg(ctx context.Context, userID, orgID uuid.UUID) error {
	args := m.Called(ctx, userID, orgID)
	return args.Error(0)
}

func (m *MockRepo) GetUserGlobalRole(ctx context.Context, userID uuid.UUID) (sqlc.Role, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(sqlc.Role), args.Error(1)
}

func (m *MockRepo) WithTx(ctx context.Context, fn func(repo organization_members.Repository) error) error {
	return fn(m)
}

func TestOrganizationMembers_List(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := organization_members.NewService(mockRepo)

	orgID := uuid.New()
	userID := uuid.New()

	mockRepo.On("List", mock.Anything, orgID, "").Return([]sqlc.ListOrganizationMembersRow{
		{
			ID:             uuid.New(),
			Role:           sqlc.RoleMEMBER,
			JoinedAt:       pgtype.Timestamp{Time: time.Now(), Valid: true},
			UserID:         userID,
			OrganizationID: orgID,
			UserEmail:      "member@example.com",
			UserFirstName:  "Jane",
			UserLastName:   "Doe",
			UserStatus:     sqlc.UserStatusACTIVE,
		},
	}, nil).Once()

	res, err := svc.FindAll(context.Background(), &orgID, "", userID)
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "member@example.com", res[0].User.Email)
}

func TestOrganizationMembers_UpdateRole(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := organization_members.NewService(mockRepo)

	orgID := uuid.New()
	ownerID := uuid.New()
	memberID := uuid.New()
	membershipID := uuid.New()

	t.Run("Owner successfully updates member role to MANAGER", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, membershipID).Return(&sqlc.GetOrganizationMemberByIDRow{
			ID:             membershipID,
			UserID:         memberID,
			OrganizationID: orgID,
			Role:           sqlc.RoleMEMBER,
			OrgOwnerID:     ownerID,
		}, nil).Once()

		mockRepo.On("GetUserGlobalRole", mock.Anything, ownerID).Return(sqlc.RoleMEMBER, nil).Once()
		mockRepo.On("GetByUserAndOrg", mock.Anything, ownerID, orgID).Return(&sqlc.GetOrganizationMemberByUserAndOrgRow{
			Role: sqlc.RoleOWNER,
		}, nil).Once()

		mockRepo.On("UpdateRole", mock.Anything, membershipID, sqlc.RoleMANAGER, ownerID).Return(&sqlc.OrganizationMember{
			ID:   membershipID,
			Role: sqlc.RoleMANAGER,
		}, nil).Once()

		// Reload for response
		mockRepo.On("GetByID", mock.Anything, membershipID).Return(&sqlc.GetOrganizationMemberByIDRow{
			ID:             membershipID,
			UserID:         memberID,
			OrganizationID: orgID,
			Role:           sqlc.RoleMANAGER,
			OrgOwnerID:     ownerID,
		}, nil).Once()

		res, err := svc.UpdateRole(context.Background(), membershipID, sqlc.RoleMANAGER, ownerID)
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "MANAGER", res.Role)
	})

	t.Run("Cannot change the role of organization owner", func(t *testing.T) {
		ownerMembershipID := uuid.New()
		mockRepo.On("GetByID", mock.Anything, ownerMembershipID).Return(&sqlc.GetOrganizationMemberByIDRow{
			ID:             ownerMembershipID,
			UserID:         ownerID,
			OrganizationID: orgID,
			Role:           sqlc.RoleOWNER,
			OrgOwnerID:     ownerID,
		}, nil).Once()

		mockRepo.On("GetUserGlobalRole", mock.Anything, ownerID).Return(sqlc.RoleMEMBER, nil).Once()
		mockRepo.On("GetByUserAndOrg", mock.Anything, ownerID, orgID).Return(&sqlc.GetOrganizationMemberByUserAndOrgRow{
			Role: sqlc.RoleOWNER,
		}, nil).Once()

		_, err := svc.UpdateRole(context.Background(), ownerMembershipID, sqlc.RoleMEMBER, ownerID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot change the role of organization owner")
	})
}

func TestOrganizationMembers_Remove(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := organization_members.NewService(mockRepo)

	orgID := uuid.New()
	ownerID := uuid.New()
	targetMemberID := uuid.New()
	membershipID := uuid.New()

	t.Run("Prevents removing the sole OWNER of an organization", func(t *testing.T) {
		ownerMembershipID := uuid.New()
		mockRepo.On("GetByID", mock.Anything, ownerMembershipID).Return(&sqlc.GetOrganizationMemberByIDRow{
			ID:             ownerMembershipID,
			UserID:         ownerID,
			OrganizationID: orgID,
			OrgOwnerID:     ownerID,
		}, nil).Once()

		err := svc.Remove(context.Background(), ownerMembershipID, ownerID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot remove organization owner")
	})

	t.Run("Owner removes normal member successfully", func(t *testing.T) {
		mockRepo.On("GetByID", mock.Anything, membershipID).Return(&sqlc.GetOrganizationMemberByIDRow{
			ID:             membershipID,
			UserID:         targetMemberID,
			OrganizationID: orgID,
			OrgOwnerID:     ownerID,
		}, nil).Once()

		mockRepo.On("GetUserGlobalRole", mock.Anything, ownerID).Return(sqlc.RoleMEMBER, nil).Once()
		mockRepo.On("GetByUserAndOrg", mock.Anything, ownerID, orgID).Return(&sqlc.GetOrganizationMemberByUserAndOrgRow{
			Role: sqlc.RoleOWNER,
		}, nil).Once()

		mockRepo.On("ClearDefaultOrg", mock.Anything, targetMemberID, orgID).Return(nil).Once()
		mockRepo.On("Delete", mock.Anything, membershipID).Return(nil).Once()

		err := svc.Remove(context.Background(), membershipID, ownerID)
		assert.NoError(t, err)
	})
}

func TestOrganizationMembers_HTTP(t *testing.T) {
	mockRepo := new(MockRepo)
	svc := organization_members.NewService(mockRepo)
	handler := organization_members.NewHandler(svc)

	userID := uuid.New()
	authMw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := ctxhelper.WithUserID(r.Context(), userID.String())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	router := organization_members.Routes(handler, authMw)

	t.Run("PATCH /:id updates role", func(t *testing.T) {
		membershipID := uuid.New()
		orgID := uuid.New()

		mockRepo.On("GetByID", mock.Anything, membershipID).Return(&sqlc.GetOrganizationMemberByIDRow{
			ID:             membershipID,
			UserID:         uuid.New(),
			OrganizationID: orgID,
			Role:           sqlc.RoleMEMBER,
			OrgOwnerID:     userID,
		}, nil).Once()

		mockRepo.On("GetUserGlobalRole", mock.Anything, userID).Return(sqlc.RoleMEMBER, nil).Once()
		mockRepo.On("GetByUserAndOrg", mock.Anything, userID, orgID).Return(&sqlc.GetOrganizationMemberByUserAndOrgRow{
			Role: sqlc.RoleOWNER,
		}, nil).Once()

		mockRepo.On("UpdateRole", mock.Anything, membershipID, sqlc.RoleMANAGER, userID).Return(&sqlc.OrganizationMember{
			ID:   membershipID,
			Role: sqlc.RoleMANAGER,
		}, nil).Once()

		mockRepo.On("GetByID", mock.Anything, membershipID).Return(&sqlc.GetOrganizationMemberByIDRow{
			ID:             membershipID,
			UserID:         uuid.New(),
			OrganizationID: orgID,
			Role:           sqlc.RoleMANAGER,
			OrgOwnerID:     userID,
		}, nil).Once()

		req := httptest.NewRequest(http.MethodPatch, "/"+membershipID.String(), strings.NewReader(`{"role":"MANAGER"}`))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), `"role":"MANAGER"`)
	})
}
