package task_statuses

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler handles HTTP endpoints for task statuses.
type Handler struct {
	service Service
}

// NewHandler constructs a task_statuses HTTP handler.
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

// Create handles POST /api/task-statuses.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req CreateTaskStatusRequest
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

// CreateFromProject handles POST /api/task-statuses/from-project.
func (h *Handler) CreateFromProject(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req CreateTaskStatusFromProjectRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.CreateFromProject(r.Context(), req, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, res)
}

// FindAll handles GET /api/task-statuses.
func (h *Handler) FindAll(w http.ResponseWriter, r *http.Request) {
	var wfID, orgID *uuid.UUID
	if s := r.URL.Query().Get("workflowId"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			httpx.RespondError(w, r, apperror.BadRequest("Invalid workflowId format"))
			return
		}
		wfID = &id
	}
	if s := r.URL.Query().Get("organizationId"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			httpx.RespondError(w, r, apperror.BadRequest("Invalid organizationId format"))
			return
		}
		orgID = &id
	}
	res, err := h.service.FindAll(r.Context(), wfID, orgID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindByProject handles GET /api/task-statuses/project.
func (h *Handler) FindByProject(w http.ResponseWriter, r *http.Request) {
	rawID := r.URL.Query().Get("projectId")
	projectID, err := uuid.Parse(rawID)
	if err != nil {
		httpx.RespondJSON(w, http.StatusOK, []TaskStatusResponse{})
		return
	}
	res, err := h.service.FindByProject(r.Context(), projectID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// UpdatePositions handles PATCH /api/task-statuses/positions.
func (h *Handler) UpdatePositions(w http.ResponseWriter, r *http.Request) {
	userID, isSuperAdmin, err := extractAuth(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req UpdatePositionsRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.UpdatePositions(r.Context(), req, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindDeleted handles GET /api/task-statuses/deleted.
func (h *Handler) FindDeleted(w http.ResponseWriter, r *http.Request) {
	var wfID *uuid.UUID
	if s := r.URL.Query().Get("workflowId"); s != "" {
		id, err := uuid.Parse(s)
		if err != nil {
			httpx.RespondError(w, r, apperror.BadRequest("Invalid workflowId format"))
			return
		}
		wfID = &id
	}
	res, err := h.service.FindDeleted(r.Context(), wfID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindOne handles GET /api/task-statuses/{id}.
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

// Restore handles PATCH /api/task-statuses/{id}/restore.
func (h *Handler) Restore(w http.ResponseWriter, r *http.Request) {
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
	res, err := h.service.Restore(r.Context(), id, userID, isSuperAdmin)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// Update handles PATCH /api/task-statuses/{id}.
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
	var req UpdateTaskStatusRequest
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

// Remove handles DELETE /api/task-statuses/{id}.
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
	w.WriteHeader(http.StatusOK)
}
