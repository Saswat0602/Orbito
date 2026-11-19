package health

import (
	"net/http"

	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler handles health check endpoints.
type Handler struct {
	service *Service
}

// NewHandler creates a new health Handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetHealth handles GET /api/health.
func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	status := h.service.CheckHealth(r.Context())
	statusCode := http.StatusOK
	if status.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	httpx.RespondJSON(w, statusCode, status)
}

// GetLiveness handles GET /api/health/liveness.
func (h *Handler) GetLiveness(w http.ResponseWriter, r *http.Request) {
	resp := LivenessStatus{
		Status:    "live",
		Timestamp: httpx.FormatTimestamp(),
	}

	httpx.RespondJSON(w, http.StatusOK, resp)
}

// GetReadiness handles GET /api/health/readiness.
func (h *Handler) GetReadiness(w http.ResponseWriter, r *http.Request) {
	status := h.service.CheckHealth(r.Context())
	statusCode := http.StatusOK
	if status.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	httpx.RespondJSON(w, statusCode, status)
}
