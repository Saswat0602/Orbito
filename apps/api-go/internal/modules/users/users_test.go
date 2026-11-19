package users

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
	"github.com/orbito/api-go/internal/platform/crypto"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
)

type memoryUsersRepo struct {
	users      map[uuid.UUID]*sqlc.User
	sharedOrgs map[string]bool
}

func newMemoryUsersRepo() *memoryUsersRepo {
	return &memoryUsersRepo{
		users:      make(map[uuid.UUID]*sqlc.User),
		sharedOrgs: make(map[string]bool),
	}
}

func (m *memoryUsersRepo) ListUsers(_ context.Context, limit, offset int32) ([]sqlc.ListUsersRow, error) {
	var rows []sqlc.ListUsersRow
	for _, u := range m.users {
		if !u.DeletedAt.Valid {
			rows = append(rows, sqlc.ListUsersRow{
				ID:          u.ID,
				Email:       u.Email,
				Username:    u.Username,
				FirstName:   u.FirstName,
				LastName:    u.LastName,
				Avatar:      u.Avatar,
				Bio:         u.Bio,
				Timezone:    u.Timezone,
				Language:    u.Language,
				Role:        u.Role,
				Status:      u.Status,
				LastLoginAt: u.LastLoginAt,
				CreatedAt:   u.CreatedAt,
				UpdatedAt:   u.UpdatedAt,
				DeletedAt:   u.DeletedAt,
			})
		}
	}
	if int(offset) >= len(rows) {
		return []sqlc.ListUsersRow{}, nil
	}
	end := int(offset + limit)
	if end > len(rows) {
		end = len(rows)
	}
	return rows[offset:end], nil
}

func (m *memoryUsersRepo) CountUsers(_ context.Context) (int64, error) {
	var count int64
	for _, u := range m.users {
		if !u.DeletedAt.Valid {
			count++
		}
	}
	return count, nil
}

func (m *memoryUsersRepo) GetUserByID(_ context.Context, id uuid.UUID) (*sqlc.User, error) {
	u, ok := m.users[id]
	if !ok || u.DeletedAt.Valid {
		return nil, nil
	}
	return u, nil
}

func (m *memoryUsersRepo) GetUserByEmail(_ context.Context, email string) (*sqlc.User, error) {
	for _, u := range m.users {
		if strings.EqualFold(u.Email, email) && !u.DeletedAt.Valid {
			return u, nil
		}
	}
	return nil, nil
}

func (m *memoryUsersRepo) GetUserByUsername(_ context.Context, username string) (*sqlc.User, error) {
	for _, u := range m.users {
		if u.Username.Valid && strings.EqualFold(u.Username.String, username) && !u.DeletedAt.Valid {
			return u, nil
		}
	}
	return nil, nil
}

func (m *memoryUsersRepo) CreateUser(_ context.Context, params sqlc.CreateUserParams) (*sqlc.User, error) {
	u := &sqlc.User{
		ID:        params.ID,
		Email:     params.Email,
		Username:  params.Username,
		FirstName: params.FirstName,
		LastName:  params.LastName,
		Password:  params.Password,
		Role:      params.Role,
		Status:    params.Status,
		Timezone:  params.Timezone,
		Language:  params.Language,
		CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
	}
	m.users[params.ID] = u
	return u, nil
}

func (m *memoryUsersRepo) UpdateUserProfile(_ context.Context, params sqlc.UpdateUserProfileParams) (*sqlc.UpdateUserProfileRow, error) {
	u, ok := m.users[params.ID]
	if !ok {
		return nil, nil
	}
	if params.FirstName != "" {
		u.FirstName = params.FirstName
	}
	if params.LastName != "" {
		u.LastName = params.LastName
	}
	if params.Bio != "" {
		u.Bio = pgtype.Text{String: params.Bio, Valid: true}
	}
	u.UpdatedAt = pgtype.Timestamp{Time: time.Now(), Valid: true}

	return &sqlc.UpdateUserProfileRow{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Avatar:    u.Avatar,
		Bio:       u.Bio,
		Timezone:  u.Timezone,
		Language:  u.Language,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}, nil
}

func (m *memoryUsersRepo) UpdateUserStatus(_ context.Context, id uuid.UUID, status sqlc.UserStatus) (*sqlc.UpdateUserStatusRow, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, nil
	}
	u.Status = status
	u.UpdatedAt = pgtype.Timestamp{Time: time.Now(), Valid: true}

	return &sqlc.UpdateUserStatusRow{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Role:      u.Role,
		Status:    u.Status,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}, nil
}

func (m *memoryUsersRepo) UpdatePassword(_ context.Context, id uuid.UUID, hash pgtype.Text) error {
	if u, ok := m.users[id]; ok {
		u.Password = hash
		u.UpdatedAt = pgtype.Timestamp{Time: time.Now(), Valid: true}
	}
	return nil
}

func (m *memoryUsersRepo) CreateUserWithProfile(_ context.Context, params sqlc.CreateUserWithProfileParams) (*sqlc.CreateUserWithProfileRow, error) {
	u := &sqlc.User{
		ID:           params.ID,
		Email:        params.Email,
		Username:     params.Username,
		FirstName:    params.FirstName,
		LastName:     params.LastName,
		Password:     params.Password,
		Role:         params.Role,
		Status:       params.Status,
		Timezone:     params.Timezone,
		Language:     params.Language,
		MobileNumber: params.MobileNumber,
		Avatar:       params.Avatar,
		Bio:          params.Bio,
		CreatedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true},
		UpdatedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true},
	}
	m.users[params.ID] = u
	return &sqlc.CreateUserWithProfileRow{
		ID:           u.ID,
		Email:        u.Email,
		Username:     u.Username,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Avatar:       u.Avatar,
		Bio:          u.Bio,
		Timezone:     u.Timezone,
		Language:     u.Language,
		MobileNumber: u.MobileNumber,
		Role:         u.Role,
		Status:       u.Status,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}, nil
}

func (m *memoryUsersRepo) CheckUsersShareOrganization(_ context.Context, u1, u2 uuid.UUID) (bool, error) {
	key1 := u1.String() + ":" + u2.String()
	key2 := u2.String() + ":" + u1.String()
	return m.sharedOrgs[key1] || m.sharedOrgs[key2], nil
}

func (m *memoryUsersRepo) GetPublicUserByID(_ context.Context, id uuid.UUID) (*sqlc.GetPublicUserByIDRow, error) {
	u, ok := m.users[id]
	if !ok || u.DeletedAt.Valid {
		return nil, nil
	}
	return &sqlc.GetPublicUserByIDRow{
		ID:           u.ID,
		Email:        u.Email,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Avatar:       u.Avatar,
		MobileNumber: u.MobileNumber,
		Bio:          u.Bio,
		Timezone:     u.Timezone,
		Role:         u.Role,
	}, nil
}

func (m *memoryUsersRepo) GetPublicUserByUsername(_ context.Context, username string) (*sqlc.GetPublicUserByUsernameRow, error) {
	for _, u := range m.users {
		if u.Username.Valid && strings.EqualFold(u.Username.String, username) && !u.DeletedAt.Valid {
			return &sqlc.GetPublicUserByUsernameRow{
				ID:           u.ID,
				Email:        u.Email,
				FirstName:    u.FirstName,
				LastName:     u.LastName,
				Avatar:       u.Avatar,
				MobileNumber: u.MobileNumber,
				Bio:          u.Bio,
				Timezone:     u.Timezone,
				Role:         u.Role,
			}, nil
		}
	}
	return nil, nil
}

func (m *memoryUsersRepo) DeleteUser(_ context.Context, id, deletedBy uuid.UUID) error {
	if u, ok := m.users[id]; ok {
		u.DeletedAt = pgtype.Timestamp{Time: time.Now(), Valid: true}
		u.DeletedBy = pgtype.UUID{Bytes: deletedBy, Valid: true}
	}
	return nil
}

func TestUsersModule(t *testing.T) {
	repo := newMemoryUsersRepo()
	svc := NewService(repo)
	handler := NewHandler(svc)
	router := Routes(handler, nil)

	// 1. Initial check: GET /exists
	req := httptest.NewRequest(http.MethodGet, "/exists", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on /exists, got %d", rec.Code)
	}
	var existsRes map[string]bool
	_ = json.Unmarshal(rec.Body.Bytes(), &existsRes)
	if existsRes["exists"] {
		t.Fatal("Expected exists=false when repository is empty")
	}

	// Seed super admin and regular user
	adminID := uuid.New()
	adminUser := &sqlc.User{
		ID:        adminID,
		Email:     "admin@orbito.io",
		FirstName: "Super",
		LastName:  "Admin",
		Role:      sqlc.RoleSUPERADMIN,
		Status:    sqlc.UserStatusACTIVE,
	}
	repo.users[adminID] = adminUser

	regularID := uuid.New()
	initPwdHash, _ := crypto.HashPassword("OldSecret123!")
	regularUser := &sqlc.User{
		ID:        regularID,
		Email:     "member@orbito.io",
		FirstName: "Regular",
		LastName:  "Member",
		Password:  pgtype.Text{String: initPwdHash, Valid: true},
		Role:      sqlc.RoleMEMBER,
		Status:    sqlc.UserStatusACTIVE,
		CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
	}
	repo.users[regularID] = regularUser

	// 2. GET /users (as SUPER_ADMIN)
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), adminID.String()), "SUPER_ADMIN"))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for admin listing users, got %d: %s", rec.Code, rec.Body.String())
	}

	// 3. GET /users (as regular member -> 403 Forbidden)
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), regularID.String()), "MEMBER"))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden for non-admin directory view, got %d", rec.Code)
	}

	// 4. Non-admin viewing someone else's profile -> 403 Forbidden
	otherID := uuid.New()
	req = httptest.NewRequest(http.MethodGet, "/"+adminID.String(), nil)
	req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), otherID.String()), "MEMBER"))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden viewing other profile, got %d", rec.Code)
	}

	// 5. Non-admin attempting privileged field self-update -> 403 Forbidden
	roleStr := "SUPER_ADMIN"
	forbiddenBody, _ := json.Marshal(UpdateUserDTO{Role: &roleStr})
	req = httptest.NewRequest(http.MethodPatch, "/"+regularID.String(), bytes.NewReader(forbiddenBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), regularID.String()), "MEMBER"))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden on privileged role self-update, got %d", rec.Code)
	}

	// 6. Valid profile update (Bio)
	newBio := "Software Engineer"
	validBody, _ := json.Marshal(UpdateUserDTO{Bio: &newBio})
	req = httptest.NewRequest(http.MethodPatch, "/"+regularID.String(), bytes.NewReader(validBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), regularID.String()), "MEMBER"))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on profile update, got %d: %s", rec.Code, rec.Body.String())
	}
	if regularUser.Bio.String != newBio {
		t.Fatalf("Expected bio updated to '%s', got '%s'", newBio, regularUser.Bio.String)
	}

	// 7. Update status: PATCH /users/{id}/status (SUPER_ADMIN)
	statusBody, _ := json.Marshal(UpdateStatusDTO{Status: "INACTIVE"})
	req = httptest.NewRequest(http.MethodPatch, "/"+regularID.String()+"/status", bytes.NewReader(statusBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), adminID.String()), "SUPER_ADMIN"))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on status update, got %d: %s", rec.Code, rec.Body.String())
	}
	if regularUser.Status != sqlc.UserStatusINACTIVE {
		t.Fatalf("Expected status INACTIVE, got %s", regularUser.Status)
	}

	// 8. Change password: POST /users/change-password
	changePwdBody, _ := json.Marshal(ChangePasswordDTO{
		CurrentPassword: "OldSecret123!",
		NewPassword:     "NewBrandPassword456!",
		ConfirmPassword: "NewBrandPassword456!",
	})
	req = httptest.NewRequest(http.MethodPost, "/change-password", bytes.NewReader(changePwdBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), regularID.String()), "MEMBER"))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on change-password, got %d: %s", rec.Code, rec.Body.String())
	}
	if !crypto.CheckPasswordHash("NewBrandPassword456!", regularUser.Password.String) {
		t.Fatal("Expected new password to verify with bcrypt")
	}

	// 9. Change password with wrong current password fails with 400 Bad Request
	wrongPwdBody, _ := json.Marshal(ChangePasswordDTO{
		CurrentPassword: "WrongPassword999!",
		NewPassword:     "AnotherPassword789!",
		ConfirmPassword: "AnotherPassword789!",
	})
	req = httptest.NewRequest(http.MethodPost, "/change-password", bytes.NewReader(wrongPwdBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), regularID.String()), "MEMBER"))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request for wrong current password, got %d: %s", rec.Code, rec.Body.String())
	}

	// 10. Change password with identical password fails with 400 Bad Request
	samePwdBody, _ := json.Marshal(ChangePasswordDTO{
		CurrentPassword: "NewBrandPassword456!",
		NewPassword:     "NewBrandPassword456!",
		ConfirmPassword: "NewBrandPassword456!",
	})
	req = httptest.NewRequest(http.MethodPost, "/change-password", bytes.NewReader(samePwdBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), regularID.String()), "MEMBER"))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request for identical new password, got %d", rec.Code)
	}

	// 11. Public profile: shared organization member can view colleague profile
	colleagueID := uuid.New()
	colleagueUser := &sqlc.User{
		ID:           colleagueID,
		Email:        "colleague@orbito.io",
		Username:     pgtype.Text{String: "colleague1", Valid: true},
		FirstName:    "Alice",
		LastName:     "Colleague",
		Timezone:     "UTC",
		Language:     "en",
		MobileNumber: pgtype.Text{String: "+1234567890", Valid: true},
		Role:         sqlc.RoleMEMBER,
		Status:       sqlc.UserStatusACTIVE,
	}
	repo.users[colleagueID] = colleagueUser
	repo.sharedOrgs[regularID.String()+":"+colleagueID.String()] = true

	req = httptest.NewRequest(http.MethodGet, "/"+colleagueID.String()+"/profile", nil)
	req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), regularID.String()), "MEMBER"))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for shared organization profile access, got %d: %s", rec.Code, rec.Body.String())
	}
	var pubProfile PublicProfileDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &pubProfile)
	if pubProfile.Email != "colleague@orbito.io" || pubProfile.FirstName != "Alice" {
		t.Fatalf("Unexpected public profile data: %+v", pubProfile)
	}

	// 12. Public profile: lookup by username
	req = httptest.NewRequest(http.MethodGet, "/colleague1/profile", nil)
	req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), regularID.String()), "MEMBER"))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for username public profile lookup, got %d: %s", rec.Code, rec.Body.String())
	}

	// 13. Public profile: stranger in different organization is rejected with 403 Forbidden
	strangerID := uuid.New()
	strangerUser := &sqlc.User{
		ID:        strangerID,
		Email:     "stranger@other.io",
		Username:  pgtype.Text{String: "stranger1", Valid: true},
		FirstName: "Bob",
		LastName:  "Stranger",
		Role:      sqlc.RoleMEMBER,
		Status:    sqlc.UserStatusACTIVE,
	}
	repo.users[strangerID] = strangerUser

	req = httptest.NewRequest(http.MethodGet, "/"+strangerID.String()+"/profile", nil)
	req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), regularID.String()), "MEMBER"))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden for non-shared organization public profile, got %d", rec.Code)
	}

	// 14. Online status endpoint: returns isOnline and lastSeen
	req = httptest.NewRequest(http.MethodGet, "/"+regularID.String()+"/status", nil)
	req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), regularID.String()), "MEMBER"))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on user status, got %d: %s", rec.Code, rec.Body.String())
	}
	var statusRes OnlineStatusDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &statusRes)
	if statusRes.LastSeen == nil || *statusRes.LastSeen == "" {
		t.Fatal("Expected lastSeen timestamp to be present in user online status")
	}
}
