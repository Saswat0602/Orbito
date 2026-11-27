package task_watchers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes registers endpoints matching NestJS TaskWatchersController.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(authMw)

	r.Post("/", h.Create)
	r.Post("/watch", h.Create)
	r.Post("/unwatch", h.Unwatch)
	r.Post("/toggle", h.Toggle)
	r.Get("/", h.FindAll)
	r.Get("/stats", h.GetStats)
	r.Get("/task/{taskId}", h.GetTaskWatchers)
	r.Get("/user/{userId}/watched-tasks", h.GetUserWatchedTasks)
	r.Get("/check/{taskId}/{userId}", h.IsUserWatchingTask)
	r.Get("/{id}", h.FindOne)
	r.Delete("/{id}", h.Remove)

	return r
}
