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
	users map[uuid.UUID]*sqlc.User
}

func newMemoryUsersRepo() *memoryUsersRepo {
	return &memoryUsersRepo{users: make(map[uuid.UUID]*sqlc.User)}
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

	// 9. Change password with wrong current password fails
	wrongPwdBody, _ := json.Marshal(ChangePasswordDTO{
		CurrentPassword: "WrongPassword999!",
		NewPassword:     "AnotherPassword789!",
	})
	req = httptest.NewRequest(http.MethodPost, "/change-password", bytes.NewReader(wrongPwdBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctxhelper.WithUserRole(ctxhelper.WithUserID(req.Context(), regularID.String()), "MEMBER"))
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401 Unauthorized for wrong current password, got %d", rec.Code)
	}
}
