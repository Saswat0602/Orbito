package public

import (
	"time"

	"github.com/google/uuid"
)

// StatusBadge holds minimal workflow status display metadata.
type StatusBadge struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Category string `json:"category,omitempty"`
}

// PersonSummary represents sanitized user fields safe for public exposure.
type PersonSummary struct {
	ID        *uuid.UUID `json:"id,omitempty"`
	FirstName string     `json:"firstName"`
	LastName  string     `json:"lastName"`
	Avatar    *string    `json:"avatar,omitempty"`
}

// SharedAttachmentItem represents minimal attachment metadata on a shared task.
type SharedAttachmentItem struct {
	ID       uuid.UUID `json:"id"`
	FileName string    `json:"fileName"`
	FileSize int32     `json:"fileSize"`
	MimeType string    `json:"mimeType"`
}

// SharedTaskResponse matches NestJS PublicSharedTaskDto.
type SharedTaskResponse struct {
	Title       string                 `json:"title"`
	Description *string                `json:"description,omitempty"`
	Status      StatusBadge            `json:"status"`
	Priority    string                 `json:"priority"`
	DueDate     *time.Time             `json:"dueDate,omitempty"`
	Assignees   []PersonSummary        `json:"assignees"`
	CreatedBy   *PersonSummary         `json:"createdBy,omitempty"`
	Attachments []SharedAttachmentItem `json:"attachments"`
}

// AttachmentURLResponse matches GET /api/public/tasks/{token}/attachments/{attachmentId}.
type AttachmentURLResponse struct {
	URL string `json:"url"`
}

// WorkspaceSummary holds public workspace metadata.
type WorkspaceSummary struct {
	ID   *uuid.UUID `json:"id,omitempty"`
	Slug string     `json:"slug"`
	Name string     `json:"name"`
}

// OrganizationSummary holds public organization metadata.
type OrganizationSummary struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

// ProjectStats holds capped public project metrics.
type ProjectStats struct {
	TaskCount        int64   `json:"taskCount"`
	CompletionRate   float64 `json:"completionRate"`
	HasActiveSprints bool    `json:"hasActiveSprints"`
}

// ProjectResponse matches NestJS PublicProjectDto.
type ProjectResponse struct {
	ID           uuid.UUID            `json:"id"`
	Name         string               `json:"name"`
	Slug         string               `json:"slug"`
	Description  *string              `json:"description,omitempty"`
	Color        string               `json:"color"`
	Avatar       *string              `json:"avatar,omitempty"`
	Status       string               `json:"status"`
	Priority     string               `json:"priority"`
	Visibility   string               `json:"visibility"`
	StartDate    *time.Time           `json:"startDate,omitempty"`
	EndDate      *time.Time           `json:"endDate,omitempty"`
	CreatedAt    time.Time            `json:"createdAt"`
	Workspace    WorkspaceSummary     `json:"workspace"`
	Organization *OrganizationSummary `json:"organization,omitempty"`
	IsPublicView bool                 `json:"isPublicView"`
	Stats        *ProjectStats        `json:"stats,omitempty"`
}

// WorkflowSummary holds minimal workflow context for public statuses.
type WorkflowSummary struct {
	ID           uuid.UUID           `json:"id"`
	Name         string              `json:"name"`
	Description  *string             `json:"description,omitempty"`
	Organization OrganizationSummary `json:"organization"`
}

// StatusCount holds Prisma-compatible _count.tasks projection.
type StatusCount struct {
	Tasks int64 `json:"tasks"`
}

// StatusResponse represents a public project workflow status item.
type StatusResponse struct {
	ID         uuid.UUID       `json:"id"`
	Name       string          `json:"name"`
	Color      string          `json:"color"`
	Category   string          `json:"category"`
	Position   int32           `json:"position"`
	WorkflowID uuid.UUID       `json:"workflowId"`
	CreatedAt  time.Time       `json:"createdAt"`
	UpdatedAt  time.Time       `json:"updatedAt"`
	Workflow   WorkflowSummary `json:"workflow"`
	Count      StatusCount     `json:"_count"`
}

// StatusFlowItem matches ProjectChartsService task-status output.
type StatusFlowItem struct {
	StatusID uuid.UUID      `json:"statusId"`
	Count    int64          `json:"count"`
	Status   StatusFlowMeta `json:"status"`
}

// StatusFlowMeta holds nested status fields inside StatusFlowItem.
type StatusFlowMeta struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Color    string    `json:"color"`
	Category string    `json:"category"`
	Position int32     `json:"position"`
}

// TaskTypeCount matches Prisma groupBy type count item.
type TaskTypeCount struct {
	Type  string           `json:"type"`
	Count map[string]int64 `json:"_count"`
}

// TaskPriorityCount matches Prisma groupBy priority count item.
type TaskPriorityCount struct {
	Priority string           `json:"priority"`
	Count    map[string]int64 `json:"_count"`
}

// KPIMetrics matches ProjectChartsService kpi-metrics output.
type KPIMetrics struct {
	TotalTasks        int64   `json:"totalTasks"`
	CompletedTasks    int64   `json:"completedTasks"`
	ActiveSprints     int64   `json:"activeSprints"`
	TotalBugs         int64   `json:"totalBugs"`
	ResolvedBugs      int64   `json:"resolvedBugs"`
	CompletionRate    float64 `json:"completionRate"`
	BugResolutionRate float64 `json:"bugResolutionRate"`
}

// SprintVelocityItem matches ProjectChartsService sprint-velocity item.
type SprintVelocityItem struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	StartDate *time.Time `json:"startDate"`
	EndDate   *time.Time `json:"endDate"`
	Velocity  int64      `json:"velocity"`
}

// CalendarEvent matches PublicCalendarService event items.
type CalendarEvent struct {
	ID       string       `json:"id"`
	Title    string       `json:"title"`
	Date     time.Time    `json:"date"`
	Type     string       `json:"type"`
	Status   *StatusBadge `json:"status,omitempty"`
	TaskID   *uuid.UUID   `json:"taskId,omitempty"`
	SprintID *uuid.UUID   `json:"sprintId,omitempty"`
}

// LabelBadge holds public label metadata.
type LabelBadge struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Color string    `json:"color"`
}

// TaskResponse matches NestJS PublicTaskDto.
type TaskResponse struct {
	ID           uuid.UUID       `json:"id"`
	Title        string          `json:"title"`
	Description  *string         `json:"description,omitempty"`
	Type         string          `json:"type"`
	Priority     string          `json:"priority"`
	Slug         string          `json:"slug,omitempty"`
	DueDate      *time.Time      `json:"dueDate,omitempty"`
	CreatedAt    time.Time       `json:"createdAt"`
	Status       StatusBadge     `json:"status"`
	Labels       []LabelBadge    `json:"labels,omitempty"`
	Subtasks     []TaskResponse  `json:"subtasks,omitempty"`
	Assignees    []PersonSummary `json:"assignees,omitempty"`
	Project      any             `json:"project,omitempty"`
	Sprint       any             `json:"sprint,omitempty"`
	ParentTask   any             `json:"parentTask,omitempty"`
	IsPublicView bool            `json:"isPublicView"`
}

// TaskFilters configures pagination and attribute filtering for public tasks.
type TaskFilters struct {
	ParentTaskID string
	Status       string
	Priority     string
	Type         string
	Page         int
	Limit        int
}

// TaskPaginationResponse matches NestJS PublicTaskPagination.
type TaskPaginationResponse struct {
	Data       []TaskResponse `json:"data"`
	Page       int            `json:"page"`
	Total      int64          `json:"total"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"totalPages"`
}

// KanbanColumn represents a single workflow column on the public Kanban board.
type KanbanColumn struct {
	ID       uuid.UUID      `json:"id"`
	Name     string         `json:"name"`
	Color    string         `json:"color"`
	Category string         `json:"category"`
	Tasks    []TaskResponse `json:"tasks"`
}

// KanbanBoardResponse wraps public Kanban columns.
type KanbanBoardResponse struct {
	Columns []KanbanColumn `json:"columns"`
}

// SprintResponse matches NestJS PublicSprintDto.
type SprintResponse struct {
	ID           uuid.UUID      `json:"id"`
	Name         string         `json:"name"`
	Description  *string        `json:"description,omitempty"`
	Status       string         `json:"status"`
	StartDate    *time.Time     `json:"startDate,omitempty"`
	EndDate      *time.Time     `json:"endDate,omitempty"`
	CreatedAt    time.Time      `json:"createdAt"`
	Progress     int            `json:"progress"`
	Tasks        []TaskResponse `json:"tasks,omitempty"`
	IsPublicView bool           `json:"isPublicView"`
}

// ActivityItem represents an anonymized public task activity item.
type ActivityItem struct {
	ID          uuid.UUID      `json:"id"`
	Type        string         `json:"type"`
	Description string         `json:"description"`
	EntityType  string         `json:"entityType"`
	EntityID    string         `json:"entityId"`
	CreatedAt   time.Time      `json:"createdAt"`
	User        *PersonSummary `json:"user,omitempty"`
}

// PaginationMeta holds pagination counters matching getTaskActivities.
type PaginationMeta struct {
	CurrentPage int   `json:"currentPage"`
	TotalPages  int   `json:"totalPages"`
	TotalCount  int64 `json:"totalCount"`
	HasNextPage bool  `json:"hasNextPage"`
	HasPrevPage bool  `json:"hasPrevPage"`
}

// ActivitiesResponse matches NestJS getTaskActivities output.
type ActivitiesResponse struct {
	Activities []ActivityItem `json:"activities"`
	Pagination PaginationMeta `json:"pagination"`
}

// CommentCount holds reply counter metadata.
type CommentCount struct {
	Replies int `json:"replies"`
}

// CommentResponse represents a public top-level task comment with nested replies.
type CommentResponse struct {
	ID              uuid.UUID         `json:"id"`
	Content         string            `json:"content"`
	ParentCommentID *uuid.UUID        `json:"parentCommentId,omitempty"`
	CreatedAt       time.Time         `json:"createdAt"`
	UpdatedAt       time.Time         `json:"updatedAt"`
	Author          PersonSummary     `json:"author"`
	Replies         []CommentResponse `json:"replies,omitempty"`
	Count           CommentCount      `json:"_count"`
}

// AttachmentItem represents a public project task attachment item.
type AttachmentItem struct {
	ID        uuid.UUID `json:"id"`
	FileName  string    `json:"fileName"`
	FileSize  int32     `json:"fileSize"`
	MimeType  string    `json:"mimeType"`
	URL       *string   `json:"url,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}
