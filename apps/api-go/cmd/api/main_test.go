package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/config"
	"github.com/orbito/api-go/internal/modules/auth"
	"github.com/orbito/api-go/internal/shared/httpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildRouter(t *testing.T) {
	cfg := &config.Config{
		AppEnv:           "test",
		JWTSecret:        "test-secret-at-least-32-chars-long!",
		JWTRefreshSecret: "test-refresh-secret-at-least-32-chars!",
		CORSOrigins:      []string{"http://localhost:3001"},
		MaxFileSize:      10485760,
	}
	appLog := slog.New(slog.NewJSONHandler(io.Discard, nil))
	router := BuildRouter(cfg, nil, appLog)

	t.Run("GET /api returns exact NestJS welcome text and security headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api", nil)
		req.Header.Set("x-correlation-id", "test-corr-id-123")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Header().Get("Content-Type"), "text/html")
		assert.Equal(t, "test-corr-id-123", rr.Header().Get("x-correlation-id"))
		assert.Equal(t, "nosniff", rr.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "Welcome to Orbito Backend API!", rr.Body.String())
	})

	t.Run("GET /api/health/liveness routes correctly through mounted subrouter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/health/liveness", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), `"status":"live"`)
	})

	t.Run("GET /api/settings/registration-status routes correctly through mounted settings subrouter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/settings/registration-status", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), `"enabled":true`)
	})

	t.Run("GET /api/email/queue/stats routes correctly through mounted email subrouter", func(t *testing.T) {
		tokenSvc := auth.NewTokenService(cfg.JWTSecret, cfg.JWTRefreshSecret, 15*time.Minute, 7*24*time.Hour)
		tok, _ := tokenSvc.MintAccessToken(uuid.New(), "test@orbito.io", "MEMBER")

		req := httptest.NewRequest(http.MethodGet, "/api/email/queue/stats", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), `"waiting":0`)
	})

	t.Run("GET /api/auth/registration-status routes correctly through mounted auth subrouter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/auth/registration-status", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), `"enabled":true`)
	})

	t.Run("GET /api/users/exists routes correctly through mounted users subrouter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/users/exists", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), `"exists":false`)
	})

	t.Run("GET /api/admin/dashboard routes correctly through mounted admin subrouter and checks role", func(t *testing.T) {
		tokenSvc := auth.NewTokenService(cfg.JWTSecret, cfg.JWTRefreshSecret, 15*time.Minute, 7*24*time.Hour)
		tok, _ := tokenSvc.MintAccessToken(uuid.New(), "admin@orbito.io", "SUPER_ADMIN")

		req := httptest.NewRequest(http.MethodGet, "/api/admin/dashboard", nil)
		req.Header.Set("Authorization", "Bearer "+tok)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), `"totalUsers":`)
	})

	t.Run("GET /api/workflows requires authentication", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/workflows", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Unauthorized")
	})

	for _, path := range []string{"/api/task-statuses", "/api/labels", "/api/sprints"} {
		t.Run("GET "+path+" requires authentication", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusUnauthorized, rr.Code)
			assert.Contains(t, rr.Body.String(), "Unauthorized")
		})
	}

	t.Run("unknown route returns exact NestJS 404 JSON error envelope", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/nonexistent-route", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var errResp httpx.ErrorResponse
		err := json.Unmarshal(rr.Body.Bytes(), &errResp)
		require.NoError(t, err)

		assert.Equal(t, 404, errResp.StatusCode)
		assert.Equal(t, "Not Found", errResp.Error)
		assert.Equal(t, "Cannot GET /api/nonexistent-route", errResp.Message)
		assert.Equal(t, "/api/nonexistent-route", errResp.Path)
		assert.Equal(t, "GET", errResp.Method)
		assert.NotEmpty(t, errResp.Timestamp)
	})
}
