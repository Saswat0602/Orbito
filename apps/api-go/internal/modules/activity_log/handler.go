package activity_log

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

var validEntityTypes = map[string]struct{}{
	"Task":         {},
	"Project":      {},
	"Workspace":    {},
	"Organization": {},
	"User":         {},
}

// Handler handles HTTP requests for /api/activity-logs and /api/activity-log.
type Handler struct {
	service Service
}

// NewHandler constructs an activity_log HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

func parsePositiveInt(raw string, def int) int {
	if raw == "" {
		return def
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v == 0 {
		return def
	}
	return v
}

// List handles GET /api/activity-logs and GET /api/activity-log.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit := parsePositiveInt(q.Get("limit"), 20)
	if limit < 1 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}
	page := parsePositiveInt(q.Get("page"), 1)
	if page < 1 {
		page = 1
	}

	if wsIDRaw := strings.TrimSpace(q.Get("workspaceId")); wsIDRaw != "" {
		wsID, err := uuid.Parse(wsIDRaw)
		if err != nil {
			httpx.RespondError(w, r, apperror.BadRequest("Invalid workspaceId format"))
			return
		}
		res, err := h.service.GetRecentActivityByWorkspaceOptimized(r.Context(), wsID, limit, page)
		if err != nil {
			httpx.RespondError(w, r, err)
			return
		}
		httpx.RespondJSON(w, http.StatusOK, res.Activities)
		return
	}

	if projID := strings.TrimSpace(q.Get("projectId")); projID != "" {
		res, err := h.service.GetRecentActivityByProject(r.Context(), projID, limit, page)
		if err != nil {
			httpx.RespondError(w, r, err)
			return
		}
		httpx.RespondJSON(w, http.StatusOK, res.Activities)
		return
	}

	if orgIDRaw := strings.TrimSpace(q.Get("organizationId")); orgIDRaw != "" {
		orgID, err := uuid.Parse(orgIDRaw)
		if err != nil {
			httpx.RespondError(w, r, apperror.BadRequest("Invalid organizationId format"))
			return
		}
		var filters *OrgFilters
		if et := strings.TrimSpace(q.Get("entityType")); et != "" {
			filters = &OrgFilters{EntityType: &et}
		}
		res, err := h.service.GetRecentActivityByOrganizationComprehensive(r.Context(), orgID, limit, page, filters)
		if err != nil {
			httpx.RespondError(w, r, err)
			return
		}
		httpx.RespondJSON(w, http.StatusOK, res.Activities)
		return
	}

	userIDStr := ctxhelper.GetUserID(r.Context())
	if userID, err := uuid.Parse(userIDStr); err == nil {
		res, svcErr := h.service.GetUserRecentActivity(r.Context(), userID, limit, page, nil)
		if svcErr != nil {
			httpx.RespondError(w, r, svcErr)
			return
		}
		httpx.RespondJSON(w, http.StatusOK, res.Activities)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, []any{})
}

// GetProjectActivities handles GET /project/{projectId}.
func (h *Handler) GetProjectActivities(w http.ResponseWriter, r *http.Request) {
	projID := chi.URLParam(r, "projectId")
	limit := parsePositiveInt(r.URL.Query().Get("limit"), 20)
	if limit < 1 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}
	page := parsePositiveInt(r.URL.Query().Get("page"), 1)
	if page < 1 {
		page = 1
	}

	res, err := h.service.GetRecentActivityByProject(r.Context(), projID, limit, page)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res.Activities)
}

// GetOrganizationRecentActivity handles GET /organization/{organizationId}/recent.
func (h *Handler) GetOrganizationRecentActivity(w http.ResponseWriter, r *http.Request) {
	orgID, err := httpx.URLParamUUID(r, "organizationId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	q := r.URL.Query()
	limitStr := q.Get("limit")
	if limitStr == "" {
		limitStr = "10"
	}
	limitNum, lErr := strconv.Atoi(limitStr)
	if lErr != nil || limitNum < 1 {
		httpx.RespondError(w, r, apperror.BadRequest("Limit must be a positive number"))
		return
	}
	if limitNum > 50 {
		limitNum = 50
	}

	pageStr := q.Get("page")
	if pageStr == "" {
		pageStr = "1"
	}
	pageNum, pErr := strconv.Atoi(pageStr)
	if pErr != nil || pageNum < 1 {
		httpx.RespondError(w, r, apperror.BadRequest("Page must be a positive number"))
		return
	}

	var filters OrgFilters
	if et := strings.TrimSpace(q.Get("entityType")); et != "" {
		if _, ok := validEntityTypes[et]; !ok {
			httpx.RespondError(w, r, apperror.BadRequest("Invalid entity type. Must be one of: Task, Project, Workspace, Organization, User"))
			return
		}
		filters.EntityType = &et
	}
	if uidStr := strings.TrimSpace(q.Get("userId")); uidStr != "" {
		uid, uErr := uuid.Parse(uidStr)
		if uErr != nil {
			httpx.RespondError(w, r, apperror.BadRequest("Invalid userId format"))
			return
		}
		filters.UserID = &uid
	}

	res, err := h.service.GetRecentActivityByOrganizationComprehensive(r.Context(), orgID, limitNum, pageNum, &filters)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetOrganizationActivityStats handles GET /organization/{organizationId}/stats.
func (h *Handler) GetOrganizationActivityStats(w http.ResponseWriter, r *http.Request) {
	orgID, err := httpx.URLParamUUID(r, "organizationId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	days := parsePositiveInt(r.URL.Query().Get("days"), 30)
	if days < 1 {
		days = 1
	}
	if days > 365 {
		days = 365
	}

	res, err := h.service.GetOrganizationActivityStats(r.Context(), orgID, days)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetTaskActivities handles GET /task/{taskId}/activities.
func (h *Handler) GetTaskActivities(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "taskId")
	limit := parsePositiveInt(r.URL.Query().Get("limit"), 50)
	if limit < 1 {
		limit = 1
	}
	if limit > 50 {
		limit = 50
	}
	page := parsePositiveInt(r.URL.Query().Get("page"), 1)
	if page < 1 {
		page = 1
	}

	res, err := h.service.GetTaskActivities(r.Context(), taskID, page, limit)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

