package sprints

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// CreateSprintRequest represents POST /api/sprints.
type CreateSprintRequest struct {
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Goal        *string `json:"goal" validate:"omitempty,max=500"`
	Status      *string `json:"status" validate:"omitempty,oneof=PLANNING ACTIVE COMPLETED CANCELLED"`
	StartDate   *string `json:"startDate"`
	EndDate     *string `json:"endDate"`
	ProjectSlug *string `json:"projectSlug"`
	ProjectID   *string `json:"projectId" validate:"omitempty,uuid"`
}

// UpdateSprintRequest represents PATCH /api/sprints/:id.
type UpdateSprintRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=1,max=100"`
	Goal        *string `json:"goal" validate:"omitempty,max=500"`
	Status      *string `json:"status" validate:"omitempty,oneof=PLANNING ACTIVE COMPLETED CANCELLED"`
	StartDate   *string `json:"startDate"`
	EndDate     *string `json:"endDate"`
	ProjectSlug *string `json:"projectSlug"`
	ProjectID   *string `json:"projectId" validate:"omitempty,uuid"`
}

// OrgBrief represents condensed organization info.
type OrgBrief struct {
	ID   string `json:"id"`
	Name string `json:"name,omitempty"`
	Slug string `json:"slug,omitempty"`
}

// WorkspaceBrief represents condensed workspace info.
type WorkspaceBrief struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Organization *OrgBrief `json:"organization,omitempty"`
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
	Email     string  `json:"email"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
}

// TaskUserBrief represents condensed assignee/reporter info.
type TaskUserBrief struct {
	ID        string  `json:"id"`
	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`
	Avatar    *string `json:"avatar"`
}

// TaskUserWrapper mirrors Prisma's nested `{ user: { ... } }` join shape for sprint tasks.
type TaskUserWrapper struct {
	User TaskUserBrief `json:"user"`
}

// TaskStatusBrief represents condensed status info nested on sprint tasks.
type TaskStatusBrief struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Color    string `json:"color"`
	Category string `json:"category"`
}

// SprintTaskCount represents Prisma `_count` for a task inside a sprint.
type SprintTaskCount struct {
	ChildTasks int64 `json:"childTasks"`
	Comments   int64 `json:"comments"`
}

// SprintTaskItem represents a task included in sprint detail responses.
type SprintTaskItem struct {
	ID                string            `json:"id"`
	Title             string            `json:"title"`
	Description       *string           `json:"description"`
	Type              string            `json:"type"`
	Priority          string            `json:"priority"`
	TaskNumber        int32             `json:"taskNumber"`
	Slug              string            `json:"slug"`
	StartDate         *string           `json:"startDate"`
	DueDate           *string           `json:"dueDate"`
	CompletedAt       *string           `json:"completedAt"`
	StoryPoints       *int32            `json:"storyPoints"`
	OriginalEstimate  *int32            `json:"originalEstimate"`
	RemainingEstimate *int32            `json:"remainingEstimate"`
	ProjectID         string            `json:"projectId"`
	StatusID          string            `json:"statusId"`
	SprintID          *string           `json:"sprintId"`
	ParentTaskID      *string           `json:"parentTaskId"`
	CreatedAt         string            `json:"createdAt"`
	UpdatedAt         string            `json:"updatedAt"`
	IsArchived        bool              `json:"isArchived"`
	Status            *TaskStatusBrief  `json:"status,omitempty"`
	Assignees         []TaskUserWrapper `json:"assignees"`
	Reporters         []TaskUserWrapper `json:"reporters,omitempty"`
	Count             *SprintTaskCount  `json:"_count,omitempty"`
}

// SprintCount represents the Prisma-compatible `_count` wrapper.
type SprintCount struct {
	Tasks int64 `json:"tasks"`
}

// SprintResponse represents serialized sprint data matching NestJS responses.
type SprintResponse struct {
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	Slug          *string          `json:"slug"`
	Goal          *string          `json:"goal"`
	Status        string           `json:"status"`
	IsDefault     bool             `json:"isDefault"`
	Archive       bool             `json:"archive"`
	StartDate     *string          `json:"startDate"`
	EndDate       *string          `json:"endDate"`
	ProjectID     string           `json:"projectId"`
	CreatedByID   *string          `json:"createdById"`
	UpdatedByID   *string          `json:"updatedById"`
	CreatedAt     string           `json:"createdAt"`
	UpdatedAt     string           `json:"updatedAt"`
	Project       *ProjectBrief    `json:"project,omitempty"`
	CreatedByUser *UserBrief       `json:"createdByUser,omitempty"`
	UpdatedByUser *UserBrief       `json:"updatedByUser,omitempty"`
	Tasks         []SprintTaskItem `json:"tasks,omitempty"`
	Count         *SprintCount     `json:"_count,omitempty"`
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

func parseDateString(raw string) (time.Time, bool) {
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
