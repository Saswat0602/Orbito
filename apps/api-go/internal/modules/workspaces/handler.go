package workspaces

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler processes HTTP requests for workspaces endpoints.
type Handler struct {
	service Service
}

// NewHandler constructs a workspaces HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

// Create handles POST /api/workspaces.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	var req CreateWorkspaceRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}

	res, err := h.service.Create(r.Context(), req, currentUserID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusCreated, res)
}

// FindAll handles GET /api/workspaces.
func (h *Handler) FindAll(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	var orgID *uuid.UUID
	orgIDStr := r.URL.Query().Get("organizationId")
	if orgIDStr != "" {
		parsed, pErr := uuid.Parse(orgIDStr)
		if pErr == nil {
			orgID = &parsed
		}
	}

	search := r.URL.Query().Get("search")
	res, err := h.service.FindAll(r.Context(), currentUserID, orgID, search)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindArchived handles GET /api/workspaces/archived.
func (h *Handler) FindArchived(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	orgIDStr := r.URL.Query().Get("organizationId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("organizationId query parameter required"))
		return
	}

	res, err := h.service.FindArchived(r.Context(), orgID, currentUserID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindTree handles GET /api/workspaces/tree.
func (h *Handler) FindTree(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	orgIDStr := r.URL.Query().Get("organizationId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("organizationId query parameter required"))
		return
	}

	res, err := h.service.GetTree(r.Context(), orgID, currentUserID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindBySlugOnly handles GET /api/workspaces/by-slug/{slug}.
func (h *Handler) FindBySlugOnly(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	slug := chi.URLParam(r, "slug")
	res, err := h.service.FindBySlugOnly(r.Context(), slug, currentUserID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindBySlug handles GET /api/workspaces/organization/{organizationId}/slug/{slug}.
func (h *Handler) FindBySlug(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	orgID, err := uuid.Parse(chi.URLParam(r, "organizationId"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid organization ID"))
		return
	}

	slug := chi.URLParam(r, "slug")
	res, err := h.service.FindBySlug(r.Context(), orgID, slug, currentUserID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindOne handles GET /api/workspaces/{id}.
func (h *Handler) FindOne(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid workspace ID"))
		return
	}

	res, err := h.service.FindOne(r.Context(), id, currentUserID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// Update handles PATCH /api/workspaces/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid workspace ID"))
		return
	}

	var req UpdateWorkspaceRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}

	res, err := h.service.Update(r.Context(), id, req, currentUserID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// Archive handles PATCH /api/workspaces/archive/{id}.
func (h *Handler) Archive(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid workspace ID"))
		return
	}

	if err := h.service.Archive(r.Context(), id, currentUserID); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Unarchive handles PATCH /api/workspaces/unarchive/{id}.
func (h *Handler) Unarchive(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid workspace ID"))
		return
	}

	if err := h.service.Unarchive(r.Context(), id, currentUserID); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Delete handles DELETE /api/workspaces/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid workspace ID"))
		return
	}

	if err := h.service.Delete(r.Context(), id, currentUserID); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, map[string]string{"message": "Workspace deleted successfully"})
}

// GetMembers handles GET /api/workspaces/{id}/members.
func (h *Handler) GetMembers(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid workspace ID"))
		return
	}

	res, err := h.service.GetMembers(r.Context(), id, currentUserID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetCharts handles GET /api/workspaces/{id}/charts.
func (h *Handler) GetCharts(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid workspace ID"))
		return
	}

	res, err := h.service.GetCharts(r.Context(), id, currentUserID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}
