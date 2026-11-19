package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
)

func TestAuthRegisterAndLogin(t *testing.T) {
	repo := newMemoryAuthRepo()
	jwtSecret := "test-jwt-secret-at-least-32-chars-long!"
	jwtRefreshSecret := "test-jwt-refresh-secret-at-least-32-chars!"
	tokenSvc := NewTokenService(jwtSecret, jwtRefreshSecret, 15*time.Minute, 7*24*time.Hour)
	authSvc := NewService(repo, tokenSvc, nil, nil, "http://localhost:3001")
	authHandler := NewHandler(authSvc, jwtSecret, false, "http://localhost:3001")

	authMw := middleware.RequireAuth(tokenSvc, repo.GetUserByID)
	router := Routes(authHandler, authMw)

	// 1. Register first account (should receive MEMBER role, and on login elevate to SUPER_ADMIN)
	regBody, _ := json.Marshal(RegisterDTO{
		Email:     "admin@orbito.io",
		Password:  "Secret1234!",
		FirstName: "Super",
		LastName:  "Admin",
	})
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created on register, got %d: %s", rec.Code, rec.Body.String())
	}

	var regRes ResponseDTO
	if err := json.Unmarshal(rec.Body.Bytes(), &regRes); err != nil {
		t.Fatalf("Failed to parse register response: %v", err)
	}
	if regRes.AccessToken == "" || regRes.RefreshToken == "" {
		t.Fatal("Tokens missing in register response")
	}
	if regRes.User.Username == nil || *regRes.User.Username != "admin" {
		t.Fatalf("Expected auto-assigned username 'admin', got %v", regRes.User.Username)
	}

	// Verify cookies
	cookies := rec.Result().Cookies()
	var hasAccessCookie, hasRefreshCookie bool
	for _, c := range cookies {
		if c.Name == "access_token" && c.Value != "" {
			hasAccessCookie = true
		}
		if c.Name == "refresh_token" && c.Value != "" {
			hasRefreshCookie = true
		}
	}
	if !hasAccessCookie || !hasRefreshCookie {
		t.Fatal("Expected access_token and refresh_token cookies to be set")
	}

	// 2. Reject duplicate registration with 409 Conflict
	req = httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("Expected 409 Conflict on duplicate email, got %d", rec.Code)
	}

	// 3. Login with password (bcrypt cost 12 verification + bootstrap rule to SUPER_ADMIN)
	loginBody, _ := json.Marshal(LoginDTO{
		Email:    "admin@orbito.io",
		Password: "Secret1234!",
	})
	req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(loginBody))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on login, got %d: %s", rec.Code, rec.Body.String())
	}
	var loginRes ResponseDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &loginRes)
	if loginRes.User.Role != string(sqlc.RoleSUPERADMIN) {
		t.Fatalf("Expected bootstrap elevation to SUPER_ADMIN, got %s", loginRes.User.Role)
	}

	// 4. GET /me (using access token cookie)
	req = httptest.NewRequest(http.MethodGet, "/me", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: loginRes.AccessToken})
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on GET /me, got %d: %s", rec.Code, rec.Body.String())
	}
	var meRes UserDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &meRes)
	if meRes.Email != "admin@orbito.io" {
		t.Fatalf("Expected email admin@orbito.io, got %s", meRes.Email)
	}
}

func TestContextHelpersUsage(t *testing.T) {
	ctx := ctxhelper.WithUserID(context.Background(), "some-id")
	if ctxhelper.GetUserID(ctx) != "some-id" {
		t.Fatal("Failed context user id helper")
	}
}
