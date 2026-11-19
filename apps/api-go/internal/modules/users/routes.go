package users

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes sets up Chi routing for user directory and profile endpoints.
func Routes(h *Handler, authMiddleware func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	// Public check
	r.Get("/exists", h.CheckUsersExist)

	// Protected routes
	r.Group(func(sub chi.Router) {
		if authMiddleware != nil {
			sub.Use(authMiddleware)
		}
		sub.Get("/", h.FindAll)
		sub.Post("/", h.Create)
		sub.Patch("/profile", h.UpdateCurrentUserProfile)
		sub.Post("/change-password", h.ChangePassword)
		sub.Get("/status/bulk", h.GetUsersStatus)
		sub.Get("/{id}", h.FindOne)
		sub.Get("/{id}/profile", h.GetPublicProfile)
		sub.Patch("/{id}", h.Update)
		sub.Patch("/{id}/status", h.UpdateStatus)
		sub.Patch("/{id}/change-password", h.ChangePassword)
		sub.Delete("/{id}", h.Remove)
		sub.Get("/{id}/status", h.GetUserStatus)
	})

	return r
}
