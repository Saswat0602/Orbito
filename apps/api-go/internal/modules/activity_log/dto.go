package activity_log

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/httpx"
	"github.com/orbito/api-go/internal/shared/pagination"
)

// LogInput defines parameters for recording an audit trail entry.
type LogInput struct {
	Type           sqlc.ActivityType
	Description    string
	EntityType     string
	EntityID       string
	UserID         uuid.UUID
	OrganizationID *uuid.UUID
	OldValue       any
	NewValue       any
}

// OrgFilters defines query filters for organization activity queries.
type OrgFilters struct {
	EntityType *string
	UserID     *uuid.UUID
	StartDate  *time.Time
	EndDate    *time.Time
}

// ActorSummary represents the simplified user object returned in workspace/project/org activity queries.
type ActorSummary struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Email  string  `json:"email"`
	Avatar *string `json:"avatar"`
}

// ActorUser represents the Prisma-style user object returned in task/user activity queries.
type ActorUser struct {
	ID        string  `json:"id"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Email     string  `json:"email"`
	Avatar    *string `json:"avatar"`
}

// EnrichedActivityItem matches NestJS getRecentActivityByWorkspaceOptimized & getRecentActivityByOrganizationComprehensive items.
type EnrichedActivityItem struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	Description   string          `json:"description"`
	EntityType    string          `json:"entityType"`
	EntityID      string          `json:"entityId"`
	NewValue      json.RawMessage `json:"newValue,omitempty"`
	CreatedAt     string          `json:"createdAt"`
	TaskSlug      *string         `json:"taskSlug"`
	ProjectSlug   *string         `json:"projectSlug"`
	WorkspaceSlug *string         `json:"workspaceSlug"`
	SprintSlug    *string         `json:"sprintSlug,omitempty"`
	User          ActorSummary    `json:"user"`
}

// ProjectActivityItem matches NestJS getRecentActivityByProject items.
type ProjectActivityItem struct {
	ID            string       `json:"id"`
	Type          string       `json:"type"`
	Description   string       `json:"description"`
	EntityType    string       `json:"entityType"`
	EntityID      string       `json:"entityId"`
	CreatedAt     string       `json:"createdAt"`
	ProjectSlug   string       `json:"projectSlug"`
	WorkspaceSlug *string      `json:"workspaceSlug"`
	User          ActorSummary `json:"user"`
}

// RawActivityItem matches NestJS ActivityLog Prisma items (with user include).
type RawActivityItem struct {
	ID             string          `json:"id"`
	Type           string          `json:"type"`
	Description    string          `json:"description"`
	EntityType     string          `json:"entityType"`
	EntityID       string          `json:"entityId"`
	OldValue       json.RawMessage `json:"oldValue"`
	NewValue       json.RawMessage `json:"newValue"`
	UserID         string          `json:"userId"`
	OrganizationID *string         `json:"organizationId"`
	CreatedBy      *string         `json:"createdBy"`
	UpdatedBy      *string         `json:"updatedBy"`
	CreatedAt      string          `json:"createdAt"`
	UpdatedAt      string          `json:"updatedAt"`
	User           ActorUser       `json:"user"`
}

// PaginatedEnrichedResponse matches paginated workspace and org recent activity responses.
type PaginatedEnrichedResponse struct {
	Activities []EnrichedActivityItem `json:"activities"`
	Pagination pagination.Pagination  `json:"pagination"`
}

// PaginatedProjectResponse matches paginated project activity responses.
type PaginatedProjectResponse struct {
	Activities []ProjectActivityItem `json:"activities"`
	Pagination pagination.Pagination `json:"pagination"`
}

// PaginatedRawResponse matches paginated task and user activity responses.
type PaginatedRawResponse struct {
	Activities []RawActivityItem     `json:"activities"`
	Pagination pagination.Pagination `json:"pagination"`
}

// UserActivityCount represents per-user activity totals in organization stats.
type UserActivityCount struct {
	UserID   string `json:"userId"`
	UserName string `json:"userName"`
	Count    int64  `json:"count"`
}

// DateActivityCount represents per-day activity totals in organization stats.
type DateActivityCount struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// OrganizationStatsResponse matches GET /api/activity-log/organization/:organizationId/stats.
type OrganizationStatsResponse struct {
	TotalActivities  int64               `json:"totalActivities"`
	ActivitiesByType map[string]int64    `json:"activitiesByType"`
	ActivitiesByUser []UserActivityCount `json:"activitiesByUser"`
	ActivitiesByDate []DateActivityCount `json:"activitiesByDate"`
}

func pgTextPtr(t pgtype.Text) *string {
	if !t.Valid || t.String == "" {
		return nil
	}
	v := t.String
	return &v
}

func nonEmptyPtr(s string) *string {
	if s == "" {
		return nil
	}
	v := s
	return &v
}

func firstNonEmpty(vals ...string) *string {
	for _, v := range vals {
		if v != "" {
			cp := v
			return &cp
		}
	}
	return nil
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

func rawJSONOrNull(b []byte) json.RawMessage {
	if len(b) == 0 {
		return json.RawMessage("null")
	}
	return json.RawMessage(b)
}

func fullName(first, last string) string {
	return strings.TrimSpace(first + " " + last)
}
