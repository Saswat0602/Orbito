package email

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes configures the Chi router for the email endpoints.
func Routes(h *Handler, authMiddleware func(http.Handler) http.Handler) chi.Router {
	r := chi.NewRouter()

	r.Group(func(sub chi.Router) {
		if authMiddleware != nil {
			sub.Use(authMiddleware)
		}
		sub.Post("/send", h.SendEmail)
		sub.Post("/send-bulk", h.SendBulkEmail)
		sub.Get("/queue/stats", h.GetQueueStats)
	})

	return r
}
