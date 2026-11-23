package project_members

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes configures and returns an HTTP handler for project members endpoints.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(authMw)

	r.Post("/", h.Create)
	r.Get("/", h.FindAll)
	r.Get("/project/{id}", h.FindAllByProject)
	r.Get("/project/{projectId}/stats", h.GetStats)
	r.Get("/workspace/{workspaceId}", h.ListByWorkspace)
	r.Get("/user/{userId}/projects", h.ListUserProjects)
	r.Get("/user/{userId}/project/{projectId}", h.FindByUserAndProject)
	r.Get("/{id}", h.FindOne)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Remove)

	return r
}
