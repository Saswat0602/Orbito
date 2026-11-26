package tasks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/modules/task_ranks"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// AccessContext captures resolved organization or project access for listing queries.
type AccessContext struct {
	IsSuperAdmin bool
	IsElevated   bool
	Role         middleware.Role
}

// CreateTxInput holds validated task creation inputs executed inside a transaction.
type CreateTxInput struct {
	TaskID            uuid.UUID
	Title             string
	Description       pgtype.Text
	Type              sqlc.TaskType
	Priority          sqlc.TaskPriority
	StartDate         pgtype.Timestamp
	DueDate           pgtype.Timestamp
	CompletedAt       pgtype.Timestamp
	StoryPoints       pgtype.Int4
	OriginalEstimate  pgtype.Int4
	RemainingEstimate pgtype.Int4
	CustomFields      []byte
	ProjectID         uuid.UUID
	WorkflowID        uuid.UUID
	WorkspaceID       uuid.UUID
	OrganizationID    uuid.UUID
	StatusID          *uuid.UUID
	SprintID          pgtype.UUID
	ParentTaskRaw     *string
	AllowEmailReplies bool
	IsRecurring       bool
	Recurrence        *sqlc.CreateRecurringTaskConfigParams
	AssigneeIDs       []uuid.UUID
	ReporterIDs       []uuid.UUID
	UserID            uuid.UUID
}

// CreatedAttachmentInput describes an uploaded attachment row to persist with a task.
type CreatedAttachmentInput struct {
	ID         uuid.UUID
	FileName   string
	FilePath   pgtype.Text
	FileSize   int32
	MimeType   string
	URL        pgtype.Text
	StorageKey pgtype.Text
	TaskID     uuid.UUID
	UserID     uuid.UUID
}

// ProjectWorkflowInfo holds project metadata + statuses for by-status board queries.
type ProjectWorkflowInfo struct {
	ProjectID   uuid.UUID
	ProjectName string
	ProjectSlug string
	WorkspaceID uuid.UUID
	OrgID       uuid.UUID
	OrgOwnerID  uuid.UUID
	OrgArchived bool
	Statuses    []sqlc.TaskStatus
}

// Repository defines data access operations for tasks and task shares.
type Repository interface {
	GetProjectMeta(ctx context.Context, projectID uuid.UUID) (*sqlc.GetProjectMetaForTaskCreateRow, error)
	GetProjectAccessForCreate(ctx context.Context, projectID, userID uuid.UUID, isSuperAdmin bool) (canChange bool, err error)
	GetDefaultSprintID(ctx context.Context, projectID uuid.UUID) (pgtype.UUID, error)
	CreateTaskTx(ctx context.Context, in CreateTxInput) (uuid.UUID, error)
	BulkCreateTasksTx(ctx context.Context, projectID, statusID uuid.UUID, sprintID pgtype.UUID, items []CreateTxInput) ([]uuid.UUID, error)
	CreateAttachment(ctx context.Context, in CreatedAttachmentInput) (*sqlc.CreateTaskAttachmentForTaskRow, error)
	GetTaskAccessDetails(ctx context.Context, idOrSlug string, userID uuid.UUID) (*sqlc.GetTaskAccessDetailsRow, error)
	GetTaskFullByID(ctx context.Context, id uuid.UUID) (*sqlc.GetTaskFullByIDRow, error)
	ListAssigneesBatch(ctx context.Context, taskIDs []uuid.UUID) ([]sqlc.ListTaskAssigneesBatchRow, error)
	ListReportersBatch(ctx context.Context, taskIDs []uuid.UUID) ([]sqlc.ListTaskReportersBatchRow, error)
	ListLabelsBatch(ctx context.Context, taskIDs []uuid.UUID) ([]sqlc.ListTaskLabelsBatchRow, error)
	ListAttachmentsByTask(ctx context.Context, taskID uuid.UUID) ([]sqlc.ListTaskAttachmentsByTaskIDRow, error)
	ListWatchersByTask(ctx context.Context, taskID uuid.UUID) ([]sqlc.ListTaskWatchersByTaskIDRow, error)
	ListCommentsByTask(ctx context.Context, taskID uuid.UUID) ([]sqlc.ListTaskCommentsByTaskIDRow, error)
	ListTimeEntriesByTask(ctx context.Context, taskID uuid.UUID) ([]sqlc.ListTaskTimeEntriesByTaskIDRow, error)
	ListChildTasks(ctx context.Context, parentID uuid.UUID) ([]sqlc.ListTaskChildrenForDetailRow, error)
	GetProjectInbox(ctx context.Context, projectID uuid.UUID) (*sqlc.GetProjectInboxForTaskDetailRow, error)
	GetStatusByID(ctx context.Context, id uuid.UUID) (*sqlc.GetTaskStatusBasicByIDRow, error)
	ResolveTaskIDBySlug(ctx context.Context, slug string) (uuid.UUID, error)
	CheckUsersExist(ctx context.Context, userIDs []uuid.UUID) ([]uuid.UUID, error)
	UpdateTaskTx(ctx context.Context, params sqlc.UpdateTaskCoreByIDParams, assigneeIDs, reporterIDs *[]uuid.UUID) error
	DeleteTaskByID(ctx context.Context, id uuid.UUID) (int64, error)
	BulkDeleteByIDs(ctx context.Context, ids []uuid.UUID) (int64, error)
	CreateQuickComment(ctx context.Context, taskID, authorID uuid.UUID, content string) (*CommentBrief, error)
	GetRecurringConfig(ctx context.Context, taskID uuid.UUID) (*sqlc.RecurringTask, error)
	CreateRecurrenceConfig(ctx context.Context, params sqlc.CreateRecurringTaskConfigParams) (*sqlc.RecurringTask, error)
	UpdateRecurrenceConfig(ctx context.Context, params sqlc.UpdateRecurringTaskConfigByIDParams) (*sqlc.RecurringTask, error)
	IncrementRecurrenceOccurrence(ctx context.Context, recID uuid.UUID, nextOcc int32, nextTime time.Time) error
	DeleteRecurrenceConfig(ctx context.Context, recID uuid.UUID) error
	SetTaskIsRecurring(ctx context.Context, taskID uuid.UUID, isRecurring bool) (*sqlc.Task, error)
	ListRecurringByProject(ctx context.Context, projectID uuid.UUID) ([]sqlc.ListRecurringTasksByProjectIDRow, error)
	CreateShareLink(ctx context.Context, params sqlc.CreatePublicTaskShareParams) (*sqlc.PublicTaskShare, error)
	ListActiveShares(ctx context.Context, taskID uuid.UUID) ([]sqlc.PublicTaskShare, error)
	GetShareWithAccess(ctx context.Context, shareID, userID uuid.UUID) (*sqlc.GetTaskShareWithAccessRow, error)
	RevokeShare(ctx context.Context, shareID uuid.UUID) error

	// Dynamic listing, grouped, status-board, and bulk query helpers.
	ResolveScopeContext(ctx context.Context, orgRaw string, projectRaw, workspaceRaw []string, userID uuid.UUID) (orgID uuid.UUID, projectIDs, workspaceIDs []uuid.UUID, err error)
	GetOrgAccess(ctx context.Context, orgID, userID uuid.UUID, isSuperAdmin bool) (AccessContext, error)
	FindTasksDynamic(ctx context.Context, orgID uuid.UUID, projectIDs, workspaceIDs []uuid.UUID, acc AccessContext, userID uuid.UUID, q TaskFilterQuery, excludeDoneWhenNoStatus bool) ([]uuid.UUID, int64, FilterCountsResponse, error)
	GetProjectWorkflowInfoBySlug(ctx context.Context, slug string) (*ProjectWorkflowInfo, error)
	GetProjectAccessByID(ctx context.Context, projectID, userID uuid.UUID, isSuperAdmin bool) (AccessContext, error)
	QueryColumnTasks(ctx context.Context, projectID, statusID uuid.UUID, sprintID *uuid.UUID, includeSubtasks, restrictedToUser bool, userID uuid.UUID, limit, offset int) ([]uuid.UUID, int64, error)
	FindTodayTaskIDs(ctx context.Context, orgID uuid.UUID, acc AccessContext, userID uuid.UUID, dayStart, dayEnd time.Time, limit, offset int) ([]uuid.UUID, int64, error)
	FindOrgTaskIDs(ctx context.Context, orgID uuid.UUID, assigneeID uuid.UUID, priority, search string, limit, offset int) ([]uuid.UUID, int64, error)
	QueryGroupedTasks(ctx context.Context, orgID uuid.UUID, projectIDs, workspaceIDs []uuid.UUID, acc AccessContext, userID uuid.UUID, q GroupedTasksQuery) ([]TaskGroupBucket, error)
	ResolveBulkTaskIDs(ctx context.Context, all bool, explicitIDs, excludedIDs []uuid.UUID, projectID, workspaceID *uuid.UUID, sprintID *uuid.UUID, search string, statuses, priorities, types, assignees, reporters []string) ([]uuid.UUID, error)
	FindStatusInWorkflowByIDOrName(ctx context.Context, projectID uuid.UUID, statusID *uuid.UUID) (*sqlc.GetTaskStatusBasicByIDRow, error)
	BulkReplaceAssignees(ctx context.Context, taskIDs []uuid.UUID, assigneeIDs []uuid.UUID) error
}

type pgRepository struct {
	db      *database.DB
	queries *sqlc.Queries
}

// NewRepository constructs a PostgreSQL tasks repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{db: db, queries: db.Queries()}
}

func (r *pgRepository) GetProjectMeta(ctx context.Context, projectID uuid.UUID) (*sqlc.GetProjectMetaForTaskCreateRow, error) {
	row, err := r.queries.GetProjectMetaForTaskCreate(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetProjectAccessForCreate(ctx context.Context, projectID, userID uuid.UUID, isSuperAdmin bool) (bool, error) {
	if isSuperAdmin {
		return true, nil
	}
	acc, err := r.GetProjectAccessByID(ctx, projectID, userID, isSuperAdmin)
	if err != nil {
		return false, err
	}
	return acc.IsElevated || acc.Role == middleware.RoleMember, nil
}

func (r *pgRepository) GetDefaultSprintID(ctx context.Context, projectID uuid.UUID) (pgtype.UUID, error) {
	id, err := r.queries.GetDefaultSprintForProject(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgtype.UUID{}, nil
		}
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: id, Valid: true}, nil
}

func (r *pgRepository) CreateTaskTx(ctx context.Context, in CreateTxInput) (uuid.UUID, error) {
	var createdID uuid.UUID
	err := r.db.WithTx(ctx, func(q *sqlc.Queries) error {
		lockInfo, err := q.LockProjectRowForTaskNumber(ctx, in.ProjectID)
		if err != nil {
			return apperror.NotFound("Project not found")
		}
		prefix := lockInfo.Slug
		if lockInfo.TaskPrefix.Valid && lockInfo.TaskPrefix.String != "" {
			prefix = lockInfo.TaskPrefix.String
		}
		maxNum, err := q.GetMaxTaskNumberForProject(ctx, in.ProjectID)
		if err != nil {
			return fmt.Errorf("max task number: %w", err)
		}
		taskNum := maxNum + 1
		slug := fmt.Sprintf("%s-%d", prefix, taskNum)

		var statusID uuid.UUID
		if in.StatusID != nil {
			statusID = *in.StatusID
		} else {
			statusID, err = q.GetFirstStatusInWorkflow(ctx, in.WorkflowID)
			if err != nil {
				statusID, err = q.GetAnyFirstStatus(ctx)
				if err != nil {
					return apperror.BadRequest("No task status available for project workflow")
				}
			}
		}

		var parentUUID pgtype.UUID
		if in.ParentTaskRaw != nil && *in.ParentTaskRaw != "" {
			if pid, pErr := uuid.Parse(*in.ParentTaskRaw); pErr == nil {
				parentUUID = pgtype.UUID{Bytes: pid, Valid: true}
			} else {
				resolvedID, rErr := q.ResolveTaskIDBySlug(ctx, *in.ParentTaskRaw)
				if rErr != nil {
					return apperror.NotFound(fmt.Sprintf("Parent task with slug %s not found", *in.ParentTaskRaw))
				}
				parentUUID = pgtype.UUID{Bytes: resolvedID, Valid: true}
			}
		}

		t, err := q.CreateTaskRecord(ctx, sqlc.CreateTaskRecordParams{
			ID:                in.TaskID,
			Title:             in.Title,
			Description:       in.Description,
			Type:              in.Type,
			Priority:          in.Priority,
			TaskNumber:        taskNum,
			Slug:              slug,
			StartDate:         in.StartDate,
			DueDate:           in.DueDate,
			CompletedAt:       in.CompletedAt,
			StoryPoints:       in.StoryPoints,
			OriginalEstimate:  in.OriginalEstimate,
			RemainingEstimate: in.RemainingEstimate,
			CustomFields:      in.CustomFields,
			ProjectID:         in.ProjectID,
			StatusID:          statusID,
			SprintID:          in.SprintID,
			ParentTaskID:      parentUUID,
			CreatedByID:       pgtype.UUID{Bytes: in.UserID, Valid: true},
			UpdatedByID:       pgtype.UUID{Bytes: in.UserID, Valid: true},
			AllowEmailReplies: in.AllowEmailReplies,
			IsRecurring:       in.IsRecurring,
		})
		if err != nil {
			return fmt.Errorf("insert task record: %w", err)
		}
		createdID = t.ID

		for _, uid := range in.AssigneeIDs {
			if err := q.InsertTaskAssignee(ctx, sqlc.InsertTaskAssigneeParams{TaskID: t.ID, UserID: uid}); err != nil {
				return err
			}
		}
		for _, uid := range in.ReporterIDs {
			if err := q.InsertTaskReporter(ctx, sqlc.InsertTaskReporterParams{TaskID: t.ID, UserID: uid}); err != nil {
				return err
			}
		}

		if err := task_ranks.SeedForTasksBatchWithQueries(ctx, q, []uuid.UUID{t.ID}, in.ProjectID, in.WorkspaceID, in.OrganizationID); err != nil {
			return err
		}

		if in.IsRecurring && in.Recurrence != nil {
			in.Recurrence.TaskID = t.ID
			if _, err := q.CreateRecurringTaskConfig(ctx, *in.Recurrence); err != nil {
				return fmt.Errorf("insert recurring task config: %w", err)
			}
		}
		return nil
	})
	return createdID, err
}

func (r *pgRepository) BulkCreateTasksTx(ctx context.Context, projectID, statusID uuid.UUID, sprintID pgtype.UUID, items []CreateTxInput) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(items))
	if len(items) == 0 {
		return ids, nil
	}
	err := r.db.WithTx(ctx, func(q *sqlc.Queries) error {
		lockInfo, err := q.LockProjectRowForTaskNumber(ctx, projectID)
		if err != nil {
			return apperror.NotFound("Project not found")
		}
		prefix := lockInfo.Slug
		if lockInfo.TaskPrefix.Valid && lockInfo.TaskPrefix.String != "" {
			prefix = lockInfo.TaskPrefix.String
		}
		maxNum, err := q.GetMaxTaskNumberForProject(ctx, projectID)
		if err != nil {
			return err
		}

		for i, item := range items {
			delta := i + 1
			if delta > 1000000 {
				delta = 1000000
			}
			taskNum := maxNum + int32(delta) // #nosec G115 -- bounded bulk batch size
			slug := fmt.Sprintf("%s-%d", prefix, taskNum)
			t, err := q.CreateTaskRecord(ctx, sqlc.CreateTaskRecordParams{
				ID:          item.TaskID,
				Title:       item.Title,
				Description: item.Description,
				Type:        item.Type,
				Priority:    item.Priority,
				TaskNumber:  taskNum,
				Slug:        slug,
				DueDate:     item.DueDate,
				ProjectID:   projectID,
				StatusID:    statusID,
				SprintID:    sprintID,
				CreatedByID: pgtype.UUID{Bytes: item.UserID, Valid: true},
				UpdatedByID: pgtype.UUID{Bytes: item.UserID, Valid: true},
			})
			if err != nil {
				return err
			}
			ids = append(ids, t.ID)
		}
		return task_ranks.SeedForTasksBatchWithQueries(ctx, q, ids, projectID, items[0].WorkspaceID, items[0].OrganizationID)
	})
	return ids, err
}

func (r *pgRepository) CreateAttachment(ctx context.Context, in CreatedAttachmentInput) (*sqlc.CreateTaskAttachmentForTaskRow, error) {
	row, err := r.queries.CreateTaskAttachmentForTask(ctx, sqlc.CreateTaskAttachmentForTaskParams{
		ID:          in.ID,
		FileName:    in.FileName,
		FilePath:    in.FilePath,
		FileSize:    in.FileSize,
		MimeType:    in.MimeType,
		Url:         in.URL,
		StorageKey:  in.StorageKey,
		TaskID:      in.TaskID,
		CreatedByID: pgtype.UUID{Bytes: in.UserID, Valid: true},
		UpdatedByID: pgtype.UUID{Bytes: in.UserID, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetTaskAccessDetails(ctx context.Context, idOrSlug string, userID uuid.UUID) (*sqlc.GetTaskAccessDetailsRow, error) {
	params := sqlc.GetTaskAccessDetailsParams{UserID: userID}
	if parsed, err := uuid.Parse(idOrSlug); err == nil {
		params.IsUuid = true
		params.TaskUuid = pgtype.UUID{Bytes: parsed, Valid: true}
	} else {
		params.IsUuid = false
		params.TaskSlug = idOrSlug
	}
	row, err := r.queries.GetTaskAccessDetails(ctx, params)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetTaskFullByID(ctx context.Context, id uuid.UUID) (*sqlc.GetTaskFullByIDRow, error) {
	row, err := r.queries.GetTaskFullByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) ListAssigneesBatch(ctx context.Context, taskIDs []uuid.UUID) ([]sqlc.ListTaskAssigneesBatchRow, error) {
	if len(taskIDs) == 0 {
		return nil, nil
	}
	return r.queries.ListTaskAssigneesBatch(ctx, taskIDs)
}

func (r *pgRepository) ListReportersBatch(ctx context.Context, taskIDs []uuid.UUID) ([]sqlc.ListTaskReportersBatchRow, error) {
	if len(taskIDs) == 0 {
		return nil, nil
	}
	return r.queries.ListTaskReportersBatch(ctx, taskIDs)
}

func (r *pgRepository) ListLabelsBatch(ctx context.Context, taskIDs []uuid.UUID) ([]sqlc.ListTaskLabelsBatchRow, error) {
	if len(taskIDs) == 0 {
		return nil, nil
	}
	return r.queries.ListTaskLabelsBatch(ctx, taskIDs)
}

func (r *pgRepository) ListAttachmentsByTask(ctx context.Context, taskID uuid.UUID) ([]sqlc.ListTaskAttachmentsByTaskIDRow, error) {
	return r.queries.ListTaskAttachmentsByTaskID(ctx, taskID)
}

func (r *pgRepository) ListWatchersByTask(ctx context.Context, taskID uuid.UUID) ([]sqlc.ListTaskWatchersByTaskIDRow, error) {
	return r.queries.ListTaskWatchersByTaskID(ctx, taskID)
}

func (r *pgRepository) ListCommentsByTask(ctx context.Context, taskID uuid.UUID) ([]sqlc.ListTaskCommentsByTaskIDRow, error) {
	return r.queries.ListTaskCommentsByTaskID(ctx, taskID)
}

func (r *pgRepository) ListTimeEntriesByTask(ctx context.Context, taskID uuid.UUID) ([]sqlc.ListTaskTimeEntriesByTaskIDRow, error) {
	return r.queries.ListTaskTimeEntriesByTaskID(ctx, taskID)
}

func (r *pgRepository) ListChildTasks(ctx context.Context, parentID uuid.UUID) ([]sqlc.ListTaskChildrenForDetailRow, error) {
	return r.queries.ListTaskChildrenForDetail(ctx, pgtype.UUID{Bytes: parentID, Valid: true})
}

func (r *pgRepository) GetProjectInbox(ctx context.Context, projectID uuid.UUID) (*sqlc.GetProjectInboxForTaskDetailRow, error) {
	row, err := r.queries.GetProjectInboxForTaskDetail(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetStatusByID(ctx context.Context, id uuid.UUID) (*sqlc.GetTaskStatusBasicByIDRow, error) {
	row, err := r.queries.GetTaskStatusBasicByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) ResolveTaskIDBySlug(ctx context.Context, slug string) (uuid.UUID, error) {
	return r.queries.ResolveTaskIDBySlug(ctx, slug)
}

func (r *pgRepository) CheckUsersExist(ctx context.Context, userIDs []uuid.UUID) ([]uuid.UUID, error) {
	return r.queries.CheckUsersExistByIDs(ctx, userIDs)
}

func (r *pgRepository) UpdateTaskTx(ctx context.Context, params sqlc.UpdateTaskCoreByIDParams, assigneeIDs, reporterIDs *[]uuid.UUID) error {
	return r.db.WithTx(ctx, func(q *sqlc.Queries) error {
		if _, err := q.UpdateTaskCoreByID(ctx, params); err != nil {
			return err
		}
		if assigneeIDs != nil {
			if err := q.DeleteTaskAssigneesByTaskID(ctx, params.ID); err != nil {
				return err
			}
			for _, uid := range *assigneeIDs {
				if err := q.InsertTaskAssignee(ctx, sqlc.InsertTaskAssigneeParams{TaskID: params.ID, UserID: uid}); err != nil {
					return err
				}
			}
		}
		if reporterIDs != nil {
			if err := q.DeleteTaskReportersByTaskID(ctx, params.ID); err != nil {
				return err
			}
			for _, uid := range *reporterIDs {
				if err := q.InsertTaskReporter(ctx, sqlc.InsertTaskReporterParams{TaskID: params.ID, UserID: uid}); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (r *pgRepository) DeleteTaskByID(ctx context.Context, id uuid.UUID) (int64, error) {
	return r.queries.DeleteTaskByID(ctx, id)
}

func (r *pgRepository) BulkDeleteByIDs(ctx context.Context, ids []uuid.UUID) (int64, error) {
	return r.queries.BulkDeleteTasksByIDs(ctx, ids)
}

func (r *pgRepository) CreateQuickComment(ctx context.Context, taskID, authorID uuid.UUID, content string) (*CommentBrief, error) {
	c, err := r.queries.CreateQuickTaskComment(ctx, sqlc.CreateQuickTaskCommentParams{
		ID:       uuid.New(),
		Content:  content,
		TaskID:   taskID,
		AuthorID: authorID,
	})
	if err != nil {
		return nil, err
	}
	u, err := r.queries.GetCommentAuthorBrief(ctx, authorID)
	var author *UserBrief
	if err == nil {
		f, l := u.FirstName, u.LastName
		author = &UserBrief{ID: u.ID.String(), FirstName: &f, LastName: &l, Avatar: pgTextPtr(u.Avatar)}
	}
	return &CommentBrief{
		ID:              c.ID.String(),
		Content:         c.Content,
		TaskID:          c.TaskID.String(),
		AuthorID:        c.AuthorID.String(),
		ParentCommentID: pgUUIDPtr(c.ParentCommentID),
		CreatedAt:       formatTS(c.CreatedAt),
		UpdatedAt:       formatTS(c.UpdatedAt),
		Author:          author,
	}, nil
}

func (r *pgRepository) GetRecurringConfig(ctx context.Context, taskID uuid.UUID) (*sqlc.RecurringTask, error) {
	row, err := r.queries.GetRecurringTaskConfigByTaskID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) CreateRecurrenceConfig(ctx context.Context, params sqlc.CreateRecurringTaskConfigParams) (*sqlc.RecurringTask, error) {
	row, err := r.queries.CreateRecurringTaskConfig(ctx, params)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) UpdateRecurrenceConfig(ctx context.Context, params sqlc.UpdateRecurringTaskConfigByIDParams) (*sqlc.RecurringTask, error) {
	row, err := r.queries.UpdateRecurringTaskConfigByID(ctx, params)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) IncrementRecurrenceOccurrence(ctx context.Context, recID uuid.UUID, nextOcc int32, nextTime time.Time) error {
	return r.queries.IncrementRecurringOccurrence(ctx, sqlc.IncrementRecurringOccurrenceParams{
		ID:                recID,
		CurrentOccurrence: nextOcc,
		NextOccurrence:    pgtype.Timestamp{Time: nextTime, Valid: true},
	})
}

func (r *pgRepository) DeleteRecurrenceConfig(ctx context.Context, recID uuid.UUID) error {
	return r.queries.DeleteRecurringConfigByID(ctx, recID)
}

func (r *pgRepository) SetTaskIsRecurring(ctx context.Context, taskID uuid.UUID, isRecurring bool) (*sqlc.Task, error) {
	t, err := r.queries.SetTaskIsRecurring(ctx, sqlc.SetTaskIsRecurringParams{ID: taskID, IsRecurring: isRecurring})
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *pgRepository) ListRecurringByProject(ctx context.Context, projectID uuid.UUID) ([]sqlc.ListRecurringTasksByProjectIDRow, error) {
	return r.queries.ListRecurringTasksByProjectID(ctx, projectID)
}

func (r *pgRepository) CreateShareLink(ctx context.Context, params sqlc.CreatePublicTaskShareParams) (*sqlc.PublicTaskShare, error) {
	row, err := r.queries.CreatePublicTaskShare(ctx, params)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) ListActiveShares(ctx context.Context, taskID uuid.UUID) ([]sqlc.PublicTaskShare, error) {
	return r.queries.ListActiveSharesForTask(ctx, taskID)
}

func (r *pgRepository) GetShareWithAccess(ctx context.Context, shareID, userID uuid.UUID) (*sqlc.GetTaskShareWithAccessRow, error) {
	row, err := r.queries.GetTaskShareWithAccess(ctx, sqlc.GetTaskShareWithAccessParams{ID: shareID, UserID: userID})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) RevokeShare(ctx context.Context, shareID uuid.UUID) error {
	return r.queries.RevokePublicTaskShare(ctx, shareID)
}

// argBuilder accumulates positional $N parameters cleanly for dynamic pgx queries.
type argBuilder struct {
	args []any
}

func (b *argBuilder) add(val any) string {
	b.args = append(b.args, val)
	return fmt.Sprintf("$%d", len(b.args))
}

func (r *pgRepository) ResolveScopeContext(ctx context.Context, orgRaw string, projectRaw, workspaceRaw []string, userID uuid.UUID) (uuid.UUID, []uuid.UUID, []uuid.UUID, error) {
	pool := r.db.Pool()
	var projectIDs []uuid.UUID
	for _, raw := range projectRaw {
		if pid, err := uuid.Parse(raw); err == nil {
			projectIDs = append(projectIDs, pid)
		} else if raw != "" {
			var id uuid.UUID
			if err := pool.QueryRow(ctx, `SELECT id FROM "projects" WHERE slug = $1 LIMIT 1`, raw).Scan(&id); err == nil {
				projectIDs = append(projectIDs, id)
			}
		}
	}

	var workspaceIDs []uuid.UUID
	for _, raw := range workspaceRaw {
		if wid, err := uuid.Parse(raw); err == nil {
			workspaceIDs = append(workspaceIDs, wid)
		} else if raw != "" {
			var id uuid.UUID
			if err := pool.QueryRow(ctx, `SELECT id FROM "workspaces" WHERE slug = $1 LIMIT 1`, raw).Scan(&id); err == nil {
				workspaceIDs = append(workspaceIDs, id)
			}
		}
	}

	var orgID uuid.UUID
	if orgRaw != "" {
		if id, err := uuid.Parse(orgRaw); err == nil {
			orgID = id
		} else {
			_ = pool.QueryRow(ctx, `SELECT id FROM "organizations" WHERE slug = $1 LIMIT 1`, orgRaw).Scan(&orgID)
		}
	}
	if orgID == uuid.Nil && len(projectIDs) > 0 {
		_ = pool.QueryRow(ctx, `SELECT w.organization_id FROM "projects" p JOIN "workspaces" w ON p.workspace_id = w.id WHERE p.id = $1 LIMIT 1`, projectIDs[0]).Scan(&orgID)
	}
	if orgID == uuid.Nil && len(workspaceIDs) > 0 {
		_ = pool.QueryRow(ctx, `SELECT organization_id FROM "workspaces" WHERE id = $1 LIMIT 1`, workspaceIDs[0]).Scan(&orgID)
	}
	if orgID == uuid.Nil {
		_ = pool.QueryRow(ctx, `SELECT organization_id FROM "organization_members" WHERE user_id = $1 LIMIT 1`, userID).Scan(&orgID)
	}
	if orgID == uuid.Nil {
		return uuid.Nil, nil, nil, apperror.BadRequest("Organization context could not be determined. Please specify organizationId, workspaceId, or projectId.")
	}
	return orgID, projectIDs, workspaceIDs, nil
}

func (r *pgRepository) GetOrgAccess(ctx context.Context, orgID, userID uuid.UUID, isSuperAdmin bool) (AccessContext, error) {
	if isSuperAdmin {
		return AccessContext{IsSuperAdmin: true, IsElevated: true, Role: middleware.RoleSuperAdmin}, nil
	}
	var ownerID uuid.UUID
	var archived bool
	var memberRole pgtype.Text
	err := r.db.Pool().QueryRow(ctx, `
		SELECT o.owner_id, o.archive, om.role::text
		FROM "organizations" o
		LEFT JOIN "organization_members" om ON om.organization_id = o.id AND om.user_id = $2
		WHERE o.id = $1`, orgID, userID).Scan(&ownerID, &archived, &memberRole)
	if err != nil {
		return AccessContext{}, apperror.NotFound("Organization not found")
	}
	if ownerID == userID && !archived {
		return AccessContext{IsElevated: true, Role: middleware.RoleOwner}, nil
	}
	if !memberRole.Valid {
		return AccessContext{}, apperror.Forbidden("You are not a member of this organization")
	}
	role := middleware.Role(memberRole.String)
	return AccessContext{
		IsElevated: role == middleware.RoleManager || role == middleware.RoleOwner,
		Role:       role,
	}, nil
}

func buildVisibilityClause(b *argBuilder, acc AccessContext, userID uuid.UUID) string {
	if acc.IsSuperAdmin || acc.IsElevated {
		return ""
	}
	uRef := b.add(userID)
	return fmt.Sprintf(`(p.visibility = 'PUBLIC' OR (p.visibility = 'INTERNAL' AND EXISTS(SELECT 1 FROM "workspace_members" wm WHERE wm.workspace_id = p.workspace_id AND wm.user_id = %s)) OR EXISTS(SELECT 1 FROM "project_members" pm WHERE pm.project_id = p.id AND pm.user_id = %s) OR EXISTS(SELECT 1 FROM "workspace_members" wm WHERE wm.workspace_id = p.workspace_id AND wm.user_id = %s AND wm.role IN ('MANAGER','OWNER')))`, uRef, uRef, uRef)
}

func buildDynamicTaskWhere(
	b *argBuilder,
	orgID uuid.UUID,
	projectIDs, workspaceIDs []uuid.UUID,
	acc AccessContext,
	userID uuid.UUID,
	q TaskFilterQuery,
	excludeDoneWhenNoStatus bool,
	defaultTopLevel bool,
) []string {
	conds := []string{fmt.Sprintf("w.organization_id = %s", b.add(orgID))}
	if vis := buildVisibilityClause(b, acc, userID); vis != "" {
		conds = append(conds, vis)
	}
	if len(workspaceIDs) > 0 {
		conds = append(conds, fmt.Sprintf("p.workspace_id = ANY(%s::uuid[])", b.add(workspaceIDs)))
	}
	if len(projectIDs) > 0 {
		conds = append(conds, fmt.Sprintf("t.project_id = ANY(%s::uuid[])", b.add(projectIDs)))
	}
	if q.SprintID != "" {
		if sid, err := uuid.Parse(q.SprintID); err == nil {
			conds = append(conds, fmt.Sprintf("t.sprint_id = %s", b.add(sid)))
		}
	}
	if q.ParentTaskID != nil {
		switch *q.ParentTaskID {
		case "all":
		case "", "null":
			conds = append(conds, "t.parent_task_id IS NULL")
		default:
			if pid, err := uuid.Parse(*q.ParentTaskID); err == nil {
				conds = append(conds, fmt.Sprintf("t.parent_task_id = %s", b.add(pid)))
			} else {
				conds = append(conds, fmt.Sprintf("t.parent_task_id = COALESCE((SELECT id FROM \"tasks\" WHERE slug = %s LIMIT 1), '00000000-0000-0000-0000-000000000000'::uuid)", b.add(*q.ParentTaskID)))
			}
		}
	} else if defaultTopLevel {
		conds = append(conds, "t.parent_task_id IS NULL")
	}

	if len(q.Priorities) > 0 {
		conds = append(conds, fmt.Sprintf("t.priority::text = ANY(%s::text[])", b.add(q.Priorities)))
	}
	if len(q.Statuses) > 0 {
		conds = append(conds, fmt.Sprintf("t.status_id::text = ANY(%s::text[])", b.add(q.Statuses)))
	} else if excludeDoneWhenNoStatus {
		conds = append(conds, "ts.category != 'DONE'")
	}
	if len(q.Types) > 0 {
		conds = append(conds, fmt.Sprintf("t.type::text = ANY(%s::text[])", b.add(q.Types)))
	}
	if len(q.AssigneeIDs) > 0 {
		conds = append(conds, fmt.Sprintf(`EXISTS(SELECT 1 FROM "task_assignees" ta WHERE ta.task_id = t.id AND ta.user_id::text = ANY(%s::text[]))`, b.add(q.AssigneeIDs)))
	}
	if len(q.ReporterIDs) > 0 {
		conds = append(conds, fmt.Sprintf(`EXISTS(SELECT 1 FROM "task_reporters" tr WHERE tr.task_id = t.id AND tr.user_id::text = ANY(%s::text[]))`, b.add(q.ReporterIDs)))
	}
	if strings.TrimSpace(q.Search) != "" {
		pat := "%" + strings.TrimSpace(q.Search) + "%"
		ref := b.add(pat)
		conds = append(conds, fmt.Sprintf("(t.title ILIKE %s OR t.description ILIKE %s OR t.slug ILIKE %s)", ref, ref, ref))
	}
	if q.From != nil && q.To != nil {
		col := "t.due_date"
		switch q.DateField {
		case "createdAt":
			col = "t.created_at"
		case "updatedAt":
			col = "t.updated_at"
		case "startDate":
			col = "t.start_date"
		}
		conds = append(conds, fmt.Sprintf("%s >= %s AND %s <= %s", col, b.add(*q.From), col, b.add(*q.To)))
	}
	return conds
}

func (r *pgRepository) FindTasksDynamic(
	ctx context.Context,
	orgID uuid.UUID,
	projectIDs, workspaceIDs []uuid.UUID,
	acc AccessContext,
	userID uuid.UUID,
	q TaskFilterQuery,
	excludeDoneWhenNoStatus bool,
) ([]uuid.UUID, int64, FilterCountsResponse, error) {
	pool := r.db.Pool()
	b := &argBuilder{}
	conds := buildDynamicTaskWhere(b, orgID, projectIDs, workspaceIDs, acc, userID, q, excludeDoneWhenNoStatus, true)
	whereSQL := strings.Join(conds, " AND ")

	fromSQL := `FROM "tasks" t JOIN "projects" p ON t.project_id = p.id JOIN "workspaces" w ON p.workspace_id = w.id JOIN "task_statuses" ts ON t.status_id = ts.id`

	var total int64
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) "+fromSQL+" WHERE "+whereSQL, b.args...).Scan(&total); err != nil {
		return nil, 0, FilterCountsResponse{}, err
	}

	dir := "DESC"
	if strings.EqualFold(q.SortOrder, "asc") {
		dir = "ASC"
	}
	orderCol := "t.created_at " + dir
	switch q.SortBy {
	case "title":
		orderCol = "t.title " + dir
	case "priority":
		orderCol = "t.priority " + dir
	case "dueDate":
		orderCol = "t.due_date " + dir + " NULLS LAST"
	case "taskNumber":
		orderCol = "t.task_number " + dir
	case "updatedAt":
		orderCol = "t.updated_at " + dir
	}

	page := max(1, q.Page)
	limit := min(max(1, q.Limit), 100)
	offset := (page - 1) * limit

	listArgs := append([]any(nil), b.args...)
	listArgs = append(listArgs, limit, offset)
	listSQL := fmt.Sprintf("SELECT t.id %s WHERE %s ORDER BY %s LIMIT $%d OFFSET $%d", fromSQL, whereSQL, orderCol, len(listArgs)-1, len(listArgs))
	rows, err := pool.Query(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, 0, FilterCountsResponse{}, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if sErr := rows.Scan(&id); sErr != nil {
			return nil, 0, FilterCountsResponse{}, sErr
		}
		ids = append(ids, id)
	}

	fc := FilterCountsResponse{
		Priorities: []ValueCount{},
		Types:      []ValueCount{},
		Statuses:   []EntityCount{},
		Assignees:  []EntityCount{},
		Reporters:  []EntityCount{},
	}

	if pRows, pErr := pool.Query(ctx, "SELECT t.priority::text, COUNT(*) "+fromSQL+" WHERE "+whereSQL+" GROUP BY t.priority", b.args...); pErr == nil {
		defer pRows.Close()
		for pRows.Next() {
			var v ValueCount
			if pRows.Scan(&v.Value, &v.Count) == nil {
				fc.Priorities = append(fc.Priorities, v)
			}
		}
	}
	if tRows, tErr := pool.Query(ctx, "SELECT t.type::text, COUNT(*) "+fromSQL+" WHERE "+whereSQL+" GROUP BY t.type", b.args...); tErr == nil {
		defer tRows.Close()
		for tRows.Next() {
			var v ValueCount
			if tRows.Scan(&v.Value, &v.Count) == nil {
				fc.Types = append(fc.Types, v)
			}
		}
	}
	if sRows, sErr := pool.Query(ctx, "SELECT t.status_id, ts.name, COUNT(*) "+fromSQL+" WHERE "+whereSQL+" GROUP BY t.status_id, ts.name", b.args...); sErr == nil {
		defer sRows.Close()
		for sRows.Next() {
			var e EntityCount
			if sRows.Scan(&e.ID, &e.Name, &e.Count) == nil {
				fc.Statuses = append(fc.Statuses, e)
			}
		}
	}
	if aRows, aErr := pool.Query(ctx, `SELECT ta.user_id, CONCAT(u.first_name, ' ', u.last_name), COUNT(*) `+fromSQL+` JOIN "task_assignees" ta ON ta.task_id = t.id JOIN "users" u ON u.id = ta.user_id WHERE `+whereSQL+` GROUP BY ta.user_id, u.first_name, u.last_name`, b.args...); aErr == nil {
		defer aRows.Close()
		for aRows.Next() {
			var e EntityCount
			if aRows.Scan(&e.ID, &e.Name, &e.Count) == nil {
				fc.Assignees = append(fc.Assignees, e)
			}
		}
	}
	if rRows, rErr := pool.Query(ctx, `SELECT tr.user_id, CONCAT(u.first_name, ' ', u.last_name), COUNT(*) `+fromSQL+` JOIN "task_reporters" tr ON tr.task_id = t.id JOIN "users" u ON u.id = tr.user_id WHERE `+whereSQL+` GROUP BY tr.user_id, u.first_name, u.last_name`, b.args...); rErr == nil {
		defer rRows.Close()
		for rRows.Next() {
			var e EntityCount
			if rRows.Scan(&e.ID, &e.Name, &e.Count) == nil {
				fc.Reporters = append(fc.Reporters, e)
			}
		}
	}

	return ids, total, fc, nil
}

func (r *pgRepository) GetProjectWorkflowInfoBySlug(ctx context.Context, slug string) (*ProjectWorkflowInfo, error) {
	pool := r.db.Pool()
	var info ProjectWorkflowInfo
	var workflowID uuid.UUID
	err := pool.QueryRow(ctx, `
		SELECT p.id, p.name, p.slug, p.workspace_id, p.workflow_id, w.organization_id, o.owner_id, o.archive
		FROM "projects" p
		JOIN "workspaces" w ON p.workspace_id = w.id
		JOIN "organizations" o ON w.organization_id = o.id
		WHERE p.slug = $1`, slug).Scan(&info.ProjectID, &info.ProjectName, &info.ProjectSlug, &info.WorkspaceID, &workflowID, &info.OrgID, &info.OrgOwnerID, &info.OrgArchived)
	if err != nil {
		return nil, apperror.NotFound("Project or project workflow not found")
	}
	statuses, err := r.queries.GetWorkflowStatusesByWorkflowID(ctx, workflowID)
	if err != nil {
		return nil, err
	}
	info.Statuses = statuses
	return &info, nil
}

func (r *pgRepository) GetProjectAccessByID(ctx context.Context, projectID, userID uuid.UUID, isSuperAdmin bool) (AccessContext, error) {
	if isSuperAdmin {
		return AccessContext{IsSuperAdmin: true, IsElevated: true, Role: middleware.RoleSuperAdmin}, nil
	}
	var visibility sqlc.ProjectVisibility
	var orgOwnerID uuid.UUID
	var orgArchived bool
	var orgRole, wsRole, projRole sqlc.NullRole
	err := r.db.Pool().QueryRow(ctx, `
		SELECT p.visibility, o.owner_id, o.archive, om.role, wm.role, pm.role
		FROM "projects" p
		JOIN "workspaces" w ON p.workspace_id = w.id
		JOIN "organizations" o ON w.organization_id = o.id
		LEFT JOIN "organization_members" om ON om.organization_id = o.id AND om.user_id = $2
		LEFT JOIN "workspace_members" wm ON wm.workspace_id = w.id AND wm.user_id = $2
		LEFT JOIN "project_members" pm ON pm.project_id = p.id AND pm.user_id = $2
		WHERE p.id = $1`, projectID, userID).Scan(&visibility, &orgOwnerID, &orgArchived, &orgRole, &wsRole, &projRole)
	if err != nil {
		return AccessContext{}, apperror.NotFound("Project not found")
	}
	if orgOwnerID == userID && !orgArchived {
		return AccessContext{IsElevated: true, Role: middleware.RoleOwner}, nil
	}
	if orgRole.Valid && (orgRole.Role == sqlc.RoleMANAGER || orgRole.Role == sqlc.RoleOWNER) {
		return AccessContext{IsElevated: true, Role: middleware.Role(orgRole.Role)}, nil
	}
	if wsRole.Valid && (wsRole.Role == sqlc.RoleMANAGER || wsRole.Role == sqlc.RoleOWNER) {
		return AccessContext{IsElevated: true, Role: middleware.Role(wsRole.Role)}, nil
	}
	if projRole.Valid {
		role := middleware.Role(projRole.Role)
		return AccessContext{IsElevated: role == middleware.RoleManager || role == middleware.RoleOwner, Role: role}, nil
	}
	if visibility == sqlc.ProjectVisibilityINTERNAL && wsRole.Valid {
		return AccessContext{IsElevated: false, Role: middleware.Role(wsRole.Role)}, nil
	}
	if visibility == sqlc.ProjectVisibilityPUBLIC {
		return AccessContext{IsElevated: false, Role: middleware.RoleViewer}, nil
	}
	if !wsRole.Valid && !orgRole.Valid {
		return AccessContext{}, apperror.Forbidden("Not a member of this project")
	}
	return AccessContext{IsElevated: false, Role: middleware.RoleViewer}, nil
}

func (r *pgRepository) QueryColumnTasks(ctx context.Context, projectID, statusID uuid.UUID, sprintID *uuid.UUID, includeSubtasks, restrictedToUser bool, userID uuid.UUID, limit, offset int) ([]uuid.UUID, int64, error) {
	pool := r.db.Pool()
	b := &argBuilder{}
	conds := []string{
		fmt.Sprintf("t.status_id = %s", b.add(statusID)),
		fmt.Sprintf("t.project_id = %s", b.add(projectID)),
	}
	if sprintID != nil {
		conds = append(conds, fmt.Sprintf("t.sprint_id = %s", b.add(*sprintID)))
	}
	if !includeSubtasks {
		conds = append(conds, "t.parent_task_id IS NULL")
	}
	if restrictedToUser {
		uRef := b.add(userID)
		conds = append(conds, fmt.Sprintf(`(t.created_by_id = %s OR EXISTS(SELECT 1 FROM "task_assignees" ta WHERE ta.task_id = t.id AND ta.user_id = %s) OR EXISTS(SELECT 1 FROM "task_reporters" tr WHERE tr.task_id = t.id AND tr.user_id = %s))`, uRef, uRef, uRef))
	}
	whereSQL := strings.Join(conds, " AND ")

	var total int64
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM "tasks" t WHERE `+whereSQL, b.args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	pRef := b.add(projectID)
	lRef := b.add(limit)
	oRef := b.add(offset)
	rows, err := pool.Query(ctx, fmt.Sprintf(`
		SELECT t.id
		FROM "tasks" t
		LEFT JOIN "task_ranks" trk ON t.id = trk.task_id
			AND trk.scope_type = 'PROJECT'::"ScopeType"
			AND trk.scope_id = %s
			AND trk.view_type = 'BOARD'::"ViewType"
		WHERE %s
		ORDER BY trk.rank DESC NULLS LAST, t.created_at DESC
		LIMIT %s OFFSET %s`, pRef, whereSQL, lRef, oRef), b.args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if sErr := rows.Scan(&id); sErr != nil {
			return nil, 0, sErr
		}
		ids = append(ids, id)
	}
	return ids, total, nil
}

func (r *pgRepository) FindTodayTaskIDs(ctx context.Context, orgID uuid.UUID, acc AccessContext, userID uuid.UUID, dayStart, dayEnd time.Time, limit, offset int) ([]uuid.UUID, int64, error) {
	pool := r.db.Pool()
	b := &argBuilder{}
	sRef := b.add(dayStart)
	eRef := b.add(dayEnd)
	conds := []string{
		fmt.Sprintf("w.organization_id = %s", b.add(orgID)),
		fmt.Sprintf("((t.due_date >= %s AND t.due_date <= %s) OR (t.created_at >= %s AND t.created_at <= %s) OR (t.updated_at >= %s AND t.updated_at <= %s) OR (t.completed_at >= %s AND t.completed_at <= %s))", sRef, eRef, sRef, eRef, sRef, eRef, sRef, eRef),
	}
	if vis := buildVisibilityClause(b, acc, userID); vis != "" {
		uRef := b.add(userID)
		conds = append(conds, fmt.Sprintf("(%s OR EXISTS(SELECT 1 FROM \"task_assignees\" ta WHERE ta.task_id = t.id AND ta.user_id = %s) OR EXISTS(SELECT 1 FROM \"task_reporters\" tr WHERE tr.task_id = t.id AND tr.user_id = %s) OR t.created_by_id = %s)", vis, uRef, uRef, uRef))
	}
	whereSQL := strings.Join(conds, " AND ")
	fromSQL := `FROM "tasks" t JOIN "projects" p ON t.project_id = p.id JOIN "workspaces" w ON p.workspace_id = w.id`

	var total int64
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) "+fromSQL+" WHERE "+whereSQL, b.args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	lRef := b.add(limit)
	oRef := b.add(offset)
	rows, err := pool.Query(ctx, fmt.Sprintf("SELECT t.id %s WHERE %s ORDER BY t.priority DESC, t.due_date ASC NULLS LAST LIMIT %s OFFSET %s", fromSQL, whereSQL, lRef, oRef), b.args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	return ids, total, nil
}

func (r *pgRepository) FindOrgTaskIDs(ctx context.Context, orgID, assigneeID uuid.UUID, priority, search string, limit, offset int) ([]uuid.UUID, int64, error) {
	pool := r.db.Pool()
	b := &argBuilder{}
	conds := []string{
		fmt.Sprintf("w.organization_id = %s", b.add(orgID)),
		fmt.Sprintf(`EXISTS(SELECT 1 FROM "task_assignees" ta WHERE ta.task_id = t.id AND ta.user_id = %s)`, b.add(assigneeID)),
	}
	if priority != "" {
		conds = append(conds, fmt.Sprintf("t.priority::text = %s", b.add(priority)))
	}
	if strings.TrimSpace(search) != "" {
		ref := b.add("%" + strings.TrimSpace(search) + "%")
		conds = append(conds, fmt.Sprintf("(t.title ILIKE %s OR t.slug ILIKE %s)", ref, ref))
	}
	whereSQL := strings.Join(conds, " AND ")
	fromSQL := `FROM "tasks" t JOIN "projects" p ON t.project_id = p.id JOIN "workspaces" w ON p.workspace_id = w.id`

	var total int64
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) "+fromSQL+" WHERE "+whereSQL, b.args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	lRef := b.add(limit)
	oRef := b.add(offset)
	rows, err := pool.Query(ctx, fmt.Sprintf("SELECT t.id %s WHERE %s ORDER BY t.priority DESC, t.created_at DESC LIMIT %s OFFSET %s", fromSQL, whereSQL, lRef, oRef), b.args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	return ids, total, nil
}

func (r *pgRepository) QueryGroupedTasks(ctx context.Context, orgID uuid.UUID, projectIDs, workspaceIDs []uuid.UUID, acc AccessContext, userID uuid.UUID, q GroupedTasksQuery) ([]TaskGroupBucket, error) {
	pool := r.db.Pool()
	b := &argBuilder{}
	fq := TaskFilterQuery{
		SprintID:    q.SprintID,
		Priorities:  q.Priorities,
		Statuses:    q.Statuses,
		Types:       q.Types,
		AssigneeIDs: q.AssigneeIDs,
		ReporterIDs: q.ReporterIDs,
		Search:      q.Search,
	}
	conds := buildDynamicTaskWhere(b, orgID, projectIDs, workspaceIDs, acc, userID, fq, false, false)
	whereSQL := strings.Join(conds, " AND ")
	fromSQL := `FROM "tasks" t JOIN "projects" p ON t.project_id = p.id JOIN "workspaces" w ON p.workspace_id = w.id JOIN "task_statuses" ts ON t.status_id = ts.id`

	keyExpr := "t.status_id::text"
	labelExpr := "ts.name"
	extraGroupCols := ", t.status_id, ts.name"
	switch q.GroupBy {
	case "priority":
		keyExpr = "t.priority::text"
		labelExpr = "t.priority::text"
		extraGroupCols = ", t.priority"
	case "project":
		keyExpr = "t.project_id::text"
		labelExpr = "p.name"
		extraGroupCols = ", t.project_id, p.name"
	case "type":
		keyExpr = "t.type::text"
		labelExpr = "t.type::text"
		extraGroupCols = ", t.type"
	case "dueDate":
		keyExpr = "COALESCE(TO_CHAR(t.due_date, 'YYYY-MM-DD'), 'no-date')"
		labelExpr = keyExpr
		extraGroupCols = ""
	case "createdAt":
		keyExpr = "COALESCE(TO_CHAR(t.created_at, 'YYYY-MM-DD'), 'no-date')"
		labelExpr = keyExpr
		extraGroupCols = ""
	}
	_ = extraGroupCols

	if q.GroupBy == "assignee" {
		keyExpr = "COALESCE(ta.user_id::text, 'unassigned')"
		labelExpr = "COALESCE(CONCAT(u.first_name, ' ', u.last_name), 'Unassigned')"
		fromSQL += ` LEFT JOIN "task_assignees" ta ON ta.task_id = t.id LEFT JOIN "users" u ON u.id = ta.user_id`
	}

	groupSQL := fmt.Sprintf("SELECT %s AS gkey, %s AS glabel, COUNT(DISTINCT t.id) %s WHERE %s GROUP BY gkey, glabel ORDER BY glabel ASC", keyExpr, labelExpr, fromSQL, whereSQL)
	gRows, err := pool.Query(ctx, groupSQL, b.args...)
	if err != nil {
		return nil, err
	}
	defer gRows.Close()

	var buckets []TaskGroupBucket
	for gRows.Next() {
		var item TaskGroupBucket
		if sErr := gRows.Scan(&item.Key, &item.Label, &item.TotalCount); sErr != nil {
			return nil, sErr
		}
		if q.GroupKey != "" && item.Key != q.GroupKey {
			continue
		}
		item.Tasks = []TaskResponse{}
		buckets = append(buckets, item)
	}

	limit := max(1, min(q.LimitPerGroup, 100))
	offset := (max(1, q.Page) - 1) * limit
	if q.GroupKey == "" {
		offset = 0
	}

	for idx := range buckets {
		bucketArgs := append([]any(nil), b.args...)
		bucketArgs = append(bucketArgs, buckets[idx].Key, limit, offset)
		idSQL := fmt.Sprintf("SELECT DISTINCT t.id, t.created_at %s WHERE %s AND %s = $%d ORDER BY t.created_at DESC LIMIT $%d OFFSET $%d", fromSQL, whereSQL, keyExpr, len(bucketArgs)-2, len(bucketArgs)-1, len(bucketArgs))
		idRows, iErr := pool.Query(ctx, idSQL, bucketArgs...)
		if iErr != nil {
			continue
		}
		var taskIDs []uuid.UUID
		for idRows.Next() {
			var tid uuid.UUID
			var _ts time.Time
			if idRows.Scan(&tid, &_ts) == nil {
				taskIDs = append(taskIDs, tid)
			}
		}
		idRows.Close()
		for _, tid := range taskIDs {
			buckets[idx].Tasks = append(buckets[idx].Tasks, TaskResponse{ID: tid.String()})
		}
	}

	return buckets, nil
}

func (r *pgRepository) ResolveBulkTaskIDs(ctx context.Context, all bool, explicitIDs, excludedIDs []uuid.UUID, projectID, workspaceID, sprintID *uuid.UUID, search string, statuses, priorities, types, assignees, reporters []string) ([]uuid.UUID, error) {
	if !all {
		excl := make(map[uuid.UUID]struct{}, len(excludedIDs))
		for _, id := range excludedIDs {
			excl[id] = struct{}{}
		}
		out := make([]uuid.UUID, 0, len(explicitIDs))
		for _, id := range explicitIDs {
			if _, skip := excl[id]; !skip {
				out = append(out, id)
			}
		}
		return out, nil
	}

	b := &argBuilder{}
	var conds []string
	if projectID != nil {
		conds = append(conds, fmt.Sprintf("t.project_id = %s", b.add(*projectID)))
	}
	if workspaceID != nil {
		conds = append(conds, fmt.Sprintf("p.workspace_id = %s", b.add(*workspaceID)))
	}
	if sprintID != nil {
		conds = append(conds, fmt.Sprintf("t.sprint_id = %s", b.add(*sprintID)))
	}
	if len(statuses) > 0 {
		conds = append(conds, fmt.Sprintf("t.status_id::text = ANY(%s::text[])", b.add(statuses)))
	}
	if len(priorities) > 0 {
		conds = append(conds, fmt.Sprintf("t.priority::text = ANY(%s::text[])", b.add(priorities)))
	}
	if len(types) > 0 {
		conds = append(conds, fmt.Sprintf("t.type::text = ANY(%s::text[])", b.add(types)))
	}
	if len(assignees) > 0 {
		conds = append(conds, fmt.Sprintf(`EXISTS(SELECT 1 FROM "task_assignees" ta WHERE ta.task_id = t.id AND ta.user_id::text = ANY(%s::text[]))`, b.add(assignees)))
	}
	if len(reporters) > 0 {
		conds = append(conds, fmt.Sprintf(`EXISTS(SELECT 1 FROM "task_reporters" tr WHERE tr.task_id = t.id AND tr.user_id::text = ANY(%s::text[]))`, b.add(reporters)))
	}
	if strings.TrimSpace(search) != "" {
		ref := b.add("%" + strings.TrimSpace(search) + "%")
		conds = append(conds, fmt.Sprintf("(t.title ILIKE %s OR t.description ILIKE %s)", ref, ref))
	}
	if len(excludedIDs) > 0 {
		conds = append(conds, fmt.Sprintf("NOT (t.id = ANY(%s::uuid[]))", b.add(excludedIDs)))
	}

	whereSQL := "TRUE"
	if len(conds) > 0 {
		whereSQL = strings.Join(conds, " AND ")
	}
	rows, err := r.db.Pool().Query(ctx, `SELECT t.id FROM "tasks" t JOIN "projects" p ON t.project_id = p.id WHERE `+whereSQL, b.args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if rows.Scan(&id) == nil {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (r *pgRepository) FindStatusInWorkflowByIDOrName(ctx context.Context, projectID uuid.UUID, statusID *uuid.UUID) (*sqlc.GetTaskStatusBasicByIDRow, error) {
	pool := r.db.Pool()
	if statusID == nil {
		var row sqlc.GetTaskStatusBasicByIDRow
		err := pool.QueryRow(ctx, `
			SELECT ts.id, ts.name, ts.color, ts.category, ts.workflow_id
			FROM "projects" p
			JOIN "task_statuses" ts ON ts.workflow_id = p.workflow_id
			WHERE p.id = $1 AND ts.category = 'DONE' AND ts.deleted_at IS NULL
			LIMIT 1`, projectID).Scan(&row.ID, &row.Name, &row.Color, &row.Category, &row.WorkflowID)
		if err != nil {
			return nil, err
		}
		return &row, nil
	}

	var row sqlc.GetTaskStatusBasicByIDRow
	err := pool.QueryRow(ctx, `
		SELECT ts.id, ts.name, ts.color, ts.category, ts.workflow_id
		FROM "projects" p
		JOIN "task_statuses" ts ON ts.workflow_id = p.workflow_id
		WHERE p.id = $1 AND ts.id = $2
		LIMIT 1`, projectID, *statusID).Scan(&row.ID, &row.Name, &row.Color, &row.Category, &row.WorkflowID)
	if err == nil {
		return &row, nil
	}

	target, tErr := r.GetStatusByID(ctx, *statusID)
	if tErr != nil {
		return nil, tErr
	}
	err = pool.QueryRow(ctx, `
		SELECT ts.id, ts.name, ts.color, ts.category, ts.workflow_id
		FROM "projects" p
		JOIN "task_statuses" ts ON ts.workflow_id = p.workflow_id
		WHERE p.id = $1 AND LOWER(ts.name) = LOWER($2) AND ts.deleted_at IS NULL
		LIMIT 1`, projectID, target.Name).Scan(&row.ID, &row.Name, &row.Color, &row.Category, &row.WorkflowID)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) BulkReplaceAssignees(ctx context.Context, taskIDs []uuid.UUID, assigneeIDs []uuid.UUID) error {
	return r.db.WithTx(ctx, func(q *sqlc.Queries) error {
		for _, tid := range taskIDs {
			if err := q.DeleteTaskAssigneesByTaskID(ctx, tid); err != nil {
				return err
			}
			for _, uid := range assigneeIDs {
				if err := q.InsertTaskAssignee(ctx, sqlc.InsertTaskAssigneeParams{TaskID: tid, UserID: uid}); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func marshalJSONField(val map[string]any) []byte {
	if val == nil {
		return nil
	}
	b, _ := json.Marshal(val)
	return b
}
