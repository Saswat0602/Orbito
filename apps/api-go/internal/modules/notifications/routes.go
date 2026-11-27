package notifications

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes returns the Chi router mounted at /api/notifications.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	if authMw != nil {
		r.Use(authMw)
	}

	r.Get("/", h.List)
	r.Get("/unread-count", h.UnreadCount)
	r.Get("/unread-by-organization", h.UnreadByOrganization)
	r.Get("/recent", h.Recent)
	r.Get("/by-type/{type}", h.ByType)
	r.Get("/user/{userId}/organization/{organizationId}", h.ByUserAndOrganization)
	r.Get("/stats/summary", h.StatsSummary)

	r.Patch("/mark-all-read", h.MarkAllAsRead)
	r.Patch("/read-all", h.MarkAllAsRead)
	r.Post("/mark-all-read", h.MarkAllAsRead)
	r.Patch("/mark-all-unread-read", h.MarkAllUnreadAsRead)
	r.Delete("/bulk", h.DeleteBulk)

	r.Get("/{id}", h.GetByID)
	r.Patch("/{id}/read", h.MarkAsRead)
	r.Delete("/{id}", h.Delete)

	return r
}
