package task_dependencies

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// CreateRequest represents the CreateTaskDependencyDto payload.
type CreateRequest struct {
	Type            *string `json:"type" validate:"omitempty,oneof=BLOCKS DEPENDS_ON RELATES_TO DUPLICATES"`
	DependentTaskID string  `json:"dependentTaskId" validate:"required"`
	BlockingTaskID  string  `json:"blockingTaskId" validate:"required"`
	CreatedBy       *string `json:"createdBy" validate:"omitempty,uuid"`
}

// BulkCreateRequest represents the BulkCreateDependenciesDto payload.
type BulkCreateRequest struct {
	Dependencies []CreateRequest `json:"dependencies" validate:"required,min=1,dive"`
}

// UpdateRequest represents the UpdateTaskDependencyDto payload.
type UpdateRequest struct {
	Type            *string `json:"type" validate:"omitempty,oneof=BLOCKS DEPENDS_ON RELATES_TO DUPLICATES"`
	DependentTaskID *string `json:"dependentTaskId" validate:"omitempty,min=1"`
	BlockingTaskID  *string `json:"blockingTaskId" validate:"omitempty,min=1"`
}

// TaskSummaryResponse describes minimal dependent/blocking task metadata.
type TaskSummaryResponse struct {
	ID        string          `json:"id"`
	Title     string          `json:"title"`
	Slug      string          `json:"slug"`
	ProjectID *string         `json:"projectId,omitempty"`
	Status    *StatusNameWrap `json:"status,omitempty"`
}

// StatusNameWrap mirrors Prisma's nested `{ status: { name } }` projection.
type StatusNameWrap struct {
	Name string `json:"name"`
}

// CreatorSummaryResponse describes minimal creator metadata.
type CreatorSummaryResponse struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

// DependencyResponse represents a full task dependency record.
type DependencyResponse struct {
	ID              string                  `json:"id"`
	Type            string                  `json:"type"`
	DependentTaskID string                  `json:"dependentTaskId"`
	BlockingTaskID  string                  `json:"blockingTaskId"`
	CreatedBy       *string                 `json:"createdBy"`
	UpdatedBy       *string                 `json:"updatedBy"`
	CreatedAt       string                  `json:"createdAt"`
	UpdatedAt       string                  `json:"updatedAt"`
	DependentTask   *TaskSummaryResponse    `json:"dependentTask,omitempty"`
	BlockingTask    *TaskSummaryResponse    `json:"blockingTask,omitempty"`
	CreatedByUser   *CreatorSummaryResponse `json:"createdByUser,omitempty"`
}

// TaskDependenciesSummaryResponse represents GET /task-dependencies/task/:taskId.
type TaskDependenciesSummaryResponse struct {
	DependsOn []DependencyResponse `json:"dependsOn"`
	Blocks    []DependencyResponse `json:"blocks"`
}

// DependencyStatsResponse represents GET /task-dependencies/stats.
type DependencyStatsResponse struct {
	TotalDependencies int      `json:"totalDependencies"`
	BlockedTasks      int      `json:"blockedTasks"`
	CriticalPath      []string `json:"criticalPath"`
}

func formatTS(ts pgtype.Timestamp) string {
	if !ts.Valid {
		return ""
	}
	return ts.Time.UTC().Format("2006-01-02T15:04:05.000Z")
}

func pgUUIDStrPtr(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	s := uuid.UUID(u.Bytes).String()
	return &s
}

func mapDetailRow(r *sqlc.GetTaskDependencyDetailsByIDRow) DependencyResponse {
	depPID := r.DepProjectID.String()
	blkPID := r.BlkProjectID.String()
	resp := DependencyResponse{
		ID:              r.ID.String(),
		Type:            string(r.Type),
		DependentTaskID: r.DependentTaskID.String(),
		BlockingTaskID:  r.BlockingTaskID.String(),
		CreatedBy:       pgUUIDStrPtr(r.CreatedByID),
		UpdatedBy:       pgUUIDStrPtr(r.UpdatedByID),
		CreatedAt:       formatTS(r.CreatedAt),
		UpdatedAt:       formatTS(r.UpdatedAt),
		DependentTask: &TaskSummaryResponse{
			ID:        r.DependentTaskID.String(),
			Title:     r.DepTitle,
			Slug:      r.DepSlug,
			ProjectID: &depPID,
			Status:    &StatusNameWrap{Name: r.DepStatusName},
		},
		BlockingTask: &TaskSummaryResponse{
			ID:        r.BlockingTaskID.String(),
			Title:     r.BlkTitle,
			Slug:      r.BlkSlug,
			ProjectID: &blkPID,
			Status:    &StatusNameWrap{Name: r.BlkStatusName},
		},
	}
	if r.CreatorID.Valid {
		resp.CreatedByUser = &CreatorSummaryResponse{
			ID:        uuid.UUID(r.CreatorID.Bytes).String(),
			FirstName: r.CreatorFirstName.String,
			LastName:  r.CreatorLastName.String,
			Email:     r.CreatorEmail.String,
		}
	}
	return resp
}

func mapListRow(r sqlc.ListTaskDependenciesRow) DependencyResponse {
	detail := sqlc.GetTaskDependencyDetailsByIDRow(r)
	return mapDetailRow(&detail)
}

func mapDependsOnRow(r sqlc.ListDependenciesTaskDependsOnRow) DependencyResponse {
	return DependencyResponse{
		ID:              r.ID.String(),
		Type:            string(r.Type),
		DependentTaskID: r.DependentTaskID.String(),
		BlockingTaskID:  r.BlockingTaskID.String(),
		CreatedBy:       pgUUIDStrPtr(r.CreatedByID),
		UpdatedBy:       pgUUIDStrPtr(r.UpdatedByID),
		CreatedAt:       formatTS(r.CreatedAt),
		UpdatedAt:       formatTS(r.UpdatedAt),
		BlockingTask: &TaskSummaryResponse{
			ID:     r.BlockingTaskID.String(),
			Title:  r.BlkTitle,
			Slug:   r.BlkSlug,
			Status: &StatusNameWrap{Name: r.BlkStatusName},
		},
	}
}

func mapBlocksRow(r sqlc.ListDependenciesTaskBlocksRow) DependencyResponse {
	return DependencyResponse{
		ID:              r.ID.String(),
		Type:            string(r.Type),
		DependentTaskID: r.DependentTaskID.String(),
		BlockingTaskID:  r.BlockingTaskID.String(),
		CreatedBy:       pgUUIDStrPtr(r.CreatedByID),
		UpdatedBy:       pgUUIDStrPtr(r.UpdatedByID),
		CreatedAt:       formatTS(r.CreatedAt),
		UpdatedAt:       formatTS(r.UpdatedAt),
		DependentTask: &TaskSummaryResponse{
			ID:     r.DependentTaskID.String(),
			Title:  r.DepTitle,
			Slug:   r.DepSlug,
			Status: &StatusNameWrap{Name: r.DepStatusName},
		},
	}
}
