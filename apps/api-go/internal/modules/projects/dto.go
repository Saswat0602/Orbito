package projects

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CreateProjectRequest contains payload for creating a new project.
type CreateProjectRequest struct {
	Name        string          `json:"name"`
	Slug        string          `json:"slug"`
	TaskPrefix  *string         `json:"taskPrefix"`
	Key         *string         `json:"key"`
	Description *string         `json:"description"`
	Avatar      *string         `json:"avatar"`
	Color       string          `json:"color"`
	Status      *string         `json:"status"`
	Priority    *string         `json:"priority"`
	Visibility  *string         `json:"visibility"`
	StartDate   *time.Time      `json:"startDate"`
	EndDate     *time.Time      `json:"endDate"`
	Settings    json.RawMessage `json:"settings"`
	WorkspaceID uuid.UUID       `json:"workspaceId"`
	WorkflowID  *uuid.UUID      `json:"workflowId"`
}

// UpdateProjectRequest contains payload for modifying an existing project.
type UpdateProjectRequest struct {
	Name        *string         `json:"name"`
	Slug        *string         `json:"slug"`
	TaskPrefix  *string         `json:"taskPrefix"`
	Description *string         `json:"description"`
	Avatar      *string         `json:"avatar"`
	Color       *string         `json:"color"`
	Status      *string         `json:"status"`
	Priority    *string         `json:"priority"`
	Visibility  *string         `json:"visibility"`
	StartDate   *time.Time      `json:"startDate"`
	EndDate     *time.Time      `json:"endDate"`
	Settings    json.RawMessage `json:"settings"`
	WorkflowID  *uuid.UUID      `json:"workflowId"`
}

// ProjectStatsResponse contains computed stats for a project.
type ProjectStatsResponse struct {
	TotalTasks     int64   `json:"totalTasks"`
	CompletedTasks int64   `json:"completedTasks"`
	CompletionRate float64 `json:"completionRate"`
	MemberCount    int64   `json:"memberCount"`
}

// ProjectResponse models the serialized project response.
type ProjectResponse struct {
	ID             uuid.UUID       `json:"id"`
	Name           string          `json:"name"`
	Slug           string          `json:"slug"`
	TaskPrefix     *string         `json:"taskPrefix"`
	Description    *string         `json:"description"`
	Avatar         *string         `json:"avatar"`
	Color          string          `json:"color"`
	Status         string          `json:"status"`
	Priority       string          `json:"priority"`
	Visibility     string          `json:"visibility"`
	StartDate      *time.Time      `json:"startDate"`
	EndDate        *time.Time      `json:"endDate"`
	Settings       json.RawMessage `json:"settings,omitempty"`
	WorkflowID     uuid.UUID       `json:"workflowId"`
	WorkspaceID    uuid.UUID       `json:"workspaceId"`
	CreatedByID    *uuid.UUID      `json:"createdById,omitempty"`
	UpdatedByID    *uuid.UUID      `json:"updatedById,omitempty"`
	CreatedAt      time.Time       `json:"createdAt"`
	UpdatedAt      time.Time       `json:"updatedAt"`
	Archive        bool            `json:"archive"`
	WorkspaceName  string          `json:"workspaceName,omitempty"`
	WorkspaceSlug  string          `json:"workspaceSlug,omitempty"`
	OrganizationID *uuid.UUID      `json:"organizationId,omitempty"`
	TaskCount      int64           `json:"taskCount,omitempty"`
	MemberCount    int64           `json:"memberCount,omitempty"`
}

// ProjectListResponse provides pagination wrappers for project lists.
type ProjectListResponse struct {
	Data       []ProjectResponse `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"pageSize"`
	TotalPages int               `json:"totalPages"`
}
