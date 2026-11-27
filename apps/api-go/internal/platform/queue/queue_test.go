package queue

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/orbito/api-go/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntimeInMemoryFallbackAndRetries(t *testing.T) {
	rt := NewRuntime(&config.Config{MaxConcurrentJobs: 2, JobRetryAttempts: 2}, nil)
	require.NoError(t, rt.Start(context.Background()))
	defer func() { _ = rt.Close() }()

	var attempts atomic.Int32
	done := make(chan struct{})
	rt.RegisterHandler("test:job", func(_ context.Context, payload []byte) error {
		var data map[string]string
		require.NoError(t, json.Unmarshal(payload, &data))
		assert.Equal(t, "ok", data["val"])
		if attempts.Add(1) == 1 {
			return errors.New("transient")
		}
		close(done)
		return nil
	})

	require.NoError(t, rt.Enqueue(context.Background(), QueueAutomation, "test:job", map[string]string{"val": "ok"}))

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for retried job completion")
	}

	health := rt.Health()
	assert.Equal(t, "in-memory", health.Backend)
	assert.Equal(t, int64(1), health.Stats.TotalJobs)
	assert.Equal(t, int64(1), health.Stats.CompletedJobs)
	assert.Equal(t, int64(1), health.Stats.FailedJobs)
	assert.Contains(t, health.Queues, QueueAutomation)
	assert.NotEmpty(t, health.Timestamp)
}

func TestQueueHTTPHandlersAndBasicAuth(t *testing.T) {
	rt := NewRuntime(&config.Config{MaxConcurrentJobs: 1, JobRetryAttempts: 1}, nil)
	require.NoError(t, rt.Start(context.Background()))
	defer func() { _ = rt.Close() }()

	r := chi.NewRouter()
	MountRoutes(r, rt, "admin@orbito.dev", "s3cr3t")

	req := httptest.NewRequest(http.MethodGet, "/health/queue", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	testReq := httptest.NewRequest(http.MethodGet, "/queues/test", nil)
	testRec := httptest.NewRecorder()
	r.ServeHTTP(testRec, testReq)
	assert.Equal(t, http.StatusOK, testRec.Code)

	redirReq := httptest.NewRequest(http.MethodGet, "/queues/health", nil)
	redirRec := httptest.NewRecorder()
	r.ServeHTTP(redirRec, redirReq)
	assert.Equal(t, http.StatusTemporaryRedirect, redirRec.Code)
	assert.Equal(t, "/api/health/queue", redirRec.Header().Get("Location"))

	unauthReq := httptest.NewRequest(http.MethodGet, "/queues/", nil)
	unauthRec := httptest.NewRecorder()
	r.ServeHTTP(unauthRec, unauthReq)
	assert.Equal(t, http.StatusUnauthorized, unauthRec.Code)

	authReq := httptest.NewRequest(http.MethodGet, "/queues/", nil)
	creds := base64.StdEncoding.EncodeToString([]byte("admin@orbito.dev:s3cr3t"))
	authReq.Header.Set("Authorization", "Basic "+creds)
	authRec := httptest.NewRecorder()
	r.ServeHTTP(authRec, authReq)
	assert.Equal(t, http.StatusOK, authRec.Code)
}
