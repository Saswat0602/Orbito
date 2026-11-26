package workflows

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// CreateWorkflowRequest represents the payload to create a new workflow.
type CreateWorkflowRequest struct {
	Name           string  `json:"name" validate:"required,min=1,max=100"`
	Description    *string `json:"description" validate:"omitempty,max=500"`
	IsDefault      *bool   `json:"isDefault"`
	OrganizationID string  `json:"organizationId" validate:"required,uuid"`
}

// UpdateWorkflowRequest represents the payload to update an existing workflow.
type UpdateWorkflowRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description" validate:"omitempty,max=500"`
	IsDefault   *bool   `json:"isDefault"`
}

// MakeDefaultWorkflowRequest represents the payload to set a workflow as default.
type MakeDefaultWorkflowRequest struct {
	OrganizationID string `json:"organizationId" validate:"required,uuid"`
}

// OrgBrief represents condensed organization details in responses.
type OrgBrief struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// UserBrief represents condensed user profile details in responses.
type UserBrief struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
}

// WorkflowCounts represents aggregate relation counts.
type WorkflowCounts struct {
	Statuses    int64 `json:"statuses"`
	Transitions int64 `json:"transitions"`
}

// TaskStatusResponse represents a single task status node.
type TaskStatusResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Color       string  `json:"color"`
	Category    string  `json:"category"`
	Position    int32   `json:"position"`
	IsDefault   bool    `json:"isDefault"`
	WorkflowID  string  `json:"workflowId"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
	CreatedByID *string `json:"createdById,omitempty"`
	UpdatedByID *string `json:"updatedById,omitempty"`
	DeletedAt   *string `json:"deletedAt"`
	DeletedBy   *string `json:"deletedBy"`
}

// TransitionStatusBrief represents condensed status info in transition responses.
type TransitionStatusBrief struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Category string `json:"category"`
}

// StatusTransitionResponse represents a allowed movement between statuses.
type StatusTransitionResponse struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  *string                `json:"description"`
	WorkflowID   string                 `json:"workflowId"`
	FromStatusID string                 `json:"fromStatusId"`
	ToStatusID   string                 `json:"toStatusId"`
	CreatedByID  *string                `json:"createdById,omitempty"`
	UpdatedByID  *string                `json:"updatedById,omitempty"`
	CreatedAt    string                 `json:"createdAt"`
	UpdatedAt    string                 `json:"updatedAt"`
	FromStatus   *TransitionStatusBrief `json:"fromStatus,omitempty"`
	ToStatus     *TransitionStatusBrief `json:"toStatus,omitempty"`
}

// WorkflowResponse represents full or listing workflow responses matching NestJS schema.
type WorkflowResponse struct {
	ID             string                     `json:"id"`
	Name           string                     `json:"name"`
	Description    *string                    `json:"description"`
	IsDefault      bool                       `json:"isDefault"`
	OrganizationID string                     `json:"organizationId"`
	CreatedByID    *string                    `json:"createdById,omitempty"`
	UpdatedByID    *string                    `json:"updatedById,omitempty"`
	CreatedAt      string                     `json:"createdAt"`
	UpdatedAt      string                     `json:"updatedAt"`
	Organization   *OrgBrief                  `json:"organization,omitempty"`
	CreatedByUser  *UserBrief                 `json:"createdByUser,omitempty"`
	UpdatedByUser  *UserBrief                 `json:"updatedByUser,omitempty"`
	Statuses       []TaskStatusResponse       `json:"statuses,omitempty"`
	Transitions    []StatusTransitionResponse `json:"transitions,omitempty"`
	Count          *WorkflowCounts            `json:"_count,omitempty"`
}

func formatTime(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

func pgTextToPtr(t pgtype.Text) *string {
	if t.Valid {
		return &t.String
	}
	return nil
}

func pgUUIDToPtr(u pgtype.UUID) *string {
	if u.Valid {
		id := uuid.UUID(u.Bytes).String()
		return &id
	}
	return nil
}

func pgTimestampToPtr(ts pgtype.Timestamp) *string {
	if ts.Valid {
		s := formatTime(ts.Time)
		return &s
	}
	return nil
}
