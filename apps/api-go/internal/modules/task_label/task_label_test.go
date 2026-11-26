package task_label

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
	tasks    map[string]sqlc.FindTaskForTaskLabelRow
	labels   map[uuid.UUID]sqlc.GetLabelWorkspaceForTaskLabelRow
	assigned map[[2]uuid.UUID]sqlc.GetTaskLabelPairDetailsRow
	projIDs  []uuid.UUID
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		tasks:    make(map[string]sqlc.FindTaskForTaskLabelRow),
		labels:   make(map[uuid.UUID]sqlc.GetLabelWorkspaceForTaskLabelRow),
		assigned: make(map[[2]uuid.UUID]sqlc.GetTaskLabelPairDetailsRow),
	}
}

func (m *mockRepo) FindTask(_ context.Context, identifier string, _ uuid.UUID) (*sqlc.FindTaskForTaskLabelRow, error) {
	if r, ok := m.tasks[identifier]; ok {
		return &r, nil
	}
	return nil, nil
}

func (m *mockRepo) GetLabelWorkspace(_ context.Context, labelID, _ uuid.UUID) (*sqlc.GetLabelWorkspaceForTaskLabelRow, error) {
	if l, ok := m.labels[labelID]; ok {
		return &l, nil
	}
	return nil, nil
}

func (m *mockRepo) FindLabelsInWorkspace(_ context.Context, labelIDs []uuid.UUID, workspaceID uuid.UUID) ([]uuid.UUID, error) {
	var out []uuid.UUID
	for _, id := range labelIDs {
		if l, ok := m.labels[id]; ok && (l.ProjectWorkspaceID == workspaceID || l.InTargetWs) {
			out = append(out, id)
		}
	}
	return out, nil
}

func (m *mockRepo) CheckAssignmentExists(_ context.Context, taskID, labelID uuid.UUID) (bool, error) {
	_, ok := m.assigned[[2]uuid.UUID{taskID, labelID}]
	return ok, nil
}

func (m *mockRepo) AssignSingleTx(_ context.Context, taskID, labelID, _, userID uuid.UUID) (*sqlc.GetTaskLabelPairDetailsRow, error) {
	now := pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
	lbl := m.labels[labelID]
	row := sqlc.GetTaskLabelPairDetailsRow{
		TaskID: taskID, LabelID: labelID,
		CreatedByID: pgtype.UUID{Bytes: userID, Valid: true}, UpdatedByID: pgtype.UUID{Bytes: userID, Valid: true},
		CreatedAt: now, UpdatedAt: now,
		LabelName: lbl.Name, LabelColor: lbl.Color, TaskTitle: "Sample", TaskSlug: "OR-10",
	}
	m.assigned[[2]uuid.UUID{taskID, labelID}] = row
	return &row, nil
}

func (m *mockRepo) ReplaceMultipleTx(_ context.Context, taskID uuid.UUID, labelIDs []uuid.UUID, wsID, userID uuid.UUID) ([]sqlc.GetTaskLabelPairDetailsRow, error) {
	for k := range m.assigned {
		if k[0] == taskID {
			delete(m.assigned, k)
		}
	}
	var out []sqlc.GetTaskLabelPairDetailsRow
	for _, lid := range labelIDs {
		r, _ := m.AssignSingleTx(context.Background(), taskID, lid, wsID, userID)
		out = append(out, *r)
	}
	return out, nil
}

func (m *mockRepo) ListAccessibleProjectIDs(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
	return m.projIDs, nil
}

func (m *mockRepo) ListByProjectIDs(_ context.Context, _ []uuid.UUID) ([]sqlc.ListTaskLabelsByProjectIDsRow, error) {
	out := make([]sqlc.ListTaskLabelsByProjectIDsRow, 0, len(m.assigned))
	for _, r := range m.assigned {
		out = append(out, sqlc.ListTaskLabelsByProjectIDsRow(r))
	}
	return out, nil
}

func (m *mockRepo) GetPairDetails(_ context.Context, taskID, labelID uuid.UUID) (*sqlc.GetTaskLabelPairDetailsRow, error) {
	if r, ok := m.assigned[[2]uuid.UUID{taskID, labelID}]; ok {
		return &r, nil
	}
	return nil, nil
}

func (m *mockRepo) DeletePair(_ context.Context, taskID, labelID uuid.UUID) (int64, error) {
	k := [2]uuid.UUID{taskID, labelID}
	if _, ok := m.assigned[k]; !ok {
		return 0, nil
	}
	delete(m.assigned, k)
	return 1, nil
}

func TestTaskLabels_Workflow(t *testing.T) {
	repo := newMockRepo()
	taskID, projID, wsID, lbl1, lbl2, userID := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	repo.projIDs = []uuid.UUID{projID}
	repo.tasks[taskID.String()] = sqlc.FindTaskForTaskLabelRow{
		TaskID: taskID, TaskSlug: "OR-10", ProjectID: projID, WorkspaceID: wsID, IsWsMember: true,
	}
	repo.labels[lbl1] = sqlc.GetLabelWorkspaceForTaskLabelRow{ID: lbl1, Name: "bug", Color: "#f00", ProjectWorkspaceID: wsID}
	repo.labels[lbl2] = sqlc.GetLabelWorkspaceForTaskLabelRow{ID: lbl2, Name: "feat", Color: "#0f0", InTargetWs: true}

	h := NewHandler(NewService(repo))
	router := Routes(h, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := ctxhelper.WithUserRole(ctxhelper.WithUserID(r.Context(), userID.String()), "MEMBER")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	// Single assign -> 201
	b1, _ := json.Marshal(AssignRequest{TaskID: taskID.String(), LabelID: lbl1.String()})
	req1 := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b1))
	rr1 := httptest.NewRecorder()
	router.ServeHTTP(rr1, req1)
	assert.Equal(t, http.StatusCreated, rr1.Code)

	// Duplicate assign -> 400
	reqDup := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b1))
	rrDup := httptest.NewRecorder()
	router.ServeHTTP(rrDup, reqDup)
	assert.Equal(t, http.StatusBadRequest, rrDup.Code)

	// Assign multiple -> 201 replacing lbl1 with lbl2
	b2, _ := json.Marshal(AssignMultipleRequest{TaskID: taskID.String(), LabelIDs: []string{lbl2.String()}})
	reqMult := httptest.NewRequest(http.MethodPost, "/assign-multiple", bytes.NewReader(b2))
	rrMult := httptest.NewRecorder()
	router.ServeHTTP(rrMult, reqMult)
	assert.Equal(t, http.StatusCreated, rrMult.Code)

	var mult []AssignmentResponse
	require.NoError(t, json.Unmarshal(rrMult.Body.Bytes(), &mult))
	require.Len(t, mult, 1)
	assert.Equal(t, lbl2.String(), mult[0].LabelID)

	// Remove lbl2 -> 200 with removed pair
	reqDel := httptest.NewRequest(http.MethodDelete, "/"+taskID.String()+"/"+lbl2.String(), nil)
	rrDel := httptest.NewRecorder()
	router.ServeHTTP(rrDel, reqDel)
	assert.Equal(t, http.StatusOK, rrDel.Code)
}
