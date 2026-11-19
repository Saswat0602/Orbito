package middleware

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/orbito/api-go/internal/platform/logger"
)

// panicErrorResponse represents the 500 error response structure matching NestJS GlobalExceptionFilter.
type panicErrorResponse struct {
	StatusCode    int    `json:"statusCode"`
	Error         string `json:"error"`
	Message       string `json:"message"`
	Path          string `json:"path"`
	Method        string `json:"method"`
	Timestamp     string `json:"timestamp"`
	CorrelationID string `json:"correlationId,omitempty"`
}

// Recoverer catches runtime panics, logs the stack trace, and returns a 500 JSON error envelope.
func Recoverer(isProduction bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					if rvr == http.ErrAbortHandler {
						panic(rvr)
					}

					ctx := r.Context()
					correlationID := GetCorrelationID(ctx)
					log := logger.FromContext(ctx)

					reqPath := r.RequestURI
					if reqPath == "" && r.URL != nil {
						reqPath = r.URL.RequestURI()
					}
					if reqPath == "" && r.URL != nil {
						reqPath = r.URL.Path
					}

					stack := string(debug.Stack())
					log.Error("panic recovered",
						slog.Any("error", rvr),
						slog.String("stack", stack),
						slog.String("path", reqPath),
						slog.String("method", r.Method),
						slog.String("correlation_id", correlationID),
					)

					message := "An unexpected internal error occurred."
					if !isProduction {
						message = fmt.Sprintf("Panic: %v", rvr)
					}

					resp := panicErrorResponse{
						StatusCode:    http.StatusInternalServerError,
						Error:         "InternalServerError",
						Message:       message,
						Path:          reqPath,
						Method:        r.Method,
						Timestamp:     time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
						CorrelationID: correlationID,
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_ = json.NewEncoder(w).Encode(resp)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
