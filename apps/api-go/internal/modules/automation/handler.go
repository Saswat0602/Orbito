package automation

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler handles HTTP endpoints under /api/automation.
type Handler struct {
	service Service
}

// NewHandler creates an automation HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

// Create handles POST /api/automation/rules.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Authentication required"))
		return
	}
	var req CreateRuleRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.Create(r.Context(), req, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, res)
}

// FindAll handles GET /api/automation/rules.
func (h *Handler) FindAll(w http.ResponseWriter, r *http.Request) {
	orgID := httpx.QueryParam(r, "organizationId", "")
	wsID := httpx.QueryParam(r, "workspaceId", "")
	projID := httpx.QueryParam(r, "projectId", "")

	res, err := h.service.FindAll(r.Context(), orgID, wsID, projID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindOne handles GET /api/automation/rules/{id}.
func (h *Handler) FindOne(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.FindOne(r.Context(), id)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// Update handles PATCH /api/automation/rules/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req UpdateRuleRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.Update(r.Context(), id, req)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// ToggleStatus handles PATCH /api/automation/rules/{id}/toggle.
func (h *Handler) ToggleStatus(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.ToggleStatus(r.Context(), id)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// Remove handles DELETE /api/automation/rules/{id}.
func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	if err := h.service.Remove(r.Context(), id); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// GetStats handles GET /api/automation/rules/{id}/stats.
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.GetExecutionStats(r.Context(), id)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// Routes builds the /api/automation subrouter behind JWT authentication.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) chi.Router {
	ar := chi.NewRouter()
	ar.Use(authMw)
	ar.Post("/rules", h.Create)
	ar.Get("/rules", h.FindAll)
	ar.Get("/rules/{id}", h.FindOne)
	ar.Patch("/rules/{id}", h.Update)
	ar.Patch("/rules/{id}/toggle", h.ToggleStatus)
	ar.Delete("/rules/{id}", h.Remove)
	ar.Get("/rules/{id}/stats", h.GetStats)
	return ar
}

// MountRoutes registers /api/automation endpoints behind JWT authentication.
func MountRoutes(r chi.Router, h *Handler, authMw func(http.Handler) http.Handler) {
	r.Mount("/automation", Routes(h, authMw))
}
