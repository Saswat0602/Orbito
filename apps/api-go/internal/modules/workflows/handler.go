package workflows

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler processes HTTP requests for workflows.
type Handler struct {
	service Service
}

// NewHandler constructs a workflows HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

// Create handles POST /api/workflows.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	var req CreateWorkflowRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	res, err := h.service.Create(r.Context(), req, currentUserID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusCreated, res)
}

// FindAll handles GET /api/workflows.
func (h *Handler) FindAll(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	var orgID *uuid.UUID
	if orgStr := r.URL.Query().Get("organizationId"); orgStr != "" {
		parsed, pErr := uuid.Parse(orgStr)
		if pErr != nil {
			httpx.RespondError(w, r, apperror.BadRequest("Invalid organization ID format"))
			return
		}
		orgID = &parsed
	}

	res, err := h.service.FindAll(r.Context(), orgID, currentUserID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindByOrganizationSlug handles GET /api/workflows/slug.
func (h *Handler) FindByOrganizationSlug(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)
	slug := r.URL.Query().Get("slug")

	res, err := h.service.FindByOrganizationSlug(r.Context(), slug, currentUserID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindOne handles GET /api/workflows/{id}.
func (h *Handler) FindOne(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	wfID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Validation failed (uuid is expected)"))
		return
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	res, err := h.service.FindOne(r.Context(), wfID, currentUserID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetDefaultWorkflow handles GET /api/workflows/organization/{organizationId}/default.
func (h *Handler) GetDefaultWorkflow(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	orgID, err := uuid.Parse(chi.URLParam(r, "organizationId"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Validation failed (uuid is expected)"))
		return
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	res, err := h.service.GetDefaultWorkflow(r.Context(), orgID, currentUserID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// Update handles PATCH /api/workflows/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	wfID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Validation failed (uuid is expected)"))
		return
	}

	var req UpdateWorkflowRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	res, err := h.service.Update(r.Context(), wfID, req, currentUserID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// MakeDefault handles PATCH /api/workflows/{id}/set-default.
func (h *Handler) MakeDefault(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	wfID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Validation failed (uuid is expected)"))
		return
	}

	var req MakeDefaultWorkflowRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}

	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Validation failed (uuid is expected)"))
		return
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	res, err := h.service.MakeDefault(r.Context(), wfID, orgID, currentUserID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// Delete handles DELETE /api/workflows/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	wfID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Validation failed (uuid is expected)"))
		return
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	if err := h.service.Delete(r.Context(), wfID, currentUserID, isSuperAdmin); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, map[string]bool{"success": true})
}
