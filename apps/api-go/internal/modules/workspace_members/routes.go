package workspace_members

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Routes sets up Chi routing for workspace-members endpoints.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(authMw)

	r.Post("/", h.Create)
	r.Get("/", h.FindAll)

	// Direct endpoint matching acceptance criteria: GET /api/workspace-members/workspace/:id
	r.Get("/workspace/{id}", func(w http.ResponseWriter, req *http.Request) {
		idStr := chi.URLParam(req, "id")
		wsID, err := uuid.Parse(idStr)
		if err != nil {
			httpx.RespondError(w, req, apperror.BadRequest("Invalid workspace ID"))
			return
		}
		q := req.URL.Query()
		q.Set("workspaceId", wsID.String())
		req.URL.RawQuery = q.Encode()
		h.FindAll(w, req)
	})

	r.Get("/user/{userId}/workspace/{workspaceId}", h.FindByUserAndWorkspace)

	r.Get("/{id}", h.FindOne)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Remove)

	return r
}
