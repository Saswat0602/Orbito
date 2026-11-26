package task_watchers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler handles HTTP requests for task watchers.
type Handler struct {
	service Service
}

// NewHandler constructs a task_watchers HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

func extractCaller(r *http.Request) (CallerContext, error) {
	uid, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		return CallerContext{}, apperror.Unauthorized("Unauthenticated")
	}
	return CallerContext{
		UserID:       uid,
		IsSuperAdmin: ctxhelper.GetUserRole(r.Context()) == string(middleware.RoleSuperAdmin),
	}, nil
}

// Create handles POST /api/task-watchers and POST /api/task-watchers/watch.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.Create(r.Context(), req)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, res)
}

// Unwatch handles POST /api/task-watchers/unwatch.
func (h *Handler) Unwatch(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	if err := h.service.Unwatch(r.Context(), req); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Toggle handles POST /api/task-watchers/toggle.
func (h *Handler) Toggle(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.ToggleWatch(r.Context(), req)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, res)
}

// FindAll handles GET /api/task-watchers.
func (h *Handler) FindAll(w http.ResponseWriter, r *http.Request) {
	caller, err := extractCaller(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	taskID := httpx.QueryParam(r, "taskId", "")
	var userID *uuid.UUID
	if raw := httpx.QueryParam(r, "userId", ""); raw != "" {
		parsed, pErr := uuid.Parse(raw)
		if pErr != nil {
			httpx.RespondError(w, r, apperror.BadRequest("Validation failed (uuid is expected)"))
			return
		}
		userID = &parsed
	}
	res, err := h.service.FindAll(r.Context(), caller, taskID, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetStats handles GET /api/task-watchers/stats.
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	var taskID, userID *uuid.UUID
	if raw := httpx.QueryParam(r, "taskId", ""); raw != "" {
		if parsed, err := uuid.Parse(raw); err == nil {
			taskID = &parsed
		}
	}
	if raw := httpx.QueryParam(r, "userId", ""); raw != "" {
		if parsed, err := uuid.Parse(raw); err == nil {
			userID = &parsed
		}
	}
	res, err := h.service.GetWatcherStats(r.Context(), taskID, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetTaskWatchers handles GET /api/task-watchers/task/{taskId}.
func (h *Handler) GetTaskWatchers(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskId")
	res, err := h.service.GetTaskWatchers(r.Context(), taskID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetUserWatchedTasks handles GET /api/task-watchers/user/{userId}/watched-tasks.
func (h *Handler) GetUserWatchedTasks(w http.ResponseWriter, r *http.Request) {
	userID, err := httpx.URLParamUUID(r, "userId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.GetUserWatchedTasks(r.Context(), userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// IsUserWatchingTask handles GET /api/task-watchers/check/{taskId}/{userId}.
func (h *Handler) IsUserWatchingTask(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskId")
	userID, err := httpx.URLParamUUID(r, "userId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.IsUserWatchingTask(r.Context(), userID, taskID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindOne handles GET /api/task-watchers/{id}.
func (h *Handler) FindOne(w http.ResponseWriter, r *http.Request) {
	caller, err := extractCaller(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.FindOne(r.Context(), id, caller)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// Remove handles DELETE /api/task-watchers/{id}.
func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	caller, err := extractCaller(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	if err := h.service.Remove(r.Context(), id, caller.UserID); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
