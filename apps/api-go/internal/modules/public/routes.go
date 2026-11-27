package public

import (
	"time"

	"github.com/go-chi/chi/v5"
)

// SharedTasksRoutes mounts unauthenticated /api/public/tasks routes.
func SharedTasksRoutes(h *Handler, rl *RateLimiter) chi.Router {
	r := chi.NewRouter()
	r.With(rl.Middleware("shared_task", 10, time.Minute)).Get("/{token}", h.GetSharedTask)
	r.With(rl.Middleware("shared_attachment", 5, time.Minute)).Get("/{token}/attachments/{attachmentId}", h.GetSharedAttachmentURL)
	return r
}

// WorkspacesRoutes mounts unauthenticated /api/public/workspaces routes.
func WorkspacesRoutes(h *Handler, rl *RateLimiter) chi.Router {
	r := chi.NewRouter()
	projLim := rl.Middleware("projects", 0, 0)
	defLim := rl.Middleware("default", 0, 0)
	calLim := rl.Middleware("calendar", 0, 0)
	sprintLim := rl.Middleware("sprints", 0, 0)

	r.With(projLim).Get("/{workspaceSlug}/projects", h.GetWorkspaceProjects)
	r.With(projLim).Get("/{workspaceSlug}/projects/{projectSlug}", h.GetProject)
	r.With(defLim).Get("/{slug}/charts", h.GetProjectCharts)
	r.With(defLim).Get("/{slug}/statuses", h.GetProjectStatuses)
	r.With(calLim).Get("/{workspaceSlug}/projects/{projectSlug}/calendar", h.GetCalendarEvents)
	r.With(sprintLim).Get("/{workspaceSlug}/projects/{projectSlug}/sprints", h.GetSprints)
	r.With(sprintLim).Get("/{workspaceSlug}/projects/{projectSlug}/sprints/{sprintId}", h.GetSprint)
	r.With(sprintLim).Get("/{workspaceSlug}/projects/{projectSlug}/sprints/{sprintId}/tasks", h.GetSprintTasks)
	return r
}

// ProjectTasksRoutes mounts unauthenticated /api/public/project-tasks routes.
func ProjectTasksRoutes(h *Handler, rl *RateLimiter) chi.Router {
	r := chi.NewRouter()
	r.Use(rl.Middleware("tasks", 0, 0))

	r.Get("/{workspaceSlug}/projects/{projectSlug}/tasks", h.GetProjectTasks)
	r.Get("/{workspaceSlug}/projects/{projectSlug}/taskskanban", h.GetProjectTasksKanban)
	r.Get("/key/{slug}", h.GetTaskBySlug)
	r.Get("/activities/{taskId}", h.GetTaskActivities)
	r.Get("/comments/{taskId}", h.GetTaskComments)
	r.Get("/attachments/{taskId}", h.GetTaskAttachments)
	r.Get("/{taskId}", h.GetTask)
	return r
}
