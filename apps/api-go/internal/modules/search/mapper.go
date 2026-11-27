package search

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

type advancedParamPair struct {
	count sqlc.CountTasksAdvancedQueryParams
	query sqlc.SearchTasksAdvancedQueryParams
}

func buildAdvancedParams(
	req AdvancedRequest,
	scopes AccessibleScopes,
	filterOrg, filterWs, filterProj pgtype.UUID,
	limit, offset int32,
) (advancedParamPair, error) {
	var qText pgtype.Text
	if strings.TrimSpace(req.Query) != "" {
		qText = pgtype.Text{String: escapeLikePattern(strings.TrimSpace(req.Query)), Valid: true}
	}

	taskTypes, err := parseTaskTypes(req.TaskTypes)
	if err != nil {
		return advancedParamPair{}, err
	}
	priorities, err := parseTaskPriorities(req.Priorities)
	if err != nil {
		return advancedParamPair{}, err
	}
	assignees, err := parseUUIDSlice(req.AssigneeIDs, "assigneeIds")
	if err != nil {
		return advancedParamPair{}, err
	}
	reporters, err := parseUUIDSlice(req.ReporterIDs, "reporterIds")
	if err != nil {
		return advancedParamPair{}, err
	}
	statuses, err := parseUUIDSlice(req.StatusIDs, "statusIds")
	if err != nil {
		return advancedParamPair{}, err
	}
	labels, err := parseUUIDSlice(req.LabelIDs, "labelIds")
	if err != nil {
		return advancedParamPair{}, err
	}
	sprints, err := parseUUIDSlice(req.SprintIDs, "sprintIds")
	if err != nil {
		return advancedParamPair{}, err
	}

	dueFrom, err := parseOptTime(req.DueDateFrom, "dueDateFrom")
	if err != nil {
		return advancedParamPair{}, err
	}
	dueTo, err := parseOptTime(req.DueDateTo, "dueDateTo")
	if err != nil {
		return advancedParamPair{}, err
	}
	createdFrom, err := parseOptTime(req.CreatedFrom, "createdFrom")
	if err != nil {
		return advancedParamPair{}, err
	}
	createdTo, err := parseOptTime(req.CreatedTo, "createdTo")
	if err != nil {
		return advancedParamPair{}, err
	}

	return advancedParamPair{
		count: sqlc.CountTasksAdvancedQueryParams{
			OrgIds: scopes.OrganizationIDs, WsIds: scopes.WorkspaceIDs, ProjIds: scopes.ProjectIDs,
			FilterOrgID: filterOrg, FilterWsID: filterWs, FilterProjID: filterProj,
			QueryText: qText, TaskTypes: taskTypes, Priorities: priorities,
			StatusIds: statuses, SprintIds: sprints, AssigneeIds: assignees,
			ReporterIds: reporters, LabelIds: labels,
			DueFrom: dueFrom, DueTo: dueTo, CreatedFrom: createdFrom, CreatedTo: createdTo,
		},
		query: sqlc.SearchTasksAdvancedQueryParams{
			OrgIds: scopes.OrganizationIDs, WsIds: scopes.WorkspaceIDs, ProjIds: scopes.ProjectIDs,
			FilterOrgID: filterOrg, FilterWsID: filterWs, FilterProjID: filterProj,
			QueryText: qText, TaskTypes: taskTypes, Priorities: priorities,
			StatusIds: statuses, SprintIds: sprints, AssigneeIds: assignees,
			ReporterIds: reporters, LabelIds: labels,
			DueFrom: dueFrom, DueTo: dueTo, CreatedFrom: createdFrom, CreatedTo: createdTo,
			LimitVal: limit, OffsetVal: offset,
		},
	}, nil
}

func parseTaskTypes(items []string) ([]sqlc.TaskType, error) {
	out := make([]sqlc.TaskType, 0, len(items))
	for _, s := range items {
		t := sqlc.TaskType(strings.ToUpper(strings.TrimSpace(s)))
		switch t {
		case sqlc.TaskTypeTASK, sqlc.TaskTypeSTORY, sqlc.TaskTypeBUG, sqlc.TaskTypeEPIC, sqlc.TaskTypeSUBTASK:
			out = append(out, t)
		default:
			return nil, apperror.BadRequest(fmt.Sprintf("invalid taskType %q", s))
		}
	}
	return out, nil
}

func parseTaskPriorities(items []string) ([]sqlc.TaskPriority, error) {
	out := make([]sqlc.TaskPriority, 0, len(items))
	for _, s := range items {
		p := sqlc.TaskPriority(strings.ToUpper(strings.TrimSpace(s)))
		switch p {
		case sqlc.TaskPriorityLOWEST, sqlc.TaskPriorityLOW, sqlc.TaskPriorityMEDIUM, sqlc.TaskPriorityHIGH, sqlc.TaskPriorityHIGHEST:
			out = append(out, p)
		default:
			return nil, apperror.BadRequest(fmt.Sprintf("invalid priority %q", s))
		}
	}
	return out, nil
}

func parseUUIDSlice(items []string, field string) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, 0, len(items))
	for _, s := range items {
		parsed, err := uuid.Parse(strings.TrimSpace(s))
		if err != nil {
			return nil, apperror.BadRequest(field + " contains an invalid UUID")
		}
		out = append(out, parsed)
	}
	return out, nil
}

func parseOptTime(raw, field string) (pgtype.Timestamp, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return pgtype.Timestamp{}, nil
	}
	parsed, err := time.Parse(time.RFC3339, trimmed)
	if err != nil {
		parsed, err = time.Parse("2006-01-02", trimmed)
		if err != nil {
			return pgtype.Timestamp{}, apperror.BadRequest(field + " must be a valid ISO 8601 date string")
		}
	}
	return pgtype.Timestamp{Time: parsed.UTC(), Valid: true}, nil
}

func strPtr(s string) *string { return &s }

func mapGlobalTasks(rows []sqlc.SearchTasksGlobalQueryRow, rawQuery string) []ResultItem {
	out := make([]ResultItem, 0, len(rows))
	for i := range rows {
		r := &rows[i]
		desc := r.Description.String
		out = append(out, ResultItem{
			ID: r.ID.String(), Type: "task", Title: r.Title, Description: desc,
			ProjectID: strPtr(r.ProjectID.String()), ProjectKey: strPtr(r.ProjectSlug),
			OrganizationID: strPtr(r.OrganizationID.String()), WorkspaceID: strPtr(r.WorkspaceID.String()),
			RelevanceScore: calculateRelevanceScore(r.Title, desc, rawQuery),
			CreatedAt:      formatTime(r.CreatedAt.Time), UpdatedAt: formatTime(r.UpdatedAt.Time),
			Metadata: map[string]any{
				"key": r.Slug, "type": string(r.Type), "priority": string(r.Priority),
			},
		})
	}
	return out
}

func mapGlobalProjects(rows []sqlc.SearchProjectsGlobalQueryRow, rawQuery string) []ResultItem {
	out := make([]ResultItem, 0, len(rows))
	for i := range rows {
		r := &rows[i]
		desc := r.Description.String
		out = append(out, ResultItem{
			ID: r.ID.String(), Type: "project", Title: r.Name, Description: desc,
			ProjectID: strPtr(r.ID.String()), ProjectKey: strPtr(r.Slug),
			OrganizationID: strPtr(r.OrganizationID.String()), WorkspaceID: strPtr(r.WorkspaceID.String()),
			RelevanceScore: calculateRelevanceScore(r.Name, desc, rawQuery),
			CreatedAt:      formatTime(r.CreatedAt.Time), UpdatedAt: formatTime(r.UpdatedAt.Time),
			Metadata: map[string]any{
				"slug": r.Slug, "status": string(r.Status), "priority": string(r.Priority),
			},
		})
	}
	return out
}

func mapGlobalUsers(rows []sqlc.SearchUsersGlobalQueryRow, rawQuery string) []ResultItem {
	out := make([]ResultItem, 0, len(rows))
	for i := range rows {
		r := &rows[i]
		fullName := strings.TrimSpace(r.FirstName + " " + r.LastName)
		out = append(out, ResultItem{
			ID: r.ID.String(), Type: "user", Title: fullName, Description: r.Email,
			RelevanceScore: calculateRelevanceScore(fullName, r.Email, rawQuery),
			CreatedAt:      formatTime(r.CreatedAt.Time), UpdatedAt: formatTime(r.UpdatedAt.Time),
			Metadata: map[string]any{
				"email": r.Email, "username": r.Username, "role": string(r.Role),
			},
		})
	}
	return out
}

func mapGlobalComments(rows []sqlc.SearchCommentsGlobalQueryRow, rawQuery string) []ResultItem {
	out := make([]ResultItem, 0, len(rows))
	for i := range rows {
		r := &rows[i]
		out = append(out, ResultItem{
			ID: r.ID.String(), Type: "comment", Title: "Comment on " + r.TaskTitle, Description: r.Content,
			ProjectID: strPtr(r.ProjectID.String()), ProjectKey: strPtr(r.ProjectSlug),
			OrganizationID: strPtr(r.OrganizationID.String()), WorkspaceID: strPtr(r.WorkspaceID.String()),
			RelevanceScore: calculateRelevanceScore(r.TaskTitle, r.Content, rawQuery),
			CreatedAt:      formatTime(r.CreatedAt.Time), UpdatedAt: formatTime(r.UpdatedAt.Time),
			Metadata: map[string]any{
				"taskId": r.TaskID.String(), "taskKey": r.TaskSlug,
			},
		})
	}
	return out
}

func mapGlobalAttachments(rows []sqlc.SearchAttachmentsGlobalQueryRow, rawQuery string) []ResultItem {
	out := make([]ResultItem, 0, len(rows))
	for i := range rows {
		r := &rows[i]
		out = append(out, ResultItem{
			ID: r.ID.String(), Type: "attachment", Title: r.FileName, Description: "Attachment on " + r.TaskTitle,
			ProjectID: strPtr(r.ProjectID.String()), ProjectKey: strPtr(r.ProjectSlug),
			OrganizationID: strPtr(r.OrganizationID.String()), WorkspaceID: strPtr(r.WorkspaceID.String()),
			RelevanceScore: calculateRelevanceScore(r.FileName, r.TaskTitle, rawQuery),
			CreatedAt:      formatTime(r.CreatedAt.Time), UpdatedAt: formatTime(r.UpdatedAt.Time),
			Metadata: map[string]any{
				"fileSize": r.FileSize, "mimeType": r.MimeType, "taskId": r.TaskID.String(), "taskKey": r.TaskSlug,
			},
		})
	}
	return out
}

func mapGlobalSprints(rows []sqlc.SearchSprintsGlobalQueryRow, rawQuery string) []ResultItem {
	out := make([]ResultItem, 0, len(rows))
	for i := range rows {
		r := &rows[i]
		desc := r.Goal.String
		out = append(out, ResultItem{
			ID: r.ID.String(), Type: "sprint", Title: r.Name, Description: desc,
			ProjectID: strPtr(r.ProjectID.String()), ProjectKey: strPtr(r.ProjectSlug),
			OrganizationID: strPtr(r.OrganizationID.String()), WorkspaceID: strPtr(r.WorkspaceID.String()),
			RelevanceScore: calculateRelevanceScore(r.Name, desc, rawQuery),
			CreatedAt:      formatTime(r.CreatedAt.Time), UpdatedAt: formatTime(r.UpdatedAt.Time),
			Metadata: map[string]any{
				"status": string(r.Status),
			},
		})
	}
	return out
}

func mapAdvancedTaskRow(
	r *sqlc.SearchTasksAdvancedQueryRow,
	rawQuery string,
	assignees, reporters []map[string]any,
) ResultItem {
	desc := r.Description.String
	meta := map[string]any{
		"key": r.Slug, "type": string(r.Type), "priority": string(r.Priority),
		"assignee": assignees, "reporter": reporters,
	}
	if r.StatusID.Valid {
		meta["status"] = map[string]any{
			"id": uuid.UUID(r.StatusID.Bytes).String(), "name": r.StatusName.String, "color": r.StatusColor.String,
		}
	}
	if r.SprintID.Valid {
		meta["sprint"] = map[string]any{
			"id": uuid.UUID(r.SprintID.Bytes).String(), "name": r.SprintName.String, "status": string(r.SprintStatus.SprintStatus),
		}
	}
	if r.DueDate.Valid {
		meta["dueDate"] = formatTime(r.DueDate.Time)
	}
	if r.StoryPoints.Valid {
		meta["storyPoints"] = r.StoryPoints.Int32
	}

	return ResultItem{
		ID: r.ID.String(), Type: "task", Title: r.Title, Description: desc,
		ProjectID: strPtr(r.ProjectID.String()), ProjectKey: strPtr(r.ProjectSlug),
		OrganizationID: strPtr(r.OrganizationID.String()), WorkspaceID: strPtr(r.WorkspaceID.String()),
		RelevanceScore: calculateRelevanceScore(r.Title, desc, rawQuery),
		CreatedAt:      formatTime(r.CreatedAt.Time), UpdatedAt: formatTime(r.UpdatedAt.Time),
		Metadata: meta,
	}
}
