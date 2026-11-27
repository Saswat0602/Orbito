package notifications

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Handler handles HTTP requests for /api/notifications endpoints.
type Handler struct {
	service Service
}

// NewHandler constructs a notifications HTTP handler.
func NewHandler(svc Service) *Handler {
	return &Handler{service: svc}
}

func parseOptionalOrgID(raw string) (*uuid.UUID, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == "undefined" || trimmed == "null" {
		return nil, nil
	}
	id, err := uuid.Parse(trimmed)
	if err != nil {
		return nil, apperror.BadRequest("Invalid organizationId format")
	}
	return &id, nil
}

func parsePageLimit(r *http.Request, defaultLimit, maxLimit int) (int, int, error) {
	page := 1
	if rawPage := strings.TrimSpace(r.URL.Query().Get("page")); rawPage != "" {
		parsed, err := strconv.Atoi(rawPage)
		if err != nil {
			parsed = 1
		}
		if parsed < 1 {
			return 0, 0, apperror.BadRequest("Page must be a positive number")
		}
		page = parsed
	}

	limit := defaultLimit
	if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit < 1 {
		limit = 1
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return page, limit, nil
}

// List handles GET /api/notifications.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthorized"))
		return
	}

	page, limit, pErr := parsePageLimit(r, 20, 100)
	if pErr != nil {
		httpx.RespondError(w, r, pErr)
		return
	}

	var filters ListFilters
	if isReadRaw := strings.TrimSpace(r.URL.Query().Get("isRead")); isReadRaw != "" && isReadRaw != "undefined" {
		b := isReadRaw == "true"
		filters.IsRead = &b
	}
	if typeRaw := strings.TrimSpace(r.URL.Query().Get("type")); typeRaw != "" {
		nt := sqlc.NotificationType(typeRaw)
		filters.Type = &nt
	}
	orgID, orgErr := parseOptionalOrgID(r.URL.Query().Get("organizationId"))
	if orgErr != nil {
		httpx.RespondError(w, r, orgErr)
		return
	}
	filters.OrganizationID = orgID

	res, err := h.service.GetUserNotifications(r.Context(), userID, filters, page, limit)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// UnreadCount handles GET /api/notifications/unread-count.
func (h *Handler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthorized"))
		return
	}

	orgID, orgErr := parseOptionalOrgID(r.URL.Query().Get("organizationId"))
	if orgErr != nil {
		httpx.RespondError(w, r, orgErr)
		return
	}

	count, err := h.service.GetUnreadCount(r.Context(), userID, orgID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]int64{"count": count})
}

// UnreadByOrganization handles GET /api/notifications/unread-by-organization.
func (h *Handler) UnreadByOrganization(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthorized"))
		return
	}

	res, err := h.service.GetUnreadCountsByOrganization(r.Context(), userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// Recent handles GET /api/notifications/recent.
func (h *Handler) Recent(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthorized"))
		return
	}

	_, limit, pErr := parsePageLimit(r, 10, 50)
	if pErr != nil {
		httpx.RespondError(w, r, pErr)
		return
	}

	orgID, orgErr := parseOptionalOrgID(r.URL.Query().Get("organizationId"))
	if orgErr != nil {
		httpx.RespondError(w, r, orgErr)
		return
	}

	since := time.Now().UTC().AddDate(0, 0, -7)
	res, err := h.service.GetUserNotifications(r.Context(), userID, ListFilters{
		OrganizationID: orgID,
		Since:          &since,
	}, 1, limit)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	httpx.RespondJSON(w, http.StatusOK, RecentResponse{
		Notifications: res.Notifications,
		Count:         len(res.Notifications),
	})
}

// ByType handles GET /api/notifications/by-type/{type}.
func (h *Handler) ByType(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthorized"))
		return
	}

	page, limit, pErr := parsePageLimit(r, 20, 100)
	if pErr != nil {
		httpx.RespondError(w, r, pErr)
		return
	}

	orgID, orgErr := parseOptionalOrgID(r.URL.Query().Get("organizationId"))
	if orgErr != nil {
		httpx.RespondError(w, r, orgErr)
		return
	}

	nt := sqlc.NotificationType(chi.URLParam(r, "type"))
	res, err := h.service.GetUserNotifications(r.Context(), userID, ListFilters{
		Type:           &nt,
		OrganizationID: orgID,
	}, page, limit)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// ByUserAndOrganization handles GET /api/notifications/user/{userId}/organization/{organizationId}.
func (h *Handler) ByUserAndOrganization(w http.ResponseWriter, r *http.Request) {
	callerID := ctxhelper.GetUserID(r.Context())
	callerRole := ctxhelper.GetUserRole(r.Context())

	targetUserID, err := httpx.URLParamUUID(r, "userId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	orgID, err := httpx.URLParamUUID(r, "organizationId")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	if callerID != targetUserID.String() && callerRole != string(sqlc.RoleSUPERADMIN) {
		httpx.RespondError(w, r, apperror.Forbidden("You can only read your own notifications"))
		return
	}

	page, limit, pErr := parsePageLimit(r, 20, 100)
	if pErr != nil {
		httpx.RespondError(w, r, pErr)
		return
	}

	var filters UserOrgFilters
	if isReadRaw := r.URL.Query().Get("isRead"); isReadRaw != "" {
		b := isReadRaw == "true"
		filters.IsRead = &b
	}
	if typeRaw := strings.TrimSpace(r.URL.Query().Get("type")); typeRaw != "" {
		nt := sqlc.NotificationType(typeRaw)
		filters.Type = &nt
	}
	if prioRaw := strings.TrimSpace(r.URL.Query().Get("priority")); prioRaw != "" {
		pr := sqlc.NotificationPriority(prioRaw)
		filters.Priority = &pr
	}
	if startRaw := strings.TrimSpace(r.URL.Query().Get("startDate")); startRaw != "" {
		t, tErr := parseFlexibleDate(startRaw)
		if tErr != nil {
			httpx.RespondError(w, r, apperror.BadRequest("Invalid startDate format"))
			return
		}
		filters.StartDate = &t
	}
	if endRaw := strings.TrimSpace(r.URL.Query().Get("endDate")); endRaw != "" {
		t, tErr := parseFlexibleDate(endRaw)
		if tErr != nil {
			httpx.RespondError(w, r, apperror.BadRequest("Invalid endDate format"))
			return
		}
		filters.EndDate = &t
	}

	res, err := h.service.GetNotificationsByUserAndOrganization(r.Context(), targetUserID, orgID, filters, page, limit)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

func parseFlexibleDate(raw string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", raw)
}

// StatsSummary handles GET /api/notifications/stats/summary.
func (h *Handler) StatsSummary(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthorized"))
		return
	}

	orgID, orgErr := parseOptionalOrgID(r.URL.Query().Get("organizationId"))
	if orgErr != nil {
		httpx.RespondError(w, r, orgErr)
		return
	}

	res, err := h.service.GetUserNotificationStats(r.Context(), userID, orgID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// GetByID handles GET /api/notifications/{id}.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthorized"))
		return
	}

	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	res, err := h.service.GetNotificationByID(r.Context(), id, userID)
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, res)
}

// MarkAsRead handles PATCH /api/notifications/{id}/read.
func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthorized"))
		return
	}

	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	if err := h.service.MarkAsRead(r.Context(), id, userID); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]string{"message": "Notification marked as read"})
}

// MarkAllAsRead handles PATCH/POST /api/notifications/mark-all-read and PATCH /read-all.
func (h *Handler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthorized"))
		return
	}

	orgID, orgErr := parseOptionalOrgID(r.URL.Query().Get("organizationId"))
	if orgErr != nil {
		httpx.RespondError(w, r, orgErr)
		return
	}

	if err := h.service.MarkAllAsRead(r.Context(), userID, orgID); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]string{"message": "All notifications marked as read"})
}

// MarkAllUnreadAsRead handles PATCH /api/notifications/mark-all-unread-read.
func (h *Handler) MarkAllUnreadAsRead(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthorized"))
		return
	}

	orgID, orgErr := parseOptionalOrgID(r.URL.Query().Get("organizationId"))
	if orgErr != nil {
		httpx.RespondError(w, r, orgErr)
		return
	}

	if err := h.service.MarkAllAsRead(r.Context(), userID, orgID); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	httpx.RespondJSON(w, http.StatusOK, map[string]string{"message": "All unread notifications marked as read"})
}

// DeleteBulk handles DELETE /api/notifications/bulk.
func (h *Handler) DeleteBulk(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthorized"))
		return
	}

	var req DeleteBulkRequest
	if bErr := httpx.BindJSON(r, &req); bErr != nil {
		httpx.RespondError(w, r, bErr)
		return
	}
	if len(req.NotificationIDs) == 0 {
		httpx.RespondError(w, r, apperror.BadRequest("notificationIds must be a non-empty array"))
		return
	}

	ids := make([]uuid.UUID, 0, len(req.NotificationIDs))
	for _, raw := range req.NotificationIDs {
		parsed, pErr := uuid.Parse(raw)
		if pErr != nil {
			httpx.RespondError(w, r, apperror.BadRequest("Invalid notification ID: "+raw))
			return
		}
		ids = append(ids, parsed)
	}

	if err := h.service.DeleteMultipleNotifications(r.Context(), ids, userID); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Delete handles DELETE /api/notifications/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := uuid.Parse(ctxhelper.GetUserID(r.Context()))
	if err != nil {
		httpx.RespondError(w, r, apperror.Unauthorized("Unauthorized"))
		return
	}

	id, err := httpx.URLParamUUID(r, "id")
	if err != nil {
		httpx.RespondError(w, r, err)
		return
	}

	if err := h.service.DeleteNotification(r.Context(), id, userID); err != nil {
		httpx.RespondError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

