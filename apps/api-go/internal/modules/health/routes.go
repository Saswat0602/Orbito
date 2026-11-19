package health

import (
	"github.com/go-chi/chi/v5"
)

// Routes mounts the health endpoints on a Chi subrouter.
func Routes(h *Handler) chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.GetHealth)
	r.Get("/liveness", h.GetLiveness)
	r.Get("/readiness", h.GetReadiness)
	return r
}
