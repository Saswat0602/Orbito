package public

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler serves unauthenticated public read routes.
type Handler struct {
	service Service
}

// NewHandler creates a Handler bound to public.Service.
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

// GetSharedTask handles GET /api/public/tasks/{token}.
func (h *Handler) GetSharedTask(w http.ResponseWriter, r *http.Request) {
	res, err := h.service.GetSharedTaskByToken(r.Context(), chi.URLParam(r, "token"))
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetSharedAttachmentURL handles GET /api/public/tasks/{token}/attachments/{attachmentId}.
func (h *Handler) GetSharedAttachmentURL(w http.ResponseWriter, r *http.Request) {
	attID, err := httpx.URLParamUUID(r, "attachmentId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.GetSharedAttachmentURL(r.Context(), chi.URLParam(r, "token"), attID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetWorkspaceProjects handles GET /api/public/workspaces/{workspaceSlug}/projects.
func (h *Handler) GetWorkspaceProjects(w http.ResponseWriter, r *http.Request) {
	res, err := h.service.GetWorkspaceProjects(r.Context(), chi.URLParam(r, "workspaceSlug"))
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetProject handles GET /api/public/workspaces/{workspaceSlug}/projects/{projectSlug}.
func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	res, err := h.service.GetPublicProject(r.Context(), chi.URLParam(r, "workspaceSlug"), chi.URLParam(r, "projectSlug"))
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetProjectCharts handles GET /api/public/workspaces/{slug}/charts.
func (h *Handler) GetProjectCharts(w http.ResponseWriter, r *http.Request) {
	rawTypes := r.URL.Query()["types"]
	types := make([]string, 0, len(rawTypes))
	for _, entry := range rawTypes {
		for _, part := range strings.Split(entry, ",") {
			if trimmed := strings.TrimSpace(part); trimmed != "" {
				types = append(types, trimmed)
			}
		}
	}
	res, err := h.service.GetProjectCharts(r.Context(), chi.URLParam(r, "slug"), types)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetProjectStatuses handles GET /api/public/workspaces/{slug}/statuses.
func (h *Handler) GetProjectStatuses(w http.ResponseWriter, r *http.Request) {
	res, err := h.service.GetProjectStatuses(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetCalendarEvents handles GET /api/public/workspaces/{workspaceSlug}/projects/{projectSlug}/calendar.
func (h *Handler) GetCalendarEvents(w http.ResponseWriter, r *http.Request) {
	res, err := h.service.GetCalendarEvents(
		r.Context(),
		chi.URLParam(r, "workspaceSlug"),
		chi.URLParam(r, "projectSlug"),
		httpx.QueryParam(r, "startDate", ""),
		httpx.QueryParam(r, "endDate", ""),
	)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetSprints handles GET /api/public/workspaces/{workspaceSlug}/projects/{projectSlug}/sprints.
func (h *Handler) GetSprints(w http.ResponseWriter, r *http.Request) {
	res, err := h.service.GetPublicSprints(r.Context(), chi.URLParam(r, "workspaceSlug"), chi.URLParam(r, "projectSlug"))
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetSprint handles GET /api/public/workspaces/{workspaceSlug}/projects/{projectSlug}/sprints/{sprintId}.
func (h *Handler) GetSprint(w http.ResponseWriter, r *http.Request) {
	sprintID, err := httpx.URLParamUUID(r, "sprintId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.GetPublicSprint(r.Context(), chi.URLParam(r, "workspaceSlug"), chi.URLParam(r, "projectSlug"), sprintID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetSprintTasks handles GET /api/public/workspaces/{workspaceSlug}/projects/{projectSlug}/sprints/{sprintId}/tasks.
func (h *Handler) GetSprintTasks(w http.ResponseWriter, r *http.Request) {
	sprintID, err := httpx.URLParamUUID(r, "sprintId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.GetPublicSprintTasks(r.Context(), chi.URLParam(r, "workspaceSlug"), chi.URLParam(r, "projectSlug"), sprintID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetProjectTasks handles GET /api/public/project-tasks/{workspaceSlug}/projects/{projectSlug}/tasks.
func (h *Handler) GetProjectTasks(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(httpx.QueryParam(r, "page", "1"))
	limit, _ := strconv.Atoi(httpx.QueryParam(r, "limit", "50"))
	filters := TaskFilters{
		ParentTaskID: httpx.QueryParam(r, "parentTaskId", ""),
		Status:       httpx.QueryParam(r, "status", ""),
		Priority:     httpx.QueryParam(r, "priority", ""),
		Type:         httpx.QueryParam(r, "type", ""),
		Page:         page,
		Limit:        limit,
	}
	res, err := h.service.GetPublicTasks(r.Context(), chi.URLParam(r, "workspaceSlug"), chi.URLParam(r, "projectSlug"), filters)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetProjectTasksKanban handles GET /api/public/project-tasks/{workspaceSlug}/projects/{projectSlug}/taskskanban.
func (h *Handler) GetProjectTasksKanban(w http.ResponseWriter, r *http.Request) {
	res, err := h.service.GetPublicTasksKanban(r.Context(), chi.URLParam(r, "workspaceSlug"), chi.URLParam(r, "projectSlug"))
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetTask handles GET /api/public/project-tasks/{taskId}.
func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	taskID, err := httpx.URLParamUUID(r, "taskId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.GetPublicTaskByID(r.Context(), taskID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetTaskBySlug handles GET /api/public/project-tasks/key/{slug}.
func (h *Handler) GetTaskBySlug(w http.ResponseWriter, r *http.Request) {
	res, err := h.service.GetPublicTaskBySlug(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetTaskActivities handles GET /api/public/project-tasks/activities/{taskId}.
func (h *Handler) GetTaskActivities(w http.ResponseWriter, r *http.Request) {
	taskID, err := httpx.URLParamUUID(r, "taskId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	page, _ := strconv.Atoi(httpx.QueryParam(r, "page", "1"))
	limit, _ := strconv.Atoi(httpx.QueryParam(r, "limit", "50"))
	res, err := h.service.GetTaskActivities(r.Context(), taskID, page, limit)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetTaskComments handles GET /api/public/project-tasks/comments/{taskId}.
func (h *Handler) GetTaskComments(w http.ResponseWriter, r *http.Request) {
	taskID, err := httpx.URLParamUUID(r, "taskId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.GetPublicTaskComments(r.Context(), taskID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetTaskAttachments handles GET /api/public/project-tasks/attachments/{taskId}.
func (h *Handler) GetTaskAttachments(w http.ResponseWriter, r *http.Request) {
	taskID, err := httpx.URLParamUUID(r, "taskId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	res, err := h.service.GetPublicTaskAttachments(r.Context(), taskID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}
