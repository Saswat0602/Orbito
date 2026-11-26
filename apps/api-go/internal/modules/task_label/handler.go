package task_label

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler handles HTTP requests for task label assignments.
type Handler struct {
	service Service
}

// NewHandler constructs a task_label HTTP handler.
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

// AssignLabel handles POST /api/task-labels.
func (h *Handler) AssignLabel(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req AssignRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.AssignLabel(r.Context(), req, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, res)
}

// AssignMultiple handles POST /api/task-labels/assign-multiple.
func (h *Handler) AssignMultiple(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req AssignMultipleRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	res, err := h.service.AssignMultiple(r.Context(), req, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusCreated, res)
}

// FindAll handles GET /api/task-labels.
func (h *Handler) FindAll(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.FindAll(r.Context(), userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// Remove handles DELETE /api/task-labels/{taskId}/{labelId}.
func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	userID, err := extractUserID(r)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	taskID := chi.URLParam(r, "taskId")
	labelID, err := httpx.URLParamUUID(r, "labelId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.Remove(r.Context(), taskID, labelID, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}
