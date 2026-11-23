package project_members

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler processes HTTP requests for project members.
type Handler struct {
	service Service
}

// NewHandler constructs a project members HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

// Create handles POST /api/project-members.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	var req CreateProjectMemberRequest
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

// FindAll handles GET /api/project-members.
func (h *Handler) FindAll(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	q := r.URL.Query()
	projStr := q.Get("projectId")
	if projStr == "" {
		httpx.RespondError(w, r, apperror.BadRequest("projectId query parameter is required"))
		return
	}
	projID, err := uuid.Parse(projStr)
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid projectId"))
		return
	}

	var search *string
	if s := q.Get("search"); s != "" {
		search = &s
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	res, err := h.service.ListMembers(r.Context(), projID, search, currentUserID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindAllByProject handles GET /api/project-members/project/{id}.
func (h *Handler) FindAllByProject(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	projID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid project ID"))
		return
	}

	var search *string
	if s := r.URL.Query().Get("search"); s != "" {
		search = &s
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	res, err := h.service.ListMembers(r.Context(), projID, search, currentUserID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindOne handles GET /api/project-members/{id}.
func (h *Handler) FindOne(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid project member ID"))
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

// FindByUserAndProject handles GET /api/project-members/user/{userId}/project/{projectId}.
func (h *Handler) FindByUserAndProject(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	targetUserID, err := uuid.Parse(chi.URLParam(r, "userId"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid userId"))
		return
	}
	projID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid projectId"))
		return
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	res, err := h.service.FindByUserAndProject(r.Context(), targetUserID, projID, currentUserID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// Update handles PATCH /api/project-members/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid project member ID"))
		return
	}

	var req UpdateProjectMemberRoleRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	res, err := h.service.UpdateRole(r.Context(), id, req, currentUserID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// Remove handles DELETE /api/project-members/{id}.
func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid project member ID"))
		return
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	if err := h.service.Remove(r.Context(), id, currentUserID, isSuperAdmin); err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetStats handles GET /api/project-members/project/{projectId}/stats.
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	currentUserID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthenticated"))
		return
	}

	projID, err := uuid.Parse(chi.URLParam(r, "projectId"))
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Invalid project ID"))
		return
	}

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	res, err := h.service.GetStats(r.Context(), projID, currentUserID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// ListByWorkspace handles GET /api/project-members/workspace/{workspaceId}.
func (h *Handler) ListByWorkspace(w http.ResponseWriter, r *http.Request) {
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

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	res, err := h.service.ListByWorkspace(r.Context(), wsID, currentUserID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}

// ListUserProjects handles GET /api/project-members/user/{userId}/projects.
func (h *Handler) ListUserProjects(w http.ResponseWriter, r *http.Request) {
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

	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)

	res, err := h.service.ListUserProjects(r.Context(), targetUserID, currentUserID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, res)
}
