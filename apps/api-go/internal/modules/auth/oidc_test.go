package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/orbito/api-go/internal/platform/crypto"
)

func TestOIDCEndpoints(t *testing.T) {
	repo := newMemoryAuthRepo()
	jwtSecret := "test-jwt-secret-at-least-32-chars-long!"
	tokenSvc := NewTokenService(jwtSecret, jwtSecret, 15*time.Minute, 7*24*time.Hour)
	authSvc := NewService(repo, tokenSvc, nil, nil, "http://localhost:3001")
	authHandler := NewHandler(authSvc, jwtSecret, false, "http://localhost:3001")
	router := Routes(authHandler, nil)

	// 1. GET /oidc/config
	req := httptest.NewRequest(http.MethodGet, "/oidc/config", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on /oidc/config, got %d", rec.Code)
	}
	var cfgRes OIDCConfigResponseDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &cfgRes)
	if cfgRes.ProviderName != "OpenID Connect" {
		t.Fatalf("Expected 'OpenID Connect', got %s", cfgRes.ProviderName)
	}

	// 2. GET /oidc/login redirect
	req = httptest.NewRequest(http.MethodGet, "/oidc/login", nil)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("Expected 302 Found redirect on /oidc/login, got %d", rec.Code)
	}

	// 3. POST /oidc/exchange with encrypted sso_auth cookie
	ssoPayload, _ := json.Marshal(ResponseDTO{
		AccessToken:  "mock-sso-access-token",
		RefreshToken: "mock-sso-refresh-token",
		User: UserDTO{
			ID:        "00000000-0000-0000-0000-000000000001",
			Email:     "sso-user@orbito.io",
			FirstName: "SSO",
			LastName:  "User",
			Role:      "MEMBER",
		},
	})
	encryptedCookie, _ := crypto.EncryptCookieValue(string(ssoPayload), jwtSecret)

	req = httptest.NewRequest(http.MethodPost, "/oidc/exchange", nil)
	req.AddCookie(&http.Cookie{Name: "sso_auth", Value: encryptedCookie})
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on /oidc/exchange, got %d: %s", rec.Code, rec.Body.String())
	}
	var exRes ResponseDTO
	_ = json.Unmarshal(rec.Body.Bytes(), &exRes)
	if exRes.AccessToken != "mock-sso-access-token" {
		t.Fatalf("Expected exchanged access token, got %s", exRes.AccessToken)
	}
}
