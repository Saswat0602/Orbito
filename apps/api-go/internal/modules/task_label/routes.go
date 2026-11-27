package task_label

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes registers endpoints matching NestJS TaskLabelsController.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(authMw)

	r.Post("/", h.AssignLabel)
	r.Post("/assign-multiple", h.AssignMultiple)
	r.Get("/", h.FindAll)
	r.Delete("/{taskId}/{labelId}", h.Remove)

	return r
}
