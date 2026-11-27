package activity_log

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes returns the Chi router mounted at /api/activity-logs and /api/activity-log.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	if authMw != nil {
		r.Use(authMw)
	}

	r.Get("/", h.List)
	r.Get("/project/{projectId}", h.GetProjectActivities)
	r.Get("/organization/{organizationId}/recent", h.GetOrganizationRecentActivity)
	r.Get("/organization/{organizationId}/stats", h.GetOrganizationActivityStats)
	r.Get("/task/{taskId}/activities", h.GetTaskActivities)

	return r
}

