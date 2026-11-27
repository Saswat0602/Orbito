package gantt

import (
	"net/http"

	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler handles HTTP requests for Gantt chart endpoints.
type Handler struct {
	service Service
}

// NewHandler constructs a gantt HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

// GetProjectGanttData handles GET /api/gantt/project/{projectId}.
func (h *Handler) GetProjectGanttData(w http.ResponseWriter, r *http.Request) {
	projectID, err := httpx.URLParamUUID(r, "projectId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.GetProjectGanttData(r.Context(), projectID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetSprintGanttData handles GET /api/gantt/sprint/{sprintId}.
func (h *Handler) GetSprintGanttData(w http.ResponseWriter, r *http.Request) {
	sprintID, err := httpx.URLParamUUID(r, "sprintId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.GetSprintGanttData(r.Context(), sprintID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetResourceAllocation handles GET /api/gantt/project/{projectId}/resources.
func (h *Handler) GetResourceAllocation(w http.ResponseWriter, r *http.Request) {
	projectID, err := httpx.URLParamUUID(r, "projectId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.GetResourceAllocation(r.Context(), projectID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

