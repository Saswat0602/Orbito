package workspaces

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CreateWorkspaceRequest contains payload for workspace creation.
type CreateWorkspaceRequest struct {
	Name              string          `json:"name" validate:"required,min=1,max=100"`
	Slug              string          `json:"slug" validate:"required,min=1,max=50"`
	Description       *string         `json:"description,omitempty" validate:"omitempty,max=500"`
	Avatar            *string         `json:"avatar,omitempty"`
	Color             *string         `json:"color,omitempty"`
	Settings          json.RawMessage `json:"settings,omitempty"`
	OrganizationID    uuid.UUID       `json:"organizationId" validate:"required"`
	ParentWorkspaceID *uuid.UUID      `json:"parentWorkspaceId,omitempty"`
	InheritMembers    *bool           `json:"inheritMembers,omitempty"`
}

// UpdateWorkspaceRequest contains fields for updating a workspace.
type UpdateWorkspaceRequest struct {
	Name        *string         `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Slug        *string         `json:"slug,omitempty" validate:"omitempty,min=1,max=50"`
	Description *string         `json:"description,omitempty"`
	Avatar      *string         `json:"avatar,omitempty"`
	Color       *string         `json:"color,omitempty"`
	Settings    json.RawMessage `json:"settings,omitempty"`
}

// WorkspaceResponse represents a single workspace response object.
type WorkspaceResponse struct {
	ID                uuid.UUID       `json:"id"`
	Name              string          `json:"name"`
	Slug              string          `json:"slug"`
	Description       *string         `json:"description,omitempty"`
	Avatar            *string         `json:"avatar,omitempty"`
	Color             *string         `json:"color,omitempty"`
	Settings          json.RawMessage `json:"settings,omitempty"`
	OrganizationID    uuid.UUID       `json:"organizationId"`
	ParentWorkspaceID *uuid.UUID      `json:"parentWorkspaceId,omitempty"`
	Path              *string         `json:"path,omitempty"`
	Role              *string         `json:"role,omitempty"`
	Archive           bool            `json:"archive"`
	CreatedAt         time.Time       `json:"createdAt"`
	UpdatedAt         time.Time       `json:"updatedAt"`
}

// WorkspaceMemberDTO represents a member belonging to a workspace.
type WorkspaceMemberDTO struct {
	ID          uuid.UUID `json:"id"`
	Role        string    `json:"role"`
	JoinedAt    time.Time `json:"joinedAt"`
	UserID      uuid.UUID `json:"userId"`
	WorkspaceID uuid.UUID `json:"workspaceId"`
	UserEmail   string    `json:"userEmail"`
	FirstName   string    `json:"firstName"`
	LastName    string    `json:"lastName"`
	Avatar      *string   `json:"avatar,omitempty"`
}

// WorkspaceKPIMetrics contains aggregated task and project health statistics.
type WorkspaceKPIMetrics struct {
	TotalProjects     int64   `json:"totalProjects"`
	ActiveProjects    int64   `json:"activeProjects"`
	CompletedProjects int64   `json:"completedProjects"`
	TotalTasks        int64   `json:"totalTasks"`
	CompletedTasks    int64   `json:"completedTasks"`
	OverdueTasks      int64   `json:"overdueTasks"`
	CompletionRate    float64 `json:"completionRate"`
}

// WorkspaceChartDataResponse contains chart datasets for a workspace.
type WorkspaceChartDataResponse struct {
	KPIMetrics    *WorkspaceKPIMetrics `json:"kpi-metrics,omitempty"`
	ProjectStatus []map[string]any     `json:"project-status,omitempty"`
	TaskPriority  []map[string]any     `json:"task-priority,omitempty"`
}
