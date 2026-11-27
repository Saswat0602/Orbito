package task_watchers

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// CreateRequest represents CreateTaskWatcherDto / WatchTaskDto / UnwatchTaskDto payloads.
type CreateRequest struct {
	TaskID string `json:"taskId" validate:"required"`
	UserID string `json:"userId" validate:"required,uuid"`
}

// UserBrief represents minimal watcher user metadata.
type UserBrief struct {
	ID        string  `json:"id"`
	Email     string  `json:"email"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Avatar    *string `json:"avatar"`
	Bio       *string `json:"bio,omitempty"`
	Status    string  `json:"status,omitempty"`
}

// StatusBrief represents task status summary on watched tasks.
type StatusBrief struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Category string `json:"category"`
}

// OrgBrief represents organization summary on watched tasks.
type OrgBrief struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// WorkspaceBrief represents workspace summary on watched tasks.
type WorkspaceBrief struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Organization *OrgBrief `json:"organization,omitempty"`
}

// ProjectBrief represents project summary on watched tasks.
type ProjectBrief struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Slug      string          `json:"slug"`
	Workspace *WorkspaceBrief `json:"workspace,omitempty"`
}

// TaskBrief represents task summary inside watcher responses.
type TaskBrief struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Slug        string        `json:"slug"`
	Description *string       `json:"description,omitempty"`
	Type        string        `json:"type,omitempty"`
	Priority    string        `json:"priority,omitempty"`
	DueDate     *string       `json:"dueDate,omitempty"`
	Status      *StatusBrief  `json:"status,omitempty"`
	Project     *ProjectBrief `json:"project,omitempty"`
	Count       *TaskCounts   `json:"_count,omitempty"`
}

// TaskCounts mirrors Prisma `_count` for comments and watchers.
type TaskCounts struct {
	Comments int64 `json:"comments"`
	Watchers int64 `json:"watchers"`
}

// WatcherResponse represents a task watcher record with optional expansions.
type WatcherResponse struct {
	ID        string     `json:"id"`
	TaskID    string     `json:"taskId"`
	UserID    string     `json:"userId"`
	CreatedBy *string    `json:"createdBy"`
	UpdatedBy *string    `json:"updatedBy"`
	CreatedAt string     `json:"createdAt"`
	UpdatedAt string     `json:"updatedAt"`
	User      *UserBrief `json:"user,omitempty"`
	Task      *TaskBrief `json:"task,omitempty"`
}

// ToggleWatchResponse represents POST /task-watchers/toggle response.
type ToggleWatchResponse struct {
	IsWatching bool             `json:"isWatching"`
	Watcher    *WatcherResponse `json:"watcher,omitempty"`
}

// WatcherGroupCount mirrors Prisma `groupBy` output in getWatcherStats.
type WatcherGroupCount struct {
	TaskID string         `json:"taskId"`
	Count  map[string]int `json:"_count"`
}

// WatcherStatsResponse represents GET /task-watchers/stats output.
type WatcherStatsResponse struct {
	TotalWatchers    int64               `json:"totalWatchers"`
	MostWatchedTasks []WatcherGroupCount `json:"mostWatchedTasks,omitempty"`
	RecentWatches    int64               `json:"recentWatches"`
}

func tsString(ts pgtype.Timestamp) string {
	if !ts.Valid {
		return ""
	}
	return ts.Time.UTC().Format("2006-01-02T15:04:05.000Z")
}

func tsPtr(ts pgtype.Timestamp) *string {
	if !ts.Valid {
		return nil
	}
	s := ts.Time.UTC().Format("2006-01-02T15:04:05.000Z")
	return &s
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

func mapDetailRow(r *sqlc.GetTaskWatcherDetailsByIDRow) WatcherResponse {
	return WatcherResponse{
		ID:        r.ID.String(),
		TaskID:    r.TaskID.String(),
		UserID:    r.UserID.String(),
		CreatedBy: uuidPtr(r.CreatedByID),
		UpdatedBy: uuidPtr(r.UpdatedByID),
		CreatedAt: tsString(r.CreatedAt),
		UpdatedAt: tsString(r.UpdatedAt),
		User: &UserBrief{
			ID:        r.UserID.String(),
			Email:     r.UserEmail,
			FirstName: r.UserFirstName,
			LastName:  r.UserLastName,
			Avatar:    textPtr(r.UserAvatar),
			Bio:       textPtr(r.UserBio),
			Status:    string(r.UserStatus),
		},
		Task: &TaskBrief{
			ID:          r.TaskID.String(),
			Title:       r.TaskTitle,
			Slug:        r.TaskSlug,
			Description: textPtr(r.TaskDescription),
			Type:        string(r.TaskType),
			Priority:    string(r.TaskPriority),
			Status: &StatusBrief{
				ID:       r.StatusID.String(),
				Name:     r.StatusName,
				Color:    r.StatusColor,
				Category: string(r.StatusCategory),
			},
			Project: &ProjectBrief{
				ID:   r.ProjectID.String(),
				Name: r.ProjectName,
				Slug: r.ProjectSlug,
				Workspace: &WorkspaceBrief{
					ID:   r.WorkspaceID.String(),
					Name: r.WorkspaceName,
					Slug: r.WorkspaceSlug,
				},
			},
		},
	}
}

func mapFilteredRow(r sqlc.ListTaskWatchersFilteredRow) WatcherResponse {
	return WatcherResponse{
		ID:        r.ID.String(),
		TaskID:    r.TaskID.String(),
		UserID:    r.UserID.String(),
		CreatedBy: uuidPtr(r.CreatedByID),
		UpdatedBy: uuidPtr(r.UpdatedByID),
		CreatedAt: tsString(r.CreatedAt),
		UpdatedAt: tsString(r.UpdatedAt),
		User: &UserBrief{
			ID:        r.UserID.String(),
			Email:     r.UserEmail,
			FirstName: r.UserFirstName,
			LastName:  r.UserLastName,
			Avatar:    textPtr(r.UserAvatar),
			Status:    string(r.UserStatus),
		},
		Task: &TaskBrief{
			ID:       r.TaskID.String(),
			Title:    r.TaskTitle,
			Slug:     r.TaskSlug,
			Type:     string(r.TaskType),
			Priority: string(r.TaskPriority),
			Status: &StatusBrief{
				ID:       r.StatusID.String(),
				Name:     r.StatusName,
				Color:    r.StatusColor,
				Category: string(r.StatusCategory),
			},
			Project: &ProjectBrief{
				ID:   r.ProjectID.String(),
				Name: r.ProjectName,
				Slug: r.ProjectSlug,
			},
		},
	}
}

func mapTaskWatcherRow(r sqlc.ListWatchersForTaskRow) WatcherResponse {
	return WatcherResponse{
		ID:        r.ID.String(),
		TaskID:    r.TaskID.String(),
		UserID:    r.UserID.String(),
		CreatedBy: uuidPtr(r.CreatedByID),
		UpdatedBy: uuidPtr(r.UpdatedByID),
		CreatedAt: tsString(r.CreatedAt),
		UpdatedAt: tsString(r.UpdatedAt),
		User: &UserBrief{
			ID:        r.UserID.String(),
			Email:     r.UserEmail,
			FirstName: r.UserFirstName,
			LastName:  r.UserLastName,
			Avatar:    textPtr(r.UserAvatar),
			Status:    string(r.UserStatus),
		},
	}
}

func mapUserWatchedTaskRow(r sqlc.ListWatchedTasksByUserRow) WatcherResponse {
	return WatcherResponse{
		ID:        r.ID.String(),
		TaskID:    r.TaskID.String(),
		UserID:    r.UserID.String(),
		CreatedBy: uuidPtr(r.CreatedByID),
		UpdatedBy: uuidPtr(r.UpdatedByID),
		CreatedAt: tsString(r.CreatedAt),
		UpdatedAt: tsString(r.UpdatedAt),
		Task: &TaskBrief{
			ID:          r.TaskID.String(),
			Title:       r.TaskTitle,
			Slug:        r.TaskSlug,
			Description: textPtr(r.TaskDescription),
			Type:        string(r.TaskType),
			Priority:    string(r.TaskPriority),
			DueDate:     tsPtr(r.TaskDueDate),
			Status: &StatusBrief{
				ID:       r.StatusID.String(),
				Name:     r.StatusName,
				Color:    r.StatusColor,
				Category: string(r.StatusCategory),
			},
			Project: &ProjectBrief{
				ID:   r.ProjectID.String(),
				Name: r.ProjectName,
				Slug: r.ProjectSlug,
				Workspace: &WorkspaceBrief{
					ID:   r.WorkspaceID.String(),
					Name: r.WorkspaceName,
					Slug: r.WorkspaceSlug,
					Organization: &OrgBrief{
						ID:   r.OrgID.String(),
						Name: r.OrgName,
						Slug: r.OrgSlug,
					},
				},
			},
			Count: &TaskCounts{
				Comments: r.CommentCount,
				Watchers: r.WatcherCount,
			},
		},
	}
}

