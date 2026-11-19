package shared

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/httpx"
	"github.com/orbito/api-go/internal/shared/pagination"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type sampleRequest struct {
	Email string `json:"email" validate:"required,email"`
	Name  string `json:"name" validate:"required,min=2"`
	Role  string `json:"role" validate:"required"`
}

func TestRespondError(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register?referral=abc", nil)
	// Inject correlation ID into request context
	req = req.WithContext(middleware.WithCorrelationID(req.Context(), "cid-test-1234"))

	rr := httptest.NewRecorder()
	testErr := apperror.BadRequest("Invalid username")

	httpx.RespondError(rr, req, testErr)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var resp httpx.ErrorResponse
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, 400, resp.StatusCode)
	assert.Equal(t, "Bad Request", resp.Error)
	assert.Equal(t, "Invalid username", resp.Message)
	assert.Equal(t, "/api/auth/register?referral=abc", resp.Path)
	assert.Equal(t, "POST", resp.Method)
	assert.Equal(t, "cid-test-1234", resp.CorrelationID)
	assert.NotEmpty(t, resp.Timestamp)
	// Verify timestamp format: must end with Z and have .000 millisecond precision
	assert.True(t, strings.HasSuffix(resp.Timestamp, "Z"))
	_, parseErr := time.Parse(httpx.ISO8601Format, resp.Timestamp)
	require.NoError(t, parseErr, "Timestamp must parse as exact ISO8601 with milliseconds")
}

func TestISOTimeSerialization(t *testing.T) {
	type TestPayload struct {
		CreatedAt httpx.ISOTime `json:"createdAt"`
	}

	fixedTime := time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC)
	payload := TestPayload{CreatedAt: httpx.ISOTime(fixedTime)}

	data, err := json.Marshal(payload)
	require.NoError(t, err)
	assert.Equal(t, `{"createdAt":"2026-11-19T12:00:00.000Z"}`, string(data))

	var decoded TestPayload
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	assert.Equal(t, fixedTime, decoded.CreatedAt.Time())
}

func TestBindJSONValidation(t *testing.T) {
	t.Run("valid payload succeeds", func(t *testing.T) {
		body := `{"email":"test@orbito.io","name":"Alex Mercer","role":"MANAGER"}`
		req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))

		var target sampleRequest
		err := httpx.BindJSON(req, &target)
		require.NoError(t, err)
		assert.Equal(t, "test@orbito.io", target.Email)
		assert.Equal(t, "Alex Mercer", target.Name)
	})

	t.Run("validation failure returns 400 with array of error messages", func(t *testing.T) {
		body := `{"email":"not-an-email","name":"A","role":""}`
		req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))

		var target sampleRequest
		err := httpx.BindJSON(req, &target)
		require.Error(t, err)

		appErr := apperror.Map(err)
		assert.Equal(t, 400, appErr.StatusCode)
		assert.Equal(t, "Bad Request", appErr.ErrorName)

		messages, ok := appErr.Message.([]string)
		require.True(t, ok, "Validation errors must be formatted as []string")
		assert.Len(t, messages, 3)
	})

	t.Run("disallows unknown fields and formats as property should not exist", func(t *testing.T) {
		body := `{"email":"test@orbito.io","name":"Valid","role":"MEMBER","hackerField":"injection"}`
		req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))

		var target sampleRequest
		err := httpx.BindJSON(req, &target)
		require.Error(t, err)

		appErr := apperror.Map(err)
		assert.Equal(t, 400, appErr.StatusCode)
		assert.Equal(t, "Bad Request", appErr.ErrorName)

		messages, ok := appErr.Message.([]string)
		require.True(t, ok)
		assert.Contains(t, messages, "property hackerField should not exist")
	})
}

func TestAppErrorMapper(t *testing.T) {
	t.Run("maps pgx.ErrNoRows to 404", func(t *testing.T) {
		appErr := apperror.Map(pgx.ErrNoRows)
		assert.Equal(t, 404, appErr.StatusCode)
		assert.Equal(t, "NotFound", appErr.ErrorName)
		assert.Equal(t, "Requested record was not found.", appErr.Message)
	})

	t.Run("maps postgres unique_violation to 409", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Code:   "23505",
			Detail: "Key (email)=(test@orbito.io) already exists.",
		}
		appErr := apperror.Map(pgErr)
		assert.Equal(t, 409, appErr.StatusCode)
		assert.Equal(t, "Conflict", appErr.ErrorName)
		assert.Contains(t, appErr.Message.(string), "Duplicate entry detected")
	})

	t.Run("maps postgres foreign_key_violation to 400", func(t *testing.T) {
		pgErr := &pgconn.PgError{
			Code:   "23503",
			Detail: "Key (project_id)=(123) is not present in table projects.",
		}
		appErr := apperror.Map(pgErr)
		assert.Equal(t, 400, appErr.StatusCode)
		assert.Equal(t, "ForeignKeyViolation", appErr.ErrorName)
		assert.Contains(t, appErr.Message.(string), "Foreign key reference constraint failed")
	})

	t.Run("maps generic error to 500", func(t *testing.T) {
		appErr := apperror.Map(errors.New("unexpected crash"))
		assert.Equal(t, 500, appErr.StatusCode)
		assert.Equal(t, "Internal Server Error", appErr.ErrorName)
	})
}

func TestPagination(t *testing.T) {
	t.Run("calculates pagination fields accurately", func(t *testing.T) {
		p := pagination.New(1, 10, 48)
		assert.Equal(t, 1, p.CurrentPage)
		assert.Equal(t, 5, p.TotalPages)
		assert.Equal(t, int64(48), p.TotalCount)
		assert.True(t, p.HasNextPage)
		assert.False(t, p.HasPrevPage)

		pLast := pagination.New(5, 10, 48)
		assert.Equal(t, 5, pLast.CurrentPage)
		assert.False(t, pLast.HasNextPage)
		assert.True(t, pLast.HasPrevPage)
	})

	t.Run("calculates correct SQL offset", func(t *testing.T) {
		assert.Equal(t, 0, pagination.Offset(1, 10))
		assert.Equal(t, 10, pagination.Offset(2, 10))
		assert.Equal(t, 40, pagination.Offset(5, 10))
	})

	t.Run("handles zero items gracefully", func(t *testing.T) {
		p := pagination.New(1, 10, 0)
		assert.Equal(t, 1, p.CurrentPage)
		assert.Equal(t, 0, p.TotalPages)
		assert.Equal(t, int64(0), p.TotalCount)
		assert.False(t, p.HasNextPage)
		assert.False(t, p.HasPrevPage)
	})
}

func TestParamsHelper(t *testing.T) {
	r := chi.NewRouter()
	var extractedID uuid.UUID
	var idErr error

	testUUID := uuid.New()

	r.Get("/items/{id}", func(w http.ResponseWriter, req *http.Request) {
		extractedID, idErr = httpx.URLParamUUID(req, "id")
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/items/"+testUUID.String(), nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	require.NoError(t, idErr)
	assert.Equal(t, testUUID, extractedID)
}
