package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
)

func TestJWTTokenService(t *testing.T) {
	tokenSvc := NewTokenService("test-access-secret-32-bytes-orbito", "test-refresh-secret-32-bytes-orbito", 1*time.Second, 5*time.Second)

	userID := uuid.New()
	email := "user@orbito.io"
	role := "MEMBER"

	// 1. Mint & Validate Access Token
	token, err := tokenSvc.MintAccessToken(userID, email, role)
	if err != nil {
		t.Fatalf("Failed to mint access token: %v", err)
	}

	claims, err := tokenSvc.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("Failed to validate access token: %v", err)
	}
	if claims.Sub != userID.String() || claims.Email != email || claims.Role != role {
		t.Fatalf("Claims mismatch: %+v", claims)
	}

	// 2. Token expiration
	time.Sleep(1200 * time.Millisecond)
	_, err = tokenSvc.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("Expected error for expired token, got nil")
	}

	// 3. Refresh token
	refToken, err := tokenSvc.MintRefreshToken(userID, email, role)
	if err != nil {
		t.Fatalf("Failed to mint refresh token: %v", err)
	}
	refClaims, err := tokenSvc.ValidateRefreshToken(refToken)
	if err != nil {
		t.Fatalf("Failed to validate refresh token: %v", err)
	}
	if refClaims.Sub != userID.String() {
		t.Fatalf("Refresh claims sub mismatch: %s", refClaims.Sub)
	}
}

func TestRequireAuthMiddleware(t *testing.T) {
	tokenSvc := NewTokenService("test-access-secret-32-bytes-orbito", "test-refresh-secret-32-bytes-orbito", 15*time.Minute, 7*24*time.Hour)

	userMap := make(map[uuid.UUID]*sqlc.User)
	userChecker := func(_ context.Context, id uuid.UUID) (*sqlc.User, error) {
		return userMap[id], nil
	}

	mw := middleware.RequireAuth(tokenSvc, userChecker)
	protectedHandler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := ctxhelper.GetUserID(r.Context())
		w.Header().Set("X-User-ID", uid)
		w.WriteHeader(http.StatusOK)
	}))

	// 1. Missing token returns 401
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	protectedHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401 for missing token, got %d", rec.Code)
	}

	// 2. Bearer token header works
	activeUser := &sqlc.User{
		ID:     uuid.New(),
		Email:  "active@orbito.io",
		Role:   sqlc.RoleMEMBER,
		Status: sqlc.UserStatusACTIVE,
	}
	userMap[activeUser.ID] = activeUser

	token, _ := tokenSvc.MintAccessToken(activeUser.ID, activeUser.Email, string(activeUser.Role))
	req = httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	protectedHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 for valid Bearer token, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("X-User-ID") != activeUser.ID.String() {
		t.Fatalf("Context user ID mismatch: %s", rec.Header().Get("X-User-ID"))
	}

	// 3. access_token cookie fallback works
	req = httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	rec = httptest.NewRecorder()
	protectedHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 for access_token cookie, got %d", rec.Code)
	}

	// 4. System user ID is barred
	sysToken, _ := tokenSvc.MintAccessToken(SystemUserID, "system@orbito.local", "SUPER_ADMIN")
	req = httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+sysToken)
	rec = httptest.NewRecorder()
	protectedHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401 for system user, got %d", rec.Code)
	}

	// 5. Inactive user returns 401 Account is not active
	inactiveUser := &sqlc.User{
		ID:     uuid.New(),
		Email:  "inactive@orbito.io",
		Role:   sqlc.RoleMEMBER,
		Status: sqlc.UserStatusINACTIVE,
	}
	userMap[inactiveUser.ID] = inactiveUser
	inactToken, _ := tokenSvc.MintAccessToken(inactiveUser.ID, inactiveUser.Email, string(inactiveUser.Role))
	req = httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+inactToken)
	rec = httptest.NewRecorder()
	protectedHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401 for inactive user, got %d", rec.Code)
	}
	if rec.Body.String() == "" || !containsSubstring(rec.Body.String(), "Account is not active") {
		t.Fatalf("Expected 'Account is not active', got: %s", rec.Body.String())
	}

	// 6. Soft-deleted user returns 401 Account is not active
	deletedUser := &sqlc.User{
		ID:        uuid.New(),
		Email:     "deleted@orbito.io",
		Role:      sqlc.RoleMEMBER,
		Status:    sqlc.UserStatusACTIVE,
		DeletedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
	}
	userMap[deletedUser.ID] = deletedUser
	delToken, _ := tokenSvc.MintAccessToken(deletedUser.ID, deletedUser.Email, string(deletedUser.Role))
	req = httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+delToken)
	rec = httptest.NewRecorder()
	protectedHandler.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Expected 401 for deleted user, got %d", rec.Code)
	}
}

func containsSubstring(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || stringIndex(s, sub) >= 0)
}

func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
