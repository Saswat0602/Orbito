package gantt

import (
	"time"
)

// AssigneeBrief represents an assignee inside Gantt task items.
type AssigneeBrief struct {
	ID        string  `json:"id"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Avatar    *string `json:"avatar,omitempty"`
}

// StatusBrief represents status display attributes for Gantt bars.
type StatusBrief struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// TaskItem represents a single task node (and nested children) in Gantt data.
type TaskItem struct {
	ID           string          `json:"id"`
	Title        string          `json:"title"`
	Start        *string         `json:"start"`
	End          *string         `json:"end"`
	Progress     int             `json:"progress"`
	Dependencies []string        `json:"dependencies"`
	Assignees    []AssigneeBrief `json:"assignees,omitempty"`
	Priority     string          `json:"priority"`
	Status       StatusBrief     `json:"status"`
	Type         string          `json:"type"`
	Key          string          `json:"key"`
	Parent       *string         `json:"parent,omitempty"`
	Children     []*TaskItem     `json:"children,omitempty"`
}

// Timeline represents the bounding start/end window and duration in days.
type Timeline struct {
	Start    string `json:"start"`
	End      string `json:"end"`
	Duration int    `json:"duration"`
}

// Milestone represents a sprint start/end or project milestone marker.
type Milestone struct {
	ID    string  `json:"id"`
	Title string  `json:"title"`
	Date  *string `json:"date"`
	Type  string  `json:"type"`
}

// ChartData represents GET /gantt/project/:projectId and /gantt/sprint/:sprintId responses.
type ChartData struct {
	Tasks        []*TaskItem `json:"tasks"`
	Timeline     Timeline    `json:"timeline"`
	CriticalPath []string    `json:"criticalPath"`
	Milestones   []Milestone `json:"milestones"`
}

// ResourceAssignee represents an assignee summary in resource allocation view.
type ResourceAssignee struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Avatar *string `json:"avatar"`
}

// ResourceTask represents a scheduled task allocated to a resource.
type ResourceTask struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Start       *string `json:"start"`
	End         *string `json:"end"`
	StoryPoints int     `json:"storyPoints"`
}

// ResourceAllocation represents GET /gantt/project/:projectId/resources entry.
type ResourceAllocation struct {
	Assignee ResourceAssignee `json:"assignee"`
	Tasks    []ResourceTask   `json:"tasks"`
	Workload float64          `json:"workload"`
}

func formatISO(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

func formatISOPtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := formatISO(*t)
	return &s
}

