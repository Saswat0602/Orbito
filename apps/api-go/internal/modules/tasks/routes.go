package tasks

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Routes registers all task endpoints matching NestJS TasksController.
func Routes(h *Handler, authMw func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(authMw)

	r.Post("/", h.Create)
	r.Post("/bulk-create", h.BulkCreate)
	r.Post("/bulk-delete", h.BulkDelete)
	r.Post("/bulk-status-update", h.BulkStatusUpdate)
	r.Post("/bulk-assign", h.BulkAssign)
	r.Post("/create-task-attachment", h.CreateWithAttachments)

	r.Get("/", h.FindAll)
	r.Get("/all-tasks", h.GetAllTasks)
	r.Get("/grouped", h.GetGroupedTasks)
	r.Get("/by-status", h.GetTasksByStatus)
	r.Get("/today", h.GetTodaysTasks)
	r.Get("/slug/{slug}", h.FindOne)
	r.Get("/key/{key}", h.FindByKey)
	r.Get("/organization/{orgId}", h.GetTasksByOrganization)
	r.Get("/recurring/project/{projectId}", h.GetRecurringTasks)

	r.Patch("/{id}/reorder", h.Reorder)
	r.Patch("/{id}/status", h.UpdateStatus)
	r.Patch("/{id}/assignees", h.UpdateAssignees)
	r.Patch("/{id}/unassign", h.UnassignTask)
	r.Patch("/{id}/priority", h.UpdatePriority)
	r.Patch("/{id}/due-date", h.UpdateDueDate)
	r.Post("/{id}/comments", h.AddComment)
	r.Post("/{id}/complete-occurrence", h.CompleteOccurrence)
	r.Post("/{id}/recurrence", h.AddRecurrence)
	r.Patch("/{id}/recurrence", h.UpdateRecurrence)
	r.Delete("/{id}/recurrence", h.StopRecurrence)

	r.Get("/{id}", h.FindOne)
	r.Patch("/{id}", h.Update)
	r.Delete("/{id}", h.Remove)

	return r
}

// ShareRoutes registers endpoints matching NestJS TaskSharesController.
func ShareRoutes(h *Handler, authMw func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()
	r.Use(authMw)

	r.Post("/", h.CreateShare)
	r.Get("/task/{taskId}", h.GetSharesForTask)
	r.Delete("/{shareId}", h.RevokeShare)

	return r
}

