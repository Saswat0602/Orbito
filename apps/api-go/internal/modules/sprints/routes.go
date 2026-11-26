package sprints

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes registers sprint endpoints matching NestJS SprintsController.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(authMw)

	r.Post("/", h.Create)
	r.Get("/", h.FindAll)
	r.Get("/slug", h.FindAllByProjectSlug)
	r.Get("/by-slug/{projectSlug}/{sprintSlug}", h.FindBySlug)
	r.Get("/project/{projectId}/active", h.GetActiveSprint)
	r.Patch("/archive/{id}", h.ArchiveSprint)
	r.Get("/{id}", h.FindOne)
	r.Patch("/{id}", h.Update)
	r.Patch("/{id}/start", h.StartSprint)
	r.Patch("/{id}/complete", h.CompleteSprint)
	r.Delete("/{id}", h.Remove)

	return r
}
