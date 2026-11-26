package tasks

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// RecurrenceConfigRequest configures task recurrence cadence and end behavior.
type RecurrenceConfigRequest struct {
	RecurrenceType  string  `json:"recurrenceType" validate:"required,oneof=DAILY WEEKLY MONTHLY QUARTERLY YEARLY CUSTOM"`
	Interval        *int32  `json:"interval" validate:"required,min=1"`
	DaysOfWeek      []int32 `json:"daysOfWeek" validate:"omitempty,dive,min=0,max=6"`
	DayOfMonth      *int32  `json:"dayOfMonth" validate:"omitempty,min=1,max=31"`
	MonthOfYear     *int32  `json:"monthOfYear" validate:"omitempty,min=1,max=12"`
	EndType         string  `json:"endType" validate:"required,oneof=NEVER ON_DATE AFTER_OCCURRENCES"`
	EndDate         *string `json:"endDate"`
	OccurrenceCount *int32  `json:"occurrenceCount" validate:"omitempty,min=1"`
}

// CreateTaskRequest represents POST /api/tasks and POST /api/tasks/create-task-attachment payloads.
type CreateTaskRequest struct {
	Title             string                   `json:"title" validate:"required,min=1,max=200"`
	Description       *string                  `json:"description" validate:"omitempty,max=20000"`
	Type              *string                  `json:"type" validate:"omitempty,oneof=TASK STORY BUG EPIC SUBTASK"`
	Priority          *string                  `json:"priority" validate:"omitempty,oneof=LOWEST LOW MEDIUM HIGH HIGHEST"`
	StartDate         *string                  `json:"startDate"`
	DueDate           *string                  `json:"dueDate"`
	CompletedAt       *string                  `json:"completedAt"`
	StoryPoints       *int32                   `json:"storyPoints"`
	OriginalEstimate  *int32                   `json:"originalEstimate"`
	RemainingEstimate *int32                   `json:"remainingEstimate"`
	CustomFields      map[string]any           `json:"customFields"`
	ProjectID         string                   `json:"projectId" validate:"required,uuid"`
	AssigneeIDs       []string                 `json:"assigneeIds" validate:"omitempty,dive,uuid"`
	AssigneeID        *string                  `json:"assigneeId" validate:"omitempty,uuid"`
	ReporterIDs       []string                 `json:"reporterIds" validate:"omitempty,dive,uuid"`
	StatusID          *string                  `json:"statusId" validate:"omitempty,uuid"`
	SprintID          *string                  `json:"sprintId" validate:"omitempty,uuid"`
	ParentTaskID      *string                  `json:"parentTaskId"`
	AllowEmailReplies *bool                    `json:"allowEmailReplies"`
	IsRecurring       *bool                    `json:"isRecurring"`
	RecurrenceConfig  *RecurrenceConfigRequest `json:"recurrenceConfig"`
}

// UpdateTaskRequest represents PATCH /api/tasks/:id.
type UpdateTaskRequest struct {
	Title             *string                  `json:"title" validate:"omitempty,min=1,max=200"`
	Description       *string                  `json:"description"`
	Type              *string                  `json:"type" validate:"omitempty,oneof=TASK STORY BUG EPIC SUBTASK"`
	Priority          *string                  `json:"priority" validate:"omitempty,oneof=LOWEST LOW MEDIUM HIGH HIGHEST"`
	StartDate         *string                  `json:"startDate"`
	DueDate           *string                  `json:"dueDate"`
	CompletedAt       *string                  `json:"completedAt"`
	ClearCompletedAt  bool                     `json:"-"`
	StoryPoints       *int32                   `json:"storyPoints"`
	OriginalEstimate  *int32                   `json:"originalEstimate"`
	RemainingEstimate *int32                   `json:"remainingEstimate"`
	CustomFields      map[string]any           `json:"customFields"`
	StatusID          *string                  `json:"statusId" validate:"omitempty,uuid"`
	SprintID          *string                  `json:"sprintId"`
	ClearSprintID     bool                     `json:"-"`
	ParentTaskID      *string                  `json:"parentTaskId"`
	ClearParentTaskID bool                     `json:"-"`
	AssigneeIDs       *[]string                `json:"assigneeIds" validate:"omitempty,dive,uuid"`
	ReporterIDs       *[]string                `json:"reporterIds" validate:"omitempty,dive,uuid"`
	AllowEmailReplies *bool                    `json:"allowEmailReplies"`
	IsRecurring       *bool                    `json:"isRecurring"`
	StopRecurrence    *bool                    `json:"stopRecurrence"`
	RecurrenceConfig  *RecurrenceConfigRequest `json:"recurrenceConfig"`
	ProjectID         *string                  `json:"projectId" validate:"omitempty,uuid"`
}

// UpdateStatusRequest represents PATCH /api/tasks/:id/status.
type UpdateStatusRequest struct {
	StatusID string `json:"statusId" validate:"required,uuid"`
}

// UpdateAssigneesRequest represents PATCH /api/tasks/:id/assignees.
type UpdateAssigneesRequest struct {
	AssigneeIDs []string `json:"assigneeIds" validate:"required,dive,uuid"`
}

// UpdatePriorityRequest represents PATCH /api/tasks/:id/priority.
type UpdatePriorityRequest struct {
	Priority string `json:"priority" validate:"required,oneof=LOWEST LOW MEDIUM HIGH HIGHEST"`
}

// UpdateDueDateRequest represents PATCH /api/tasks/:id/due-date.
type UpdateDueDateRequest struct {
	DueDate string `json:"dueDate" validate:"required"`
}

// AddCommentRequest represents POST /api/tasks/:id/comments.
type AddCommentRequest struct {
	Comment string `json:"comment" validate:"required,min=1"`
}

// BulkTaskItem represents a row in bulk task creation.
type BulkTaskItem struct {
	Title       string  `json:"title" validate:"required,min=1"`
	Description *string `json:"description"`
	Type        *string `json:"type" validate:"omitempty,oneof=TASK STORY BUG EPIC SUBTASK"`
	Priority    *string `json:"priority" validate:"omitempty,oneof=LOWEST LOW MEDIUM HIGH HIGHEST"`
	DueDate     *string `json:"dueDate"`
}

// BulkCreateTasksRequest represents POST /api/tasks/bulk-create.
type BulkCreateTasksRequest struct {
	ProjectID string         `json:"projectId" validate:"required,uuid"`
	StatusID  string         `json:"statusId" validate:"required,uuid"`
	SprintID  *string        `json:"sprintId" validate:"omitempty,uuid"`
	Tasks     []BulkTaskItem `json:"tasks" validate:"required,min=1,max=1000,dive"`
}

// BulkDeleteTasksRequest represents POST /api/tasks/bulk-delete.
type BulkDeleteTasksRequest struct {
	TaskIDs     []string `json:"taskIds" validate:"omitempty,dive,uuid"`
	ProjectID   *string  `json:"projectId"`
	All         bool     `json:"all"`
	ExcludedIDs []string `json:"excludedIds" validate:"omitempty,dive,uuid"`
}

// BulkUpdateTasksStatusRequest represents POST /api/tasks/bulk-status-update.
type BulkUpdateTasksStatusRequest struct {
	TaskIDs        []string `json:"taskIds" validate:"omitempty,dive,uuid"`
	ProjectID      *string  `json:"projectId"`
	WorkspaceID    *string  `json:"workspaceId"`
	OrganizationID *string  `json:"organizationId"`
	All            bool     `json:"all"`
	ExcludedIDs    []string `json:"excludedIds" validate:"omitempty,dive,uuid"`
	StatusID       *string  `json:"statusId" validate:"omitempty,uuid"`
	Search         *string  `json:"search"`
	Statuses       *string  `json:"statuses"`
	Priorities     *string  `json:"priorities"`
	Types          *string  `json:"types"`
	Assignees      *string  `json:"assignees"`
	Reporters      *string  `json:"reporters"`
	SprintID       *string  `json:"sprintId" validate:"omitempty,uuid"`
}

// BulkAssignTasksRequest represents POST /api/tasks/bulk-assign.
type BulkAssignTasksRequest struct {
	TaskIDs        []string `json:"taskIds" validate:"omitempty,dive,uuid"`
	ProjectID      *string  `json:"projectId"`
	WorkspaceID    *string  `json:"workspaceId"`
	OrganizationID *string  `json:"organizationId"`
	All            bool     `json:"all"`
	ExcludedIDs    []string `json:"excludedIds" validate:"omitempty,dive,uuid"`
	AssigneeIDs    []string `json:"assigneeIds" validate:"present,dive,uuid"`
	Search         *string  `json:"search"`
	Statuses       *string  `json:"statuses"`
	Priorities     *string  `json:"priorities"`
	Types          *string  `json:"types"`
	Assignees      *string  `json:"assignees"`
	Reporters      *string  `json:"reporters"`
	SprintID       *string  `json:"sprintId" validate:"omitempty,uuid"`
}

// CreateShareLinkRequest represents POST /api/task-shares.
type CreateShareLinkRequest struct {
	TaskID        string `json:"taskId" validate:"required,min=1"`
	ExpiresInDays int    `json:"expiresInDays" validate:"required,min=1,max=365"`
}

// ShareLinkResponse represents public task share response data.
type ShareLinkResponse struct {
	ID        string `json:"id"`
	Token     string `json:"token"`
	ShareURL  string `json:"shareUrl"`
	ExpiresAt string `json:"expiresAt"`
	CreatedAt string `json:"createdAt"`
}

// TaskFilterQuery collects listing and filtering parameters for GET /api/tasks & /all-tasks.
type TaskFilterQuery struct {
	OrganizationID string
	ProjectIDs     []string
	WorkspaceIDs   []string
	SprintID       string
	ParentTaskID   *string
	Priorities     []string
	Statuses       []string
	Types          []string
	AssigneeIDs    []string
	ReporterIDs    []string
	Search         string
	SortBy         string
	SortOrder      string
	ViewType       string
	Page           int
	Limit          int
	GroupBy        string
	From           *time.Time
	To             *time.Time
	DateField      string
}

// GroupedTasksQuery collects parameters for GET /api/tasks/grouped.
type GroupedTasksQuery struct {
	OrganizationID string
	WorkspaceIDs   []string
	ProjectIDs     []string
	SprintID       string
	Priorities     []string
	Statuses       []string
	Types          []string
	AssigneeIDs    []string
	ReporterIDs    []string
	Search         string
	GroupBy        string
	GroupKey       string
	Page           int
	LimitPerGroup  int
}

// ByStatusQuery collects parameters for GET /api/tasks/by-status.
type ByStatusQuery struct {
	Slug            string
	IncludeSubtasks bool
	StatusID        string
	SprintID        string
	Page            int
	Limit           int
}

// OrgBrief represents condensed organization info.
type OrgBrief struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// WorkspaceBrief represents condensed workspace info.
type WorkspaceBrief struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	OrganizationID string    `json:"organizationId,omitempty"`
	Organization   *OrgBrief `json:"organization,omitempty"`
}

// ProjectBrief represents condensed project info.
type ProjectBrief struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Slug      string          `json:"slug"`
	Workspace *WorkspaceBrief `json:"workspace,omitempty"`
}

// UserBrief represents condensed user profile info.
type UserBrief struct {
	ID        string  `json:"id"`
	Email     string  `json:"email,omitempty"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Avatar    *string `json:"avatar,omitempty"`
}

// StatusBrief represents condensed status details.
type StatusBrief struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Category string `json:"category"`
}

// SprintBrief represents condensed sprint details.
type SprintBrief struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Slug      *string `json:"slug,omitempty"`
	Status    string  `json:"status"`
	StartDate *string `json:"startDate,omitempty"`
	EndDate   *string `json:"endDate,omitempty"`
}

// ParentTaskBrief represents parent task summary.
type ParentTaskBrief struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Slug  string `json:"slug"`
	Type  string `json:"type"`
}

// TaskLabelResponse represents flattened label attached to a task.
type TaskLabelResponse struct {
	TaskID      string  `json:"taskId"`
	LabelID     string  `json:"labelId"`
	Name        string  `json:"name"`
	Color       string  `json:"color"`
	Description *string `json:"description"`
}

// AttachmentBrief represents a task attachment item.
type AttachmentBrief struct {
	ID        string  `json:"id"`
	FileName  string  `json:"fileName"`
	FileSize  int32   `json:"fileSize"`
	MimeType  string  `json:"mimeType"`
	URL       *string `json:"url,omitempty"`
	CreatedAt string  `json:"createdAt"`
}

// WatcherBrief represents a task watcher item.
type WatcherBrief struct {
	ID     string     `json:"id"`
	TaskID string     `json:"taskId"`
	UserID string     `json:"userId"`
	User   *UserBrief `json:"user,omitempty"`
}

// CommentBrief represents a task comment item.
type CommentBrief struct {
	ID              string     `json:"id"`
	Content         string     `json:"content"`
	TaskID          string     `json:"taskId"`
	AuthorID        string     `json:"authorId"`
	ParentCommentID *string    `json:"parentCommentId"`
	CreatedAt       string     `json:"createdAt"`
	UpdatedAt       string     `json:"updatedAt"`
	Author          *UserBrief `json:"author,omitempty"`
}

// TimeEntryBrief represents a logged time entry item on task detail.
type TimeEntryBrief struct {
	ID          string     `json:"id"`
	Description *string    `json:"description"`
	TimeSpent   int32      `json:"timeSpent"`
	StartTime   *string    `json:"startTime"`
	EndTime     *string    `json:"endTime"`
	Date        string     `json:"date"`
	TaskID      string     `json:"taskId"`
	UserID      string     `json:"userId"`
	User        *UserBrief `json:"user,omitempty"`
}

// RecurrenceResponse represents recurring schedule config for a task.
type RecurrenceResponse struct {
	ID                string  `json:"id"`
	TaskID            string  `json:"taskId,omitempty"`
	RecurrenceType    string  `json:"recurrenceType"`
	Interval          int32   `json:"interval"`
	DaysOfWeek        []int32 `json:"daysOfWeek"`
	DayOfMonth        *int32  `json:"dayOfMonth"`
	MonthOfYear       *int32  `json:"monthOfYear"`
	EndType           string  `json:"endType"`
	EndDate           *string `json:"endDate"`
	OccurrenceCount   *int32  `json:"occurrenceCount"`
	CurrentOccurrence int32   `json:"currentOccurrence"`
	NextOccurrence    string  `json:"nextOccurrence"`
	IsActive          bool    `json:"isActive"`
}

// TaskCounts represents Prisma `_count` aggregates.
type TaskCounts struct {
	ChildTasks  int64 `json:"childTasks"`
	Comments    int64 `json:"comments"`
	Attachments int64 `json:"attachments,omitempty"`
	Watchers    int64 `json:"watchers,omitempty"`
	TimeEntries int64 `json:"timeEntries,omitempty"`
}

// ChildTaskBrief represents child task summary in FindOne responses.
type ChildTaskBrief struct {
	ID        string       `json:"id"`
	Title     string       `json:"title"`
	Slug      string       `json:"slug"`
	Type      string       `json:"type"`
	Priority  string       `json:"priority"`
	Status    *StatusBrief `json:"status,omitempty"`
	Assignees []UserBrief  `json:"assignees"`
	Reporters []UserBrief  `json:"reporters,omitempty"`
}

// ProjectInboxBrief represents email reply configuration shown on task detail.
type ProjectInboxBrief struct {
	ID           string  `json:"id"`
	ProjectID    string  `json:"projectId"`
	Enabled      bool    `json:"enabled"`
	EmailAddress *string `json:"emailAddress"`
}

// TaskResponse represents serialized task entities matching NestJS JSON contracts.
type TaskResponse struct {
	ID                string              `json:"id"`
	Title             string              `json:"title"`
	Description       *string             `json:"description"`
	Type              string              `json:"type"`
	Priority          string              `json:"priority"`
	TaskNumber        int32               `json:"taskNumber"`
	Slug              string              `json:"slug"`
	StartDate         *string             `json:"startDate"`
	DueDate           *string             `json:"dueDate"`
	CompletedAt       *string             `json:"completedAt"`
	StoryPoints       *int32              `json:"storyPoints"`
	OriginalEstimate  *int32              `json:"originalEstimate"`
	RemainingEstimate *int32              `json:"remainingEstimate"`
	CustomFields      any                 `json:"customFields"`
	ProjectID         string              `json:"projectId"`
	StatusID          string              `json:"statusId"`
	SprintID          *string             `json:"sprintId"`
	ParentTaskID      *string             `json:"parentTaskId"`
	CreatedBy         *string             `json:"createdById,omitempty"`
	CreatedAt         string              `json:"createdAt"`
	UpdatedAt         string              `json:"updatedAt"`
	IsArchived        bool                `json:"isArchived"`
	AllowEmailReplies bool                `json:"allowEmailReplies"`
	IsRecurring       bool                `json:"isRecurring"`
	Project           *ProjectBrief       `json:"project,omitempty"`
	Status            *StatusBrief        `json:"status,omitempty"`
	Sprint            *SprintBrief        `json:"sprint,omitempty"`
	ParentTask        *ParentTaskBrief    `json:"parentTask,omitempty"`
	CreatedByUser     *UserBrief          `json:"createdByUser,omitempty"`
	Assignees         []UserBrief         `json:"assignees"`
	Reporters         []UserBrief         `json:"reporters"`
	Labels            []TaskLabelResponse `json:"labels,omitempty"`
	Attachments       []AttachmentBrief   `json:"attachments,omitempty"`
	Watchers          []WatcherBrief      `json:"watchers,omitempty"`
	Comments          []CommentBrief      `json:"comments,omitempty"`
	TimeEntries       []TimeEntryBrief    `json:"timeEntries,omitempty"`
	ChildTasks        []ChildTaskBrief    `json:"childTasks,omitempty"`
	RecurringConfig   *RecurrenceResponse `json:"recurringConfig,omitempty"`
	ShowEmailReply    *ProjectInboxBrief  `json:"showEmailReply,omitempty"`
	Count             *TaskCounts         `json:"_count,omitempty"`
}

// ValueCount represents priority/type filter facet counts.
type ValueCount struct {
	Value string `json:"value"`
	Count int64  `json:"count"`
}

// EntityCount represents status/assignee/reporter filter facet counts.
type EntityCount struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

// FilterCountsResponse represents filterCounts nested in paginated task listings.
type FilterCountsResponse struct {
	Priorities []ValueCount  `json:"priorities"`
	Types      []ValueCount  `json:"types"`
	Statuses   []EntityCount `json:"statuses"`
	Assignees  []EntityCount `json:"assignees"`
	Reporters  []EntityCount `json:"reporters"`
}

// PaginatedTasksResponse represents GET /api/tasks and GET /api/tasks/all-tasks.
type PaginatedTasksResponse struct {
	Data         []TaskResponse        `json:"data"`
	Total        int64                 `json:"total"`
	Page         int                   `json:"page"`
	Limit        int                   `json:"limit"`
	TotalPages   int                   `json:"totalPages"`
	FilterCounts *FilterCountsResponse `json:"filterCounts,omitempty"`
}

// TaskGroupBucket represents one group returned by GET /api/tasks/grouped.
type TaskGroupBucket struct {
	Key        string         `json:"key"`
	Label      string         `json:"label"`
	TotalCount int64          `json:"totalCount"`
	Tasks      []TaskResponse `json:"tasks"`
	Page       *int           `json:"page,omitempty"`
}

// GroupedTasksResponse represents GET /api/tasks/grouped.
type GroupedTasksResponse struct {
	Groups        []TaskGroupBucket `json:"groups"`
	GroupBy       string            `json:"groupBy"`
	Page          int               `json:"page,omitempty"`
	LimitPerGroup int               `json:"limitPerGroup,omitempty"`
}

// ColumnPagination holds column pagination metadata matching NestJS TasksByStatus.
type ColumnPagination struct {
	Total       int64 `json:"total"`
	Page        int   `json:"page"`
	Limit       int   `json:"limit"`
	TotalPages  int   `json:"totalPages"`
	HasNextPage bool  `json:"hasNextPage"`
	HasPrevPage bool  `json:"hasPrevPage"`
}

// StatusColumnBucket represents one status column in GET /api/tasks/by-status.
type StatusColumnBucket struct {
	StatusID       string           `json:"statusId"`
	StatusName     string           `json:"statusName"`
	StatusColor    string           `json:"statusColor"`
	StatusCategory string           `json:"statusCategory"`
	Tasks          []TaskResponse   `json:"tasks"`
	Pagination     ColumnPagination `json:"pagination"`
}

// ByStatusMeta holds aggregate metadata across status columns.
type ByStatusMeta struct {
	TotalTasks    int64  `json:"totalTasks"`
	LoadedTasks   int64  `json:"loadedTasks"`
	TotalStatuses int    `json:"totalStatuses"`
	FetchedAt     string `json:"fetchedAt"`
}

// ByStatusEnvelope represents GET /api/tasks/by-status.
type ByStatusEnvelope struct {
	Data []StatusColumnBucket `json:"data"`
	Meta ByStatusMeta         `json:"meta"`
}

// StandardPagination holds 5-field pagination matching Today/Org endpoints.
type StandardPagination struct {
	CurrentPage int   `json:"currentPage"`
	TotalPages  int   `json:"totalPages"`
	TotalCount  int64 `json:"totalCount"`
	HasNextPage bool  `json:"hasNextPage"`
	HasPrevPage bool  `json:"hasPrevPage"`
}

// TodayTasksResponse represents GET /api/tasks/today & GET /api/tasks/organization/:orgId.
type TodayTasksResponse struct {
	Tasks      []TaskResponse     `json:"tasks"`
	Pagination StandardPagination `json:"pagination"`
}

// FailedTaskItem records a partial-failure item during bulk actions.
type FailedTaskItem struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}

// BulkDeleteResponse represents POST /api/tasks/bulk-delete.
type BulkDeleteResponse struct {
	DeletedCount int64            `json:"deletedCount"`
	FailedTasks  []FailedTaskItem `json:"failedTasks"`
}

// BulkUpdateStatusResponse represents POST /api/tasks/bulk-status-update.
type BulkUpdateStatusResponse struct {
	UpdatedCount int              `json:"updatedCount"`
	UpdatedTasks []TaskResponse   `json:"updatedTasks"`
	FailedTasks  []FailedTaskItem `json:"failedTasks"`
}

// BulkAssignResponse represents POST /api/tasks/bulk-assign.
type BulkAssignResponse struct {
	AssignedCount int              `json:"assignedCount"`
	UpdatedTasks  []TaskResponse   `json:"updatedTasks"`
	FailedTasks   []FailedTaskItem `json:"failedTasks"`
}

// BulkCreateResponse represents POST /api/tasks/bulk-create.
type BulkCreateResponse struct {
	Message string `json:"message,omitempty"`
	Count   int    `json:"count"`
	Tasks   any    `json:"tasks,omitempty"`
}

// CompleteOccurrenceResponse represents POST /api/tasks/:id/complete-occurrence.
type CompleteOccurrenceResponse struct {
	CompletedTask *TaskResponse `json:"completedTask"`
	NextTask      *TaskResponse `json:"nextTask"`
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

func pgInt4Ptr(v pgtype.Int4) *int32 {
	if v.Valid {
		return &v.Int32
	}
	return nil
}

func decodeJSONField(raw []byte) any {
	if len(raw) == 0 {
		return nil
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil
	}
	return out
}

func parseISO8601(raw string) (time.Time, bool) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}
