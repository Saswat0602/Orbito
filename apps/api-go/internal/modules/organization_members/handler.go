package organization_members

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler handles HTTP requests for organization member endpoints.
type Handler struct {
	service Service
}

// NewHandler constructs an organization members HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

// Create handles POST /api/organization-members.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	var req AddMemberRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}

	res, err := h.service.AddMember(r.Context(), req, currentUserID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusCreated, res)
}

// FindAll handles GET /api/organization-members.
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
	res, err := h.service.FindAll(r.Context(), orgID, search, currentUserID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindAllBySlug handles GET /api/organization-members/slug.
func (h *Handler) FindAllBySlug(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	slug := r.URL.Query().Get("slug")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	search := r.URL.Query().Get("search")

	res, err := h.service.FindAllBySlug(r.Context(), slug, page, limit, search, currentUserID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindOne handles GET /api/organization-members/{id}.
func (h *Handler) FindOne(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid member ID"))
		return
	}

	res, err := h.service.FindOne(r.Context(), id, currentUserID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindByUserAndOrg handles GET /api/organization-members/user/{userId}/organization/{organizationId}.
func (h *Handler) FindByUserAndOrg(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	targetUserID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid user ID"))
		return
	}

	orgID, err := uuid.Parse(chi.URLParam(r, "organizationId"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid organization ID"))
		return
	}

	res, err := h.service.FindByUserAndOrg(r.Context(), targetUserID, orgID, currentUserID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// Update handles PATCH /api/organization-members/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid member ID"))
		return
	}

	var req UpdateRoleRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}

	res, err := h.service.UpdateRole(r.Context(), id, req.Role, currentUserID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// Remove handles DELETE /api/organization-members/{id}.
func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid member ID"))
		return
	}

	if err := h.service.Remove(r.Context(), id, currentUserID); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SetDefaultOrganization handles PATCH /api/organization-members/set-default.
func (h *Handler) SetDefaultOrganization(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	orgIDStr := r.URL.Query().Get("organizationId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid organizationId query parameter"))
		return
	}

	if err := h.service.SetDefaultOrganization(r.Context(), orgID, currentUserID); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, map[string]string{"message": "Default organization updated successfully"})
}
