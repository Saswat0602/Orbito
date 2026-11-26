package task_statuses

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes registers task status endpoints matching NestJS TaskStatusesController.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(authMw)

	r.Post("/", h.Create)
	r.Post("/from-project", h.CreateFromProject)
	r.Get("/", h.FindAll)
	r.Get("/project", h.FindByProject)
	r.Patch("/positions", h.UpdatePositions)
	r.Get("/deleted", h.FindDeleted)
	r.Get("/{id}", h.FindOne)
	r.Patch("/{id}/restore", h.Restore)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Remove)

	return r
}

