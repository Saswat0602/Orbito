package invitations

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes configures and returns an HTTP handler for invitations endpoints.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	// Public routes
	r.Get("/verify/{token}", h.Verify)
	r.Get("/token/{token}", h.Verify)

	// Protected routes
	r.Group(func(pr chi.Router) {
		pr.Use(authMw)

		pr.Post("/", h.Create)
		pr.Get("/user", h.GetUserInvitations)
		pr.Get("/entity/{entityType}/{entityId}", h.GetEntityInvitations)
		pr.Patch("/{token}/accept", h.Accept)
		pr.Post("/{token}/accept", h.Accept)
		pr.Patch("/{token}/decline", h.Decline)
		pr.Post("/{id}/resend", h.Resend)
		pr.Delete("/{id}", h.Delete)
	})

	return r
}
