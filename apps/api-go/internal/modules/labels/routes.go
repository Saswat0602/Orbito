package labels

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes registers label endpoints matching NestJS LabelsController.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(authMw)

	r.Post("/", h.Create)
	r.Get("/", h.FindAll)
	r.Post("/assign", h.AssignLabelToTask)
	r.Post("/assign-multiple", h.AssignMultipleLabelsToTask)
	r.Get("/task/{taskId}", h.GetTaskLabels)
	r.Delete("/task/{taskId}/label/{labelId}", h.RemoveLabelFromTask)
	r.Get("/{id}", h.FindOne)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Remove)

	return r
}

