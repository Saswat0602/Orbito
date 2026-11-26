package tasks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubRepo struct {
	accessRow *sqlc.GetTaskAccessDetailsRow
	fullRow   *sqlc.GetTaskFullByIDRow
	shares    map[uuid.UUID]*sqlc.PublicTaskShare
}

func (s *stubRepo) GetProjectMeta(_ context.Context, projectID uuid.UUID) (*sqlc.GetProjectMetaForTaskCreateRow, error) {
	return &sqlc.GetProjectMetaForTaskCreateRow{ID: projectID, Slug: "orb", WorkflowID: uuid.New(), WorkspaceID: uuid.New(), OrganizationID: uuid.New()}, nil
}
func (s *stubRepo) GetProjectAccessForCreate(_ context.Context, _, _ uuid.UUID, _ bool) (bool, error) {
	return true, nil
}
func (s *stubRepo) GetDefaultSprintID(_ context.Context, _ uuid.UUID) (pgtype.UUID, error) {
	return pgtype.UUID{}, nil
}
func (s *stubRepo) CreateTaskTx(_ context.Context, in CreateTxInput) (uuid.UUID, error) {
	return in.TaskID, nil
}
func (s *stubRepo) BulkCreateTasksTx(_ context.Context, _, _ uuid.UUID, _ pgtype.UUID, items []CreateTxInput) ([]uuid.UUID, error) {
	out := make([]uuid.UUID, 0, len(items))
	for _, it := range items {
		out = append(out, it.TaskID)
	}
	return out, nil
}
func (s *stubRepo) CreateAttachment(_ context.Context, _ CreatedAttachmentInput) (*sqlc.CreateTaskAttachmentForTaskRow, error) {
	return &sqlc.CreateTaskAttachmentForTaskRow{}, nil
}
func (s *stubRepo) GetTaskAccessDetails(_ context.Context, _ string, _ uuid.UUID) (*sqlc.GetTaskAccessDetailsRow, error) {
	if s.accessRow != nil {
		return s.accessRow, nil
	}
	return nil, pgx.ErrNoRows
}
func (s *stubRepo) GetTaskFullByID(_ context.Context, _ uuid.UUID) (*sqlc.GetTaskFullByIDRow, error) {
	if s.fullRow != nil {
		return s.fullRow, nil
	}
	return nil, pgx.ErrNoRows
}
func (s *stubRepo) ListAssigneesBatch(_ context.Context, _ []uuid.UUID) ([]sqlc.ListTaskAssigneesBatchRow, error) {
	return nil, nil
}
func (s *stubRepo) ListReportersBatch(_ context.Context, _ []uuid.UUID) ([]sqlc.ListTaskReportersBatchRow, error) {
	return nil, nil
}
func (s *stubRepo) ListLabelsBatch(_ context.Context, _ []uuid.UUID) ([]sqlc.ListTaskLabelsBatchRow, error) {
	return nil, nil
}
func (s *stubRepo) ListAttachmentsByTask(_ context.Context, _ uuid.UUID) ([]sqlc.ListTaskAttachmentsByTaskIDRow, error) {
	return nil, nil
}
func (s *stubRepo) ListWatchersByTask(_ context.Context, _ uuid.UUID) ([]sqlc.ListTaskWatchersByTaskIDRow, error) {
	return nil, nil
}
func (s *stubRepo) ListCommentsByTask(_ context.Context, _ uuid.UUID) ([]sqlc.ListTaskCommentsByTaskIDRow, error) {
	return nil, nil
}
func (s *stubRepo) ListTimeEntriesByTask(_ context.Context, _ uuid.UUID) ([]sqlc.ListTaskTimeEntriesByTaskIDRow, error) {
	return nil, nil
}
func (s *stubRepo) ListChildTasks(_ context.Context, _ uuid.UUID) ([]sqlc.ListTaskChildrenForDetailRow, error) {
	return nil, nil
}
func (s *stubRepo) GetProjectInbox(_ context.Context, _ uuid.UUID) (*sqlc.GetProjectInboxForTaskDetailRow, error) {
	return nil, nil
}
func (s *stubRepo) GetStatusByID(_ context.Context, id uuid.UUID) (*sqlc.GetTaskStatusBasicByIDRow, error) {
	return &sqlc.GetTaskStatusBasicByIDRow{ID: id, Category: sqlc.StatusCategoryINPROGRESS}, nil
}
func (s *stubRepo) ResolveTaskIDBySlug(_ context.Context, _ string) (uuid.UUID, error) {
	return uuid.Nil, pgx.ErrNoRows
}
func (s *stubRepo) CheckUsersExist(_ context.Context, userIDs []uuid.UUID) ([]uuid.UUID, error) {
	return userIDs, nil
}
func (s *stubRepo) UpdateTaskTx(_ context.Context, _ sqlc.UpdateTaskCoreByIDParams, _, _ *[]uuid.UUID) error {
	return nil
}
func (s *stubRepo) DeleteTaskByID(_ context.Context, _ uuid.UUID) (int64, error)    { return 1, nil }
func (s *stubRepo) BulkDeleteByIDs(_ context.Context, _ []uuid.UUID) (int64, error) { return 0, nil }
func (s *stubRepo) CreateQuickComment(_ context.Context, taskID, authorID uuid.UUID, content string) (*CommentBrief, error) {
	return &CommentBrief{ID: uuid.New().String(), TaskID: taskID.String(), AuthorID: authorID.String(), Content: content}, nil
}
func (s *stubRepo) GetRecurringConfig(_ context.Context, _ uuid.UUID) (*sqlc.RecurringTask, error) {
	return nil, pgx.ErrNoRows
}
func (s *stubRepo) CreateRecurrenceConfig(_ context.Context, p sqlc.CreateRecurringTaskConfigParams) (*sqlc.RecurringTask, error) {
	return &sqlc.RecurringTask{ID: p.ID, TaskID: p.TaskID, RecurrenceType: p.RecurrenceType, Interval: p.Interval, EndType: p.EndType, CurrentOccurrence: 1, NextOccurrence: p.NextOccurrence, IsActive: true}, nil
}
func (s *stubRepo) UpdateRecurrenceConfig(_ context.Context, _ sqlc.UpdateRecurringTaskConfigByIDParams) (*sqlc.RecurringTask, error) {
	return &sqlc.RecurringTask{}, nil
}
func (s *stubRepo) IncrementRecurrenceOccurrence(_ context.Context, _ uuid.UUID, _ int32, _ time.Time) error {
	return nil
}
func (s *stubRepo) DeleteRecurrenceConfig(_ context.Context, _ uuid.UUID) error { return nil }
func (s *stubRepo) SetTaskIsRecurring(_ context.Context, id uuid.UUID, _ bool) (*sqlc.Task, error) {
	return &sqlc.Task{ID: id}, nil
}
func (s *stubRepo) ListRecurringByProject(_ context.Context, _ uuid.UUID) ([]sqlc.ListRecurringTasksByProjectIDRow, error) {
	return nil, nil
}
func (s *stubRepo) CreateShareLink(_ context.Context, params sqlc.CreatePublicTaskShareParams) (*sqlc.PublicTaskShare, error) {
	now := pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
	sh := &sqlc.PublicTaskShare{ID: params.ID, Token: params.Token, ExpiresAt: params.ExpiresAt, TaskID: params.TaskID, CreatedAt: now}
	if s.shares == nil {
		s.shares = make(map[uuid.UUID]*sqlc.PublicTaskShare)
	}
	s.shares[sh.ID] = sh
	return sh, nil
}
func (s *stubRepo) ListActiveShares(_ context.Context, _ uuid.UUID) ([]sqlc.PublicTaskShare, error) {
	return []sqlc.PublicTaskShare{}, nil
}
func (s *stubRepo) GetShareWithAccess(_ context.Context, id, _ uuid.UUID) (*sqlc.GetTaskShareWithAccessRow, error) {
	return &sqlc.GetTaskShareWithAccessRow{ID: id, IsProjectMember: true}, nil
}
func (s *stubRepo) RevokeShare(_ context.Context, _ uuid.UUID) error { return nil }
func (s *stubRepo) ResolveScopeContext(_ context.Context, _ string, _, _ []string, _ uuid.UUID) (uuid.UUID, []uuid.UUID, []uuid.UUID, error) {
	return uuid.New(), nil, nil, nil
}
func (s *stubRepo) GetOrgAccess(_ context.Context, _, _ uuid.UUID, _ bool) (AccessContext, error) {
	return AccessContext{IsElevated: true, Role: middleware.RoleOwner}, nil
}
func (s *stubRepo) FindTasksDynamic(_ context.Context, _ uuid.UUID, _, _ []uuid.UUID, _ AccessContext, _ uuid.UUID, _ TaskFilterQuery, _ bool) ([]uuid.UUID, int64, FilterCountsResponse, error) {
	return nil, 0, FilterCountsResponse{}, nil
}
func (s *stubRepo) GetProjectWorkflowInfoBySlug(_ context.Context, _ string) (*ProjectWorkflowInfo, error) {
	return &ProjectWorkflowInfo{ProjectID: uuid.New()}, nil
}
func (s *stubRepo) GetProjectAccessByID(_ context.Context, _, _ uuid.UUID, _ bool) (AccessContext, error) {
	return AccessContext{IsElevated: true, Role: middleware.RoleOwner}, nil
}
func (s *stubRepo) QueryColumnTasks(_ context.Context, _, _ uuid.UUID, _ *uuid.UUID, _, _ bool, _ uuid.UUID, _, _ int) ([]uuid.UUID, int64, error) {
	return nil, 0, nil
}
func (s *stubRepo) FindTodayTaskIDs(_ context.Context, _ uuid.UUID, _ AccessContext, _ uuid.UUID, _, _ time.Time, _, _ int) ([]uuid.UUID, int64, error) {
	return nil, 0, nil
}
func (s *stubRepo) FindOrgTaskIDs(_ context.Context, _, _ uuid.UUID, _, _ string, _, _ int) ([]uuid.UUID, int64, error) {
	return nil, 0, nil
}
func (s *stubRepo) QueryGroupedTasks(_ context.Context, _ uuid.UUID, _, _ []uuid.UUID, _ AccessContext, _ uuid.UUID, _ GroupedTasksQuery) ([]TaskGroupBucket, error) {
	return []TaskGroupBucket{}, nil
}
func (s *stubRepo) ResolveBulkTaskIDs(_ context.Context, _ bool, explicitIDs, _ []uuid.UUID, _, _, _ *uuid.UUID, _ string, _, _, _, _, _ []string) ([]uuid.UUID, error) {
	return explicitIDs, nil
}
func (s *stubRepo) FindStatusInWorkflowByIDOrName(_ context.Context, _ uuid.UUID, _ *uuid.UUID) (*sqlc.GetTaskStatusBasicByIDRow, error) {
	return nil, pgx.ErrNoRows
}
func (s *stubRepo) BulkReplaceAssignees(_ context.Context, _, _ []uuid.UUID) error { return nil }

func TestRecurrenceMathAndMemberRBAC(t *testing.T) {
	base := time.Date(2026, 11, 26, 10, 0, 0, 0, time.UTC) // Thursday (weekday 4)
	iv := int32(2)
	nextDaily := CalculateNextOccurrence(base, RecurrenceConfigRequest{RecurrenceType: "DAILY", Interval: &iv})
	assert.Equal(t, 28, nextDaily.Day())

	// Weekly on Mon(1), Fri(5) from Thursday(4) -> picks Friday(5) (+1 day)
	iv1 := int32(1)
	nextWeekly := CalculateNextOccurrence(base, RecurrenceConfigRequest{
		RecurrenceType: "WEEKLY",
		Interval:       &iv1,
		DaysOfWeek:     []int32{1, 5},
	})
	assert.Equal(t, time.Friday, nextWeekly.Weekday())

	userID := uuid.New()
	taskID := uuid.New()
	repo := &stubRepo{
		accessRow: &sqlc.GetTaskAccessDetailsRow{
			ID:          taskID,
			ProjectID:   uuid.New(),
			IsAssignee:  true,
			ProjectRole: sqlc.NullRole{Role: sqlc.RoleMEMBER, Valid: true},
		},
	}
	svc := NewService(repo, nil, nil, nil, "http://localhost:3001")
	newTitle := "Hacked"
	_, err := svc.Update(context.Background(), taskID.String(), UpdateTaskRequest{Title: &newTitle}, userID, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Members are not allowed to update the title field")

	share, err := svc.CreateShareLink(context.Background(), CreateShareLinkRequest{TaskID: taskID.String(), ExpiresInDays: 7}, userID, false)
	require.NoError(t, err)
	assert.Contains(t, share.ShareURL, "http://localhost:3001/public/task/")
}

func TestTasksHandler_Validation(t *testing.T) {
	h := NewHandler(NewService(&stubRepo{}, nil, nil, nil, "http://localhost:3001"))
	router := Routes(h, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := ctxhelper.WithUserID(r.Context(), uuid.New().String())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"title":""}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	groupedReq := httptest.NewRequest(http.MethodGet, "/grouped", nil)
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, groupedReq)
	assert.Equal(t, http.StatusBadRequest, rr2.Code)
}
