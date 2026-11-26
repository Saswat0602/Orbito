package task_ranks

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes registers endpoints matching NestJS TaskRanksController.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(authMw)

	r.Patch("/{taskId}/reorder", h.Reorder)
	r.Get("/rebalance", h.RebalanceAll)

	return r
}
