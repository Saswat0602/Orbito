package sprints

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler processes HTTP requests for sprints.
type Handler struct {
	service Service
}

// NewHandler constructs a sprints HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

func extractAuth(r *http.Request) (uuid.UUID, bool, error) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		return uuid.Nil, false, apperror.Unauthorized("Unauthenticated")
	}
	isSuperAdmin := ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin)
	return userID, isSuperAdmin, nil
}

// Create handles POST /api/sprints.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req CreateSprintRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.Create(r.Context(), req, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, res)
}

// FindAll handles GET /api/sprints.
func (h *Handler) FindAll(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var projectID *uuid.UUID
	if s := r.URL.Query().Get("projectId"); s != "" {
		id, pErr := uuid.Parse(s)
		if pErr != nil {
			httpx.RespondError(w, r, apperror.BadRequest("Invalid projectId format"))
			return
		}
		projectID = &id
	}
	var status *string
	if s := r.URL.Query().Get("status"); s != "" {
		status = &s
	}
	res, err := h.service.FindAll(r.Context(), projectID, status, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindAllByProjectSlug handles GET /api/sprints/slug.
func (h *Handler) FindAllByProjectSlug(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	slug := r.URL.Query().Get("slug")
	var status *string
	if s := r.URL.Query().Get("status"); s != "" {
		status = &s
	}
	res, err := h.service.FindAllByProjectSlug(r.Context(), slug, status, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindBySlug handles GET /api/sprints/by-slug/{projectSlug}/{sprintSlug}.
func (h *Handler) FindBySlug(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	projectSlug := chi.URLParam(r, "projectSlug")
	sprintSlug := chi.URLParam(r, "sprintSlug")
	res, err := h.service.FindBySlug(r.Context(), projectSlug, sprintSlug, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindOne handles GET /api/sprints/{id}.
func (h *Handler) FindOne(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.FindOne(r.Context(), id, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetActiveSprint handles GET /api/sprints/project/{projectId}/active.
func (h *Handler) GetActiveSprint(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	projectID, err := httpx.URLParamUUID(r, "projectId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.GetActiveSprint(r.Context(), projectID, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// Update handles PATCH /api/sprints/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req UpdateSprintRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.Update(r.Context(), id, req, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// StartSprint handles PATCH /api/sprints/{id}/start.
func (h *Handler) StartSprint(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.StartSprint(r.Context(), id, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// CompleteSprint handles PATCH /api/sprints/{id}/complete.
func (h *Handler) CompleteSprint(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.CompleteSprint(r.Context(), id, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// Remove handles DELETE /api/sprints/{id}.
func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	if err := h.service.Remove(r.Context(), id, userID, isSuperAdmin); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ArchiveSprint handles PATCH /api/sprints/archive/{id}.
func (h *Handler) ArchiveSprint(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	if err := h.service.ArchiveSprint(r.Context(), id, userID, isSuperAdmin); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
