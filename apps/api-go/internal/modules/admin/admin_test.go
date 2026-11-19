package admin

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/config"
	"github.com/orbito/api-go/internal/modules/email"
	"github.com/orbito/api-go/internal/modules/settings"
	"github.com/orbito/api-go/internal/platform/crypto"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/platform/mailer"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
)

type mockAdminRepo struct {
	users         map[uuid.UUID]*sqlc.User
	organizations map[uuid.UUID]*sqlc.Organization
	memberships   map[uuid.UUID]*sqlc.OrganizationMember
}

func newMockAdminRepo() *mockAdminRepo {
	return &mockAdminRepo{
		users:         make(map[uuid.UUID]*sqlc.User),
		organizations: make(map[uuid.UUID]*sqlc.Organization),
		memberships:   make(map[uuid.UUID]*sqlc.OrganizationMember),
	}
}

func (m *mockAdminRepo) GetDashboardStats(_ context.Context, _ time.Time) (*sqlc.GetDashboardStatsCountsRow, error) {
	return &sqlc.GetDashboardStatsCountsRow{
		TotalUsers:          int64(len(m.users)),
		TotalOrganizations:  int64(len(m.organizations)),
		TotalWorkspaces:     2,
		TotalProjects:       5,
		TotalTasks:          12,
		NewUsersThisWeek:    int64(len(m.users)),
		NewOrgsThisWeek:     int64(len(m.organizations)),
		NewProjectsThisWeek: 3,
		NewTasksThisWeek:    8,
		ActiveUsers:         int64(len(m.users)),
	}, nil
}

func (m *mockAdminRepo) ListUsers(_ context.Context, search string, _ sqlc.UserStatus, _ sqlc.Role, limit, offset int32) ([]sqlc.AdminListUsersRow, error) {
	var rows []sqlc.AdminListUsersRow
	for _, u := range m.users {
		if !u.DeletedAt.Valid {
			if search != "" && !strings.Contains(strings.ToLower(u.Email), strings.ToLower(search)) {
				continue
			}
			rows = append(rows, sqlc.AdminListUsersRow{
				ID:                u.ID,
				Email:             u.Email,
				FirstName:         u.FirstName,
				LastName:          u.LastName,
				Avatar:            u.Avatar,
				Role:              u.Role,
				Status:            u.Status,
				LastLoginAt:       u.LastLoginAt,
				CreatedAt:         u.CreatedAt,
				OrganizationCount: 1,
			})
		}
	}
	if int(offset) >= len(rows) {
		return []sqlc.AdminListUsersRow{}, nil
	}
	end := int(offset + limit)
	if end > len(rows) {
		end = len(rows)
	}
	return rows[offset:end], nil
}

func (m *mockAdminRepo) CountUsers(_ context.Context, search string, _ sqlc.UserStatus, _ sqlc.Role) (int64, error) {
	var count int64
	for _, u := range m.users {
		if !u.DeletedAt.Valid {
			if search != "" && !strings.Contains(strings.ToLower(u.Email), strings.ToLower(search)) {
				continue
			}
			count++
		}
	}
	return count, nil
}

func (m *mockAdminRepo) GetUserDetail(_ context.Context, id uuid.UUID) (*sqlc.AdminGetUserDetailRow, error) {
	u, ok := m.users[id]
	if !ok || u.DeletedAt.Valid {
		return nil, nil
	}
	return &sqlc.AdminGetUserDetailRow{
		ID:            u.ID,
		Email:         u.Email,
		FirstName:     u.FirstName,
		LastName:      u.LastName,
		Avatar:        u.Avatar,
		Bio:           u.Bio,
		Timezone:      u.Timezone,
		Language:      u.Language,
		Role:          u.Role,
		Status:        u.Status,
		LastLoginAt:   u.LastLoginAt,
		EmailVerified: false,
		CreatedAt:     u.CreatedAt,
		UpdatedAt:     u.UpdatedAt,
		Source:        sqlc.UserSourceMANUAL,
	}, nil
}

func (m *mockAdminRepo) GetUserOrganizations(_ context.Context, userID uuid.UUID) ([]sqlc.AdminGetUserOrganizationsRow, error) {
	var rows []sqlc.AdminGetUserOrganizationsRow
	for _, mem := range m.memberships {
		if mem.UserID == userID {
			if org, ok := m.organizations[mem.OrganizationID]; ok {
				rows = append(rows, sqlc.AdminGetUserOrganizationsRow{
					ID:              org.ID,
					Name:            org.Name,
					Slug:            org.Slug,
					Role:            mem.Role,
					MembersCount:    1,
					WorkspacesCount: 1,
				})
			}
		}
	}
	return rows, nil
}

func (m *mockAdminRepo) CountSuperAdmins(_ context.Context) (int64, error) {
	var count int64
	for _, u := range m.users {
		if u.Role == sqlc.RoleSUPERADMIN && !u.DeletedAt.Valid {
			count++
		}
	}
	return count, nil
}

func (m *mockAdminRepo) UpdateUserRole(_ context.Context, id uuid.UUID, role sqlc.Role) (*sqlc.AdminUpdateUserRoleRow, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, nil
	}
	u.Role = role
	return &sqlc.AdminUpdateUserRoleRow{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Role:      u.Role,
	}, nil
}

func (m *mockAdminRepo) UpdateUserStatus(_ context.Context, id uuid.UUID, status sqlc.UserStatus) (*sqlc.AdminUpdateUserStatusRow, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, nil
	}
	u.Status = status
	return &sqlc.AdminUpdateUserStatusRow{
		ID:        u.ID,
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Status:    u.Status,
	}, nil
}

func (m *mockAdminRepo) SoftDeleteUser(_ context.Context, id, deletedBy uuid.UUID) error {
	u, ok := m.users[id]
	if ok {
		u.Status = sqlc.UserStatusINACTIVE
		u.DeletedAt = pgtype.Timestamp{Time: time.Now(), Valid: true}
		u.DeletedBy = pgtype.UUID{Bytes: deletedBy, Valid: true}
	}
	return nil
}

func (m *mockAdminRepo) RemoveUserMemberships(_ context.Context, userID uuid.UUID) error {
	for k, mem := range m.memberships {
		if mem.UserID == userID {
			delete(m.memberships, k)
		}
	}
	return nil
}

func (m *mockAdminRepo) GetUserOwnedOrganizations(_ context.Context, ownerID uuid.UUID) ([]sqlc.AdminGetUserOwnedOrganizationsRow, error) {
	var rows []sqlc.AdminGetUserOwnedOrganizationsRow
	for _, o := range m.organizations {
		if o.OwnerID == ownerID {
			rows = append(rows, sqlc.AdminGetUserOwnedOrganizationsRow{
				ID:   o.ID,
				Name: o.Name,
			})
		}
	}
	return rows, nil
}

func (m *mockAdminRepo) SetUserResetToken(_ context.Context, id uuid.UUID, token pgtype.Text, expiry pgtype.Timestamp) error {
	if u, ok := m.users[id]; ok {
		u.ResetToken = token
		u.ResetTokenExpiry = expiry
	}
	return nil
}

func (m *mockAdminRepo) ListOrganizations(_ context.Context, search string, limit, offset int32) ([]sqlc.AdminListOrganizationsRow, error) {
	var rows []sqlc.AdminListOrganizationsRow
	for _, o := range m.organizations {
		if search != "" && !strings.Contains(strings.ToLower(o.Name), strings.ToLower(search)) {
			continue
		}
		var ownerEmail, ownerFN, ownerLN pgtype.Text
		if u, ok := m.users[o.OwnerID]; ok {
			ownerEmail = pgtype.Text{String: u.Email, Valid: true}
			ownerFN = pgtype.Text{String: u.FirstName, Valid: true}
			ownerLN = pgtype.Text{String: u.LastName, Valid: true}
		}
		rows = append(rows, sqlc.AdminListOrganizationsRow{
			ID:              o.ID,
			Name:            o.Name,
			Slug:            o.Slug,
			Description:     o.Description,
			Archive:         o.Archive,
			CreatedAt:       o.CreatedAt,
			OwnerID:         pgtype.UUID{Bytes: o.OwnerID, Valid: true},
			OwnerFirstName:  ownerFN,
			OwnerLastName:   ownerLN,
			OwnerEmail:      ownerEmail,
			MembersCount:    1,
			WorkspacesCount: 1,
		})
	}
	if int(offset) >= len(rows) {
		return []sqlc.AdminListOrganizationsRow{}, nil
	}
	end := int(offset + limit)
	if end > len(rows) {
		end = len(rows)
	}
	return rows[offset:end], nil
}

func (m *mockAdminRepo) CountOrganizations(_ context.Context, search string) (int64, error) {
	var count int64
	for _, o := range m.organizations {
		if search != "" && !strings.Contains(strings.ToLower(o.Name), strings.ToLower(search)) {
			continue
		}
		count++
	}
	return count, nil
}

func (m *mockAdminRepo) GetOrganizationDetail(_ context.Context, id uuid.UUID) (*sqlc.AdminGetOrganizationDetailRow, error) {
	o, ok := m.organizations[id]
	if !ok {
		return nil, nil
	}
	var ownerEmail, ownerFN, ownerLN pgtype.Text
	if u, ok := m.users[o.OwnerID]; ok {
		ownerEmail = pgtype.Text{String: u.Email, Valid: true}
		ownerFN = pgtype.Text{String: u.FirstName, Valid: true}
		ownerLN = pgtype.Text{String: u.LastName, Valid: true}
	}
	return &sqlc.AdminGetOrganizationDetailRow{
		ID:              o.ID,
		Name:            o.Name,
		Slug:            o.Slug,
		Description:     o.Description,
		Archive:         o.Archive,
		CreatedAt:       o.CreatedAt,
		UpdatedAt:       o.UpdatedAt,
		OwnerID:         pgtype.UUID{Bytes: o.OwnerID, Valid: true},
		OwnerFirstName:  ownerFN,
		OwnerLastName:   ownerLN,
		OwnerEmail:      ownerEmail,
		MembersCount:    1,
		WorkspacesCount: 1,
	}, nil
}

func (m *mockAdminRepo) GetOrganizationMembers(_ context.Context, orgID uuid.UUID) ([]sqlc.AdminGetOrganizationMembersRow, error) {
	var rows []sqlc.AdminGetOrganizationMembersRow
	for _, mem := range m.memberships {
		if mem.OrganizationID == orgID {
			if u, ok := m.users[mem.UserID]; ok {
				rows = append(rows, sqlc.AdminGetOrganizationMembersRow{
					Role:      mem.Role,
					JoinedAt:  mem.JoinedAt,
					UserID:    u.ID,
					FirstName: u.FirstName,
					LastName:  u.LastName,
					Email:     u.Email,
					Avatar:    u.Avatar,
					Status:    u.Status,
				})
			}
		}
	}
	return rows, nil
}

func (m *mockAdminRepo) GetOrganizationWorkspaces(_ context.Context, _ uuid.UUID) ([]sqlc.AdminGetOrganizationWorkspacesRow, error) {
	return []sqlc.AdminGetOrganizationWorkspacesRow{
		{
			ID:            uuid.New(),
			Name:          "Default Workspace",
			Slug:          "default-ws",
			ProjectsCount: 3,
			MembersCount:  2,
		},
	}, nil
}

func (m *mockAdminRepo) ToggleOrganizationArchive(_ context.Context, id uuid.UUID) (*sqlc.AdminToggleOrganizationArchiveRow, error) {
	o, ok := m.organizations[id]
	if !ok {
		return nil, nil
	}
	o.Archive = !o.Archive
	return &sqlc.AdminToggleOrganizationArchiveRow{
		ID:      o.ID,
		Name:    o.Name,
		Archive: o.Archive,
	}, nil
}

func (m *mockAdminRepo) DeleteOrganization(_ context.Context, id uuid.UUID) error {
	delete(m.organizations, id)
	return nil
}

func (m *mockAdminRepo) ClearUsersDefaultOrg(_ context.Context, _ uuid.UUID) error {
	return nil
}

func (m *mockAdminRepo) TransferOrgOwnership(_ context.Context, orgID, newOwnerID uuid.UUID) (*sqlc.AdminTransferOrgOwnershipRow, error) {
	o, ok := m.organizations[orgID]
	if !ok {
		return nil, nil
	}
	o.OwnerID = newOwnerID
	return &sqlc.AdminTransferOrgOwnershipRow{
		ID:      o.ID,
		Name:    o.Name,
		OwnerID: o.OwnerID,
	}, nil
}

func (m *mockAdminRepo) GetOrganizationMember(_ context.Context, userID, orgID uuid.UUID) (*sqlc.AdminGetOrganizationMemberRow, error) {
	for _, mem := range m.memberships {
		if mem.UserID == userID && mem.OrganizationID == orgID {
			return &sqlc.AdminGetOrganizationMemberRow{
				ID:   mem.ID,
				Role: mem.Role,
			}, nil
		}
	}
	return nil, nil
}

func (m *mockAdminRepo) CreateOrganizationMember(_ context.Context, id uuid.UUID, role sqlc.Role, userID, orgID, createdByID uuid.UUID) error {
	m.memberships[id] = &sqlc.OrganizationMember{
		ID:             id,
		Role:           role,
		UserID:         userID,
		OrganizationID: orgID,
		JoinedAt:       pgtype.Timestamp{Time: time.Now(), Valid: true},
	}
	return nil
}

func (m *mockAdminRepo) UpdateOrganizationMemberRole(_ context.Context, id uuid.UUID, role sqlc.Role) error {
	if mem, ok := m.memberships[id]; ok {
		mem.Role = role
	}
	return nil
}

func (m *mockAdminRepo) GetUserByID(_ context.Context, id uuid.UUID) (*sqlc.User, error) {
	u, ok := m.users[id]
	if !ok || u.DeletedAt.Valid {
		return nil, nil
	}
	return u, nil
}

func TestAdminModule(t *testing.T) {
	repo := newMockAdminRepo()
	fieldCrypto, _ := crypto.NewFieldCrypto("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")
	settingsSvc := settings.NewService(nil, fieldCrypto)
	mailerClient := mailer.New(mailer.Config{Host: "smtp.example.com", Port: 587})
	emailSvc := email.NewService(mailerClient, 2)
	cfg := &config.Config{
		JWTSecret:   "secret",
		SMTPHost:    "smtp.example.com",
		SMTPPort:    587,
		SMTPUser:    "admin",
		SMTPPass:    "pass",
		SMTPFrom:    "admin@orbito.io",
		CORSOrigins: []string{"http://localhost:3000"},
	}

	adminSvc := NewService(repo, settingsSvc, emailSvc, cfg)
	adminHandler := NewHandler(adminSvc)
	router := Routes(adminHandler, nil)

	// Seed users
	superAdminID := uuid.New()
	superAdmin := &sqlc.User{
		ID:        superAdminID,
		Email:     "superadmin@orbito.io",
		FirstName: "Super",
		LastName:  "Admin",
		Role:      sqlc.RoleSUPERADMIN,
		Status:    sqlc.UserStatusACTIVE,
		CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
	}
	repo.users[superAdminID] = superAdmin

	regularUserID := uuid.New()
	regularUser := &sqlc.User{
		ID:        regularUserID,
		Email:     "regular@orbito.io",
		FirstName: "Regular",
		LastName:  "User",
		Role:      sqlc.RoleMEMBER,
		Status:    sqlc.UserStatusACTIVE,
		CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
	}
	repo.users[regularUserID] = regularUser

	// Seed organization
	orgID := uuid.New()
	org := &sqlc.Organization{
		ID:        orgID,
		Name:      "Acme Corp",
		Slug:      "acme-corp",
		OwnerID:   regularUserID,
		Archive:   false,
		CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
	}
	repo.organizations[orgID] = org

	memID := uuid.New()
	repo.memberships[memID] = &sqlc.OrganizationMember{
		ID:             memID,
		Role:           sqlc.RoleOWNER,
		UserID:         regularUserID,
		OrganizationID: orgID,
		JoinedAt:       pgtype.Timestamp{Time: time.Now(), Valid: true},
	}

	// 1. Non-admin access receives 403 Forbidden
	t.Run("non-admin receives 403 Forbidden", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), regularUserID.String()), "MEMBER"))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("Expected 403 Forbidden, got %d", rec.Code)
		}
	})

	// 2. Super admin gets dashboard stats
	t.Run("super admin gets dashboard stats", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
		req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), superAdminID.String()), "SUPER_ADMIN"))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
		var stats DashboardStatsDTO
		_ = json.Unmarshal(rec.Body.Bytes(), &stats)
		if stats.TotalUsers != 2 {
			t.Fatalf("Expected 2 total users, got %d", stats.TotalUsers)
		}
		if stats.TotalOrganizations != 1 {
			t.Fatalf("Expected 1 total organization, got %d", stats.TotalOrganizations)
		}
	})

	// 3. Super admin lists users (paginated)
	t.Run("super admin lists users", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users?page=1&limit=10", nil)
		req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), superAdminID.String()), "SUPER_ADMIN"))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
		var res PaginatedUsersDTO
		_ = json.Unmarshal(rec.Body.Bytes(), &res)
		if res.Total != 2 {
			t.Fatalf("Expected total 2, got %d", res.Total)
		}
		if !strings.Contains(rec.Body.String(), `"_count":{"organizationMembers":`) {
			t.Fatalf("Expected _count.organizationMembers object in JSON response, got: %s", rec.Body.String())
		}
	})

	// 4. Super admin gets user detail
	t.Run("super admin gets user detail", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/"+regularUserID.String(), nil)
		req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), superAdminID.String()), "SUPER_ADMIN"))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
		}
		var detail UserDetailDTO
		_ = json.Unmarshal(rec.Body.Bytes(), &detail)
		if detail.Email != "regular@orbito.io" {
			t.Fatalf("Expected email regular@orbito.io, got %s", detail.Email)
		}
		if len(detail.OrganizationMembers) != 1 {
			t.Fatalf("Expected 1 organization membership, got %d", len(detail.OrganizationMembers))
		}
	})

	// 5. Update user role: self-role change prevented
	t.Run("self-role update prevented with 403", func(t *testing.T) {
		body, _ := json.Marshal(UpdateUserRoleDTO{Role: "MEMBER"})
		req := httptest.NewRequest(http.MethodPatch, "/users/"+superAdminID.String()+"/role", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), superAdminID.String()), "SUPER_ADMIN"))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("Expected 403 Forbidden on self-demotion, got %d", rec.Code)
		}
	})

	// 6. Demoting last super admin prevented
	t.Run("demoting last super admin prevented with 403", func(t *testing.T) {
		otherAdminID := uuid.New()
		otherAdmin := &sqlc.User{
			ID:        otherAdminID,
			Email:     "otheradmin@orbito.io",
			Role:      sqlc.RoleSUPERADMIN,
			Status:    sqlc.UserStatusACTIVE,
			CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		}
		repo.users[otherAdminID] = otherAdmin

		// Now we have 2 super admins: superAdminID demotes otherAdminID -> succeeds
		body, _ := json.Marshal(UpdateUserRoleDTO{Role: "MEMBER"})
		req := httptest.NewRequest(http.MethodPatch, "/users/"+otherAdminID.String()+"/role", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), superAdminID.String()), "SUPER_ADMIN"))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK when demoting 1 of 2 admins, got %d: %s", rec.Code, rec.Body.String())
		}

		// Now only 1 super admin remains. Trying to demote it fails.
		// Set otherAdmin back to super admin, and try to demote from a 2nd admin account
		secondAdminID := uuid.New()
		secondAdmin := &sqlc.User{
			ID:        secondAdminID,
			Email:     "secondadmin@orbito.io",
			Role:      sqlc.RoleSUPERADMIN,
			Status:    sqlc.UserStatusACTIVE,
			CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		}
		repo.users[secondAdminID] = secondAdmin
		// superAdminID is the target, secondAdminID is requester
		delete(repo.users, secondAdminID) // now only superAdminID remains

		body, _ = json.Marshal(UpdateUserRoleDTO{Role: "MEMBER"})
		req = httptest.NewRequest(http.MethodPatch, "/users/"+superAdminID.String()+"/role", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), otherAdminID.String()), "SUPER_ADMIN"))
		rec = httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("Expected 403 Forbidden demoting last super admin, got %d", rec.Code)
		}
	})

	// 7. Update user status
	t.Run("update user status to INACTIVE", func(t *testing.T) {
		body, _ := json.Marshal(UpdateUserStatusDTO{Status: "INACTIVE"})
		req := httptest.NewRequest(http.MethodPatch, "/users/"+regularUserID.String()+"/status", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), superAdminID.String()), "SUPER_ADMIN"))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK on status update, got %d: %s", rec.Code, rec.Body.String())
		}
		if regularUser.Status != sqlc.UserStatusINACTIVE {
			t.Fatalf("Expected status INACTIVE, got %s", regularUser.Status)
		}
	})

	// 8. Delete user owning organization fails
	t.Run("delete user owning organization fails with 403", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/users/"+regularUserID.String(), nil)
		req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), superAdminID.String()), "SUPER_ADMIN"))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("Expected 403 Forbidden when deleting user who owns organization, got %d", rec.Code)
		}
	})

	// 9. Transfer ownership of organization
	t.Run("transfer organization ownership", func(t *testing.T) {
		superAdmin.Status = sqlc.UserStatusACTIVE
		body, _ := json.Marshal(TransferOwnershipDTO{NewOwnerID: superAdminID.String()})
		req := httptest.NewRequest(http.MethodPatch, "/organizations/"+orgID.String()+"/transfer-ownership", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), superAdminID.String()), "SUPER_ADMIN"))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK on ownership transfer, got %d: %s", rec.Code, rec.Body.String())
		}
		if org.OwnerID != superAdminID {
			t.Fatalf("Expected new owner %s, got %s", superAdminID, org.OwnerID)
		}
	})

	// 10. Now deleting regular user succeeds
	t.Run("delete user succeeds after ownership transferred", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/users/"+regularUserID.String(), nil)
		req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), superAdminID.String()), "SUPER_ADMIN"))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK on delete user, got %d: %s", rec.Code, rec.Body.String())
		}
		if !regularUser.DeletedAt.Valid {
			t.Fatal("Expected user to be soft-deleted")
		}
	})

	// 11. Toggle organization archive
	t.Run("toggle organization archive", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPatch, "/organizations/"+orgID.String()+"/archive", nil)
		req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), superAdminID.String()), "SUPER_ADMIN"))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK on toggle archive, got %d: %s", rec.Code, rec.Body.String())
		}
		if !org.Archive {
			t.Fatal("Expected organization archive to be true")
		}
	})

	// 12. List organizations
	t.Run("list organizations", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/organizations", nil)
		req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), superAdminID.String()), "SUPER_ADMIN"))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK on list organizations, got %d: %s", rec.Code, rec.Body.String())
		}
		var res PaginatedOrgsDTO
		_ = json.Unmarshal(rec.Body.Bytes(), &res)
		if res.Total != 1 {
			t.Fatalf("Expected total 1, got %d", res.Total)
		}
	})

	// 13. System config and SMTP testing
	t.Run("get system config with smtp sources", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/config", nil)
		req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), superAdminID.String()), "SUPER_ADMIN"))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK on get system config, got %d: %s", rec.Code, rec.Body.String())
		}
		var res ConfigResponseDTO
		_ = json.Unmarshal(rec.Body.Bytes(), &res)
		if len(res.SmtpSources) == 0 {
			t.Fatal("Expected smtp sources array to be populated")
		}
	})

	// 14. Test SMTP connection
	t.Run("test smtp endpoint", func(t *testing.T) {
		body, _ := json.Marshal(TestSMTPDTO{Email: "test@example.com"})
		req := httptest.NewRequest(http.MethodPost, "/config/test-smtp", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), superAdminID.String()), "SUPER_ADMIN"))
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK on test-smtp, got %d: %s", rec.Code, rec.Body.String())
		}
	})
}
