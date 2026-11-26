package task_statuses

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// CreateTaskStatusRequest represents the payload for POST /api/task-statuses.
type CreateTaskStatusRequest struct {
	Name       string `json:"name" validate:"required,min=1,max=50"`
	Color      string `json:"color" validate:"required,hexcolor"`
	Category   string `json:"category" validate:"required,oneof=TODO IN_PROGRESS DONE"`
	Position   *int32 `json:"position" validate:"required,min=0"`
	WorkflowID string `json:"workflowId" validate:"required,min=1"`
}

// CreateTaskStatusFromProjectRequest represents POST /api/task-statuses/from-project.
type CreateTaskStatusFromProjectRequest struct {
	Name      string `json:"name" validate:"required,min=1"`
	ProjectID string `json:"projectId" validate:"required,uuid"`
}

// UpdateTaskStatusRequest represents PATCH /api/task-statuses/:id.
type UpdateTaskStatusRequest struct {
	Name       *string `json:"name" validate:"omitempty,min=1,max=50"`
	Color      *string `json:"color" validate:"omitempty,hexcolor"`
	Category   *string `json:"category" validate:"omitempty,oneof=TODO IN_PROGRESS DONE"`
	Position   *int32  `json:"position" validate:"omitempty,min=0"`
	WorkflowID *string `json:"workflowId" validate:"omitempty"`
	DeletedAt  *string `json:"deletedAt"`
	DeletedBy  *string `json:"deletedBy" validate:"omitempty,uuid"`
}

// StatusPositionItem represents one item in batch position updates.
type StatusPositionItem struct {
	ID       string `json:"id" validate:"required,min=1"`
	Position int32  `json:"position"`
}

// UpdatePositionsRequest represents PATCH /api/task-statuses/positions.
type UpdatePositionsRequest struct {
	StatusUpdates []StatusPositionItem `json:"statusUpdates" validate:"required,dive"`
}

// OrgBrief represents condensed organization info.
type OrgBrief struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// WorkflowBrief represents condensed workflow info nested in task status responses.
type WorkflowBrief struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  *string   `json:"description"`
	Organization *OrgBrief `json:"organization,omitempty"`
}

// UserBrief represents condensed user info.
type UserBrief struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
}

// TaskCount represents the Prisma-compatible _count wrapper.
type TaskCount struct {
	Tasks int64 `json:"tasks"`
}

// TransitionStatusBrief represents condensed status info nested in transitions.
type TransitionStatusBrief struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Category string `json:"category"`
}

// StatusTransitionItem represents an outgoing or incoming status transition.
type StatusTransitionItem struct {
	ID           string                 `json:"id"`
	Name         *string                `json:"name"`
	Description  *string                `json:"description"`
	WorkflowID   string                 `json:"workflowId"`
	FromStatusID string                 `json:"fromStatusId"`
	ToStatusID   string                 `json:"toStatusId"`
	CreatedByID  *string                `json:"createdById"`
	UpdatedByID  *string                `json:"updatedById"`
	CreatedAt    string                 `json:"createdAt"`
	UpdatedAt    string                 `json:"updatedAt"`
	ToStatus     *TransitionStatusBrief `json:"toStatus,omitempty"`
	FromStatus   *TransitionStatusBrief `json:"fromStatus,omitempty"`
}

// TaskStatusResponse represents the serialized task status response matching NestJS/Prisma shapes.
type TaskStatusResponse struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Color           string                 `json:"color"`
	Category        string                 `json:"category"`
	Position        int32                  `json:"position"`
	IsDefault       bool                   `json:"isDefault"`
	WorkflowID      string                 `json:"workflowId"`
	CreatedAt       string                 `json:"createdAt"`
	UpdatedAt       string                 `json:"updatedAt"`
	CreatedByID     *string                `json:"createdById"`
	UpdatedByID     *string                `json:"updatedById"`
	DeletedAt       *string                `json:"deletedAt"`
	DeletedBy       *string                `json:"deletedBy"`
	Workflow        *WorkflowBrief         `json:"workflow,omitempty"`
	CreatedByUser   *UserBrief             `json:"createdByUser,omitempty"`
	UpdatedByUser   *UserBrief             `json:"updatedByUser,omitempty"`
	DeletedByUser   *UserBrief             `json:"deletedByUser,omitempty"`
	FromTransitions []StatusTransitionItem `json:"fromTransitions,omitempty"`
	ToTransitions   []StatusTransitionItem `json:"toTransitions,omitempty"`
	Count           *TaskCount             `json:"_count,omitempty"`
}

func formatTS(ts pgtype.Timestamp) string {
	if ts.Valid {
		return ts.Time.UTC().Format("2006-01-02T15:04:05.000Z")
	}
	return ""
}

func pgTextPtr(t pgtype.Text) *string {
	if t.Valid {
		return &t.String
	}
	return nil
}

func pgUUIDPtr(u pgtype.UUID) *string {
	if u.Valid {
		s := uuid.UUID(u.Bytes).String()
		return &s
	}
	return nil
}

func pgTSPtr(ts pgtype.Timestamp) *string {
	if ts.Valid {
		s := ts.Time.UTC().Format("2006-01-02T15:04:05.000Z")
		return &s
	}
	return nil
}

