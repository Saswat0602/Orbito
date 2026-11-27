package task_label

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// AssignRequest represents AssignTaskLabelDto.
type AssignRequest struct {
	TaskID  string `json:"taskId" validate:"required"`
	LabelID string `json:"labelId" validate:"required,uuid"`
}

// AssignMultipleRequest represents AssignMultipleTaskLabelsDto.
type AssignMultipleRequest struct {
	TaskID   string   `json:"taskId" validate:"required"`
	LabelIDs []string `json:"labelIds" validate:"required,min=1,dive,uuid"`
}

// LabelBrief represents label metadata returned in TaskLabel assignments.
type LabelBrief struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Color       string  `json:"color"`
	Description *string `json:"description"`
}

// TaskBrief represents task metadata returned in TaskLabel assignments.
type TaskBrief struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Slug  string `json:"slug"`
}

// AssignmentResponse represents a TaskLabel record with label and task expansions.
type AssignmentResponse struct {
	TaskID    string      `json:"taskId"`
	LabelID   string      `json:"labelId"`
	CreatedBy *string     `json:"createdBy"`
	UpdatedBy *string     `json:"updatedBy"`
	CreatedAt string      `json:"createdAt"`
	UpdatedAt string      `json:"updatedAt"`
	Label     *LabelBrief `json:"label,omitempty"`
	Task      *TaskBrief  `json:"task,omitempty"`
}

func uuidPtr(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	s := uuid.UUID(u.Bytes).String()
	return &s
}

func textPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	s := t.String
	return &s
}

func tsString(ts pgtype.Timestamp) string {
	if !ts.Valid {
		return ""
	}
	return ts.Time.UTC().Format("2006-01-02T15:04:05.000Z")
}

func mapPairRow(r sqlc.GetTaskLabelPairDetailsRow) AssignmentResponse {
	return AssignmentResponse{
		TaskID:    r.TaskID.String(),
		LabelID:   r.LabelID.String(),
		CreatedBy: uuidPtr(r.CreatedByID),
		UpdatedBy: uuidPtr(r.UpdatedByID),
		CreatedAt: tsString(r.CreatedAt),
		UpdatedAt: tsString(r.UpdatedAt),
		Label: &LabelBrief{
			ID:          r.LabelID.String(),
			Name:        r.LabelName,
			Color:       r.LabelColor,
			Description: textPtr(r.LabelDescription),
		},
		Task: &TaskBrief{
			ID:    r.TaskID.String(),
			Title: r.TaskTitle,
			Slug:  r.TaskSlug,
		},
	}
}
