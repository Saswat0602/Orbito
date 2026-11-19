package httpx

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// ISO8601Format defines the UTC timestamp format with 3-digit millisecond precision matching JavaScript toISOString().
const ISO8601Format = "2006-01-02T15:04:05.000Z"

// FormatTimestamp returns the current UTC timestamp formatted in exact ISO 8601 with milliseconds.
func FormatTimestamp(t ...time.Time) string {
	ts := time.Now().UTC()
	if len(t) > 0 && !t[0].IsZero() {
		ts = t[0].UTC()
	}
	return ts.Format(ISO8601Format)
}

// ErrorResponse represents the JSON error envelope matching NestJS GlobalExceptionFilter.
type ErrorResponse struct {
	StatusCode    int    `json:"statusCode"`
	Error         string `json:"error,omitempty"`
	Message       any    `json:"message"` // string or []string
	Path          string `json:"path"`
	Method        string `json:"method"`
	Timestamp     string `json:"timestamp"`
	CorrelationID string `json:"correlationId,omitempty"`
}

// RespondJSON marshals data to JSON and sets Content-Type header.
func RespondJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// RespondError maps an error to an AppError and serializes the exact NestJS error envelope.
func RespondError(w http.ResponseWriter, r *http.Request, err error) {
	appErr := apperror.Map(err)
	correlationID := middleware.GetCorrelationID(r.Context())

	payload := ErrorResponse{
		StatusCode:    appErr.StatusCode,
		Error:         appErr.ErrorName,
		Message:       appErr.Message,
		Path:          r.URL.Path,
		Method:        r.Method,
		Timestamp:     FormatTimestamp(),
		CorrelationID: correlationID,
	}

	RespondJSON(w, appErr.StatusCode, payload)
}
