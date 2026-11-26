package task_attachments

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes registers endpoints matching NestJS TaskAttachmentsController.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(authMw)

	r.Post("/upload/{taskId}", h.Upload)
	r.Get("/", h.FindAll)
	r.Get("/stats", h.GetStats)
	r.Get("/task/{taskId}", h.GetTaskAttachments)
	r.Get("/{id}/download", h.Download)
	r.Get("/{id}/preview", h.Preview)
	r.Get("/{id}", h.FindOne)
	r.Delete("/{id}", h.Remove)

	return r
}
