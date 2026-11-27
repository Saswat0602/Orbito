package public

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePublicRepo struct {
	shareByToken map[string]*sqlc.GetPublicSharedTaskByTokenRow
	shareAtts    map[string]*sqlc.GetPublicShareAttachmentByIDRow
	projBySlugs  map[string]*sqlc.GetPublicProjectBySlugsRow
	projBySlug   map[string]*sqlc.GetPublicProjectBySlugOnlyRow
	kpis         map[uuid.UUID]*sqlc.GetPublicProjectKPIsRow
	statuses     map[uuid.UUID][]sqlc.ListPublicProjectStatusesByWorkflowRow
	tasks        []sqlc.ListPublicProjectTasksRow
	comments     []sqlc.ListPublicTaskCommentsRow
	assignees    []sqlc.ListPublicTaskAssigneesBatchRow
	attachments  []sqlc.ListPublicTaskAttachmentsRow
}

func (f *fakePublicRepo) GetSharedTaskByToken(_ context.Context, tok string) (*sqlc.GetPublicSharedTaskByTokenRow, error) {
	return f.shareByToken[tok], nil
}
func (f *fakePublicRepo) GetShareAttachmentByID(_ context.Context, tok string, attID uuid.UUID) (*sqlc.GetPublicShareAttachmentByIDRow, error) {
	return f.shareAtts[tok+":"+attID.String()], nil
}
func (f *fakePublicRepo) ListWorkspaceProjects(_ context.Context, _ string) ([]sqlc.ListWorkspacePublicProjectsRow, error) {
	return nil, nil
}
func (f *fakePublicRepo) GetProjectBySlugs(_ context.Context, ws, ps string) (*sqlc.GetPublicProjectBySlugsRow, error) {
	return f.projBySlugs[ws+"/"+ps], nil
}
func (f *fakePublicRepo) GetProjectBySlugOnly(_ context.Context, ps string) (*sqlc.GetPublicProjectBySlugOnlyRow, error) {
	return f.projBySlug[ps], nil
}
func (f *fakePublicRepo) ListStatusesByWorkflow(_ context.Context, wfID uuid.UUID) ([]sqlc.ListPublicProjectStatusesByWorkflowRow, error) {
	return f.statuses[wfID], nil
}
func (f *fakePublicRepo) GetProjectKPIs(_ context.Context, pid uuid.UUID) (*sqlc.GetPublicProjectKPIsRow, error) {
	return f.kpis[pid], nil
}
func (f *fakePublicRepo) ListTaskCountsByStatus(_ context.Context, _, _ uuid.UUID) ([]sqlc.ListPublicProjectTaskCountsByStatusRow, error) {
	return nil, nil
}
func (f *fakePublicRepo) ListTaskCountsByType(_ context.Context, _ uuid.UUID) ([]sqlc.ListPublicProjectTaskCountsByTypeRow, error) {
	return nil, nil
}
func (f *fakePublicRepo) ListTaskCountsByPriority(_ context.Context, _ uuid.UUID) ([]sqlc.ListPublicProjectTaskCountsByPriorityRow, error) {
	return nil, nil
}
func (f *fakePublicRepo) ListSprintVelocity(_ context.Context, _ uuid.UUID) ([]sqlc.ListPublicProjectSprintVelocityRow, error) {
	return nil, nil
}
func (f *fakePublicRepo) ListCalendarTasks(_ context.Context, _ sqlc.ListPublicCalendarTasksParams) ([]sqlc.ListPublicCalendarTasksRow, error) {
	return nil, nil
}
func (f *fakePublicRepo) ListSprints(_ context.Context, _ uuid.UUID) ([]sqlc.ListPublicProjectSprintsRow, error) {
	return nil, nil
}
func (f *fakePublicRepo) GetSprintByID(_ context.Context, _ uuid.UUID, _, _ string) (*sqlc.GetPublicSprintByIDRow, error) {
	return nil, nil
}
func (f *fakePublicRepo) ListSprintTasks(_ context.Context, _ uuid.UUID, _, _ string) ([]sqlc.ListPublicSprintTasksRow, error) {
	return nil, nil
}
func (f *fakePublicRepo) CountProjectTasks(_ context.Context, _ sqlc.CountPublicProjectTasksParams) (int64, error) {
	return int64(len(f.tasks)), nil
}
func (f *fakePublicRepo) ListProjectTasks(_ context.Context, _ sqlc.ListPublicProjectTasksParams) ([]sqlc.ListPublicProjectTasksRow, error) {
	return f.tasks, nil
}
func (f *fakePublicRepo) ListSubtasksBatch(_ context.Context, _ []uuid.UUID) ([]sqlc.ListPublicSubtasksBatchRow, error) {
	return nil, nil
}
func (f *fakePublicRepo) ListTaskLabelsBatch(_ context.Context, _ []uuid.UUID) ([]sqlc.ListPublicTaskLabelsBatchRow, error) {
	return nil, nil
}
func (f *fakePublicRepo) ListTaskAssigneesBatch(_ context.Context, _ []uuid.UUID) ([]sqlc.ListPublicTaskAssigneesBatchRow, error) {
	return f.assignees, nil
}
func (f *fakePublicRepo) GetTaskByIDOrSlug(_ context.Context, _ sqlc.GetPublicTaskByIDOrSlugParams) (*sqlc.GetPublicTaskByIDOrSlugRow, error) {
	return &sqlc.GetPublicTaskByIDOrSlugRow{ID: uuid.New(), Title: "T", ProjectVisibility: sqlc.ProjectVisibilityPUBLIC}, nil
}
func (f *fakePublicRepo) CountTaskActivities(_ context.Context, _ uuid.UUID) (int64, error) {
	return 0, nil
}
func (f *fakePublicRepo) ListTaskActivities(_ context.Context, _ uuid.UUID, _, _ int32) ([]sqlc.ListPublicTaskActivitiesRow, error) {
	return nil, nil
}
func (f *fakePublicRepo) ListTaskComments(_ context.Context, _ uuid.UUID) ([]sqlc.ListPublicTaskCommentsRow, error) {
	return f.comments, nil
}
func (f *fakePublicRepo) ListTaskAttachments(_ context.Context, _ uuid.UUID) ([]sqlc.ListPublicTaskAttachmentsRow, error) {
	return f.attachments, nil
}

type fakeURLResolver struct{}

func (fakeURLResolver) PresignGetURL(_ context.Context, key string) (string, error) {
	return "https://cdn.example.com/" + key, nil
}

func TestSharedTaskLifecycleAndRateLimits(t *testing.T) {
	now := time.Now().UTC()
	taskID := uuid.New()
	attID := uuid.New()
	repo := &fakePublicRepo{
		shareByToken: map[string]*sqlc.GetPublicSharedTaskByTokenRow{
			"valid": {
				ID: taskID, Token: "valid", Title: "Public Bug", Priority: sqlc.TaskPriorityHIGH,
				ExpiresAt:  pgtype.Timestamp{Time: now.Add(time.Hour), Valid: true},
				StatusName: pgtype.Text{String: "Open", Valid: true}, StatusColor: pgtype.Text{String: "#22c55e", Valid: true},
			},
			"revoked": {
				ID: taskID, Token: "revoked", RevokedAt: pgtype.Timestamp{Time: now.Add(-time.Minute), Valid: true},
				ExpiresAt: pgtype.Timestamp{Time: now.Add(time.Hour), Valid: true},
			},
			"expired": {
				ID: taskID, Token: "expired", ExpiresAt: pgtype.Timestamp{Time: now.Add(-time.Hour), Valid: true},
			},
		},
		shareAtts: map[string]*sqlc.GetPublicShareAttachmentByIDRow{
			"valid:" + attID.String(): {ID: attID, StorageKey: pgtype.Text{String: "uploads/log.txt", Valid: true}},
		},
		assignees:   []sqlc.ListPublicTaskAssigneesBatchRow{{TaskID: taskID, FirstName: "Ada", LastName: "Lovelace"}},
		attachments: []sqlc.ListPublicTaskAttachmentsRow{{ID: attID, FileName: "log.txt", FileSize: 42, MimeType: "text/plain"}},
	}

	svc := NewService(repo, fakeURLResolver{})
	h := NewHandler(svc)
	rl := NewRateLimiter()

	r := chi.NewRouter()
	r.Mount("/api/public/tasks", SharedTasksRoutes(h, rl))

	req := httptest.NewRequest(http.MethodGet, "/api/public/tasks/valid", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "noindex, nofollow", rr.Header().Get("X-Robots-Tag"))
	assert.Equal(t, "10", rr.Header().Get("X-RateLimit-Limit"))
	assert.Contains(t, rr.Body.String(), `"Public Bug"`)

	attReq := httptest.NewRequest(http.MethodGet, "/api/public/tasks/valid/attachments/"+attID.String(), nil)
	attRR := httptest.NewRecorder()
	r.ServeHTTP(attRR, attReq)
	assert.Equal(t, http.StatusOK, attRR.Code)
	assert.Contains(t, attRR.Body.String(), `https://cdn.example.com/uploads/log.txt`)

	for _, tok := range []string{"revoked", "expired", "missing"} {
		badReq := httptest.NewRequest(http.MethodGet, "/api/public/tasks/"+tok, nil)
		badRR := httptest.NewRecorder()
		r.ServeHTTP(badRR, badReq)
		assert.Equal(t, http.StatusNotFound, badRR.Code)
	}

	// Exhaust 5/min attachment limit (1 used above, 4 more allowed, 6th hits 429).
	for i := 0; i < 4; i++ {
		okReq := httptest.NewRequest(http.MethodGet, "/api/public/tasks/valid/attachments/"+attID.String(), nil)
		okRR := httptest.NewRecorder()
		r.ServeHTTP(okRR, okReq)
		assert.Equal(t, http.StatusOK, okRR.Code)
	}
	overReq := httptest.NewRequest(http.MethodGet, "/api/public/tasks/valid/attachments/"+attID.String(), nil)
	overRR := httptest.NewRecorder()
	r.ServeHTTP(overRR, overReq)
	assert.Equal(t, http.StatusTooManyRequests, overRR.Code)
	assert.Contains(t, overRR.Body.String(), `"Too Many Requests"`)
}

func TestPublicStatsCappingChartsKanbanAndCommentTree(t *testing.T) {
	projID, wfID, stID := uuid.New(), uuid.New(), uuid.New()
	rootID, replyID, authID := uuid.New(), uuid.New(), uuid.New()
	repo := &fakePublicRepo{
		projBySlugs: map[string]*sqlc.GetPublicProjectBySlugsRow{
			"acme/web": {
				ID: projID, Name: "Web", Slug: "web", WorkflowID: wfID,
				TasksCount: 240, DoneCount: 120, SprintsCount: 2,
			},
		},
		projBySlug: map[string]*sqlc.GetPublicProjectBySlugOnlyRow{
			"web": {ID: projID, Name: "Web", Slug: "web", WorkflowID: wfID, Visibility: sqlc.ProjectVisibilityPUBLIC},
		},
		kpis: map[uuid.UUID]*sqlc.GetPublicProjectKPIsRow{
			projID: {TotalTasks: 10, CompletedTasks: 5, ActiveSprints: 1, TotalBugs: 4, ResolvedBugs: 2},
		},
		statuses: map[uuid.UUID][]sqlc.ListPublicProjectStatusesByWorkflowRow{
			wfID: {{ID: stID, Name: "To Do", Color: "#999", Category: sqlc.StatusCategoryTODO}},
		},
		tasks: []sqlc.ListPublicProjectTasksRow{
			{ID: uuid.New(), Title: "Landing page", StatusID: pgtype.UUID{Bytes: stID, Valid: true}},
		},
		comments: []sqlc.ListPublicTaskCommentsRow{
			{ID: rootID, Content: "Parent", AuthorID: authID, FirstName: "Sam", LastName: "Lee"},
			{ID: replyID, Content: "Reply", ParentCommentID: pgtype.UUID{Bytes: rootID, Valid: true}, AuthorID: authID, FirstName: "Sam", LastName: "Lee"},
		},
	}

	svc := NewService(repo, nil)
	proj, err := svc.GetPublicProject(context.Background(), "acme", "web")
	require.NoError(t, err)
	require.NotNil(t, proj.Stats)
	assert.Equal(t, int64(100), proj.Stats.TaskCount)
	assert.Equal(t, 50.0, proj.Stats.CompletionRate)

	charts, err := svc.GetProjectCharts(context.Background(), "web", []string{"kpi-metrics"})
	require.NoError(t, err)
	kpi, ok := charts["kpi-metrics"].(KPIMetrics)
	require.True(t, ok)
	assert.Equal(t, 50.0, kpi.CompletionRate)

	kanban, err := svc.GetPublicTasksKanban(context.Background(), "acme", "web")
	require.NoError(t, err)
	require.Len(t, kanban.Columns, 1)
	require.Len(t, kanban.Columns[0].Tasks, 1)

	comments, err := svc.GetPublicTaskComments(context.Background(), uuid.New())
	require.NoError(t, err)
	require.Len(t, comments, 1)
	require.Len(t, comments[0].Replies, 1)
	assert.Equal(t, 1, comments[0].Count.Replies)
}
