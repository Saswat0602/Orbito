package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/crypto"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

func TestPasswordResetAndSetup(t *testing.T) {
	repo := newMemoryAuthRepo()
	jwtSecret := "test-jwt-secret-at-least-32-chars-long!"
	jwtRefreshSecret := "test-jwt-refresh-secret-at-least-32-chars!"
	tokenSvc := NewTokenService(jwtSecret, jwtRefreshSecret, 15*time.Minute, 7*24*time.Hour)
	authSvc := NewService(repo, tokenSvc, nil, nil, "http://localhost:3001")
	authHandler := NewHandler(authSvc, jwtSecret, false, "http://localhost:3001")
	router := Routes(authHandler, nil)

	// 1. Initial setup required check (0 users -> required: true)
	req := httptest.NewRequest(http.MethodGet, "/setup/required", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on /setup/required, got %d", rec.Code)
	}
	var setupReq SetupRequiredResponseDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &setupReq)
	if !setupReq.Required || !setupReq.CanSetup {
		t.Fatalf("Expected required=true and canSetup=true, got %+v", setupReq)
	}

	// 2. Perform /setup
	setupBody, _ := json.Marshal(SetupAdminDTO{
		Email:     "founding-admin@orbito.io",
		Password:  "AdminPassword123!",
		FirstName: "Admin",
		LastName:  "Root",
	})
	req = httptest.NewRequest(http.MethodPost, "/setup", bytes.NewReader(setupBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created on /setup, got %d: %s", rec.Code, rec.Body.String())
	}
	var setupRes ResponseDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &setupRes)
	if setupRes.User.Role != string(sqlc.RoleSUPERADMIN) {
		t.Fatalf("Expected SUPER_ADMIN role from setup, got %s", setupRes.User.Role)
	}

	// 3. Second /setup attempt is rejected with 409 Conflict
	req = httptest.NewRequest(http.MethodPost, "/setup", bytes.NewReader(setupBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("Expected 409 Conflict on repeated setup, got %d", rec.Code)
	}

	// 4. Forgot password flow
	user := &sqlc.User{
		ID:        uuid.New(),
		Email:     "user-reset@orbito.io",
		FirstName: "Reset",
		LastName:  "User",
		Password:  pgtype.Text{String: "old-hash", Valid: true},
		Status:    sqlc.UserStatusACTIVE,
		Role:      sqlc.RoleMEMBER,
	}
	repo.users[user.ID] = user

	forgotBody, _ := json.Marshal(ForgotPasswordDTO{Email: user.Email})
	req = httptest.NewRequest(http.MethodPost, "/forgot-password", bytes.NewReader(forgotBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on forgot-password, got %d: %s", rec.Code, rec.Body.String())
	}

	// Retrieve generated token from repo
	resetToken := user.ResetToken.String
	if resetToken == "" {
		t.Fatal("Expected reset token to be set on user in database")
	}

	// 5. Verify reset token
	req = httptest.NewRequest(http.MethodGet, "/verify-reset-token/"+resetToken, nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on verify-reset-token, got %d", rec.Code)
	}
	var verifyRes VerifyResetTokenResponseDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &verifyRes)
	if !verifyRes.Valid {
		t.Fatal("Expected reset token to be valid")
	}

	// 6. Reset password
	newPassword := "NewBrandPassword2026!"
	resetBody, _ := json.Marshal(ResetPasswordDTO{
		Token:           resetToken,
		Password:        newPassword,
		ConfirmPassword: newPassword,
	})
	req = httptest.NewRequest(http.MethodPost, "/reset-password", bytes.NewReader(resetBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on reset-password, got %d: %s", rec.Code, rec.Body.String())
	}

	// Assert new password verifies via bcrypt
	if !crypto.CheckPasswordHash(newPassword, user.Password.String) {
		t.Fatal("Expected new password to verify with bcrypt")
	}

	// Assert reset token cleared
	if user.ResetToken.Valid {
		t.Fatal("Expected reset token to be cleared after successful password reset")
	}
}
