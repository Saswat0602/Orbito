package organizations

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/orbito/api-go/internal/middleware"
)

// Routes sets up Chi routing for organizations endpoints.
func Routes(h *Handler, authMw func(http.Handler) http.Handler, rbacResolver middleware.RBACResolver) http.Handler {
	r := chi.NewRouter()

	r.Use(authMw)

	// List & Create
	r.Post("/", h.Create)
	r.Get("/", h.FindAll)

	// Slug resolution
	r.With(middleware.RequireOrgRole(rbacResolver, middleware.RoleMember, "slug")).Get("/slug/{slug}", h.FindBySlug)

	// Scoped Organization operations
	r.Route("/{id}", func(r chi.Router) {
		r.With(middleware.RequireOrgRole(rbacResolver, middleware.RoleMember, "id")).Get("/", h.FindOne)
		r.With(middleware.RequireOrgRole(rbacResolver, middleware.RoleManager, "id")).Patch("/", h.Update)
		r.With(middleware.RequireOrgRole(rbacResolver, middleware.RoleOwner, "id")).Delete("/", h.Delete)
		r.With(middleware.RequireOrgRole(rbacResolver, middleware.RoleOwner, "id")).Patch("/transfer-ownership", h.TransferOwnership)
		r.With(middleware.RequireOrgRole(rbacResolver, middleware.RoleViewer, "id")).Get("/stats", h.GetStats)
	})

	r.With(middleware.RequireOrgRole(rbacResolver, middleware.RoleOwner, "id")).Patch("/archive/{id}", h.Archive)

	return r
}
