package health

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/orbito/api-go/internal/config"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHealthEndpoints(t *testing.T) {
	cfg := &config.Config{
		RedisHost: "127.0.0.1",
		RedisPort: 6379,
	}

	t.Run("liveness endpoint always returns 200 with status live", func(t *testing.T) {
		svc := NewService(nil, cfg)
		handler := NewHandler(svc)
		router := chi.NewRouter()
		router.Mount("/api/health", Routes(handler))

		req := httptest.NewRequest(http.MethodGet, "/api/health/liveness", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var resp LivenessStatus
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "live", resp.Status)
		assert.NotEmpty(t, resp.Timestamp)
	})

	t.Run("health endpoint returns 503 when database is down", func(t *testing.T) {
		svc := NewService(nil, cfg) // nil db simulates database down
		handler := NewHandler(svc)
		router := chi.NewRouter()
		router.Mount("/api/health", Routes(handler))

		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusServiceUnavailable, rr.Code)

		var resp Status
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "unhealthy", resp.Status)
		assert.Equal(t, "down", resp.Services.Database.Status)
	})

	t.Run("readiness endpoint returns 503 when database is down", func(t *testing.T) {
		svc := NewService(nil, cfg)
		handler := NewHandler(svc)
		router := chi.NewRouter()
		router.Mount("/api/health", Routes(handler))

		req := httptest.NewRequest(http.MethodGet, "/api/health/readiness", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusServiceUnavailable, rr.Code)

		var resp Status
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "unhealthy", resp.Status)
	})

	t.Run("health check succeeds against live running services", func(t *testing.T) {
		dbURL := "postgresql://root:root@localhost:5433/orbito"
		db, err := database.New(context.Background(), dbURL)
		if err != nil {
			t.Skipf("skipping live database health test: %v", err)
		}
		defer db.Close()

		svc := NewService(db, cfg)
		handler := NewHandler(svc)
		router := chi.NewRouter()
		router.Mount("/api/health", Routes(handler))

		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		var resp Status
		err = json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "healthy", resp.Status)
		assert.Equal(t, "up", resp.Services.Database.Status)
		assert.NotNil(t, resp.Services.Database.ResponseTime)
	})
}
