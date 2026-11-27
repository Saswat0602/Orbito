package time_entries

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// CreateRequest represents CreateTimeEntryDto.
type CreateRequest struct {
	Description *string  `json:"description"`
	TimeSpent   *int     `json:"timeSpent" validate:"omitempty,min=1"`
	Duration    *float64 `json:"duration"`
	StartTime   *string  `json:"startTime"`
	EndTime     *string  `json:"endTime"`
	Date        *string  `json:"date"`
	TaskID      string   `json:"taskId" validate:"required,uuid"`
	UserID      *string  `json:"userId" validate:"omitempty,uuid"`
}

// UpdateRequest represents UpdateTimeEntryDto.
type UpdateRequest struct {
	Description *string  `json:"description"`
	TimeSpent   *int     `json:"timeSpent" validate:"omitempty,min=1"`
	Duration    *float64 `json:"duration"`
	StartTime   *string  `json:"startTime"`
	EndTime     *string  `json:"endTime"`
	Date        *string  `json:"date"`
}

// StartTimerRequest represents StartTimerDto.
type StartTimerRequest struct {
	TaskID      string  `json:"taskId" validate:"required,uuid"`
	UserID      *string `json:"userId" validate:"omitempty,uuid"`
	Description *string `json:"description"`
}

// StopTimerRequest represents StopTimerDto.
type StopTimerRequest struct {
	UserID      *string `json:"userId" validate:"omitempty,uuid"`
	Description *string `json:"description"`
}

// WorkspaceBrief represents minimal workspace details on time entry findOne.
type WorkspaceBrief struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// ProjectBrief represents project details embedded in time entry task.
type ProjectBrief struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Slug      string          `json:"slug"`
	Workspace *WorkspaceBrief `json:"workspace,omitempty"`
}

// TaskBrief represents task details embedded in time entries.
type TaskBrief struct {
	ID      string        `json:"id"`
	Title   string        `json:"title"`
	Slug    string        `json:"slug"`
	Project *ProjectBrief `json:"project,omitempty"`
}

// UserBrief represents user details embedded in time entries.
type UserBrief struct {
	ID        string  `json:"id"`
	Email     *string `json:"email,omitempty"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Avatar    *string `json:"avatar,omitempty"`
}

// EntryResponse represents a time entry record.
type EntryResponse struct {
	ID          string     `json:"id"`
	Description *string    `json:"description"`
	TimeSpent   int32      `json:"timeSpent"`
	StartTime   *string    `json:"startTime"`
	EndTime     *string    `json:"endTime"`
	Date        string     `json:"date"`
	TaskID      string     `json:"taskId"`
	UserID      string     `json:"userId"`
	CreatedBy   *string    `json:"createdBy"`
	UpdatedBy   *string    `json:"updatedBy"`
	CreatedAt   string     `json:"createdAt"`
	UpdatedAt   string     `json:"updatedAt"`
	Task        *TaskBrief `json:"task,omitempty"`
	User        *UserBrief `json:"user,omitempty"`
}

// StartTimerResponse matches POST /time-entries/timer/start response.
type StartTimerResponse struct {
	Message     string         `json:"message"`
	ActiveTimer *EntryResponse `json:"activeTimer"`
}

// TaskSummaryItem matches grouped task time totals.
type TaskSummaryItem struct {
	Task      TaskBrief `json:"task"`
	TotalTime int32     `json:"totalTime"`
	Entries   int       `json:"entries"`
}

// UserSummaryItem matches grouped user time totals.
type UserSummaryItem struct {
	User      UserBrief `json:"user"`
	TotalTime int32     `json:"totalTime"`
	Entries   int       `json:"entries"`
}

// SummaryResponse matches GET /time-entries/summary output.
type SummaryResponse struct {
	TotalTimeSpent      int32             `json:"totalTimeSpent"`
	TotalTimeSpentHours float64           `json:"totalTimeSpentHours"`
	TotalEntries        int               `json:"totalEntries"`
	TaskSummary         []TaskSummaryItem `json:"taskSummary"`
	UserSummary         []UserSummaryItem `json:"userSummary"`
	Entries             []EntryResponse   `json:"entries"`
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

func tsStr(ts pgtype.Timestamp) string {
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

func mapDetailRow(r *sqlc.GetTimeEntryDetailByIDRow, includeWorkspace, includeEmail bool) EntryResponse {
	var emailPtr *string
	if includeEmail {
		e := r.UserEmail
		emailPtr = &e
	}
	proj := &ProjectBrief{
		ID:   r.ProjectID.String(),
		Name: r.ProjectName,
		Slug: r.ProjectSlug,
	}
	if includeWorkspace {
		proj.Workspace = &WorkspaceBrief{
			ID:   r.WorkspaceID.String(),
			Name: r.WorkspaceName,
			Slug: r.WorkspaceSlug,
		}
	}
	return EntryResponse{
		ID:          r.ID.String(),
		Description: textPtr(r.Description),
		TimeSpent:   r.TimeSpent,
		StartTime:   tsPtr(r.StartTime),
		EndTime:     tsPtr(r.EndTime),
		Date:        tsStr(r.Date),
		TaskID:      r.TaskID.String(),
		UserID:      r.UserID.String(),
		CreatedBy:   uuidPtr(r.CreatedByID),
		UpdatedBy:   uuidPtr(r.UpdatedByID),
		CreatedAt:   tsStr(r.CreatedAt),
		UpdatedAt:   tsStr(r.UpdatedAt),
		Task: &TaskBrief{
			ID:      r.TaskID.String(),
			Title:   r.TaskTitle,
			Slug:    r.TaskSlug,
			Project: proj,
		},
		User: &UserBrief{
			ID:        r.UserID.String(),
			Email:     emailPtr,
			FirstName: r.UserFirstName,
			LastName:  r.UserLastName,
			Avatar:    textPtr(r.UserAvatar),
		},
	}
}

func mapFilteredRow(r sqlc.ListTimeEntriesFilteredRow) EntryResponse {
	return EntryResponse{
		ID:          r.ID.String(),
		Description: textPtr(r.Description),
		TimeSpent:   r.TimeSpent,
		StartTime:   tsPtr(r.StartTime),
		EndTime:     tsPtr(r.EndTime),
		Date:        tsStr(r.Date),
		TaskID:      r.TaskID.String(),
		UserID:      r.UserID.String(),
		CreatedBy:   uuidPtr(r.CreatedByID),
		UpdatedBy:   uuidPtr(r.UpdatedByID),
		CreatedAt:   tsStr(r.CreatedAt),
		UpdatedAt:   tsStr(r.UpdatedAt),
		Task: &TaskBrief{
			ID:    r.TaskID.String(),
			Title: r.TaskTitle,
			Slug:  r.TaskSlug,
			Project: &ProjectBrief{
				ID:   r.ProjectID.String(),
				Name: r.ProjectName,
				Slug: r.ProjectSlug,
			},
		},
		User: &UserBrief{
			ID:        r.UserID.String(),
			FirstName: r.UserFirstName,
			LastName:  r.UserLastName,
			Avatar:    textPtr(r.UserAvatar),
		},
	}
}

