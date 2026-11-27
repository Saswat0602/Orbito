package task_comments

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler handles HTTP requests for task comments.
type Handler struct {
	service Service
}

// NewHandler constructs a task_comments HTTP handler.
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

// Create handles POST /api/task-comments.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req CreateRequest
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

// FindAll handles GET /api/task-comments.
func (h *Handler) FindAll(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	taskID := httpx.QueryParam(r, "taskId", "")
	page := httpx.QueryParamInt(r, "page", 1)
	limit := httpx.QueryParamInt(r, "limit", 10)
	sort := httpx.QueryParam(r, "sort", "desc")
	res, err := h.service.FindAll(r.Context(), taskID, userID, page, limit, sort)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindWithMiddlePagination handles GET /api/task-comments/middle-pagination.
func (h *Handler) FindWithMiddlePagination(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	taskID := httpx.QueryParam(r, "taskId", "")
	page := httpx.QueryParamInt(r, "page", 1)
	limit := httpx.QueryParamInt(r, "limit", 5)
	oldest := httpx.QueryParamInt(r, "oldestCount", 2)
	newest := httpx.QueryParamInt(r, "newestCount", 2)
	res, err := h.service.FindWithMiddlePagination(r.Context(), taskID, userID, page, limit, oldest, newest)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// FindOne handles GET /api/task-comments/{id}.
func (h *Handler) FindOne(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.FindOne(r.Context(), id, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetReplies handles GET /api/task-comments/{id}/replies.
func (h *Handler) GetReplies(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.GetReplies(r.Context(), id, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetTaskCommentTree handles GET /api/task-comments/task/{taskId}/tree.
func (h *Handler) GetTaskCommentTree(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	taskID := chi.URLParam(r, "taskId")
	res, err := h.service.GetTaskCommentTree(r.Context(), taskID, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// Update handles PATCH /api/task-comments/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
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
	res, err := h.service.Update(r.Context(), id, req, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// Remove handles DELETE /api/task-comments/{id}.
func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	if err := h.service.Remove(r.Context(), id, userID); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
