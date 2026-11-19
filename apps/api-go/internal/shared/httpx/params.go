package httpx

import (
	"fmt"
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

// URLParamUUID extracts and parses a path parameter as an RFC 4122 UUID.
func URLParamUUID(r *http.Request, key string) (uuid.UUID, error) {
	val := chi.URLParam(r, key)
	if val == "" {
		return uuid.Nil, apperror.BadRequest(fmt.Sprintf("Missing required path parameter: %s", key))
	}

	id, err := uuid.Parse(val)
	if err != nil {
		return uuid.Nil, apperror.BadRequest(fmt.Sprintf("Invalid UUID format for parameter '%s': %s", key, val))
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
