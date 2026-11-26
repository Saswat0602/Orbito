package task_statuses

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

type mockRepo struct {
	wfAuth            map[uuid.UUID]*sqlc.GetWorkflowOrgOwnerAndUserRoleRow
	statusAuth        map[uuid.UUID]*sqlc.GetTaskStatusAuthByIDRow
	softDeleted       *sqlc.TaskStatus
	statuses          map[uuid.UUID]*sqlc.GetTaskStatusFullByIDRow
	activeFull        map[uuid.UUID]*sqlc.GetActiveTaskStatusFullByIDRow
	projWF            map[uuid.UUID]*sqlc.GetProjectWorkflowForStatusCreateRow
	conflict          bool
	softDeletedCalled bool
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		wfAuth:     make(map[uuid.UUID]*sqlc.GetWorkflowOrgOwnerAndUserRoleRow),
		statusAuth: make(map[uuid.UUID]*sqlc.GetTaskStatusAuthByIDRow),
		statuses:   make(map[uuid.UUID]*sqlc.GetTaskStatusFullByIDRow),
		activeFull: make(map[uuid.UUID]*sqlc.GetActiveTaskStatusFullByIDRow),
		projWF:     make(map[uuid.UUID]*sqlc.GetProjectWorkflowForStatusCreateRow),
	}
}

func (m *mockRepo) GetWorkflowAuth(_ context.Context, workflowID, _ uuid.UUID) (*sqlc.GetWorkflowOrgOwnerAndUserRoleRow, error) {
	if r, ok := m.wfAuth[workflowID]; ok {
		return r, nil
	}
	return nil, pgx.ErrNoRows
}

func (m *mockRepo) GetStatusAuth(_ context.Context, statusID, _ uuid.UUID) (*sqlc.GetTaskStatusAuthByIDRow, error) {
	if r, ok := m.statusAuth[statusID]; ok {
		return r, nil
	}
	return nil, pgx.ErrNoRows
}

func (m *mockRepo) GetProjectWorkflow(_ context.Context, projectID uuid.UUID) (*sqlc.GetProjectWorkflowForStatusCreateRow, error) {
	if r, ok := m.projWF[projectID]; ok {
		return r, nil
	}
	return nil, pgx.ErrNoRows
}

func (m *mockRepo) FindSoftDeleted(_ context.Context, _ uuid.UUID, _ string) (*sqlc.TaskStatus, error) {
	return m.softDeleted, nil
}

func (m *mockRepo) RestoreOnCreate(_ context.Context, id, _ uuid.UUID) (*sqlc.TaskStatus, error) {
	return &sqlc.TaskStatus{ID: id, Name: "QA", Color: "#111111", Category: sqlc.StatusCategoryINPROGRESS, Position: 2}, nil
}

func (m *mockRepo) CountWorkflowStatuses(_ context.Context, _ uuid.UUID) (int64, error) {
	return int64(len(m.statuses)), nil
}

func (m *mockRepo) CreateNode(_ context.Context, arg sqlc.CreateTaskStatusNodeParams) (*sqlc.TaskStatus, error) {
	now := pgtype.Timestamp{Time: time.Now(), Valid: true}
	row := &sqlc.GetTaskStatusFullByIDRow{
		ID:         arg.ID,
		Name:       arg.Name,
		Color:      arg.Color,
		Category:   arg.Category,
		Position:   arg.Position,
		IsDefault:  arg.IsDefault,
		WorkflowID: arg.WorkflowID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	m.statuses[arg.ID] = row
	return &sqlc.TaskStatus{
		ID:         arg.ID,
		Name:       arg.Name,
		Color:      arg.Color,
		Category:   arg.Category,
		Position:   arg.Position,
		IsDefault:  arg.IsDefault,
		WorkflowID: arg.WorkflowID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (m *mockRepo) GetFullByID(_ context.Context, id uuid.UUID) (*sqlc.GetTaskStatusFullByIDRow, error) {
	if r, ok := m.statuses[id]; ok {
		return r, nil
	}
	return nil, pgx.ErrNoRows
}

func (m *mockRepo) GetActiveFullByID(_ context.Context, id uuid.UUID) (*sqlc.GetActiveTaskStatusFullByIDRow, error) {
	if r, ok := m.activeFull[id]; ok {
		return r, nil
	}
	return nil, pgx.ErrNoRows
}

func (m *mockRepo) ListFiltered(_ context.Context, _ pgtype.UUID, _ pgtype.UUID) ([]sqlc.ListTaskStatusesFilteredRow, error) {
	return []sqlc.ListTaskStatusesFilteredRow{}, nil
}

func (m *mockRepo) ListByProject(_ context.Context, _ uuid.UUID) ([]sqlc.ListTaskStatusesByProjectIDRow, error) {
	return []sqlc.ListTaskStatusesByProjectIDRow{}, nil
}

func (m *mockRepo) ListFromTransitions(_ context.Context, _ uuid.UUID) ([]sqlc.ListFromTransitionsForStatusRow, error) {
	return nil, nil
}

func (m *mockRepo) ListToTransitions(_ context.Context, _ uuid.UUID) ([]sqlc.ListToTransitionsForStatusRow, error) {
	return nil, nil
}

func (m *mockRepo) ListDeleted(_ context.Context, _ pgtype.UUID) ([]sqlc.ListDeletedTaskStatusesRow, error) {
	return []sqlc.ListDeletedTaskStatusesRow{}, nil
}

func (m *mockRepo) UpdateByID(_ context.Context, arg sqlc.UpdateTaskStatusByIDParams) (uuid.UUID, error) {
	r, ok := m.statuses[arg.ID]
	if !ok {
		return uuid.Nil, pgx.ErrNoRows
	}
	if arg.Name.Valid {
		r.Name = arg.Name.String
	}
	return arg.ID, nil
}

func (m *mockRepo) UpdatePositionsTx(_ context.Context, updates []PositionUpdate, _ uuid.UUID) ([]sqlc.GetTaskStatusFullByIDRow, error) {
	res := make([]sqlc.GetTaskStatusFullByIDRow, 0, len(updates))
	for _, u := range updates {
		if r, ok := m.statuses[u.ID]; ok {
			r.Position = u.Position
			res = append(res, *r)
		}
	}
	return res, nil
}

func (m *mockRepo) SoftDeleteTx(_ context.Context, _ uuid.UUID, _ uuid.UUID, _ int32, _ uuid.UUID) error {
	m.softDeletedCalled = true
	return nil
}

func (m *mockRepo) CheckActiveNameConflict(_ context.Context, _ uuid.UUID, _ string, _ uuid.UUID) (bool, error) {
	return m.conflict, nil
}

func (m *mockRepo) RestoreByID(_ context.Context, id, _ uuid.UUID) (uuid.UUID, error) {
	if _, ok := m.statuses[id]; ok {
		return id, nil
	}
	return uuid.Nil, pgx.ErrNoRows
}

func TestTaskStatusesService_RBACAndLifecycle(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	ownerID := uuid.New()
	memberID := uuid.New()
	outsiderID := uuid.New()
	wfID := uuid.New()
	repo.wfAuth[wfID] = &sqlc.GetWorkflowOrgOwnerAndUserRoleRow{
		WorkflowID: wfID,
		OwnerID:    ownerID,
	}

	pos := int32(1)
	_, err := svc.Create(ctx, CreateTaskStatusRequest{
		Name: "QA", Color: "#AABBCC", Category: "IN_PROGRESS", Position: &pos, WorkflowID: wfID.String(),
	}, outsiderID, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a member")

	created, err := svc.Create(ctx, CreateTaskStatusRequest{
		Name: "QA", Color: "#AABBCC", Category: "IN_PROGRESS", Position: &pos, WorkflowID: wfID.String(),
	}, ownerID, false)
	require.NoError(t, err)
	assert.Equal(t, "QA", created.Name)

	statusID := uuid.MustParse(created.ID)
	repo.statusAuth[statusID] = &sqlc.GetTaskStatusAuthByIDRow{
		ID:         statusID,
		WorkflowID: wfID,
		Name:       "QA",
		Position:   1,
		OwnerID:    ownerID,
		MemberRole: sqlc.NullRole{Role: sqlc.RoleMEMBER, Valid: true},
	}

	newName := "Verified"
	_, err = svc.Update(ctx, statusID, UpdateTaskStatusRequest{Name: &newName}, memberID, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MANAGER")

	updated, err := svc.Update(ctx, statusID, UpdateTaskStatusRequest{Name: &newName}, ownerID, false)
	require.NoError(t, err)
	assert.Equal(t, "Verified", updated.Name)

	err = svc.Remove(ctx, statusID, ownerID, false)
	require.NoError(t, err)
	assert.True(t, repo.softDeletedCalled)
}

func TestTaskStatusesHandler_ValidationAndRoutes(t *testing.T) {
	repo := newMockRepo()
	h := NewHandler(NewService(repo))
	router := Routes(h, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := ctxhelper.WithUserID(r.Context(), uuid.New().String())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/not-a-uuid", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	badBody := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":""}`))
	badBody.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, badBody)
	assert.Equal(t, http.StatusBadRequest, rr2.Code)
}
