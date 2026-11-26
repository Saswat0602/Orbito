package task_watchers

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
	scopes   map[string]sqlc.GetTaskProjectScopeForWatcherRow
	userAcc  map[uuid.UUID]sqlc.CheckTargetUserTenantAccessRow
	watchers map[uuid.UUID]sqlc.GetTaskWatcherDetailsByIDRow
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		scopes:   make(map[string]sqlc.GetTaskProjectScopeForWatcherRow),
		userAcc:  make(map[uuid.UUID]sqlc.CheckTargetUserTenantAccessRow),
		watchers: make(map[uuid.UUID]sqlc.GetTaskWatcherDetailsByIDRow),
	}
}

func (m *mockRepo) GetTaskScope(_ context.Context, identifier string) (*sqlc.GetTaskProjectScopeForWatcherRow, error) {
	if s, ok := m.scopes[identifier]; ok {
		return &s, nil
	}
	return nil, nil
}

func (m *mockRepo) CheckCallerMembership(_ context.Context, _, _, _ uuid.UUID) (sqlc.CheckCallerTaskMembershipRow, error) {
	return sqlc.CheckCallerTaskMembershipRow{IsPm: true}, nil
}

func (m *mockRepo) CheckTargetUserAccess(_ context.Context, _, _, _, userID uuid.UUID) (*sqlc.CheckTargetUserTenantAccessRow, error) {
	if u, ok := m.userAcc[userID]; ok {
		return &u, nil
	}
	return nil, nil
}

func (m *mockRepo) CheckUserExists(_ context.Context, userID uuid.UUID) (bool, error) {
	_, ok := m.userAcc[userID]
	return ok, nil
}

func (m *mockRepo) CreateWatcher(_ context.Context, id, taskID, userID uuid.UUID) (uuid.UUID, error) {
	for _, w := range m.watchers {
		if w.TaskID == taskID && w.UserID == userID {
			return uuid.Nil, ErrUniqueViolation
		}
	}
	now := pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
	m.watchers[id] = sqlc.GetTaskWatcherDetailsByIDRow{
		ID:            id,
		TaskID:        taskID,
		UserID:        userID,
		CreatedAt:     now,
		UpdatedAt:     now,
		UserEmail:     "watcher@orbito.io",
		UserFirstName: "Sam",
		UserLastName:  "Lee",
		UserStatus:    sqlc.UserStatusACTIVE,
		TaskTitle:     "Fix API",
		TaskSlug:      "OR-1",
		TaskType:      sqlc.TaskTypeTASK,
		TaskPriority:  sqlc.TaskPriorityMEDIUM,
		StatusID:      uuid.New(),
		StatusName:    "To Do",
		StatusColor:   "#fff",
		ProjectID:     uuid.New(),
		ProjectName:   "Orbito",
		ProjectSlug:   "orbito",
		WorkspaceID:   uuid.New(),
		WorkspaceName: "Core",
		WorkspaceSlug: "core",
	}
	return id, nil
}

func (m *mockRepo) GetByID(_ context.Context, id uuid.UUID) (*sqlc.GetTaskWatcherDetailsByIDRow, error) {
	w, ok := m.watchers[id]
	if !ok {
		return nil, nil
	}
	return &w, nil
}

func (m *mockRepo) FindByPair(_ context.Context, taskID, userID uuid.UUID) (*sqlc.FindTaskWatcherByTaskAndUserRow, error) {
	for _, w := range m.watchers {
		if w.TaskID == taskID && w.UserID == userID {
			return &sqlc.FindTaskWatcherByTaskAndUserRow{ID: w.ID, TaskID: w.TaskID, UserID: w.UserID}, nil
		}
	}
	return nil, nil
}

func (m *mockRepo) DeleteByPair(_ context.Context, taskID, userID uuid.UUID) (int64, error) {
	for id, w := range m.watchers {
		if w.TaskID == taskID && w.UserID == userID {
			delete(m.watchers, id)
			return 1, nil
		}
	}
	return 0, nil
}

func (m *mockRepo) DeleteByID(_ context.Context, id uuid.UUID) (int64, error) {
	if _, ok := m.watchers[id]; !ok {
		return 0, nil
	}
	delete(m.watchers, id)
	return 1, nil
}

func (m *mockRepo) ListFiltered(_ context.Context, taskID, userID *uuid.UUID) ([]sqlc.ListTaskWatchersFilteredRow, error) {
	var out []sqlc.ListTaskWatchersFilteredRow
	for _, w := range m.watchers {
		if taskID != nil && w.TaskID != *taskID {
			continue
		}
		if userID != nil && w.UserID != *userID {
			continue
		}
		out = append(out, sqlc.ListTaskWatchersFilteredRow{
			ID: w.ID, TaskID: w.TaskID, UserID: w.UserID, CreatedAt: w.CreatedAt, UpdatedAt: w.UpdatedAt,
			UserEmail: w.UserEmail, UserFirstName: w.UserFirstName, UserLastName: w.UserLastName,
			TaskTitle: w.TaskTitle, TaskSlug: w.TaskSlug, ProjectID: w.ProjectID, ProjectName: w.ProjectName, ProjectSlug: w.ProjectSlug,
		})
	}
	return out, nil
}

func (m *mockRepo) ListForTask(_ context.Context, taskID uuid.UUID) ([]sqlc.ListWatchersForTaskRow, error) {
	var out []sqlc.ListWatchersForTaskRow
	for _, w := range m.watchers {
		if w.TaskID == taskID {
			out = append(out, sqlc.ListWatchersForTaskRow{
				ID: w.ID, TaskID: w.TaskID, UserID: w.UserID, CreatedAt: w.CreatedAt, UpdatedAt: w.UpdatedAt,
				UserEmail: w.UserEmail, UserFirstName: w.UserFirstName, UserLastName: w.UserLastName,
			})
		}
	}
	return out, nil
}

func (m *mockRepo) ListWatchedByUser(_ context.Context, _ uuid.UUID) ([]sqlc.ListWatchedTasksByUserRow, error) {
	return nil, nil
}

func (m *mockRepo) GetStatsCounts(_ context.Context, _ time.Time, _, _ *uuid.UUID) (sqlc.CountTaskWatchersTotalAndRecentRow, error) {
	return sqlc.CountTaskWatchersTotalAndRecentRow{TotalWatchers: int64(len(m.watchers)), RecentWatches: int64(len(m.watchers))}, nil
}

func (m *mockRepo) GroupMostWatched(_ context.Context, _ *uuid.UUID) ([]sqlc.GroupMostWatchedTasksRow, error) {
	return nil, nil
}

func TestTaskWatchers_LifecycleAndOwnership(t *testing.T) {
	repo := newMockRepo()
	taskID, userID, otherID := uuid.New(), uuid.New(), uuid.New()
	scope := sqlc.GetTaskProjectScopeForWatcherRow{
		TaskID: taskID, TaskSlug: "OR-1", ProjectID: uuid.New(), Visibility: sqlc.ProjectVisibilityPRIVATE,
	}
	repo.scopes[taskID.String()] = scope
	repo.scopes["OR-1"] = scope
	repo.userAcc[userID] = sqlc.CheckTargetUserTenantAccessRow{ID: userID, IsPm: true}
	repo.userAcc[otherID] = sqlc.CheckTargetUserTenantAccessRow{ID: otherID, IsPm: true}

	h := NewHandler(NewService(repo))
	router := Routes(h, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := ctxhelper.WithUserRole(ctxhelper.WithUserID(r.Context(), otherID.String()), "MEMBER")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	// Toggle ON
	body, _ := json.Marshal(CreateRequest{TaskID: "OR-1", UserID: userID.String()})
	reqToggle := httptest.NewRequest(http.MethodPost, "/toggle", bytes.NewReader(body))
	rrToggle := httptest.NewRecorder()
	router.ServeHTTP(rrToggle, reqToggle)
	assert.Equal(t, http.StatusCreated, rrToggle.Code)

	var tog ToggleWatchResponse
	require.NoError(t, json.Unmarshal(rrToggle.Body.Bytes(), &tog))
	assert.True(t, tog.IsWatching)
	require.NotNil(t, tog.Watcher)

	// Other user cannot DELETE userID's watcher (403)
	reqDelOther := httptest.NewRequest(http.MethodDelete, "/"+tog.Watcher.ID, nil)
	rrDelOther := httptest.NewRecorder()
	router.ServeHTTP(rrDelOther, reqDelOther)
	assert.Equal(t, http.StatusForbidden, rrDelOther.Code)

	// Toggle OFF
	reqToggleOff := httptest.NewRequest(http.MethodPost, "/toggle", bytes.NewReader(body))
	rrToggleOff := httptest.NewRecorder()
	router.ServeHTTP(rrToggleOff, reqToggleOff)
	assert.Equal(t, http.StatusCreated, rrToggleOff.Code)

	var togOff ToggleWatchResponse
	require.NoError(t, json.Unmarshal(rrToggleOff.Body.Bytes(), &togOff))
	assert.False(t, togOff.IsWatching)
}
