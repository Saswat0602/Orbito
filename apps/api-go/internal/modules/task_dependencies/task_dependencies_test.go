package task_dependencies

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	tasksByKey map[string]uuid.UUID
	existsMap  map[[2]uuid.UUID]bool
	graph      map[uuid.UUID][]uuid.UUID
	records    map[uuid.UUID]sqlc.GetTaskDependencyDetailsByIDRow
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		tasksByKey: make(map[string]uuid.UUID),
		existsMap:  make(map[[2]uuid.UUID]bool),
		graph:      make(map[uuid.UUID][]uuid.UUID),
		records:    make(map[uuid.UUID]sqlc.GetTaskDependencyDetailsByIDRow),
	}
}

func (m *mockRepo) FindTaskID(_ context.Context, identifier string) (uuid.UUID, bool, error) {
	if id, ok := m.tasksByKey[identifier]; ok {
		return id, true, nil
	}
	return uuid.Nil, false, nil
}

func (m *mockRepo) CheckExists(_ context.Context, dependentTaskID, blockingTaskID uuid.UUID) (bool, error) {
	return m.existsMap[[2]uuid.UUID{dependentTaskID, blockingTaskID}], nil
}

func (m *mockRepo) ListBlockingIDs(_ context.Context, dependentTaskID uuid.UUID) ([]uuid.UUID, error) {
	return m.graph[dependentTaskID], nil
}

func (m *mockRepo) Create(_ context.Context, id uuid.UUID, depType sqlc.DependencyType, dependentTaskID, blockingTaskID uuid.UUID, createdBy *uuid.UUID) (uuid.UUID, error) {
	now := pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
	var cb pgtype.UUID
	if createdBy != nil {
		cb = pgtype.UUID{Bytes: *createdBy, Valid: true}
	}
	m.existsMap[[2]uuid.UUID{dependentTaskID, blockingTaskID}] = true
	m.graph[dependentTaskID] = append(m.graph[dependentTaskID], blockingTaskID)
	m.records[id] = sqlc.GetTaskDependencyDetailsByIDRow{
		ID:              id,
		Type:            depType,
		DependentTaskID: dependentTaskID,
		BlockingTaskID:  blockingTaskID,
		CreatedByID:     cb,
		UpdatedByID:     cb,
		CreatedAt:       now,
		UpdatedAt:       now,
		DepTitle:        "Dep Task",
		DepSlug:         "DEP-1",
		DepProjectID:    uuid.New(),
		DepStatusName:   "To Do",
		BlkTitle:        "Blk Task",
		BlkSlug:         "BLK-1",
		BlkProjectID:    uuid.New(),
		BlkStatusName:   "In Progress",
	}
	return id, nil
}

func (m *mockRepo) GetByID(_ context.Context, id uuid.UUID) (*sqlc.GetTaskDependencyDetailsByIDRow, error) {
	r, ok := m.records[id]
	if !ok {
		return nil, nil
	}
	return &r, nil
}

func (m *mockRepo) ListAll(_ context.Context, _ *uuid.UUID) ([]sqlc.ListTaskDependenciesRow, error) {
	out := make([]sqlc.ListTaskDependenciesRow, 0, len(m.records))
	for _, r := range m.records {
		out = append(out, sqlc.ListTaskDependenciesRow(r))
	}
	return out, nil
}

func (m *mockRepo) ListDependsOn(_ context.Context, dependentTaskID uuid.UUID) ([]sqlc.ListDependenciesTaskDependsOnRow, error) {
	var out []sqlc.ListDependenciesTaskDependsOnRow
	for _, r := range m.records {
		if r.DependentTaskID == dependentTaskID {
			out = append(out, sqlc.ListDependenciesTaskDependsOnRow{
				ID:              r.ID,
				Type:            r.Type,
				DependentTaskID: r.DependentTaskID,
				BlockingTaskID:  r.BlockingTaskID,
				CreatedAt:       r.CreatedAt,
				UpdatedAt:       r.UpdatedAt,
				BlkTitle:        r.BlkTitle,
				BlkSlug:         r.BlkSlug,
				BlkStatusName:   r.BlkStatusName,
			})
		}
	}
	return out, nil
}

func (m *mockRepo) ListBlocks(_ context.Context, blockingTaskID uuid.UUID) ([]sqlc.ListDependenciesTaskBlocksRow, error) {
	var out []sqlc.ListDependenciesTaskBlocksRow
	for _, r := range m.records {
		if r.BlockingTaskID == blockingTaskID {
			out = append(out, sqlc.ListDependenciesTaskBlocksRow{
				ID:              r.ID,
				Type:            r.Type,
				DependentTaskID: r.DependentTaskID,
				BlockingTaskID:  r.BlockingTaskID,
				CreatedAt:       r.CreatedAt,
				UpdatedAt:       r.UpdatedAt,
				DepTitle:        r.DepTitle,
				DepSlug:         r.DepSlug,
				DepStatusName:   r.DepStatusName,
			})
		}
	}
	return out, nil
}

func (m *mockRepo) Update(_ context.Context, id uuid.UUID, depType *sqlc.DependencyType, dependentTaskID, blockingTaskID *uuid.UUID) (uuid.UUID, error) {
	r := m.records[id]
	if depType != nil {
		r.Type = *depType
	}
	if dependentTaskID != nil {
		r.DependentTaskID = *dependentTaskID
	}
	if blockingTaskID != nil {
		r.BlockingTaskID = *blockingTaskID
	}
	m.records[id] = r
	return id, nil
}

func (m *mockRepo) DeleteByID(_ context.Context, id uuid.UUID) (int64, error) {
	if _, ok := m.records[id]; !ok {
		return 0, nil
	}
	delete(m.records, id)
	return 1, nil
}

func (m *mockRepo) DeleteByPair(_ context.Context, dependentTaskID, blockingTaskID uuid.UUID) (int64, error) {
	for id, r := range m.records {
		if r.DependentTaskID == dependentTaskID && r.BlockingTaskID == blockingTaskID {
			delete(m.records, id)
			return 1, nil
		}
	}
	return 0, nil
}

func TestTaskDependencies_CreateSelfAndCycleDetection(t *testing.T) {
	repo := newMockRepo()
	tA, tB, tC := uuid.New(), uuid.New(), uuid.New()
	callerID := uuid.New()
	repo.tasksByKey[tA.String()] = tA
	repo.tasksByKey["TASK-A"] = tA
	repo.tasksByKey[tB.String()] = tB
	repo.tasksByKey[tC.String()] = tC

	svc := NewService(repo)
	ctx := context.Background()

	// Self-dependency check (same identifier and slug->UUID resolution)
	_, err := svc.Create(ctx, CreateRequest{DependentTaskID: "TASK-A", BlockingTaskID: tA.String()}, callerID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "depend on itself")

	// A depends on B, B depends on C
	depAB, err := svc.Create(ctx, CreateRequest{DependentTaskID: tA.String(), BlockingTaskID: tB.String()}, callerID)
	require.NoError(t, err)
	assert.Equal(t, "BLOCKS", depAB.Type)

	_, err = svc.Create(ctx, CreateRequest{DependentTaskID: tB.String(), BlockingTaskID: tC.String()}, callerID)
	require.NoError(t, err)

	// Duplicate fails with Conflict
	_, err = svc.Create(ctx, CreateRequest{DependentTaskID: tA.String(), BlockingTaskID: tB.String()}, callerID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Indirect cycle: C depends on A -> fails!
	_, err = svc.Create(ctx, CreateRequest{DependentTaskID: tC.String(), BlockingTaskID: tA.String()}, callerID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestTaskDependencies_HTTPHandlers(t *testing.T) {
	repo := newMockRepo()
	tA, tB := uuid.New(), uuid.New()
	callerID := uuid.New()
	repo.tasksByKey[tA.String()] = tA
	repo.tasksByKey[tB.String()] = tB

	h := NewHandler(NewService(repo))
	authMw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := ctxhelper.WithUserRole(ctxhelper.WithUserID(r.Context(), callerID.String()), "MEMBER")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
	router := Routes(h, authMw)

	body, _ := json.Marshal(map[string]string{
		"dependentTaskId": tA.String(),
		"blockingTaskId":  tB.String(),
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	var created DependencyResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &created))

	// Get task dependencies
	reqTask := httptest.NewRequest(http.MethodGet, "/task/"+tA.String(), nil)
	rrTask := httptest.NewRecorder()
	router.ServeHTTP(rrTask, reqTask)
	assert.Equal(t, http.StatusOK, rrTask.Code)

	// Delete dependency returns 204
	reqDel := httptest.NewRequest(http.MethodDelete, "/"+created.ID, nil)
	rrDel := httptest.NewRecorder()
	router.ServeHTTP(rrDel, reqDel)
	assert.Equal(t, http.StatusNoContent, rrDel.Code)
}
