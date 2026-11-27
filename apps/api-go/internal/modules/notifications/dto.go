package notifications

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/httpx"
	"github.com/orbito/api-go/internal/shared/pagination"
)

// CreateInput holds parameters for creating a notification record.
type CreateInput struct {
	Title          string
	Message        string
	Type           sqlc.NotificationType
	UserID         uuid.UUID
	OrganizationID *uuid.UUID
	Priority       *sqlc.NotificationPriority
	EntityType     *string
	EntityID       *uuid.UUID
	ActionURL      *string
	CreatedBy      *uuid.UUID
}

// ListFilters holds optional query filters for listing notifications.
type ListFilters struct {
	IsRead         *bool
	Type           *sqlc.NotificationType
	OrganizationID *uuid.UUID
	Since          *time.Time
}

// UserOrgFilters holds optional query filters for user-within-org notification queries.
type UserOrgFilters struct {
	IsRead    *bool
	Type      *sqlc.NotificationType
	Priority  *sqlc.NotificationPriority
	StartDate *time.Time
	EndDate   *time.Time
}

// DeleteBulkRequest represents the payload for DELETE /api/notifications/bulk.
type DeleteBulkRequest struct {
	NotificationIDs []string `json:"notificationIds" validate:"required,min=1"`
}

// CreatorBrief represents the embedded createdByUser object on notifications.
type CreatorBrief struct {
	ID        string  `json:"id"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Email     *string `json:"email,omitempty"`
	Avatar    *string `json:"avatar"`
}

// RecipientBrief represents the embedded user object on detailed notifications.
type RecipientBrief struct {
	ID        string  `json:"id"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Email     string  `json:"email"`
	Username  *string `json:"username"`
	Avatar    *string `json:"avatar"`
}

// OrgBrief represents the embedded organization object on detailed notifications.
type OrgBrief struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Slug   string  `json:"slug"`
	Avatar *string `json:"avatar"`
}

// EntityDetails represents hydrated entity metadata returned by getEntityDetails.
type EntityDetails struct {
	ID     string         `json:"id"`
	Type   string         `json:"type"`
	Name   string         `json:"name"`
	Slug   *string        `json:"slug,omitempty"`
	Parent map[string]any `json:"parent,omitempty"`
	Extra  map[string]any `json:"extra,omitempty"`
}

// NotificationDTO represents a notification item matching NestJS Prisma JSON.
type NotificationDTO struct {
	ID             string          `json:"id"`
	Title          string          `json:"title"`
	Message        string          `json:"message"`
	Type           string          `json:"type"`
	Priority       string          `json:"priority"`
	IsRead         bool            `json:"isRead"`
	EntityType     *string         `json:"entityType"`
	EntityID       *string         `json:"entityId"`
	ActionURL      *string         `json:"actionUrl"`
	UserID         string          `json:"userId"`
	OrganizationID *string         `json:"organizationId"`
	CreatedBy      *string         `json:"createdBy"`
	UpdatedBy      *string         `json:"updatedBy"`
	CreatedAt      string          `json:"createdAt"`
	UpdatedAt      string          `json:"updatedAt"`
	ReadAt         *string         `json:"readAt"`
	CreatedByUser  *CreatorBrief   `json:"createdByUser,omitempty"`
	User           *RecipientBrief `json:"user,omitempty"`
	Organization   *OrgBrief       `json:"organization,omitempty"`
	Entity         *EntityDetails  `json:"entity,omitempty"`
}

// ListResponse matches GET /api/notifications and /by-type/:type.
type ListResponse struct {
	Notifications []NotificationDTO     `json:"notifications"`
	Pagination    pagination.Pagination `json:"pagination"`
}

// RecentResponse matches GET /api/notifications/recent.
type RecentResponse struct {
	Notifications []NotificationDTO `json:"notifications"`
	Count         int               `json:"count"`
}

// OrgUnreadCountDTO matches items in GET /api/notifications/unread-by-organization.
type OrgUnreadCountDTO struct {
	OrganizationID   string `json:"organizationId"`
	OrganizationName string `json:"organizationName"`
	UnreadCount      int64  `json:"unreadCount"`
}

// StatsSummaryDTO matches GET /api/notifications/stats/summary.
type StatsSummaryDTO struct {
	Total  int64            `json:"total"`
	Unread int64            `json:"unread"`
	Read   int64            `json:"read"`
	Recent int64            `json:"recent"`
	ByType map[string]int64 `json:"byType"`
}

// UserOrgSummaryDTO matches summary block in GET /user/:userId/organization/:organizationId.
type UserOrgSummaryDTO struct {
	Total      int64            `json:"total"`
	Unread     int64            `json:"unread"`
	ByType     map[string]int64 `json:"byType"`
	ByPriority map[string]int64 `json:"byPriority"`
}

// UserOrgNotificationsResponse matches GET /api/notifications/user/:userId/organization/:organizationId.
type UserOrgNotificationsResponse struct {
	Notifications []NotificationDTO     `json:"notifications"`
	Pagination    pagination.Pagination `json:"pagination"`
	Summary       UserOrgSummaryDTO     `json:"summary"`
}

func pgTextPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	v := t.String
	return &v
}

func pgUUIDPtr(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	v := uuid.UUID(u.Bytes).String()
	return &v
}

func pgTSISO(ts pgtype.Timestamp) string {
	if !ts.Valid {
		return time.Now().UTC().Format(httpx.ISO8601Format)
	}
	return ts.Time.UTC().Format(httpx.ISO8601Format)
}

func pgTSISOPtr(ts pgtype.Timestamp) *string {
	if !ts.Valid {
		return nil
	}
	v := ts.Time.UTC().Format(httpx.ISO8601Format)
	return &v
}

func mapNotificationModel(n sqlc.Notification) NotificationDTO {
	return NotificationDTO{
		ID:             n.ID.String(),
		Title:          n.Title,
		Message:        n.Message,
		Type:           string(n.Type),
		Priority:       string(n.Priority),
		IsRead:         n.IsRead,
		EntityType:     pgTextPtr(n.EntityType),
		EntityID:       pgUUIDPtr(n.EntityID),
		ActionURL:      pgTextPtr(n.ActionUrl),
		UserID:         n.UserID.String(),
		OrganizationID: pgUUIDPtr(n.OrganizationID),
		CreatedBy:      pgUUIDPtr(n.CreatedByID),
		UpdatedBy:      pgUUIDPtr(n.UpdatedByID),
		CreatedAt:      pgTSISO(n.CreatedAt),
		UpdatedAt:      pgTSISO(n.UpdatedAt),
		ReadAt:         pgTSISOPtr(n.ReadAt),
	}
}

