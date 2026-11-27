package tasks

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/modules/task_ranks"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/platform/storage"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/sanitize"
)

// UploadedFile holds buffered file upload metadata from multipart requests.
type UploadedFile struct {
	FileName string
	MimeType string
	Content  []byte
}

// Service defines business operations for tasks and task shares.
type Service interface {
	Create(ctx context.Context, req CreateTaskRequest, userID uuid.UUID, isSuperAdmin bool) (*TaskResponse, error)
	CreateWithAttachments(ctx context.Context, req CreateTaskRequest, files []UploadedFile, userID uuid.UUID, isSuperAdmin bool) (*TaskResponse, error)
	FindAll(ctx context.Context, q TaskFilterQuery, userID uuid.UUID, isSuperAdmin bool) (*PaginatedTasksResponse, error)
	GetTasks(ctx context.Context, q TaskFilterQuery, userID uuid.UUID, isSuperAdmin bool) (*PaginatedTasksResponse, error)
	GetTasksGrouped(ctx context.Context, q GroupedTasksQuery, userID uuid.UUID, isSuperAdmin bool) (*GroupedTasksResponse, error)
	GetTasksGroupedByStatus(ctx context.Context, q ByStatusQuery, userID uuid.UUID, isSuperAdmin bool) (*ByStatusEnvelope, error)
	FindTodaysTasks(ctx context.Context, orgRaw string, page, limit int, userID uuid.UUID, isSuperAdmin bool) (*TodayTasksResponse, error)
	FindByOrganization(ctx context.Context, orgID uuid.UUID, priority, search string, page, limit int, userID uuid.UUID, isSuperAdmin bool) (*TodayTasksResponse, error)
	FindOne(ctx context.Context, idOrSlug string, userID uuid.UUID, isSuperAdmin bool) (*TaskResponse, error)
	FindByKey(ctx context.Context, key string, userID uuid.UUID, isSuperAdmin bool) (*TaskResponse, error)
	Update(ctx context.Context, idOrSlug string, req UpdateTaskRequest, userID uuid.UUID, isSuperAdmin bool) (*TaskResponse, error)
	Remove(ctx context.Context, idOrSlug string, userID uuid.UUID, isSuperAdmin bool) (*TaskResponse, error)
	AddComment(ctx context.Context, idOrSlug string, comment string, userID uuid.UUID, isSuperAdmin bool) (*CommentBrief, error)
	ReorderTask(ctx context.Context, taskIDOrSlug string, req task_ranks.ReorderRequest, userID uuid.UUID, isSuperAdmin bool) error
	BulkCreate(ctx context.Context, req BulkCreateTasksRequest, userID uuid.UUID, isSuperAdmin bool) (*BulkCreateResponse, error)
	BulkDeleteTasks(ctx context.Context, req BulkDeleteTasksRequest, userID uuid.UUID, isSuperAdmin bool) (*BulkDeleteResponse, error)
	BulkUpdateTasksStatus(ctx context.Context, req BulkUpdateTasksStatusRequest, userID uuid.UUID, isSuperAdmin bool) (*BulkUpdateStatusResponse, error)
	BulkAssignTasks(ctx context.Context, req BulkAssignTasksRequest, userID uuid.UUID, isSuperAdmin bool) (*BulkAssignResponse, error)
	CompleteOccurrenceAndGenerateNext(ctx context.Context, taskIDOrSlug string, userID uuid.UUID, isSuperAdmin bool) (*CompleteOccurrenceResponse, error)
	AddRecurrence(ctx context.Context, taskIDOrSlug string, req RecurrenceConfigRequest, userID uuid.UUID, isSuperAdmin bool) (*RecurrenceResponse, error)
	UpdateRecurrenceConfig(ctx context.Context, taskIDOrSlug string, req RecurrenceConfigRequest, userID uuid.UUID, isSuperAdmin bool) (*RecurrenceResponse, error)
	StopRecurrence(ctx context.Context, taskIDOrSlug string, userID uuid.UUID, isSuperAdmin bool) (*TaskResponse, error)
	GetRecurringTasks(ctx context.Context, projectID, userID uuid.UUID, isSuperAdmin bool) ([]TaskResponse, error)
	CreateShareLink(ctx context.Context, req CreateShareLinkRequest, userID uuid.UUID, isSuperAdmin bool) (*ShareLinkResponse, error)
	GetSharesForTask(ctx context.Context, taskIDOrSlug string, userID uuid.UUID, isSuperAdmin bool) ([]ShareLinkResponse, error)
	RevokeShare(ctx context.Context, shareID uuid.UUID, userID uuid.UUID, isSuperAdmin bool) error
}

type service struct {
	repo        Repository
	rankSvc     task_ranks.Service
	storage     storage.Provider
	emitter     EventEmitter
	activity    ActivityLogger
	notifier    TaskNotifier
	frontendURL string
}

// NewService constructs a tasks service with optional rank/storage/event collaborators.
func NewService(repo Repository, rankSvc task_ranks.Service, stor storage.Provider, emitter EventEmitter, frontendURL string) Service {
	return &service{
		repo:        repo,
		rankSvc:     rankSvc,
		storage:     stor,
		emitter:     emitter,
		frontendURL: frontendURL,
	}
}

// AttachPhase5Hooks attaches optional activity logging and notification dispatchers.
func AttachPhase5Hooks(svc Service, activity ActivityLogger, notifier TaskNotifier) {
	if s, ok := svc.(*service); ok {
		s.activity = activity
		s.notifier = notifier
	}
}

func (s *service) parseCreateInput(ctx context.Context, req CreateTaskRequest, userID uuid.UUID, isSuperAdmin bool) (CreateTxInput, error) {
	projectID := uuid.MustParse(req.ProjectID)
	meta, err := s.repo.GetProjectMeta(ctx, projectID)
	if err != nil {
		return CreateTxInput{}, apperror.NotFound("Project not found")
	}

	canChange, err := s.repo.GetProjectAccessForCreate(ctx, projectID, userID, isSuperAdmin)
	if err != nil {
		return CreateTxInput{}, err
	}
	if !canChange {
		return CreateTxInput{}, apperror.Forbidden("Insufficient permissions to create task in this project")
	}

	var startTS, dueTS, completedTS pgtype.Timestamp
	var startT, dueT time.Time
	if req.StartDate != nil && *req.StartDate != "" {
		parsed, ok := parseISO8601(*req.StartDate)
		if !ok {
			return CreateTxInput{}, apperror.BadRequest("startDate must be a valid ISO 8601 date string")
		}
		startT = parsed
		startTS = pgtype.Timestamp{Time: parsed, Valid: true}
	}
	if req.DueDate != nil && *req.DueDate != "" {
		parsed, ok := parseISO8601(*req.DueDate)
		if !ok {
			return CreateTxInput{}, apperror.BadRequest("dueDate must be a valid ISO 8601 date string")
		}
		dueT = parsed
		dueTS = pgtype.Timestamp{Time: parsed, Valid: true}
	}
	if startTS.Valid && dueTS.Valid && startT.After(dueT) {
		return CreateTxInput{}, apperror.BadRequest("Start date must be before the due date")
	}
	if req.CompletedAt != nil && *req.CompletedAt != "" {
		parsed, ok := parseISO8601(*req.CompletedAt)
		if !ok {
			return CreateTxInput{}, apperror.BadRequest("completedAt must be a valid ISO 8601 date string")
		}
		completedTS = pgtype.Timestamp{Time: parsed, Valid: true}
	}

	var sprintID pgtype.UUID
	if req.SprintID != nil && *req.SprintID != "" {
		sid := uuid.MustParse(*req.SprintID)
		sprintID = pgtype.UUID{Bytes: sid, Valid: true}
	} else {
		sprintID, _ = s.repo.GetDefaultSprintID(ctx, projectID)
	}

	taskType := sqlc.TaskTypeTASK
	if req.Type != nil && *req.Type != "" {
		taskType = sqlc.TaskType(*req.Type)
	}
	priority := sqlc.TaskPriorityMEDIUM
	if req.Priority != nil && *req.Priority != "" {
		priority = sqlc.TaskPriority(*req.Priority)
	}

	var desc pgtype.Text
	if req.Description != nil {
		desc = pgtype.Text{String: sanitize.HTML(*req.Description), Valid: true}
	}
	var statusPtr *uuid.UUID
	if req.StatusID != nil && *req.StatusID != "" {
		sid := uuid.MustParse(*req.StatusID)
		statusPtr = &sid
	}
	var sp, oe, re pgtype.Int4
	if req.StoryPoints != nil {
		sp = pgtype.Int4{Int32: *req.StoryPoints, Valid: true}
	}
	if req.OriginalEstimate != nil {
		oe = pgtype.Int4{Int32: *req.OriginalEstimate, Valid: true}
	}
	if req.RemainingEstimate != nil {
		re = pgtype.Int4{Int32: *req.RemainingEstimate, Valid: true}
	}

	rawAssignees := req.AssigneeIDs
	if len(rawAssignees) == 0 && req.AssigneeID != nil && *req.AssigneeID != "" {
		rawAssignees = []string{*req.AssigneeID}
	}
	assignees := make([]uuid.UUID, 0, len(rawAssignees))
	for _, raw := range rawAssignees {
		if id, pErr := uuid.Parse(raw); pErr == nil {
			assignees = append(assignees, id)
		}
	}
	reporters := make([]uuid.UUID, 0, len(req.ReporterIDs))
	for _, raw := range req.ReporterIDs {
		if id, pErr := uuid.Parse(raw); pErr == nil {
			reporters = append(reporters, id)
		}
	}

	var customBytes []byte
	if req.CustomFields != nil {
		if sanitized, ok := sanitize.Object(req.CustomFields).(map[string]any); ok {
			customBytes = marshalJSONField(sanitized)
		}
	}

	isRec := req.IsRecurring != nil && *req.IsRecurring
	var recParams *sqlc.CreateRecurringTaskConfigParams
	if isRec && req.RecurrenceConfig != nil {
		baseTime := time.Now().UTC()
		if dueTS.Valid {
			baseTime = dueTS.Time.UTC()
		}
		built, rErr := buildRecurrenceParams(uuid.Nil, *req.RecurrenceConfig, baseTime)
		if rErr != nil {
			return CreateTxInput{}, rErr
		}
		recParams = &built
	}

	return CreateTxInput{
		TaskID:            uuid.New(),
		Title:             sanitize.Text(req.Title),
		Description:       desc,
		Type:              taskType,
		Priority:          priority,
		StartDate:         startTS,
		DueDate:           dueTS,
		CompletedAt:       completedTS,
		StoryPoints:       sp,
		OriginalEstimate:  oe,
		RemainingEstimate: re,
		CustomFields:      customBytes,
		ProjectID:         projectID,
		WorkflowID:        meta.WorkflowID,
		WorkspaceID:       meta.WorkspaceID,
		OrganizationID:    meta.OrganizationID,
		StatusID:          statusPtr,
		SprintID:          sprintID,
		ParentTaskRaw:     req.ParentTaskID,
		AllowEmailReplies: req.AllowEmailReplies != nil && *req.AllowEmailReplies,
		IsRecurring:       isRec,
		Recurrence:        recParams,
		AssigneeIDs:       assignees,
		ReporterIDs:       reporters,
		UserID:            userID,
	}, nil
}

func (s *service) Create(ctx context.Context, req CreateTaskRequest, userID uuid.UUID, isSuperAdmin bool) (*TaskResponse, error) {
	in, err := s.parseCreateInput(ctx, req, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}
	createdID, err := s.repo.CreateTaskTx(ctx, in)
	if err != nil {
		return nil, err
	}
	res, err := s.FindOne(ctx, createdID.String(), userID, true)
	if err != nil {
		return nil, err
	}
	if s.emitter != nil {
		s.emitter.EmitTaskCreated(res.ProjectID, res)
	}
	if s.activity != nil {
		_ = s.activity.LogTaskCreated(ctx, res.ID, res.Title, res.Slug, res, userID)
	}
	if s.notifier != nil && len(in.AssigneeIDs) > 0 {
		_ = s.notifier.NotifyTaskAssigned(ctx, createdID, in.AssigneeIDs, userID)
	}
	return res, nil
}

func (s *service) CreateWithAttachments(ctx context.Context, req CreateTaskRequest, files []UploadedFile, userID uuid.UUID, isSuperAdmin bool) (*TaskResponse, error) {
	in, err := s.parseCreateInput(ctx, req, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}
	createdID, err := s.repo.CreateTaskTx(ctx, in)
	if err != nil {
		return nil, err
	}

	if s.storage != nil && len(files) > 0 {
		folder := "uploads/tasks/" + createdID.String()
		for _, f := range files {
			ext := filepath.Ext(f.FileName)
			uniqueName := uuid.New().String() + ext
			saved, sErr := s.storage.SaveFile(folder, uniqueName, f.Content)
			if sErr != nil {
				return nil, sErr
			}
			var urlText pgtype.Text
			if saved.URL != nil {
				urlText = pgtype.Text{String: *saved.URL, Valid: true}
			}
			capped := saved.Size
			if capped < 0 {
				capped = 0
			} else if capped > 2147483647 {
				capped = 2147483647
			}
			sizeInt := int32(capped) // #nosec G115 -- explicitly clamped to [0, math.MaxInt32]
			if _, aErr := s.repo.CreateAttachment(ctx, CreatedAttachmentInput{
				ID:         uuid.New(),
				FileName:   f.FileName,
				FilePath:   pgtype.Text{String: filepath.Join(s.storage.UploadDir(), saved.Key), Valid: true},
				FileSize:   sizeInt,
				MimeType:   f.MimeType,
				URL:        urlText,
				StorageKey: pgtype.Text{String: saved.Key, Valid: true},
				TaskID:     createdID,
				UserID:     userID,
			}); aErr != nil {
				return nil, fmt.Errorf("create attachment record: %w", aErr)
			}
		}
	}

	res, err := s.FindOne(ctx, createdID.String(), userID, true)
	if err != nil {
		return nil, err
	}
	if s.emitter != nil {
		s.emitter.EmitTaskCreated(res.ProjectID, res)
	}
	if s.activity != nil {
		_ = s.activity.LogTaskCreated(ctx, res.ID, res.Title, res.Slug, res, userID)
	}
	if s.notifier != nil && len(in.AssigneeIDs) > 0 {
		_ = s.notifier.NotifyTaskAssigned(ctx, createdID, in.AssigneeIDs, userID)
	}
	return res, nil
}

func (s *service) FindOne(ctx context.Context, idOrSlug string, userID uuid.UUID, isSuperAdmin bool) (*TaskResponse, error) {
	accessRow, acc, err := s.requireTaskAccess(ctx, idOrSlug, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}

	full, err := s.repo.GetTaskFullByID(ctx, accessRow.ID)
	if err != nil {
		return nil, apperror.NotFound("Task not found")
	}

	resp := mapTaskRowBase(full)
	aMap, rMap, lMap := s.loadAllListRelations(ctx, []uuid.UUID{full.ID})
	if a := aMap[full.ID]; a != nil {
		resp.Assignees = a
	}
	if r := rMap[full.ID]; r != nil {
		resp.Reporters = r
	}
	if l := lMap[full.ID]; l != nil {
		resp.Labels = l
	}

	if atts, aErr := s.repo.ListAttachmentsByTask(ctx, full.ID); aErr == nil {
		resp.Attachments = make([]AttachmentBrief, 0, len(atts))
		for _, a := range atts {
			resp.Attachments = append(resp.Attachments, AttachmentBrief{
				ID: a.ID.String(), FileName: a.FileName, FileSize: a.FileSize,
				MimeType: a.MimeType, URL: pgTextPtr(a.Url), CreatedAt: formatTS(a.CreatedAt),
			})
		}
	}
	if watchers, wErr := s.repo.ListWatchersByTask(ctx, full.ID); wErr == nil {
		resp.Watchers = make([]WatcherBrief, 0, len(watchers))
		for _, w := range watchers {
			f, l := w.FirstName, w.LastName
			resp.Watchers = append(resp.Watchers, WatcherBrief{
				ID: w.ID.String(), TaskID: w.TaskID.String(), UserID: w.UserID.String(),
				User: &UserBrief{ID: w.UserID.String(), FirstName: &f, LastName: &l, Avatar: pgTextPtr(w.Avatar)},
			})
		}
	}
	if comments, cErr := s.repo.ListCommentsByTask(ctx, full.ID); cErr == nil {
		resp.Comments = make([]CommentBrief, 0, len(comments))
		for _, c := range comments {
			f, l := c.FirstName, c.LastName
			resp.Comments = append(resp.Comments, CommentBrief{
				ID: c.ID.String(), Content: c.Content, TaskID: c.TaskID.String(), AuthorID: c.AuthorID.String(),
				ParentCommentID: pgUUIDPtr(c.ParentCommentID), CreatedAt: formatTS(c.CreatedAt), UpdatedAt: formatTS(c.UpdatedAt),
				Author: &UserBrief{ID: c.AuthorID.String(), FirstName: &f, LastName: &l, Avatar: pgTextPtr(c.Avatar)},
			})
		}
	}
	if entries, teErr := s.repo.ListTimeEntriesByTask(ctx, full.ID); teErr == nil {
		resp.TimeEntries = make([]TimeEntryBrief, 0, len(entries))
		for _, te := range entries {
			f, l := te.FirstName, te.LastName
			resp.TimeEntries = append(resp.TimeEntries, TimeEntryBrief{
				ID: te.ID.String(), Description: pgTextPtr(te.Description), TimeSpent: te.TimeSpent,
				StartTime: pgTSPtr(te.StartTime), EndTime: pgTSPtr(te.EndTime), Date: formatTS(te.Date),
				TaskID: te.TaskID.String(), UserID: te.UserID.String(),
				User: &UserBrief{ID: te.UserID.String(), FirstName: &f, LastName: &l, Avatar: pgTextPtr(te.Avatar)},
			})
		}
	}
	if childRows, chErr := s.repo.ListChildTasks(ctx, full.ID); chErr == nil {
		childIDs := make([]uuid.UUID, 0, len(childRows))
		for _, c := range childRows {
			childIDs = append(childIDs, c.ID)
		}
		caMap, crMap, _ := s.loadAssigneesAndReporters(ctx, childIDs)
		resp.ChildTasks = make([]ChildTaskBrief, 0, len(childRows))
		for _, c := range childRows {
			assignees := caMap[c.ID]
			if assignees == nil {
				assignees = []UserBrief{}
			}
			reporters := crMap[c.ID]
			hasUserRel := (c.CreatedByID.Valid && uuid.UUID(c.CreatedByID.Bytes) == userID) || hasUser(assignees, userID) || hasUser(reporters, userID)
			if !acc.isElevated && !hasUserRel {
				continue
			}
			resp.ChildTasks = append(resp.ChildTasks, ChildTaskBrief{
				ID: c.ID.String(), Title: c.Title, Slug: c.Slug, Type: string(c.Type), Priority: string(c.Priority),
				Status:    &StatusBrief{Name: c.StatusName, Color: c.StatusColor, Category: string(c.StatusCategory)},
				Assignees: assignees,
				Reporters: reporters,
			})
		}
	}
	if rt, rErr := s.repo.GetRecurringConfig(ctx, full.ID); rErr == nil && rt != nil {
		resp.RecurringConfig = mapRecurrenceModel(rt)
	}
	if inbox, iErr := s.repo.GetProjectInbox(ctx, full.ProjectID); iErr == nil && inbox != nil {
		resp.ShowEmailReply = &ProjectInboxBrief{
			ID: inbox.ID.String(), ProjectID: inbox.ProjectID.String(), Enabled: inbox.Enabled, EmailAddress: pgTextPtr(inbox.EmailAddress),
		}
	}
	return &resp, nil
}

func hasUser(users []UserBrief, target uuid.UUID) bool {
	tid := target.String()
	for _, u := range users {
		if u.ID == tid {
			return true
		}
	}
	return false
}

func (s *service) FindByKey(ctx context.Context, key string, userID uuid.UUID, isSuperAdmin bool) (*TaskResponse, error) {
	id, err := s.repo.ResolveTaskIDBySlug(ctx, key)
	if err != nil {
		return nil, apperror.NotFound("Task not found")
	}
	return s.FindOne(ctx, id.String(), userID, isSuperAdmin)
}

func (s *service) Update(ctx context.Context, idOrSlug string, req UpdateTaskRequest, userID uuid.UUID, isSuperAdmin bool) (*TaskResponse, error) {
	taskRow, acc, err := s.requireTaskAccess(ctx, idOrSlug, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}
	if !acc.canChange {
		return nil, apperror.Forbidden("Insufficient permissions to update this task")
	}
	if mErr := validateRoleForMember(acc.role, req); mErr != nil {
		return nil, mErr
	}

	var effStart, effDue *time.Time
	if taskRow.StartDate.Valid {
		t := taskRow.StartDate.Time.UTC()
		effStart = &t
	}
	if req.StartDate != nil && *req.StartDate != "" {
		parsed, ok := parseISO8601(*req.StartDate)
		if !ok {
			return nil, apperror.BadRequest("startDate must be a valid ISO 8601 date string")
		}
		effStart = &parsed
	}
	if taskRow.DueDate.Valid {
		t := taskRow.DueDate.Time.UTC()
		effDue = &t
	}
	if req.DueDate != nil && *req.DueDate != "" {
		parsed, ok := parseISO8601(*req.DueDate)
		if !ok {
			return nil, apperror.BadRequest("dueDate must be a valid ISO 8601 date string")
		}
		effDue = &parsed
	}
	if effStart != nil && effDue != nil && effStart.After(*effDue) {
		return nil, apperror.BadRequest("Start date must be before the due date")
	}

	params := sqlc.UpdateTaskCoreByIDParams{
		ID:          taskRow.ID,
		UpdatedByID: pgtype.UUID{Bytes: userID, Valid: true},
	}

	if req.Title != nil {
		params.Title = pgtype.Text{String: sanitize.Text(*req.Title), Valid: true}
	}
	if req.Description != nil {
		params.Description = pgtype.Text{String: sanitize.HTML(*req.Description), Valid: true}
	}
	if req.Type != nil {
		params.Type = sqlc.NullTaskType{TaskType: sqlc.TaskType(*req.Type), Valid: true}
	}
	if req.Priority != nil {
		params.Priority = sqlc.NullTaskPriority{TaskPriority: sqlc.TaskPriority(*req.Priority), Valid: true}
	}
	if req.StartDate != nil && *req.StartDate != "" {
		params.StartDate = pgtype.Timestamp{Time: *effStart, Valid: true}
	}
	if req.DueDate != nil && *req.DueDate != "" {
		params.DueDate = pgtype.Timestamp{Time: *effDue, Valid: true}
	}
	if req.StoryPoints != nil {
		params.StoryPoints = pgtype.Int4{Int32: *req.StoryPoints, Valid: true}
	}
	if req.OriginalEstimate != nil {
		params.OriginalEstimate = pgtype.Int4{Int32: *req.OriginalEstimate, Valid: true}
	}
	if req.RemainingEstimate != nil {
		params.RemainingEstimate = pgtype.Int4{Int32: *req.RemainingEstimate, Valid: true}
	}
	if req.CustomFields != nil {
		if sanitized, ok := sanitize.Object(req.CustomFields).(map[string]any); ok {
			params.CustomFields = marshalJSONField(sanitized)
		}
	}
	if req.AllowEmailReplies != nil {
		params.AllowEmailReplies = pgtype.Bool{Bool: *req.AllowEmailReplies, Valid: true}
	}

	var oldStatusName, newStatusName string
	if req.StatusID != nil {
		if oldSt, oldErr := s.repo.GetStatusByID(ctx, taskRow.StatusID); oldErr == nil {
			oldStatusName = oldSt.Name
		}
		sid := uuid.MustParse(*req.StatusID)
		st, stErr := s.repo.GetStatusByID(ctx, sid)
		if stErr != nil {
			return nil, apperror.NotFound("Task status not found")
		}
		newStatusName = st.Name
		params.StatusID = pgtype.UUID{Bytes: sid, Valid: true}
		if st.Category == sqlc.StatusCategoryDONE {
			params.CompletedAt = pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
		} else {
			params.ClearCompletedAt = true
		}
	} else if req.ClearCompletedAt {
		params.ClearCompletedAt = true
	} else if req.CompletedAt != nil && *req.CompletedAt != "" {
		parsed, ok := parseISO8601(*req.CompletedAt)
		if !ok {
			return nil, apperror.BadRequest("completedAt must be a valid ISO 8601 date string")
		}
		params.CompletedAt = pgtype.Timestamp{Time: parsed, Valid: true}
	}

	if req.ClearSprintID {
		params.ClearSprintID = true
	} else if req.SprintID != nil && *req.SprintID != "" {
		sid, sErr := uuid.Parse(*req.SprintID)
		if sErr != nil {
			return nil, apperror.BadRequest("sprintId must be a UUID")
		}
		params.SprintID = pgtype.UUID{Bytes: sid, Valid: true}
	}

	if req.ClearParentTaskID {
		params.ClearParentTaskID = true
	} else if req.ParentTaskID != nil && *req.ParentTaskID != "" {
		if pid, pErr := uuid.Parse(*req.ParentTaskID); pErr == nil {
			params.ParentTaskID = pgtype.UUID{Bytes: pid, Valid: true}
		} else {
			resolvedID, rErr := s.repo.ResolveTaskIDBySlug(ctx, *req.ParentTaskID)
			if rErr != nil {
				return nil, apperror.NotFound(fmt.Sprintf("Parent task with slug %s not found", *req.ParentTaskID))
			}
			params.ParentTaskID = pgtype.UUID{Bytes: resolvedID, Valid: true}
		}
	}

	var assigneeUUIDs *[]uuid.UUID
	if req.AssigneeIDs != nil {
		ids, vErr := s.validateUserIDs(ctx, *req.AssigneeIDs)
		if vErr != nil {
			return nil, vErr
		}
		assigneeUUIDs = &ids
	}
	var reporterUUIDs *[]uuid.UUID
	if req.ReporterIDs != nil {
		ids, vErr := s.validateUserIDs(ctx, *req.ReporterIDs)
		if vErr != nil {
			return nil, vErr
		}
		reporterUUIDs = &ids
	}

	if uErr := s.repo.UpdateTaskTx(ctx, params, assigneeUUIDs, reporterUUIDs); uErr != nil {
		return nil, fmt.Errorf("update task: %w", uErr)
	}
	res, err := s.FindOne(ctx, taskRow.ID.String(), userID, true)
	if err != nil {
		return nil, err
	}
	if s.emitter != nil {
		s.emitter.EmitTaskUpdated(res.ProjectID, res.ID, res)
		if ext, ok := s.emitter.(ExtendedEventEmitter); ok {
			if req.StatusID != nil && newStatusName != "" && newStatusName != oldStatusName {
				ext.EmitTaskStatusChanged(res.ProjectID, res.ID, map[string]any{
					"oldStatus": oldStatusName,
					"newStatus": newStatusName,
					"statusId":  *req.StatusID,
				})
			}
			if assigneeUUIDs != nil && len(*assigneeUUIDs) > 0 {
				ext.EmitTaskAssigned(res.ProjectID, res.ID, map[string]any{
					"assigneeId": (*assigneeUUIDs)[0].String(),
				})
			}
		}
	}
	if s.activity != nil {
		switch {
		case req.StatusID != nil && newStatusName != "":
			_ = s.activity.LogTaskStatusChanged(ctx, res.ID, res.Title, oldStatusName, newStatusName, userID)
		case assigneeUUIDs != nil && len(*assigneeUUIDs) > 0:
			_ = s.activity.LogTaskAssigned(ctx, res.ID, res.Title, (*assigneeUUIDs)[0].String(), userID)
		default:
			_ = s.activity.LogTaskUpdated(ctx, res.ID, res.Title, nil, res, userID)
		}
	}
	if s.notifier != nil {
		if req.StatusID != nil && newStatusName != "" && newStatusName != oldStatusName {
			_ = s.notifier.NotifyTaskStatusChanged(ctx, taskRow.ID, oldStatusName, newStatusName, userID)
		}
		if assigneeUUIDs != nil && len(*assigneeUUIDs) > 0 {
			_ = s.notifier.NotifyTaskAssigned(ctx, taskRow.ID, *assigneeUUIDs, userID)
		}
	}
	return res, nil
}

func (s *service) validateUserIDs(ctx context.Context, rawIDs []string) ([]uuid.UUID, error) {
	if len(rawIDs) == 0 {
		return []uuid.UUID{}, nil
	}
	ids := make([]uuid.UUID, 0, len(rawIDs))
	for _, raw := range rawIDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			return nil, apperror.BadRequest("Invalid user UUID: " + raw)
		}
		ids = append(ids, id)
	}
	found, err := s.repo.CheckUsersExist(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("verify user ids: %w", err)
	}
	foundSet := make(map[uuid.UUID]struct{}, len(found))
	for _, id := range found {
		foundSet[id] = struct{}{}
	}
	var missing []string
	for _, id := range ids {
		if _, ok := foundSet[id]; !ok {
			missing = append(missing, id.String())
		}
	}
	if len(missing) > 0 {
		return nil, apperror.NotFound("Users not found: " + strings.Join(missing, ", "))
	}
	return ids, nil
}

func (s *service) Remove(ctx context.Context, idOrSlug string, userID uuid.UUID, isSuperAdmin bool) (*TaskResponse, error) {
	taskRow, acc, err := s.requireTaskAccess(ctx, idOrSlug, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}
	if !acc.isElevated {
		return nil, apperror.Forbidden("Only managers and owners can delete tasks")
	}

	existing, err := s.FindOne(ctx, taskRow.ID.String(), userID, true)
	if err != nil {
		return nil, err
	}
	affected, err := s.repo.DeleteTaskByID(ctx, taskRow.ID)
	if err != nil {
		return nil, fmt.Errorf("delete task: %w", err)
	}
	if affected == 0 {
		return nil, apperror.NotFound("Task not found")
	}
	if s.emitter != nil {
		s.emitter.EmitTaskDeleted(existing.ProjectID, existing.ID)
	}
	if s.activity != nil {
		_ = s.activity.LogTaskDeleted(ctx, existing.ID, existing.Title, existing, userID)
	}
	return existing, nil
}

func (s *service) AddComment(ctx context.Context, idOrSlug string, comment string, userID uuid.UUID, isSuperAdmin bool) (*CommentBrief, error) {
	taskRow, _, err := s.requireTaskAccess(ctx, idOrSlug, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}
	return s.repo.CreateQuickComment(ctx, taskRow.ID, userID, sanitize.HTML(comment))
}
