package task_ranks

import (
	"net/http"

	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler processes HTTP endpoints for task_ranks.
type Handler struct {
	service Service
}

// NewHandler constructs a task_ranks HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

// Reorder handles PATCH /api/task-ranks/{taskId}/reorder.
func (h *Handler) Reorder(w http.ResponseWriter, r *http.Request) {
	taskID, err := httpx.URLParamUUID(r, "taskId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	var req ReorderRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	if err := h.service.Reorder(r.Context(), taskID, req); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// RebalanceAll handles GET /api/task-ranks/rebalance.
func (h *Handler) RebalanceAll(w http.ResponseWriter, _ *http.Request) {
	httpx.RespondJSON(w, http.StatusOK, RebalanceMessageResponse{
		Message: "Rebalance utility would go here",
	})
}

