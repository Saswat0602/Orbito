package settings

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes sets up Chi routing for settings endpoints.
func Routes(h *Handler, authMiddleware func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	// Public check
	r.Get("/registration-status", h.GetRegistrationStatus)

	// Authenticated routes
	r.Group(func(sub chi.Router) {
		if authMiddleware != nil {
			sub.Use(authMiddleware)
		}
		sub.Get("/", h.GetAllSettings)
		sub.Post("/", h.SetSetting)
		sub.Patch("/", h.SetSetting)
		sub.Post("/bulk", h.BulkSetSettings)
		sub.Get("/{key}", h.GetSetting)
		sub.Delete("/{key}", h.DeleteSetting)
	})

	return r
}
