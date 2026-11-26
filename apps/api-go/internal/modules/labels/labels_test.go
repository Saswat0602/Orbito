package labels

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
	projWS      map[uuid.UUID]*sqlc.GetProjectWorkspaceForLabelRow
	details     map[uuid.UUID]*sqlc.GetLabelDetailsByIDRow
	tasks       map[string]*sqlc.FindTaskWorkspaceAccessByIDOrSlugRow
	labelCheck  map[uuid.UUID]*sqlc.GetLabelWorkspaceCheckRow
	assigned    map[string]bool
	replacedIDs []uuid.UUID
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		projWS:     make(map[uuid.UUID]*sqlc.GetProjectWorkspaceForLabelRow),
		details:    make(map[uuid.UUID]*sqlc.GetLabelDetailsByIDRow),
		tasks:      make(map[string]*sqlc.FindTaskWorkspaceAccessByIDOrSlugRow),
		labelCheck: make(map[uuid.UUID]*sqlc.GetLabelWorkspaceCheckRow),
		assigned:   make(map[string]bool),
	}
}

func (m *mockRepo) GetProjectWorkspace(_ context.Context, projectID, userID uuid.UUID) (*sqlc.GetProjectWorkspaceForLabelRow, error) {
	if r, ok := m.projWS[projectID]; ok {
		return r, nil
	}
	return nil, pgx.ErrNoRows
}

func (m *mockRepo) CreateWithWorkspaceTx(_ context.Context, arg sqlc.CreateLabelRecordParams, workspaceID uuid.UUID) (uuid.UUID, error) {
	now := pgtype.Timestamp{Time: time.Now(), Valid: true}
	m.details[arg.ID] = &sqlc.GetLabelDetailsByIDRow{
		ID:            arg.ID,
		Name:          arg.Name,
		Color:         arg.Color,
		Description:   arg.Description,
		ProjectID:     arg.ProjectID,
		CreatedAt:     now,
		UpdatedAt:     now,
		WorkspaceID:   workspaceID,
		IsWsMember:    true,
		ProjectName:   "Demo",
		ProjectSlug:   "demo",
		WorkspaceName: "WS",
		WorkspaceSlug: "ws",
		OrgName:       "Org",
		OrgSlug:       "org",
	}
	m.labelCheck[arg.ID] = &sqlc.GetLabelWorkspaceCheckRow{
		ID:                 arg.ID,
		ProjectWorkspaceID: workspaceID,
		InTargetWorkspace:  true,
	}
	return arg.ID, nil
}

func (m *mockRepo) GetDetailsByID(_ context.Context, labelID, _ uuid.UUID) (*sqlc.GetLabelDetailsByIDRow, error) {
	if r, ok := m.details[labelID]; ok {
		return r, nil
	}
	return nil, pgx.ErrNoRows
}

func (m *mockRepo) ListByWorkspaceID(_ context.Context, _ uuid.UUID) ([]sqlc.ListLabelsByWorkspaceIDRow, error) {
	return []sqlc.ListLabelsByWorkspaceIDRow{}, nil
}

func (m *mockRepo) ListAccessibleForUser(_ context.Context, _ uuid.UUID) ([]sqlc.ListAccessibleLabelsForUserRow, error) {
	return []sqlc.ListAccessibleLabelsForUserRow{}, nil
}

func (m *mockRepo) ListTaskLabelEntries(_ context.Context, _ uuid.UUID) ([]sqlc.ListTaskLabelEntriesByLabelIDRow, error) {
	return nil, nil
}

func (m *mockRepo) ListTaskAssignees(_ context.Context, _ []uuid.UUID) ([]sqlc.ListTaskAssigneesForTaskIDsRow, error) {
	return nil, nil
}

func (m *mockRepo) UpdateByID(_ context.Context, arg sqlc.UpdateLabelByIDParams) (uuid.UUID, error) {
	d, ok := m.details[arg.ID]
	if !ok {
		return uuid.Nil, pgx.ErrNoRows
	}
	if arg.Name.Valid {
		d.Name = arg.Name.String
	}
	return arg.ID, nil
}

func (m *mockRepo) DeleteByID(_ context.Context, id uuid.UUID) (int64, error) {
	if _, ok := m.details[id]; ok {
		delete(m.details, id)
		return 1, nil
	}
	return 0, nil
}

func (m *mockRepo) FindTaskWorkspaceAccess(_ context.Context, taskIDOrSlug string, _ uuid.UUID) (*sqlc.FindTaskWorkspaceAccessByIDOrSlugRow, error) {
	if r, ok := m.tasks[taskIDOrSlug]; ok {
		return r, nil
	}
	return nil, pgx.ErrNoRows
}

func (m *mockRepo) GetLabelWorkspaceCheck(_ context.Context, labelID, _ uuid.UUID) (*sqlc.GetLabelWorkspaceCheckRow, error) {
	if r, ok := m.labelCheck[labelID]; ok {
		return r, nil
	}
	return nil, pgx.ErrNoRows
}

func (m *mockRepo) AssignSingleTx(_ context.Context, taskID, labelID, _ uuid.UUID) error {
	m.assigned[taskID.String()+":"+labelID.String()] = true
	return nil
}

func (m *mockRepo) RemoveTaskLabel(_ context.Context, taskID, labelID uuid.UUID) (int64, error) {
	key := taskID.String() + ":" + labelID.String()
	if m.assigned[key] {
		delete(m.assigned, key)
		return 1, nil
	}
	return 0, nil
}

func (m *mockRepo) ListWorkspaceValidLabelIDs(_ context.Context, labelIDs []uuid.UUID, workspaceID uuid.UUID) ([]uuid.UUID, error) {
	valid := make([]uuid.UUID, 0, len(labelIDs))
	for _, id := range labelIDs {
		if lc, ok := m.labelCheck[id]; ok && (lc.ProjectWorkspaceID == workspaceID || lc.InTargetWorkspace) {
			valid = append(valid, id)
		}
	}
	return valid, nil
}

func (m *mockRepo) ReplaceTaskLabelsTx(_ context.Context, _ uuid.UUID, _ uuid.UUID, labelIDs []uuid.UUID) error {
	m.replacedIDs = labelIDs
	return nil
}

func (m *mockRepo) ListLabelsByTaskID(_ context.Context, _ uuid.UUID) ([]sqlc.ListLabelsByTaskIDRow, error) {
	return []sqlc.ListLabelsByTaskIDRow{}, nil
}

func TestLabelsService_CRUDAndTaskAssignment(t *testing.T) {
	repo := newMockRepo()
	svc := NewService(repo)
	ctx := context.Background()

	userID := uuid.New()
	wsID := uuid.New()
	projID := uuid.New()
	taskID := uuid.New()

	repo.projWS[projID] = &sqlc.GetProjectWorkspaceForLabelRow{
		ID:          projID,
		WorkspaceID: wsID,
		IsWsMember:  false,
	}

	_, err := svc.Create(ctx, CreateLabelRequest{Name: "Bug", Color: "#FF5733", ProjectID: projID.String()}, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission")

	repo.projWS[projID].IsWsMember = true
	created, err := svc.Create(ctx, CreateLabelRequest{Name: "Bug", Color: "#FF5733", ProjectID: projID.String()}, userID)
	require.NoError(t, err)
	assert.Equal(t, "Bug", created.Name)

	repo.tasks["ORB-1"] = &sqlc.FindTaskWorkspaceAccessByIDOrSlugRow{
		TaskID:      taskID,
		ProjectID:   projID,
		WorkspaceID: wsID,
		IsWsMember:  true,
	}

	err = svc.AssignLabelToTask(ctx, AssignLabelRequest{TaskID: "ORB-1", LabelID: created.ID}, userID)
	require.NoError(t, err)
	assert.True(t, repo.assigned[taskID.String()+":"+created.ID])

	err = svc.AssignMultipleLabelsToTask(ctx, AssignMultipleLabelsRequest{TaskID: "ORB-1", LabelIDs: []string{created.ID}}, userID)
	require.NoError(t, err)
	assert.Len(t, repo.replacedIDs, 1)

	err = svc.RemoveLabelFromTask(ctx, "ORB-1", uuid.MustParse(created.ID), userID)
	require.NoError(t, err)
}

func TestLabelsHandler_Validation(t *testing.T) {
	h := NewHandler(NewService(newMockRepo()))
	router := Routes(h, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := ctxhelper.WithUserID(r.Context(), uuid.New().String())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Bug","color":"not-hex"}`))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
