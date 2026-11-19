package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/orbito/api-go/internal/config"
	"github.com/orbito/api-go/internal/shared/httpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildRouter(t *testing.T) {
	cfg := &config.Config{
		AppEnv:      "test",
		CORSOrigins: []string{"http://localhost:3001"},
		MaxFileSize: 10485760,
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
