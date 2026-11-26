package sprints

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
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockEmitter struct {
	started   int
	completed int
}

func (e *mockEmitter) EmitSprintStarted(string, any)   { e.started++ }
func (e *mockEmitter) EmitSprintCompleted(string, any) { e.completed++ }

type mockRepo struct {
	projByID   map[uuid.UUID]*sqlc.GetProjectAccessInfoForSprintRow
	projBySlug map[string]*sqlc.GetProjectAccessInfoForSprintBySlugRow
	sprints    map[uuid.UUID]*sqlc.GetSprintDetailsByIDRow
	logged     int
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		projByID:   make(map[uuid.UUID]*sqlc.GetProjectAccessInfoForSprintRow),
		projBySlug: make(map[string]*sqlc.GetProjectAccessInfoForSprintBySlugRow),
		sprints:    make(map[uuid.UUID]*sqlc.GetSprintDetailsByIDRow),
	}
}

func (m *mockRepo) GetProjectAccessByID(_ context.Context, projectID, _ uuid.UUID) (*sqlc.GetProjectAccessInfoForSprintRow, error) {
	if r, ok := m.projByID[projectID]; ok {
		return r, nil
	}
	return nil, pgx.ErrNoRows
}

func (m *mockRepo) GetProjectAccessBySlug(_ context.Context, slug string, _ uuid.UUID) (*sqlc.GetProjectAccessInfoForSprintBySlugRow, error) {
	if r, ok := m.projBySlug[slug]; ok {
		return r, nil
	}
	return nil, pgx.ErrNoRows
}

func (m *mockRepo) CheckSlugExists(_ context.Context, projectID uuid.UUID, slug string, _ *uuid.UUID) (bool, error) {
	for _, s := range m.sprints {
		if s.ProjectID == projectID && s.Slug.String == slug {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockRepo) UpdateSlugByID(_ context.Context, sprintID uuid.UUID, slug string) error {
	if s, ok := m.sprints[sprintID]; ok {
		s.Slug = pgtype.Text{String: slug, Valid: true}
	}
	return nil
}

func (m *mockRepo) HasActiveSprint(_ context.Context, projectID uuid.UUID, excludeID *uuid.UUID) (bool, error) {
	for _, s := range m.sprints {
		if s.ProjectID == projectID && s.Status == sqlc.SprintStatusACTIVE && (excludeID == nil || s.ID != *excludeID) {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockRepo) Create(_ context.Context, arg sqlc.CreateSprintRecordParams) (uuid.UUID, error) {
	now := pgtype.Timestamp{Time: time.Now(), Valid: true}
	m.sprints[arg.ID] = &sqlc.GetSprintDetailsByIDRow{
		ID:             arg.ID,
		Name:           arg.Name,
		Slug:           arg.Slug,
		Goal:           arg.Goal,
		Status:         arg.Status,
		StartDate:      arg.StartDate,
		EndDate:        arg.EndDate,
		ProjectID:      arg.ProjectID,
		CreatedAt:      now,
		UpdatedAt:      now,
		OrganizationID: uuid.New(),
		ProjectName:    "Core",
		ProjectSlug:    "core",
		WorkspaceName:  "WS",
		WorkspaceSlug:  "ws",
		OrgName:        "Org",
		OrgSlug:        "org",
	}
	return arg.ID, nil
}

func (m *mockRepo) GetDetailsByID(_ context.Context, id uuid.UUID) (*sqlc.GetSprintDetailsByIDRow, error) {
	if r, ok := m.sprints[id]; ok {
		return r, nil
	}
	return nil, pgx.ErrNoRows
}

func (m *mockRepo) GetByProjectAndSlug(_ context.Context, _, sprintSlug string) (*sqlc.GetSprintByProjectAndSlugRow, error) {
	for _, s := range m.sprints {
		if s.Slug.String == sprintSlug {
			return &sqlc.GetSprintByProjectAndSlugRow{
				ID:          s.ID,
				Name:        s.Name,
				Slug:        s.Slug,
				Status:      s.Status,
				ProjectID:   s.ProjectID,
				ProjectName: s.ProjectName,
				ProjectSlug: s.ProjectSlug,
				CreatedAt:   s.CreatedAt,
				UpdatedAt:   s.UpdatedAt,
			}, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (m *mockRepo) GetActiveByProjectID(_ context.Context, projectID uuid.UUID) (*sqlc.GetActiveSprintByProjectIDRow, error) {
	for _, s := range m.sprints {
		if s.ProjectID == projectID && s.Status == sqlc.SprintStatusACTIVE {
			return &sqlc.GetActiveSprintByProjectIDRow{
				ID:        s.ID,
				Name:      s.Name,
				Slug:      s.Slug,
				Status:    s.Status,
				ProjectID: s.ProjectID,
				CreatedAt: s.CreatedAt,
				UpdatedAt: s.UpdatedAt,
			}, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (m *mockRepo) ListFiltered(_ context.Context, _ pgtype.UUID, _, _ pgtype.Text) ([]sqlc.ListSprintsFilteredRow, error) {
	return []sqlc.ListSprintsFilteredRow{}, nil
}

func (m *mockRepo) ListSprintTasks(_ context.Context, _ uuid.UUID) ([]sqlc.ListSprintTasksBySprintIDRow, error) {
	return nil, nil
}

func (m *mockRepo) ListTaskAssignees(_ context.Context, _ []uuid.UUID) ([]sqlc.ListSprintTaskAssigneesRow, error) {
	return nil, nil
}

func (m *mockRepo) ListTaskReporters(_ context.Context, _ []uuid.UUID) ([]sqlc.ListSprintTaskReportersRow, error) {
	return nil, nil
}

func (m *mockRepo) UpdateByID(_ context.Context, arg sqlc.UpdateSprintByIDParams) (uuid.UUID, error) {
	s, ok := m.sprints[arg.ID]
	if !ok {
		return uuid.Nil, pgx.ErrNoRows
	}
	if arg.Status.Valid {
		s.Status = arg.Status.SprintStatus
	}
	if arg.Name.Valid {
		s.Name = arg.Name.String
	}
	return arg.ID, nil
}

func (m *mockRepo) DeleteByID(_ context.Context, id uuid.UUID) (int64, error) {
	if _, ok := m.sprints[id]; ok {
		delete(m.sprints, id)
		return 1, nil
	}
	return 0, nil
}

func (m *mockRepo) ArchiveByID(_ context.Context, id uuid.UUID) (int64, error) {
	if s, ok := m.sprints[id]; ok {
		s.Archive = true
		return 1, nil
	}
	return 0, nil
}

func (m *mockRepo) LogActivity(_ context.Context, _ sqlc.CreateSprintActivityLogParams) error {
	m.logged++
	return nil
}

func TestSprintsService_LifecycleAndRBAC(t *testing.T) {
	repo := newMockRepo()
	emitter := &mockEmitter{}
	svc := NewService(repo, emitter)
	ctx := context.Background()

	ownerID := uuid.New()
	memberID := uuid.New()
	projID := uuid.New()
	orgID := uuid.New()

	repo.projByID[projID] = &sqlc.GetProjectAccessInfoForSprintRow{
		ID:             projID,
		Name:           "Core",
		Slug:           "core",
		Visibility:     sqlc.ProjectVisibilityPRIVATE,
		OrganizationID: orgID,
		OrgOwnerID:     ownerID,
		ProjectRole:    sqlc.NullRole{Role: sqlc.RoleMEMBER, Valid: true},
	}

	pid := projID.String()
	startDate := "2026-12-01T09:00:00.000Z"
	endDate := "2026-12-14T17:00:00.000Z"

	_, err := svc.Create(ctx, CreateSprintRequest{Name: "Sprint 1", ProjectID: &pid}, memberID, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Only managers and owners")

	created, err := svc.Create(ctx, CreateSprintRequest{
		Name:      "Sprint 1",
		ProjectID: &pid,
		StartDate: &startDate,
		EndDate:   &endDate,
	}, ownerID, false)
	require.NoError(t, err)
	assert.Equal(t, "sprint-1", *created.Slug)
	assert.Equal(t, 1, repo.logged)

	sprintID := uuid.MustParse(created.ID)
	started, err := svc.StartSprint(ctx, sprintID, ownerID, false)
	require.NoError(t, err)
	assert.Equal(t, "ACTIVE", started.Status)
	assert.Equal(t, 1, emitter.started)

	// Cannot delete active sprint
	err = svc.Remove(ctx, sprintID, ownerID, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "active sprint")

	completed, err := svc.CompleteSprint(ctx, sprintID, ownerID, false)
	require.NoError(t, err)
	assert.Equal(t, "COMPLETED", completed.Status)
	assert.Equal(t, 1, emitter.completed)
}

func TestSprintsHandler_Validation(t *testing.T) {
	h := NewHandler(NewService(newMockRepo(), nil))
	router := Routes(h, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := ctxhelper.WithUserID(r.Context(), uuid.New().String())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":""}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
