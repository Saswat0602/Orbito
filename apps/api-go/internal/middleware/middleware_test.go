package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/orbito/api-go/internal/platform/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestIDMiddleware(t *testing.T) {
	t.Run("preserves existing correlation ID", func(t *testing.T) {
		handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cid := GetCorrelationID(r.Context())
			assert.Equal(t, "custom-cid-12345", cid)
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		req.Header.Set("x-correlation-id", "custom-cid-12345")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "custom-cid-12345", rr.Header().Get("x-correlation-id"))
		assert.Equal(t, "custom-cid-12345", rr.Header().Get("x-request-id"))
	})

	t.Run("generates UUID when correlation ID is absent", func(t *testing.T) {
		var capturedCID string
		handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedCID = GetCorrelationID(r.Context())
			assert.NotEmpty(t, capturedCID)
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NotEmpty(t, rr.Header().Get("x-correlation-id"))
		assert.Equal(t, capturedCID, rr.Header().Get("x-correlation-id"))
	})
}

func TestLoggerMiddleware(t *testing.T) {
	var buf bytes.Buffer
	testLogger := logger.New(false, &buf)

	stack := RequestID(Logger(testLogger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"success":true}`))
	})))

	req := httptest.NewRequest(http.MethodPost, "/api/tasks", strings.NewReader(`{}`))
	req.Header.Set("x-correlation-id", "log-test-cid")
	rr := httptest.NewRecorder()

	stack.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var logEntry map[string]any
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, "http request", logEntry["msg"])
	assert.Equal(t, "POST", logEntry["method"])
	assert.Equal(t, "/api/tasks", logEntry["path"])
	assert.Equal(t, float64(201), logEntry["status"])
	assert.Equal(t, "log-test-cid", logEntry["correlation_id"])
}

func TestRecovererMiddleware(t *testing.T) {
	var buf bytes.Buffer
	testLogger := logger.New(false, &buf)

	stack := RequestID(Logger(testLogger)(Recoverer(true)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("database connection failure")
	}))))

	req := httptest.NewRequest(http.MethodGet, "/api/panic-test", nil)
	req.Header.Set("x-correlation-id", "panic-cid-999")
	rr := httptest.NewRecorder()

	stack.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var errResp panicErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &errResp)
	require.NoError(t, err)

	assert.Equal(t, 500, errResp.StatusCode)
	assert.Equal(t, "InternalServerError", errResp.Error)
	assert.Equal(t, "An unexpected internal error occurred.", errResp.Message)
	assert.Equal(t, "/api/panic-test", errResp.Path)
	assert.Equal(t, "GET", errResp.Method)
	assert.Equal(t, "panic-cid-999", errResp.CorrelationID)
	assert.NotEmpty(t, errResp.Timestamp)
}

func TestCORSMiddleware(t *testing.T) {
	whitelist := []string{"http://localhost:3001", "https://app.orbito.io"}
	handler := CORS(whitelist)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("whitelisted origin allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/workspaces", nil)
		req.Header.Set("Origin", "http://localhost:3001")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "http://localhost:3001", rr.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", rr.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("preflight OPTIONS request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/workspaces", nil)
		req.Header.Set("Origin", "https://app.orbito.io")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		assert.Equal(t, "https://app.orbito.io", rr.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, rr.Header().Get("Access-Control-Allow-Methods"), "POST")
	})

	t.Run("non-whitelisted origin not echoed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/workspaces", nil)
		req.Header.Set("Origin", "https://malicious-site.com")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestSecureHeadersMiddleware(t *testing.T) {
	handler := SecureHeaders(true)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, "nosniff", rr.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "SAMEORIGIN", rr.Header().Get("X-Frame-Options"))
	assert.Equal(t, "0", rr.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "no-referrer", rr.Header().Get("Referrer-Policy"))
	assert.NotEmpty(t, rr.Header().Get("Strict-Transport-Security"))
}

func TestBodyLimitMiddleware(t *testing.T) {
	limit := int64(100) // 100 bytes limit
	handler := BodyLimit(limit)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "request entity too large", http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("body within limit succeeds", func(t *testing.T) {
		smallBody := strings.Repeat("a", 50)
		req := httptest.NewRequest(http.MethodPost, "/api/upload", strings.NewReader(smallBody))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("body exceeding limit fails", func(t *testing.T) {
		largeBody := strings.Repeat("a", 200)
		req := httptest.NewRequest(http.MethodPost, "/api/upload", strings.NewReader(largeBody))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusRequestEntityTooLarge, rr.Code)
	})
}
