package labels

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// CreateLabelRequest represents POST /api/labels.
type CreateLabelRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Color       string  `json:"color" validate:"required,hexcolor"`
	Description *string `json:"description" validate:"omitempty,max=500"`
	ProjectID   string  `json:"projectId" validate:"required,uuid"`
}

// UpdateLabelRequest represents PATCH /api/labels/:id.
type UpdateLabelRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=100"`
	Color       *string `json:"color" validate:"omitempty,hexcolor"`
	Description *string `json:"description" validate:"omitempty,max=500"`
	ProjectID   *string `json:"projectId" validate:"omitempty,uuid"`
}

// AssignLabelRequest represents POST /api/labels/assign.
type AssignLabelRequest struct {
	TaskID  string `json:"taskId" validate:"required,min=1"`
	LabelID string `json:"labelId" validate:"required,uuid"`
}

// AssignMultipleLabelsRequest represents POST /api/labels/assign-multiple.
type AssignMultipleLabelsRequest struct {
	TaskID   string   `json:"taskId" validate:"required,min=1"`
	LabelIDs []string `json:"labelIds" validate:"required,dive,uuid"`
}

// OrgBrief represents condensed organization details.
type OrgBrief struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// WorkspaceBrief represents condensed workspace details.
type WorkspaceBrief struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Organization *OrgBrief `json:"organization,omitempty"`
}

// ProjectBrief represents condensed project details.
type ProjectBrief struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Slug      string          `json:"slug"`
	Workspace *WorkspaceBrief `json:"workspace,omitempty"`
}

// UserBrief represents condensed creator/updater info.
type UserBrief struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
}

// AssigneeBrief represents condensed task assignee info.
type AssigneeBrief struct {
	ID        string  `json:"id"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Avatar    *string `json:"avatar"`
}

// TaskStatusBrief represents condensed task status info.
type TaskStatusBrief struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Category string `json:"category"`
}

// TaskSummary represents condensed task info nested under taskLabels.
type TaskSummary struct {
	ID        string           `json:"id"`
	Title     string           `json:"title"`
	Slug      string           `json:"slug"`
	Type      string           `json:"type"`
	Priority  string           `json:"priority"`
	Status    *TaskStatusBrief `json:"status,omitempty"`
	Assignees []AssigneeBrief  `json:"assignees"`
}

// TaskLabelEntry represents a label-to-task link with nested task details.
type TaskLabelEntry struct {
	TaskID      string       `json:"taskId"`
	LabelID     string       `json:"labelId"`
	CreatedByID *string      `json:"createdById"`
	UpdatedByID *string      `json:"updatedById"`
	CreatedAt   string       `json:"createdAt"`
	UpdatedAt   string       `json:"updatedAt"`
	Task        *TaskSummary `json:"task,omitempty"`
}

// LabelCount represents the Prisma-compatible _count wrapper.
type LabelCount struct {
	TaskLabels int64 `json:"taskLabels"`
}

// LabelResponse represents serialized label data matching NestJS responses.
type LabelResponse struct {
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	Color         string           `json:"color"`
	Description   *string          `json:"description"`
	ProjectID     string           `json:"projectId"`
	CreatedByID   *string          `json:"createdById"`
	UpdatedByID   *string          `json:"updatedById"`
	CreatedAt     string           `json:"createdAt"`
	UpdatedAt     string           `json:"updatedAt"`
	Project       *ProjectBrief    `json:"project,omitempty"`
	CreatedByUser *UserBrief       `json:"createdByUser,omitempty"`
	UpdatedByUser *UserBrief       `json:"updatedByUser,omitempty"`
	TaskLabels    []TaskLabelEntry `json:"taskLabels,omitempty"`
	Count         *LabelCount      `json:"_count,omitempty"`
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
