package httpx

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// URLParam extracts a path parameter from chi routing context.
func URLParam(r *http.Request, key string) string {
	return chi.URLParam(r, key)
}

// URLParamUUID extracts and parses a path parameter as an RFC 4122 UUID matching NestJS ParseUUIDPipe.
func URLParamUUID(r *http.Request, key string) (uuid.UUID, error) {
	val := chi.URLParam(r, key)
	if val == "" {
		return uuid.Nil, apperror.BadRequest("Validation failed (uuid is expected)")
	}

	id, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, apperror.BadRequest("Validation failed (uuid is expected)")
	}

	return id, nil
}

// QueryParam extracts a query parameter with a fallback default.
func QueryParam(r *http.Request, key, fallback string) string {
	val := strings.TrimSpace(r.URL.Query().Get(key))
	if val == "" {
		return fallback
	}
	return val
}

// QueryParamInt extracts an integer query parameter with a fallback default.
func QueryParamInt(r *http.Request, key string, fallback int) int {
	valStr := strings.TrimSpace(r.URL.Query().Get(key))
	if valStr == "" {
		return fallback
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return fallback
	}
	return val
}
