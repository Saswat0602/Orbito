package task_comments

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/orbito/api-go/internal/middleware"
)

// Routes registers endpoints matching NestJS TaskCommentsController.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(authMw)

	r.With(middleware.UserRateLimit(20, time.Minute)).Post("/", h.Create)
	r.Get("/", h.FindAll)
	r.Get("/middle-pagination", h.FindWithMiddlePagination)
	r.Get("/task/{taskId}/tree", h.GetTaskCommentTree)
	r.Get("/{id}", h.FindOne)
	r.Get("/{id}/replies", h.GetReplies)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Remove)

	return r
}

