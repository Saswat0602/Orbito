package workflows

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes configures and returns an HTTP handler for workflows endpoints.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(authMw)

	r.Post("/", h.Create)
	r.Get("/", h.FindAll)
	r.Get("/slug", h.FindByOrganizationSlug)
	r.Get("/organization/{organizationId}/default", h.GetDefaultWorkflow)
	r.Get("/{id}", h.FindOne)
	r.Patch("/{id}", h.Update)
	r.Patch("/{id}/set-default", h.MakeDefault)
	r.Delete("/{id}", h.Delete)

	return r
}
