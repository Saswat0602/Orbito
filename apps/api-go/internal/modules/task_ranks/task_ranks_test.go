package task_ranks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	ranks      map[uuid.UUID]float64
	rebalanced bool
	maxRank    float64
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		ranks: make(map[uuid.UUID]float64),
	}
}

func (m *mockRepo) ListIDsForRebalance(_ context.Context, _ sqlc.ScopeType, _ uuid.UUID, _ sqlc.ViewType) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(m.ranks))
	for id := range m.ranks {
		ids = append(ids, id)
	}
	return ids, nil
}

func (m *mockRepo) RebalanceRowsTx(_ context.Context, ids []uuid.UUID) error {
	m.rebalanced = true
	for i, id := range ids {
		m.ranks[id] = float64(i + 1)
	}
	return nil
}

func (m *mockRepo) ListUnrankedTopTasks(_ context.Context, _ sqlc.ScopeType, _ uuid.UUID, _ sqlc.ViewType) ([]uuid.UUID, error) {
	return nil, nil
}

func (m *mockRepo) GetMaxRankSingleScope(_ context.Context, _ sqlc.ScopeType, _ uuid.UUID, _ sqlc.ViewType) (float64, error) {
	return m.maxRank, nil
}

func (m *mockRepo) InsertRanksSkipConflictTx(_ context.Context, _ []sqlc.InsertTaskRankSkipConflictParams) error {
	return nil
}

func (m *mockRepo) GetTaskRank(_ context.Context, taskID uuid.UUID, _ sqlc.ScopeType, _ uuid.UUID, _ sqlc.ViewType) (*float64, error) {
	if v, ok := m.ranks[taskID]; ok {
		return &v, nil
	}
	return nil, nil
}

func (m *mockRepo) FindTrueAfterRank(_ context.Context, _ sqlc.ScopeType, _ uuid.UUID, _ sqlc.ViewType, _ float64) (*float64, error) {
	return nil, nil
}

func (m *mockRepo) FindTrueBeforeRank(_ context.Context, _ sqlc.ScopeType, _ uuid.UUID, _ sqlc.ViewType, _ float64) (*float64, error) {
	return nil, nil
}

func (m *mockRepo) UpsertTaskRank(_ context.Context, taskID uuid.UUID, _ sqlc.ScopeType, _ uuid.UUID, _ sqlc.ViewType, rank float64) error {
	m.ranks[taskID] = rank
	return nil
}

func TestComputeRankAndRebalance(t *testing.T) {
	assert.Equal(t, 1.0, ComputeRank(nil, nil))
	top := 2.0
	assert.Equal(t, 1.0, ComputeRank(nil, &top))
	bottom := 5.0
	assert.Equal(t, 6.0, ComputeRank(&bottom, nil))
	a, b := 1.0, 2.0
	assert.Equal(t, 1.5, ComputeRank(&a, &b))

	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	scopeID := uuid.New().String()
	t1, t2, t3 := uuid.New(), uuid.New(), uuid.New()
	repo.ranks[t1] = 1.0
	repo.ranks[t2] = 1.0001 // closer than 0.0001 midpoint spacing -> triggers rebalance

	afterStr := t1.String()
	beforeStr := t2.String()
	err := svc.Reorder(ctx, t3, ReorderRequest{
		ScopeType:    "PROJECT",
		ScopeID:      scopeID,
		ViewType:     "BOARD",
		AfterTaskID:  &afterStr,
		BeforeTaskID: &beforeStr,
	})
	require.NoError(t, err)
	assert.True(t, repo.rebalanced)
}

func TestTaskRanksHandler_Routes(t *testing.T) {
	h := NewHandler(NewService(newMockRepo()))
	router := Routes(h, func(next http.Handler) http.Handler { return next })

	req := httptest.NewRequest(http.MethodGet, "/rebalance", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "Rebalance utility would go here")

	badReq := httptest.NewRequest(http.MethodPatch, "/"+uuid.New().String()+"/reorder", strings.NewReader(`{}`))
	badReq.Header.Set("Content-Type", "application/json")
	rr2 := httptest.NewRecorder()
	router.ServeHTTP(rr2, badReq)
	assert.Equal(t, http.StatusBadRequest, rr2.Code)
}
