package gantt

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes registers endpoints matching NestJS GanttController.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(authMw)

	r.Get("/project/{projectId}/resources", h.GetResourceAllocation)
	r.Get("/project/{projectId}", h.GetProjectGanttData)
	r.Get("/sprint/{sprintId}", h.GetSprintGanttData)

	return r
}
