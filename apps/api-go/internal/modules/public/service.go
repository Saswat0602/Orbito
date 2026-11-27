package public

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// URLResolver resolves presigned attachment URLs when Attachment.URL is empty.
type URLResolver interface {
	PresignGetURL(ctx context.Context, key string) (string, error)
}

// Service implements read-only public endpoint queries and sanitization.
type Service interface {
	GetSharedTaskByToken(ctx context.Context, token string) (*SharedTaskResponse, error)
	GetSharedAttachmentURL(ctx context.Context, token string, attachmentID uuid.UUID) (*AttachmentURLResponse, error)
	GetWorkspaceProjects(ctx context.Context, workspaceSlug string) ([]ProjectResponse, error)
	GetPublicProject(ctx context.Context, workspaceSlug, projectSlug string) (*ProjectResponse, error)
	GetProjectStatuses(ctx context.Context, projectSlug string) ([]StatusResponse, error)
	GetProjectCharts(ctx context.Context, projectSlug string, chartTypes []string) (map[string]any, error)
	GetCalendarEvents(ctx context.Context, workspaceSlug, projectSlug, startDate, endDate string) ([]CalendarEvent, error)
	GetPublicSprints(ctx context.Context, workspaceSlug, projectSlug string) ([]SprintResponse, error)
	GetPublicSprint(ctx context.Context, workspaceSlug, projectSlug string, sprintID uuid.UUID) (*SprintResponse, error)
	GetPublicSprintTasks(ctx context.Context, workspaceSlug, projectSlug string, sprintID uuid.UUID) ([]TaskResponse, error)
	GetPublicTasks(ctx context.Context, workspaceSlug, projectSlug string, filters TaskFilters) (*TaskPaginationResponse, error)
	GetPublicTasksKanban(ctx context.Context, workspaceSlug, projectSlug string) (*KanbanBoardResponse, error)
	GetPublicTaskByID(ctx context.Context, taskID uuid.UUID) (*TaskResponse, error)
	GetPublicTaskBySlug(ctx context.Context, slug string) (*TaskResponse, error)
	GetTaskActivities(ctx context.Context, taskID uuid.UUID, page, limit int) (*ActivitiesResponse, error)
	GetPublicTaskComments(ctx context.Context, taskID uuid.UUID) ([]CommentResponse, error)
	GetPublicTaskAttachments(ctx context.Context, taskID uuid.UUID) ([]AttachmentItem, error)
}

type service struct {
	repo        Repository
	urlResolver URLResolver
	nowFn       func() time.Time
}

// NewService constructs a public service instance.
func NewService(repo Repository, urlResolver URLResolver) Service {
	return &service{repo: repo, urlResolver: urlResolver, nowFn: time.Now}
}

func (s *service) ensureRepo() error {
	if s.repo == nil {
		return apperror.NotFound("Public resource unavailable")
	}
	return nil
}

func (s *service) GetSharedTaskByToken(ctx context.Context, token string) (*SharedTaskResponse, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	share, err := s.repo.GetSharedTaskByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if share == nil {
		return nil, apperror.NotFound("Share link not found")
	}
	if share.RevokedAt.Valid {
		return nil, apperror.NotFound("Share link has been revoked")
	}
	if share.ExpiresAt.Valid && s.nowFn().After(share.ExpiresAt.Time) {
		return nil, apperror.NotFound("Share link has expired")
	}
	assignees, _ := s.repo.ListTaskAssigneesBatch(ctx, []uuid.UUID{share.ID})
	attachments, _ := s.repo.ListTaskAttachments(ctx, share.ID)
	out := mapSharedTask(share, assignees, attachments)
	return &out, nil
}

func (s *service) GetSharedAttachmentURL(ctx context.Context, token string, attachmentID uuid.UUID) (*AttachmentURLResponse, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	att, err := s.repo.GetShareAttachmentByID(ctx, token, attachmentID)
	if err != nil {
		return nil, err
	}
	if att == nil {
		return nil, apperror.NotFound("Share link or attachment not found or expired")
	}
	if att.Url.Valid && strings.TrimSpace(att.Url.String) != "" {
		return &AttachmentURLResponse{URL: att.Url.String}, nil
	}
	if att.StorageKey.Valid && strings.TrimSpace(att.StorageKey.String) != "" && s.urlResolver != nil {
		u, uErr := s.urlResolver.PresignGetURL(ctx, att.StorageKey.String)
		if uErr == nil && u != "" {
			return &AttachmentURLResponse{URL: u}, nil
		}
	}
	return nil, apperror.BadRequest("Attachment URL not available")
}

func (s *service) GetWorkspaceProjects(ctx context.Context, workspaceSlug string) ([]ProjectResponse, error) {
	if s.repo == nil {
		return []ProjectResponse{}, nil
	}
	rows, err := s.repo.ListWorkspaceProjects(ctx, workspaceSlug)
	if err != nil {
		return nil, err
	}
	out := make([]ProjectResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapWorkspaceProjectRow(r))
	}
	return out, nil
}

func (s *service) GetPublicProject(ctx context.Context, workspaceSlug, projectSlug string) (*ProjectResponse, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	row, err := s.repo.GetProjectBySlugs(ctx, workspaceSlug, projectSlug)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, apperror.NotFound("Public project not found")
	}
	out := mapProjectDetailRow(row)
	return &out, nil
}

func (s *service) GetProjectStatuses(ctx context.Context, projectSlug string) ([]StatusResponse, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	proj, err := s.repo.GetProjectBySlugOnly(ctx, projectSlug)
	if err != nil {
		return nil, err
	}
	if proj == nil {
		return nil, apperror.NotFound("Public project not found")
	}
	rows, err := s.repo.ListStatusesByWorkflow(ctx, proj.WorkflowID)
	if err != nil {
		return nil, err
	}
	out := make([]StatusResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, StatusResponse{
			ID: r.ID, Name: r.Name, Color: r.Color, Category: string(r.Category),
			Position: r.Position, WorkflowID: r.WorkflowID, CreatedAt: tsVal(r.CreatedAt), UpdatedAt: tsVal(r.UpdatedAt),
			Workflow: WorkflowSummary{
				ID: r.WorkflowID, Name: r.WorkflowName, Description: textPtr(r.WorkflowDescription),
				Organization: OrganizationSummary{ID: r.OrgID, Name: r.OrgName, Slug: r.OrgSlug},
			},
			Count: StatusCount{Tasks: r.TaskCount},
		})
	}
	return out, nil
}

func (s *service) GetProjectCharts(ctx context.Context, projectSlug string, chartTypes []string) (map[string]any, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	if len(chartTypes) == 0 {
		return nil, apperror.BadRequest("types query parameter is required")
	}
	proj, err := s.repo.GetProjectBySlugOnly(ctx, projectSlug)
	if err != nil {
		return nil, err
	}
	if proj == nil {
		return nil, apperror.NotFound("Project not found")
	}

	res := make(map[string]any, len(chartTypes))
	for _, ct := range chartTypes {
		val, cErr := s.fetchSingleChart(ctx, proj.ID, proj.WorkflowID, strings.TrimSpace(ct))
		if cErr != nil {
			return nil, cErr
		}
		res[ct] = val
	}
	return res, nil
}

func (s *service) fetchSingleChart(ctx context.Context, projectID, workflowID uuid.UUID, ct string) (any, error) {
	switch ct {
	case "task-status":
		rows, err := s.repo.ListTaskCountsByStatus(ctx, projectID, workflowID)
		if err != nil {
			return nil, err
		}
		out := make([]StatusFlowItem, 0, len(rows))
		for i, r := range rows {
			out = append(out, StatusFlowItem{
				StatusID: r.StatusID, Count: r.Count,
				Status: StatusFlowMeta{ID: r.StatusID, Name: r.StatusName, Color: r.StatusColor, Category: string(r.StatusCategory), Position: int32(i)}, // #nosec G115
			})
		}
		return out, nil
	case "task-type":
		rows, err := s.repo.ListTaskCountsByType(ctx, projectID)
		if err != nil {
			return nil, err
		}
		out := make([]TaskTypeCount, 0, len(rows))
		for _, r := range rows {
			out = append(out, TaskTypeCount{Type: string(r.Type), Count: map[string]int64{"type": r.Count}})
		}
		return out, nil
	case "task-priority":
		rows, err := s.repo.ListTaskCountsByPriority(ctx, projectID)
		if err != nil {
			return nil, err
		}
		out := make([]TaskPriorityCount, 0, len(rows))
		for _, r := range rows {
			out = append(out, TaskPriorityCount{Priority: string(r.Priority), Count: map[string]int64{"priority": r.Count}})
		}
		return out, nil
	case "kpi-metrics":
		kpi, err := s.repo.GetProjectKPIs(ctx, projectID)
		if err != nil {
			return nil, err
		}
		if kpi == nil {
			return KPIMetrics{}, nil
		}
		return KPIMetrics{
			TotalTasks: kpi.TotalTasks, CompletedTasks: kpi.CompletedTasks, ActiveSprints: kpi.ActiveSprints,
			TotalBugs: kpi.TotalBugs, ResolvedBugs: kpi.ResolvedBugs,
			CompletionRate: roundPct(kpi.CompletedTasks, kpi.TotalTasks), BugResolutionRate: roundPct(kpi.ResolvedBugs, kpi.TotalBugs),
		}, nil
	case "sprint-velocity":
		rows, err := s.repo.ListSprintVelocity(ctx, projectID)
		if err != nil {
			return nil, err
		}
		out := make([]SprintVelocityItem, 0, len(rows))
		for _, r := range rows {
			out = append(out, SprintVelocityItem{
				ID: r.ID, Name: r.Name, StartDate: tsPtr(r.StartDate), EndDate: tsPtr(r.EndDate), Velocity: r.CompletedPoints,
			})
		}
		return out, nil
	default:
		return nil, apperror.BadRequest("Unsupported chart type: " + ct)
	}
}

func (s *service) GetCalendarEvents(ctx context.Context, workspaceSlug, projectSlug, startDate, endDate string) ([]CalendarEvent, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	proj, err := s.repo.GetProjectBySlugs(ctx, workspaceSlug, projectSlug)
	if err != nil {
		return nil, err
	}
	if proj == nil {
		return nil, apperror.NotFound("Public project not found")
	}
	params := sqlc.ListPublicCalendarTasksParams{ProjectID: proj.ID}
	if startDate != "" && endDate != "" {
		st, sErr := parseISODate(startDate)
		en, eErr := parseISODate(endDate)
		if sErr == nil && eErr == nil {
			params.StartDate = pgtype.Timestamp{Time: st, Valid: true}
			params.EndDate = pgtype.Timestamp{Time: en, Valid: true}
		}
	}
	rows, err := s.repo.ListCalendarTasks(ctx, params)
	if err != nil {
		return nil, err
	}
	return buildCalendarEvents(rows), nil
}

func (s *service) GetPublicSprints(ctx context.Context, workspaceSlug, projectSlug string) ([]SprintResponse, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	proj, err := s.repo.GetProjectBySlugs(ctx, workspaceSlug, projectSlug)
	if err != nil {
		return nil, err
	}
	if proj == nil {
		return nil, apperror.NotFound("Public project not found")
	}
	rows, err := s.repo.ListSprints(ctx, proj.ID)
	if err != nil {
		return nil, err
	}
	out := make([]SprintResponse, 0, len(rows))
	for _, r := range rows {
		prog := 0
		if r.TotalTasks > 0 {
			prog = int(math.Round(float64(r.CompletedTasks) / float64(r.TotalTasks) * 100))
		}
		out = append(out, SprintResponse{
			ID: r.ID, Name: r.Name, Description: textPtr(r.Goal), Status: string(r.Status),
			StartDate: tsPtr(r.StartDate), EndDate: tsPtr(r.EndDate), CreatedAt: tsVal(r.CreatedAt),
			Progress: prog, IsPublicView: true,
		})
	}
	return out, nil
}

func (s *service) GetPublicSprint(ctx context.Context, workspaceSlug, projectSlug string, sprintID uuid.UUID) (*SprintResponse, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	r, err := s.repo.GetSprintByID(ctx, sprintID, workspaceSlug, projectSlug)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, apperror.NotFound("Public sprint not found")
	}
	tasks, _ := s.GetPublicSprintTasks(ctx, workspaceSlug, projectSlug, sprintID)
	prog := 0
	if r.TotalTasks > 0 {
		prog = int(math.Round(float64(r.CompletedTasks) / float64(r.TotalTasks) * 100))
	}
	return &SprintResponse{
		ID: r.ID, Name: r.Name, Description: textPtr(r.Goal), Status: string(r.Status),
		StartDate: tsPtr(r.StartDate), EndDate: tsPtr(r.EndDate), CreatedAt: tsVal(r.CreatedAt),
		Progress: prog, Tasks: tasks, IsPublicView: true,
	}, nil
}

func (s *service) GetPublicSprintTasks(ctx context.Context, workspaceSlug, projectSlug string, sprintID uuid.UUID) ([]TaskResponse, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	proj, err := s.repo.GetProjectBySlugs(ctx, workspaceSlug, projectSlug)
	if err != nil {
		return nil, err
	}
	if proj == nil {
		return nil, apperror.NotFound("Public project not found")
	}
	rows, err := s.repo.ListSprintTasks(ctx, sprintID, workspaceSlug, projectSlug)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	labels := s.loadLabelsMap(ctx, ids)
	out := make([]TaskResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapTaskRow(r.ID, r.Title, r.Description, r.Type, r.Priority, r.Slug, r.DueDate, r.CreatedAt, r.StatusID, r.StatusName, r.StatusColor, r.StatusCategory, labels[r.ID]))
	}
	return out, nil
}

func (s *service) GetPublicTasks(ctx context.Context, workspaceSlug, projectSlug string, filters TaskFilters) (*TaskPaginationResponse, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	proj, err := s.repo.GetProjectBySlugs(ctx, workspaceSlug, projectSlug)
	if err != nil {
		return nil, err
	}
	if proj == nil {
		return nil, apperror.NotFound("Public project not found")
	}
	page, limit := max(1, filters.Page), min(max(1, filters.Limit), 100)
	if filters.Limit <= 0 {
		limit = 50
	}
	countArgs, listArgs := buildTaskFilterParams(proj.ID, filters, page, limit)
	total, err := s.repo.CountProjectTasks(ctx, countArgs)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.ListProjectTasks(ctx, listArgs)
	if err != nil {
		return nil, err
	}
	tasks := s.hydrateTaskRows(ctx, rows)
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	return &TaskPaginationResponse{Data: tasks, Page: page, Total: total, Limit: limit, TotalPages: totalPages}, nil
}

func (s *service) GetPublicTasksKanban(ctx context.Context, workspaceSlug, projectSlug string) (*KanbanBoardResponse, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	proj, err := s.repo.GetProjectBySlugs(ctx, workspaceSlug, projectSlug)
	if err != nil {
		return nil, err
	}
	if proj == nil {
		return nil, apperror.NotFound("Public project not found")
	}
	statuses, err := s.repo.ListStatusesByWorkflow(ctx, proj.WorkflowID)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.ListProjectTasks(ctx, sqlc.ListPublicProjectTasksParams{ProjectID: proj.ID, LimitVal: 500, OffsetVal: 0})
	if err != nil {
		return nil, err
	}
	tasks := s.hydrateTaskRows(ctx, rows)
	byStatus := make(map[string][]TaskResponse)
	for _, t := range tasks {
		byStatus[t.Status.ID] = append(byStatus[t.Status.ID], t)
	}
	cols := make([]KanbanColumn, 0, len(statuses))
	for _, st := range statuses {
		colTasks := byStatus[st.ID.String()]
		if colTasks == nil {
			colTasks = []TaskResponse{}
		}
		cols = append(cols, KanbanColumn{ID: st.ID, Name: st.Name, Color: st.Color, Category: string(st.Category), Tasks: colTasks})
	}
	return &KanbanBoardResponse{Columns: cols}, nil
}

func (s *service) GetPublicTaskByID(ctx context.Context, taskID uuid.UUID) (*TaskResponse, error) {
	return s.findPublicTask(ctx, sqlc.GetPublicTaskByIDOrSlugParams{IsUuid: true, TaskUuid: pgtype.UUID{Bytes: taskID, Valid: true}})
}

func (s *service) GetPublicTaskBySlug(ctx context.Context, slug string) (*TaskResponse, error) {
	return s.findPublicTask(ctx, sqlc.GetPublicTaskByIDOrSlugParams{IsUuid: false, TaskSlug: slug})
}

func (s *service) findPublicTask(ctx context.Context, arg sqlc.GetPublicTaskByIDOrSlugParams) (*TaskResponse, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	row, err := s.repo.GetTaskByIDOrSlug(ctx, arg)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, apperror.NotFound("Task not found")
	}
	labels := s.loadLabelsMap(ctx, []uuid.UUID{row.ID})
	out := mapTaskRow(row.ID, row.Title, row.Description, row.Type, row.Priority, row.Slug, row.DueDate, row.CreatedAt, row.StatusRowID, row.StatusName, row.StatusColor, row.StatusCategory, labels[row.ID])
	if assRows, aErr := s.repo.ListTaskAssigneesBatch(ctx, []uuid.UUID{row.ID}); aErr == nil && len(assRows) > 0 {
		for _, a := range assRows {
			id := a.ID
			out.Assignees = append(out.Assignees, PersonSummary{ID: &id, FirstName: a.FirstName, LastName: a.LastName, Avatar: textPtr(a.Avatar)})
		}
	}
	out.Project = map[string]any{
		"name": row.ProjectName, "slug": row.ProjectSlug, "visibility": string(row.ProjectVisibility),
		"workspace": map[string]any{"id": row.WorkspaceID, "name": row.WorkspaceName, "slug": row.WorkspaceSlug},
	}
	return &out, nil
}

func (s *service) GetTaskActivities(ctx context.Context, taskID uuid.UUID, page, limit int) (*ActivitiesResponse, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	taskRow, err := s.repo.GetTaskByIDOrSlug(ctx, sqlc.GetPublicTaskByIDOrSlugParams{IsUuid: true, TaskUuid: pgtype.UUID{Bytes: taskID, Valid: true}})
	if err != nil {
		return nil, err
	}
	if taskRow == nil {
		return nil, apperror.NotFound("Task not found")
	}
	page, limit = max(1, page), min(max(1, limit), 100)
	offset := (page - 1) * limit
	total, err := s.repo.CountTaskActivities(ctx, taskID)
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.ListTaskActivities(ctx, taskID, int32(limit), int32(offset)) // #nosec G115
	if err != nil {
		return nil, err
	}
	items := make([]ActivityItem, 0, len(rows))
	for _, r := range rows {
		item := ActivityItem{
			ID: r.ID, Type: string(r.Type), Description: r.Description,
			EntityType: "Task", EntityID: taskID.String(), CreatedAt: tsVal(r.CreatedAt),
		}
		if r.UserID.Valid {
			uid := uuid.UUID(r.UserID.Bytes)
			item.User = &PersonSummary{ID: &uid, FirstName: textVal(r.FirstName), LastName: textVal(r.LastName), Avatar: textPtr(r.Avatar)}
		}
		items = append(items, item)
	}
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	return &ActivitiesResponse{
		Activities: items,
		Pagination: PaginationMeta{CurrentPage: page, TotalPages: totalPages, TotalCount: total, HasNextPage: page < totalPages, HasPrevPage: page > 1},
	}, nil
}

func (s *service) GetPublicTaskComments(ctx context.Context, taskID uuid.UUID) ([]CommentResponse, error) {
	if err := s.ensureRepo(); err != nil {
		return nil, err
	}
	taskRow, err := s.repo.GetTaskByIDOrSlug(ctx, sqlc.GetPublicTaskByIDOrSlugParams{IsUuid: true, TaskUuid: pgtype.UUID{Bytes: taskID, Valid: true}})
	if err != nil {
		return nil, err
	}
	if taskRow == nil {
		return nil, apperror.NotFound("Task not found")
	}
	rows, err := s.repo.ListTaskComments(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return buildCommentTree(rows), nil
}

func (s *service) GetPublicTaskAttachments(ctx context.Context, taskID uuid.UUID) ([]AttachmentItem, error) {
	if s.repo == nil {
		return []AttachmentItem{}, nil
	}
	rows, err := s.repo.ListTaskAttachments(ctx, taskID)
	if err != nil {
		return nil, err
	}
	out := make([]AttachmentItem, 0, len(rows))
	for _, r := range rows {
		out = append(out, AttachmentItem{
			ID: r.ID, FileName: r.FileName, FileSize: r.FileSize, MimeType: r.MimeType, URL: textPtr(r.Url), CreatedAt: tsVal(r.CreatedAt),
		})
	}
	return out, nil
}

func (s *service) loadLabelsMap(ctx context.Context, ids []uuid.UUID) map[uuid.UUID][]LabelBadge {
	if len(ids) == 0 {
		return map[uuid.UUID][]LabelBadge{}
	}
	rows, err := s.repo.ListTaskLabelsBatch(ctx, ids)
	if err != nil {
		return map[uuid.UUID][]LabelBadge{}
	}
	return mapLabelsByTask(rows)
}

func (s *service) hydrateTaskRows(ctx context.Context, rows []sqlc.ListPublicProjectTasksRow) []TaskResponse {
	ids := make([]uuid.UUID, 0, len(rows))
	for _, r := range rows {
		ids = append(ids, r.ID)
	}
	labels := s.loadLabelsMap(ctx, ids)
	subtasksMap := make(map[uuid.UUID][]TaskResponse)
	if len(ids) > 0 {
		if subs, err := s.repo.ListSubtasksBatch(ctx, ids); err == nil {
			subIDs := make([]uuid.UUID, 0, len(subs))
			for _, sub := range subs {
				subIDs = append(subIDs, sub.ID)
			}
			subLabels := s.loadLabelsMap(ctx, subIDs)
			for _, sub := range subs {
				if sub.ParentTaskID.Valid {
					pid := uuid.UUID(sub.ParentTaskID.Bytes)
					subtasksMap[pid] = append(subtasksMap[pid], mapTaskRow(
						sub.ID, sub.Title, sub.Description, sub.Type, sub.Priority, sub.Slug,
						sub.DueDate, sub.CreatedAt, sub.StatusID, sub.StatusName, sub.StatusColor, sub.StatusCategory, subLabels[sub.ID],
					))
				}
			}
		}
	}

	out := make([]TaskResponse, 0, len(rows))
	for _, r := range rows {
		t := mapTaskRow(r.ID, r.Title, r.Description, r.Type, r.Priority, r.Slug, r.DueDate, r.CreatedAt, r.StatusID, r.StatusName, r.StatusColor, r.StatusCategory, labels[r.ID])
		t.Subtasks = subtasksMap[r.ID]
		out = append(out, t)
	}
	return out
}

func buildTaskFilterParams(projectID uuid.UUID, f TaskFilters, page, limit int) (sqlc.CountPublicProjectTasksParams, sqlc.ListPublicProjectTasksParams) {
	c := sqlc.CountPublicProjectTasksParams{ProjectID: projectID}
	l := sqlc.ListPublicProjectTasksParams{ProjectID: projectID, LimitVal: int32(limit), OffsetVal: int32((page - 1) * limit)} // #nosec G115

	if f.Status != "" {
		c.StatusName = pgtype.Text{String: f.Status, Valid: true}
		l.StatusName = pgtype.Text{String: f.Status, Valid: true}
	}
	if f.Priority != "" {
		prio := sqlc.NullTaskPriority{TaskPriority: sqlc.TaskPriority(strings.ToUpper(f.Priority)), Valid: true}
		c.Priority, l.Priority = prio, prio
	}
	if f.Type != "" {
		tt := sqlc.NullTaskType{TaskType: sqlc.TaskType(strings.ToUpper(f.Type)), Valid: true}
		c.TaskType, l.TaskType = tt, tt
	}
	if f.ParentTaskID == "all" {
		c.ParentMode = pgtype.Text{String: "all", Valid: true}
		l.ParentMode = pgtype.Text{String: "all", Valid: true}
	} else if f.ParentTaskID != "" {
		if pid, err := uuid.Parse(f.ParentTaskID); err == nil {
			pu := pgtype.UUID{Bytes: pid, Valid: true}
			c.ParentTaskID, l.ParentTaskID = pu, pu
		}
	}
	return c, l
}

func parseISODate(raw string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", raw)
}
