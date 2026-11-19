package admin

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes sets up Chi routing for admin dashboard and platform oversight endpoints.
func Routes(h *Handler, authMiddleware func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	r.Group(func(sub chi.Router) {
		if authMiddleware != nil {
			sub.Use(authMiddleware)
		}

		sub.Get("/dashboard", h.GetDashboardStats)

		// User controls
		sub.Get("/users", h.GetUsers)
		sub.Get("/users/{id}", h.GetUserDetail)
		sub.Patch("/users/{id}/role", h.UpdateUserRole)
		sub.Patch("/users/{id}/status", h.UpdateUserStatus)
		sub.Delete("/users/{id}", h.DeleteUser)
		sub.Post("/users/{id}/reset-password", h.ResetUserPassword)

		// Organization controls
		sub.Get("/organizations", h.GetOrganizations)
		sub.Get("/organizations/{id}", h.GetOrganizationDetail)
		sub.Delete("/organizations/{id}", h.DeleteOrganization)
		sub.Patch("/organizations/{id}/archive", h.ToggleOrganizationArchive)
		sub.Patch("/organizations/{id}/transfer-ownership", h.TransferOrganizationOwnership)

		// System configuration & SMTP testing
		sub.Get("/config", h.GetSystemConfig)
		sub.Post("/config", h.SaveSystemConfig)
		sub.Post("/config/test-smtp", h.TestSMTP)
	})

	return r
}
