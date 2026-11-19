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
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/platform/crypto"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

func TestTokenRefreshAndLogout(t *testing.T) {
	repo := newMemoryAuthRepo()
	jwtSecret := "test-jwt-secret-at-least-32-chars-long!"
	jwtRefreshSecret := "test-jwt-refresh-secret-at-least-32-chars!"
	tokenSvc := NewTokenService(jwtSecret, jwtRefreshSecret, 15*time.Minute, 7*24*time.Hour)
	authSvc := NewService(repo, tokenSvc, nil, nil, "http://localhost:3001")
	authHandler := NewHandler(authSvc, jwtSecret, false, "http://localhost:3001")
	authMw := middleware.RequireAuth(tokenSvc, repo.GetUserByID)
	router := Routes(authHandler, authMw)

	// Seed user with active refresh token
	userID := uuid.New()
	userEmail := "refresh-user@orbito.io"
	userRole := sqlc.RoleMEMBER
	initialRefreshToken, _ := tokenSvc.MintRefreshToken(userID, userEmail, string(userRole))

	user := &sqlc.User{
		ID:           userID,
		Email:        userEmail,
		FirstName:    "Refresh",
		LastName:     "User",
		Role:         userRole,
		Status:       sqlc.UserStatusACTIVE,
		RefreshToken: pgtype.Text{String: initialRefreshToken, Valid: true},
	}
	repo.users[userID] = user

	// 1. Refresh via encrypted cookie
	encryptedRefreshCookie, err := crypto.EncryptCookieValue(initialRefreshToken, jwtSecret)
	if err != nil {
		t.Fatalf("Failed to encrypt cookie: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: encryptedRefreshCookie})
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on refresh via cookie, got %d: %s", rec.Code, rec.Body.String())
	}

	var refreshRes ResponseDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &refreshRes)
	if refreshRes.RefreshToken == initialRefreshToken {
		t.Fatal("Expected refresh token to be rotated, but received identical token")
	}

	// Verify database was updated with rotated token
	if user.RefreshToken.String != refreshRes.RefreshToken {
		t.Fatal("Expected database user refresh token to match newly minted token")
	}

	// 2. Old refresh token is now rejected
	req = httptest.NewRequest(http.MethodPost, "/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: encryptedRefreshCookie})
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401 Unauthorized for invalidated refresh token, got %d", rec.Code)
	}

	// 3. Logout clears refresh token in DB and clears cookies
	req = httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.Header.Set("Authorization", "Bearer "+refreshRes.AccessToken)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on logout, got %d: %s", rec.Code, rec.Body.String())
	}
	if user.RefreshToken.Valid {
		t.Fatal("Expected user refresh_token column to be null/invalid after logout")
	}

	cookies := rec.Result().Cookies()
	var refreshExpired, accessExpired bool
	for _, c := range cookies {
		if c.Name == "refresh_token" && c.MaxAge == -1 {
			refreshExpired = true
		}
		if c.Name == "access_token" && c.MaxAge == -1 {
			accessExpired = true
		}
	}
	if !refreshExpired || !accessExpired {
		t.Fatal("Expected client cookies to be cleared on logout")
	}

	// 4. Refresh via body fallback
	newToken, _ := tokenSvc.MintRefreshToken(userID, userEmail, string(userRole))
	user.RefreshToken = pgtype.Text{String: newToken, Valid: true}
	body, _ := json.Marshal(RefreshTokenDTO{RefreshToken: newToken})
	req = httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on refresh via body, got %d: %s", rec.Code, rec.Body.String())
	}
}
