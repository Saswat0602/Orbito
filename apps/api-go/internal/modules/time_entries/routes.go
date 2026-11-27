package time_entries

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes registers endpoints matching NestJS TimeEntriesController.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(authMw)

	r.Post("/", h.Create)
	r.Get("/", h.FindAll)
	r.Get("/summary", h.GetSummary)
	r.Post("/timer/start", h.StartTimer)
	r.Post("/timer/stop", h.StopTimer)
	r.Get("/timer/active/{userId}", h.GetActiveTimer)
	r.Get("/{id}", h.FindOne)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Remove)

	return r
}

