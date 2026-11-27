package public

import (
	"encoding/json"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

func tsPtr(ts pgtype.Timestamp) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}

func tsVal(ts pgtype.Timestamp) time.Time {
	if !ts.Valid {
		return time.Time{}
	}
	return ts.Time
}

func textPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	s := t.String
	return &s
}

func textVal(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

func roundPct(num, den int64) float64 {
	if den == 0 {
		return 0
	}
	return math.Round((float64(num)/float64(den))*10000) / 100
}

func allowPublicStats(rawSettings []byte) bool {
	if len(rawSettings) == 0 {
		return true
	}
	var parsed struct {
		AllowPublicStats *bool `json:"allowPublicStats"`
	}
	if err := json.Unmarshal(rawSettings, &parsed); err != nil {
		return true
	}
	if parsed.AllowPublicStats != nil && !*parsed.AllowPublicStats {
		return false
	}
	return true
}

func mapWorkspaceProjectRow(r sqlc.ListWorkspacePublicProjectsRow) ProjectResponse {
	return ProjectResponse{
		ID: r.ID, Name: r.Name, Slug: r.Slug,
		Description: textPtr(r.Description), Color: r.Color, Avatar: textPtr(r.Avatar),
		Status: string(r.Status), Priority: string(r.Priority), Visibility: string(r.Visibility),
		StartDate: tsPtr(r.StartDate), EndDate: tsPtr(r.EndDate), CreatedAt: tsVal(r.CreatedAt),
		Workspace:    WorkspaceSummary{Slug: r.WorkspaceSlug, Name: r.WorkspaceName},
		Organization: &OrganizationSummary{ID: r.OrgID, Name: r.OrgName, Slug: r.OrgSlug},
		IsPublicView: true,
	}
}

func mapProjectDetailRow(r *sqlc.GetPublicProjectBySlugsRow) ProjectResponse {
	resp := ProjectResponse{
		ID: r.ID, Name: r.Name, Slug: r.Slug,
		Description: textPtr(r.Description), Color: r.Color, Avatar: textPtr(r.Avatar),
		Status: string(r.Status), Priority: string(r.Priority), Visibility: string(r.Visibility),
		StartDate: tsPtr(r.StartDate), EndDate: tsPtr(r.EndDate), CreatedAt: tsVal(r.CreatedAt),
		Workspace:    WorkspaceSummary{Slug: r.WorkspaceSlug, Name: r.WorkspaceName},
		Organization: &OrganizationSummary{ID: r.OrgID, Name: r.OrgName, Slug: r.OrgSlug},
		IsPublicView: true,
	}
	if allowPublicStats(r.Settings) {
		resp.Stats = &ProjectStats{
			TaskCount:        min(r.TasksCount, int64(100)),
			CompletionRate:   roundPct(r.DoneCount, r.TasksCount),
			HasActiveSprints: r.SprintsCount > 0,
		}
	}
	return resp
}

func mapSharedTask(
	row *sqlc.GetPublicSharedTaskByTokenRow,
	assignees []sqlc.ListPublicTaskAssigneesBatchRow,
	attachments []sqlc.ListPublicTaskAttachmentsRow,
) SharedTaskResponse {
	resp := SharedTaskResponse{
		Title:       row.Title,
		Description: textPtr(row.Description),
		Status:      StatusBadge{Name: textVal(row.StatusName), Color: textVal(row.StatusColor)},
		Priority:    string(row.Priority),
		DueDate:     tsPtr(row.DueDate),
		Assignees:   make([]PersonSummary, 0, len(assignees)),
		Attachments: make([]SharedAttachmentItem, 0, len(attachments)),
	}
	if row.CreatorID.Valid {
		resp.CreatedBy = &PersonSummary{
			FirstName: textVal(row.CreatorFirstName),
			LastName:  textVal(row.CreatorLastName),
			Avatar:    textPtr(row.CreatorAvatar),
		}
	}
	for _, a := range assignees {
		resp.Assignees = append(resp.Assignees, PersonSummary{FirstName: a.FirstName, LastName: a.LastName})
	}
	for _, att := range attachments {
		resp.Attachments = append(resp.Attachments, SharedAttachmentItem{
			ID: att.ID, FileName: att.FileName, FileSize: att.FileSize, MimeType: att.MimeType,
		})
	}
	return resp
}

func buildCalendarEvents(rows []sqlc.ListPublicCalendarTasksRow) []CalendarEvent {
	seen := make(map[string]struct{}, len(rows)*2)
	events := make([]CalendarEvent, 0, len(rows))
	push := func(ev CalendarEvent) {
		if _, ok := seen[ev.ID]; ok {
			return
		}
		seen[ev.ID] = struct{}{}
		events = append(events, ev)
	}

	for _, r := range rows {
		if r.DueDate.Valid {
			tid := r.ID
			var st *StatusBadge
			if r.StatusID.Valid {
				st = &StatusBadge{
					ID: uuid.UUID(r.StatusID.Bytes).String(), Name: textVal(r.StatusName),
					Color: textVal(r.StatusColor), Category: string(r.StatusCategory.StatusCategory),
				}
			}
			push(CalendarEvent{
				ID: "task-" + r.ID.String(), Title: r.Title, Date: r.DueDate.Time,
				Type: "task-due", Status: st, TaskID: &tid,
			})
		}
		if r.SprintID.Valid {
			sid := uuid.UUID(r.SprintID.Bytes)
			if r.SprintStartDate.Valid {
				push(CalendarEvent{
					ID: "sprint-start-" + sid.String(), Title: textVal(r.SprintName) + " starts",
					Date: r.SprintStartDate.Time, Type: "sprint-start", SprintID: &sid,
				})
			}
			if r.SprintEndDate.Valid {
				push(CalendarEvent{
					ID: "sprint-end-" + sid.String(), Title: textVal(r.SprintName) + " ends",
					Date: r.SprintEndDate.Time, Type: "sprint-end", SprintID: &sid,
				})
			}
		}
	}

	sort.Slice(events, func(i, j int) bool { return events[i].Date.Before(events[j].Date) })
	return events
}

func mapLabelsByTask(rows []sqlc.ListPublicTaskLabelsBatchRow) map[uuid.UUID][]LabelBadge {
	out := make(map[uuid.UUID][]LabelBadge)
	for _, l := range rows {
		out[l.TaskID] = append(out[l.TaskID], LabelBadge{ID: l.ID, Name: l.Name, Color: l.Color})
	}
	return out
}

func mapTaskRow(
	id uuid.UUID, title string, desc pgtype.Text, tType sqlc.TaskType, prio sqlc.TaskPriority,
	slug string, due pgtype.Timestamp, created pgtype.Timestamp,
	statusID pgtype.UUID, statusName, statusColor pgtype.Text, statusCat sqlc.NullStatusCategory,
	labels []LabelBadge,
) TaskResponse {
	sid := ""
	if statusID.Valid {
		sid = uuid.UUID(statusID.Bytes).String()
	}
	if labels == nil {
		labels = []LabelBadge{}
	}
	return TaskResponse{
		ID: id, Title: title, Description: textPtr(desc),
		Type: string(tType), Priority: string(prio), Slug: slug,
		DueDate: tsPtr(due), CreatedAt: tsVal(created),
		Status: StatusBadge{
			ID: sid, Name: textVal(statusName), Color: textVal(statusColor), Category: string(statusCat.StatusCategory),
		},
		Labels:       labels,
		IsPublicView: true,
	}
}

func buildCommentTree(rows []sqlc.ListPublicTaskCommentsRow) []CommentResponse {
	byParent := make(map[uuid.UUID][]CommentResponse)
	roots := make([]CommentResponse, 0, len(rows))

	toItem := func(r sqlc.ListPublicTaskCommentsRow) CommentResponse {
		id := r.AuthorID
		item := CommentResponse{
			ID: r.ID, Content: r.Content, CreatedAt: tsVal(r.CreatedAt), UpdatedAt: tsVal(r.UpdatedAt),
			Author: PersonSummary{ID: &id, FirstName: r.FirstName, LastName: r.LastName, Avatar: textPtr(r.Avatar)},
		}
		if r.ParentCommentID.Valid {
			pid := uuid.UUID(r.ParentCommentID.Bytes)
			item.ParentCommentID = &pid
		}
		return item
	}

	for _, r := range rows {
		item := toItem(r)
		if r.ParentCommentID.Valid {
			pid := uuid.UUID(r.ParentCommentID.Bytes)
			byParent[pid] = append(byParent[pid], item)
		} else {
			roots = append(roots, item)
		}
	}

	for i := range roots {
		replies := byParent[roots[i].ID]
		if replies == nil {
			replies = []CommentResponse{}
		}
		roots[i].Replies = replies
		roots[i].Count = CommentCount{Replies: len(replies)}
	}
	return roots
}
