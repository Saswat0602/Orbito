package task_dependencies

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes registers endpoints matching NestJS TaskDependenciesController.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(authMw)

	r.Post("/", h.Create)
	r.Post("/bulk", h.CreateBulk)
	r.Get("/", h.FindAll)
	r.Get("/stats", h.GetDependencyStats)
	r.Get("/task/{taskId}", h.GetTaskDependencies)
	r.Get("/task/{taskId}/blocked", h.GetBlockedTasks)
	r.Get("/{id}", h.FindOne)
	r.Patch("/{id}", h.Update)
	r.Delete("/tasks/{dependentTaskId}/{blockingTaskId}", h.RemoveByTasks)
	r.Delete("/{id}", h.Remove)

	return r
}

