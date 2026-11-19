package organization_members

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes configures the Chi router for organization-members endpoints.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(authMw)

	r.Post("/", h.Create)
	r.Get("/", h.FindAll)
	r.Get("/slug", h.FindAllBySlug)
	r.Patch("/set-default", h.SetDefaultOrganization)
	r.Get("/user/{userId}/organization/{organizationId}", h.FindByUserAndOrg)

	r.Get("/{id}", h.FindOne)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Remove)

	return r
}
