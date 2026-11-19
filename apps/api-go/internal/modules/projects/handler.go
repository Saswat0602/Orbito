package projects

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler processes HTTP requests for projects.
type Handler struct {
	service Service
}

// NewHandler constructs a projects HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

// Create handles POST /api/projects.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	var req CreateProjectRequest
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

// FindAll handles GET /api/projects.
func (h *Handler) FindAll(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	q := r.URL.Query()
	var wsID *uuid.UUID
	if wsStr := q.Get("workspaceId"); wsStr != "" {
		if id, parseErr := uuid.Parse(wsStr); parseErr == nil {
			wsID = &id
		}
	}

	var status, priority, search *string
	if s := q.Get("status"); s != "" {
		status = &s
	}
	if p := q.Get("priority"); p != "" {
		priority = &p
	}
	if sc := q.Get("search"); sc != "" {
		search = &sc
	}

	page := 1
	if pStr := q.Get("page"); pStr != "" {
		if pVal, pErr := strconv.Atoi(pStr); pErr == nil && pVal > 0 {
			page = pVal
		}
	}
	pageSize := 10
	if psStr := q.Get("pageSize"); psStr != "" {
		if psVal, psErr := strconv.Atoi(psStr); psErr == nil && psVal > 0 {
			pageSize = psVal
		}
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	res, err := h.service.FindAll(r.Context(), wsID, status, priority, search, page, pageSize, currentUserID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindByOrganizationID handles GET /api/projects/by-organization.
func (h *Handler) FindByOrganizationID(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	q := r.URL.Query()
	orgStr := q.Get("organizationId")
	orgID, err := uuid.Parse(orgStr)
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("organizationId query parameter is required and must be a valid UUID"))
		return
	}

	var wsID *uuid.UUID
	if wsStr := q.Get("workspaceId"); wsStr != "" {
		if id, parseErr := uuid.Parse(wsStr); parseErr == nil {
			wsID = &id
		}
	}

	var status, priority, search *string
	if s := q.Get("status"); s != "" {
		status = &s
	}
	if p := q.Get("priority"); p != "" {
		priority = &p
	}
	if sc := q.Get("search"); sc != "" {
		search = &sc
	}

	page := 1
	if pStr := q.Get("page"); pStr != "" {
		if pVal, pErr := strconv.Atoi(pStr); pErr == nil && pVal > 0 {
			page = pVal
		}
	}
	pageSize := 10
	if psStr := q.Get("pageSize"); psStr != "" {
		if psVal, psErr := strconv.Atoi(psStr); psErr == nil && psVal > 0 {
			pageSize = psVal
		}
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	res, err := h.service.FindByOrganizationID(r.Context(), orgID, wsID, status, priority, search, page, pageSize, currentUserID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindOne handles GET /api/projects/{id}.
func (h *Handler) FindOne(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid project ID"))
		return
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	res, err := h.service.FindOne(r.Context(), id, currentUserID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetBySlug handles GET /api/projects/by-slug/{slug}.
func (h *Handler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	slug := chi.URLParam(r, "slug")
	if slug == "" {
		httpx.RespondError(w, r, apperror.BadRequest("Project slug is required"))
		return
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	res, err := h.service.GetBySlug(r.Context(), slug, currentUserID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindByKey handles GET /api/projects/workspace/{workspaceId}/key/{key}.
func (h *Handler) FindByKey(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	wsID, err := uuid.Parse(chi.URLParam(r, "workspaceId"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid workspace ID"))
		return
	}
	key := chi.URLParam(r, "key")
	if key == "" {
		httpx.RespondError(w, r, apperror.BadRequest("Project key is required"))
		return
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	res, err := h.service.FindByKey(r.Context(), wsID, key, currentUserID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// Update handles PATCH /api/projects/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid project ID"))
		return
	}

	var req UpdateProjectRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	res, err := h.service.Update(r.Context(), id, req, currentUserID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// Archive handles PATCH /api/projects/archive/{id}.
func (h *Handler) Archive(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid project ID"))
		return
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	if err := h.service.Archive(r.Context(), id, currentUserID, isSuperAdmin); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Unarchive handles PATCH /api/projects/unarchive/{id}.
func (h *Handler) Unarchive(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid project ID"))
		return
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	if err := h.service.Unarchive(r.Context(), id, currentUserID, isSuperAdmin); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Delete handles DELETE /api/projects/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid project ID"))
		return
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	if err := h.service.Delete(r.Context(), id, currentUserID, isSuperAdmin); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, map[string]string{"message": "Project deleted successfully"})
}

// GetStats handles GET /api/projects/{id}/stats.
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid project ID"))
		return
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	res, err := h.service.GetStats(r.Context(), id, currentUserID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetCharts handles GET /api/projects/{slug}/charts.
func (h *Handler) GetCharts(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	slug := chi.URLParam(r, "slug")
	if slug == "" {
		httpx.RespondError(w, r, apperror.BadRequest("Project slug is required"))
		return
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	res, err := h.service.GetCharts(r.Context(), slug, currentUserID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}
