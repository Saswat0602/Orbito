package search

import "time"

const iso8601UTC = "2006-01-02T15:04:05.000Z"

// GlobalRequest represents POST /api/search/global or GET /api/search/quick parameters.
type GlobalRequest struct {
	Query          string `json:"query"`
	EntityType     string `json:"entityType"`
	OrganizationID string `json:"organizationId"`
	WorkspaceID    string `json:"workspaceId"`
	ProjectID      string `json:"projectId"`
	Page           int    `json:"page"`
	Limit          int    `json:"limit"`
	SortBy         string `json:"sortBy"`
	SortOrder      string `json:"sortOrder"`
}

// AdvancedRequest represents POST /api/search/advanced filters.
type AdvancedRequest struct {
	Query          string   `json:"query"`
	TaskTypes      []string `json:"taskTypes"`
	Priorities     []string `json:"priorities"`
	AssigneeIDs    []string `json:"assigneeIds"`
	ReporterIDs    []string `json:"reporterIds"`
	StatusIDs      []string `json:"statusIds"`
	LabelIDs       []string `json:"labelIds"`
	SprintIDs      []string `json:"sprintIds"`
	DueDateFrom    string   `json:"dueDateFrom"`
	DueDateTo      string   `json:"dueDateTo"`
	CreatedFrom    string   `json:"createdFrom"`
	CreatedTo      string   `json:"createdTo"`
	OrganizationID string   `json:"organizationId"`
	WorkspaceID    string   `json:"workspaceId"`
	ProjectID      string   `json:"projectId"`
	Page           int      `json:"page"`
	Limit          int      `json:"limit"`
	SortBy         string   `json:"sortBy"`
	SortOrder      string   `json:"sortOrder"`
}

// ResultItem represents a unified search hit across entities.
type ResultItem struct {
	ID             string         `json:"id"`
	Type           string         `json:"type"`
	Title          string         `json:"title"`
	Description    string         `json:"description,omitempty"`
	ProjectID      *string        `json:"projectId,omitempty"`
	ProjectKey     *string        `json:"projectKey,omitempty"`
	OrganizationID *string        `json:"organizationId,omitempty"`
	WorkspaceID    *string        `json:"workspaceId,omitempty"`
	RelevanceScore int            `json:"relevanceScore"`
	CreatedAt      string         `json:"createdAt"`
	UpdatedAt      string         `json:"updatedAt"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

// Response represents paginated search output with query timing in milliseconds.
type Response struct {
	Results    []ResultItem `json:"results"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	Limit      int          `json:"limit"`
	TotalPages int          `json:"totalPages"`
	SearchTime int64        `json:"searchTime"`
}

func formatTime(t time.Time) string {
	return t.UTC().Format(iso8601UTC)
}
