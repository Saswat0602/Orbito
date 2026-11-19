package httpx

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
)

// ISO8601Format defines the UTC timestamp format with 3-digit millisecond precision matching JavaScript toISOString().
const ISO8601Format = "2006-01-02T15:04:05.000Z"

// ISOTime wraps time.Time to enforce JSON marshaling and unmarshaling
// into exact ISO 8601 UTC with 3-digit millisecond precision (.000Z).
type ISOTime time.Time

// Time returns the underlying time.Time.
func (t ISOTime) Time() time.Time {
	return time.Time(t)
}

// MarshalJSON serializes the timestamp to ISO 8601 UTC with milliseconds.
func (t ISOTime) MarshalJSON() ([]byte, error) {
	ts := time.Time(t)
	if ts.IsZero() {
		return []byte("null"), nil
	}
	formatted := ts.UTC().Format(ISO8601Format)
	return json.Marshal(formatted)
}

// UnmarshalJSON deserializes an ISO 8601 string into ISOTime.
func (t *ISOTime) UnmarshalJSON(data []byte) error {
	str := string(data)
	if str == "null" || str == `""` {
		*t = ISOTime(time.Time{})
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		parsed, err = time.Parse(ISO8601Format, s)
		if err != nil {
			return err
		}
	}
	*t = ISOTime(parsed.UTC())
	return nil
}

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

// RespondText writes a plain text or HTML string matching Express res.send().
func RespondText(w http.ResponseWriter, statusCode int, text string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(text))
}

// RequestPath returns the exact request path including query parameters matching Express req.url.
func RequestPath(r *http.Request) string {
	if r.RequestURI != "" {
		return r.RequestURI
	}
	if r.URL != nil {
		uri := r.URL.RequestURI()
		if uri != "" {
			return uri
		}
		return r.URL.Path
	}
	return "/"
}

// RespondError maps an error to an AppError and serializes the exact NestJS error envelope.
func RespondError(w http.ResponseWriter, r *http.Request, err error) {
	appErr := apperror.Map(err)
	correlationID := ctxhelper.GetCorrelationID(r.Context())

	payload := ErrorResponse{
		StatusCode:    appErr.StatusCode,
		Error:         appErr.ErrorName,
		Message:       appErr.Message,
		Path:          RequestPath(r),
		Method:        r.Method,
		Timestamp:     FormatTimestamp(),
		CorrelationID: correlationID,
	}

	RespondJSON(w, appErr.StatusCode, payload)
}
