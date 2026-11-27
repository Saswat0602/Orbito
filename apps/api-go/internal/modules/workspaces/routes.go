package workspaces

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/orbito/api-go/internal/middleware"
)

// Routes sets up Chi routing for workspaces endpoints.
func Routes(h *Handler, authMw func(http.Handler) http.Handler, rbacResolver middleware.RBACResolver) http.Handler {
	r := chi.NewRouter()

	r.Use(authMw)

	// Directory & Collection routes
	r.Post("/", h.Create)
	r.Get("/", h.FindAll)
	r.Get("/archived", h.FindArchived)
	r.Get("/tree", h.FindTree)
	r.Get("/by-slug/{slug}", h.FindBySlugOnly)
	r.Get("/organization/{organizationId}/slug/{slug}", h.FindBySlug)

	// Workspace lifecycle operations
	r.Patch("/archive/{id}", h.Archive)
	r.Patch("/unarchive/{id}", h.Unarchive)
	r.With(middleware.RequireWorkspaceRole(rbacResolver, middleware.RoleViewer, "workspaceId")).Get("/recent/{workspaceId}", h.GetRecentActivity)

	// Scoped workspace operations
	r.Route("/{id}", func(r chi.Router) {
		r.With(middleware.RequireWorkspaceRole(rbacResolver, middleware.RoleViewer, "id")).Get("/", h.FindOne)
		r.With(middleware.RequireWorkspaceRole(rbacResolver, middleware.RoleManager, "id")).Patch("/", h.Update)
		r.With(middleware.RequireWorkspaceRole(rbacResolver, middleware.RoleOwner, "id")).Delete("/", h.Delete)
		r.With(middleware.RequireWorkspaceRole(rbacResolver, middleware.RoleViewer, "id")).Get("/members", h.GetMembers)
		r.With(middleware.RequireWorkspaceRole(rbacResolver, middleware.RoleViewer, "id")).Get("/charts", h.GetCharts)
	})

	return r
}
