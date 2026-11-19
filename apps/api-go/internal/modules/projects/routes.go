package projects

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes configures and returns an HTTP handler for projects endpoints.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(authMw)

	r.Post("/", h.Create)
	r.Get("/", h.FindAll)
	r.Get("/by-organization", h.FindByOrganizationID)
	r.Get("/by-slug/{slug}", h.GetBySlug)
	r.Get("/workspace/{workspaceId}/key/{key}", h.FindByKey)
	r.Get("/{slug}/charts", h.GetCharts)
	r.Get("/{id}/stats", h.GetStats)
	r.Get("/{id}", h.FindOne)
	r.Patch("/{id}", h.Update)
	r.Patch("/archive/{id}", h.Archive)
	r.Patch("/unarchive/{id}", h.Unarchive)
	r.Delete("/{id}", h.Delete)

	return r
}
