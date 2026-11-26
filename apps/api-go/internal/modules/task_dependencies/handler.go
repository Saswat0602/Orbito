package task_dependencies

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler handles HTTP requests for task dependencies.
type Handler struct {
	service Service
}

// NewHandler constructs a task_dependencies HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

func extractUserID(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		return uuid.Nil, apperror.Unauthorized("Unauthenticated")
	}
	return id, nil
}

// Create handles POST /api/task-dependencies.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	callerID, err := extractUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req CreateRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.Create(r.Context(), req, callerID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, res)
}

// CreateBulk handles POST /api/task-dependencies/bulk.
func (h *Handler) CreateBulk(w http.ResponseWriter, r *http.Request) {
	callerID, err := extractUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req BulkCreateRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.CreateBulk(r.Context(), req, callerID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, res)
}

// FindAll handles GET /api/task-dependencies.
func (h *Handler) FindAll(w http.ResponseWriter, r *http.Request) {
	var projectID *uuid.UUID
	if raw := httpx.QueryParam(r, "projectId", ""); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			httpx.RespondError(w, r, apperror.BadRequest("Validation failed (uuid is expected)"))
			return
		}
		projectID = &parsed
	}
	res, err := h.service.FindAll(r.Context(), projectID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetDependencyStats handles GET /api/task-dependencies/stats.
func (h *Handler) GetDependencyStats(w http.ResponseWriter, r *http.Request) {
	raw := httpx.QueryParam(r, "projectId", "")
	projectID, err := uuid.Parse(raw)
	if err != nil {
		httpx.RespondError(w, r, apperror.BadRequest("Validation failed (uuid is expected)"))
		return
	}
	res, err := h.service.GetDependencyStats(r.Context(), projectID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetTaskDependencies handles GET /api/task-dependencies/task/{taskId}.
func (h *Handler) GetTaskDependencies(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskId")
	res, err := h.service.GetTaskDependencies(r.Context(), taskID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetBlockedTasks handles GET /api/task-dependencies/task/{taskId}/blocked.
func (h *Handler) GetBlockedTasks(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskId")
	res, err := h.service.GetBlockedTasks(r.Context(), taskID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindOne handles GET /api/task-dependencies/{id}.
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

// Update handles PATCH /api/task-dependencies/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req UpdateRequest
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

// Remove handles DELETE /api/task-dependencies/{id}.
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

// RemoveByTasks handles DELETE /api/task-dependencies/tasks/{dependentTaskId}/{blockingTaskId}.
func (h *Handler) RemoveByTasks(w http.ResponseWriter, r *http.Request) {
	dependentTaskID := chi.URLParam(r, "dependentTaskId")
	blockingTaskID := chi.URLParam(r, "blockingTaskId")
	if err := h.service.RemoveByTasks(r.Context(), dependentTaskID, blockingTaskID); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
