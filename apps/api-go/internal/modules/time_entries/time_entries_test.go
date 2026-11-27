package time_entries

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
	tasks   map[uuid.UUID]sqlc.CheckTimeEntryTaskProjectAccessRow
	entries map[uuid.UUID]sqlc.GetTimeEntryDetailByIDRow
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		tasks:   make(map[uuid.UUID]sqlc.CheckTimeEntryTaskProjectAccessRow),
		entries: make(map[uuid.UUID]sqlc.GetTimeEntryDetailByIDRow),
	}
}

func (m *mockRepo) CheckTaskAccess(_ context.Context, taskID, _ uuid.UUID) (*sqlc.CheckTimeEntryTaskProjectAccessRow, error) {
	if r, ok := m.tasks[taskID]; ok {
		return &r, nil
	}
	return nil, nil
}

func (m *mockRepo) Create(_ context.Context, p CreateParams) (uuid.UUID, error) {
	var st, et pgtype.Timestamp
	if p.StartTime != nil {
		st = pgtype.Timestamp{Time: *p.StartTime, Valid: true}
	}
	if p.EndTime != nil {
		et = pgtype.Timestamp{Time: *p.EndTime, Valid: true}
	}
	var desc pgtype.Text
	if p.Description != nil {
		desc = pgtype.Text{String: *p.Description, Valid: true}
	}
	m.entries[p.ID] = sqlc.GetTimeEntryDetailByIDRow{
		ID: p.ID, Description: desc, TimeSpent: p.TimeSpent, StartTime: st, EndTime: et,
		Date: pgtype.Timestamp{Time: p.Date, Valid: true}, TaskID: p.TaskID, UserID: p.UserID,
		CreatedAt: pgtype.Timestamp{Time: time.Now().UTC(), Valid: true},
		UpdatedAt: pgtype.Timestamp{Time: time.Now().UTC(), Valid: true},
		TaskTitle: "Task 1", TaskSlug: "OR-1", ProjectID: uuid.New(), ProjectName: "Orbito", ProjectSlug: "orbito",
		WorkspaceID: uuid.New(), WorkspaceName: "Core", WorkspaceSlug: "core",
		UserEmail: "u@orbito.io", UserFirstName: "Dev", UserLastName: "User",
	}
	return p.ID, nil
}

func (m *mockRepo) GetByID(_ context.Context, id uuid.UUID) (*sqlc.GetTimeEntryDetailByIDRow, error) {
	e, ok := m.entries[id]
	if !ok {
		return nil, nil
	}
	return &e, nil
}

func (m *mockRepo) ListFiltered(_ context.Context, taskID, userID *uuid.UUID, _, _ *time.Time) ([]sqlc.ListTimeEntriesFilteredRow, error) {
	var out []sqlc.ListTimeEntriesFilteredRow
	for _, e := range m.entries {
		if taskID != nil && e.TaskID != *taskID {
			continue
		}
		if userID != nil && e.UserID != *userID {
			continue
		}
		out = append(out, sqlc.ListTimeEntriesFilteredRow{
			ID: e.ID, Description: e.Description, TimeSpent: e.TimeSpent, StartTime: e.StartTime, EndTime: e.EndTime,
			Date: e.Date, TaskID: e.TaskID, UserID: e.UserID, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
			TaskTitle: e.TaskTitle, TaskSlug: e.TaskSlug, ProjectID: e.ProjectID, ProjectName: e.ProjectName, ProjectSlug: e.ProjectSlug,
			UserFirstName: e.UserFirstName, UserLastName: e.UserLastName,
		})
	}
	return out, nil
}

func (m *mockRepo) GetActiveTimer(_ context.Context, userID uuid.UUID) (*sqlc.GetActiveTimerForUserRow, error) {
	for _, e := range m.entries {
		if e.UserID == userID && e.StartTime.Valid && !e.EndTime.Valid {
			return &sqlc.GetActiveTimerForUserRow{
				ID: e.ID, Description: e.Description, TimeSpent: e.TimeSpent, StartTime: e.StartTime, EndTime: e.EndTime,
				Date: e.Date, TaskID: e.TaskID, UserID: e.UserID, CreatedAt: e.CreatedAt, UpdatedAt: e.UpdatedAt,
				TaskTitle: e.TaskTitle, TaskSlug: e.TaskSlug, ProjectID: e.ProjectID, ProjectName: e.ProjectName, ProjectSlug: e.ProjectSlug,
				UserFirstName: e.UserFirstName, UserLastName: e.UserLastName,
			}, nil
		}
	}
	return nil, nil
}

func (m *mockRepo) Update(_ context.Context, p UpdateParams) (uuid.UUID, error) {
	e := m.entries[p.ID]
	if p.TimeSpent != nil {
		e.TimeSpent = *p.TimeSpent
	}
	m.entries[p.ID] = e
	return p.ID, nil
}

func (m *mockRepo) StopTimer(_ context.Context, id uuid.UUID, endTime time.Time, timeSpent int32, _ *string) (uuid.UUID, error) {
	e := m.entries[id]
	e.EndTime = pgtype.Timestamp{Time: endTime, Valid: true}
	e.TimeSpent = timeSpent
	m.entries[id] = e
	return id, nil
}

func (m *mockRepo) Delete(_ context.Context, id uuid.UUID) (int64, error) {
	if _, ok := m.entries[id]; !ok {
		return 0, nil
	}
	delete(m.entries, id)
	return 1, nil
}

func TestTimeEntries_StartStopTimerAndSummary(t *testing.T) {
	repo := newMockRepo()
	taskID, userID, otherID := uuid.New(), uuid.New(), uuid.New()
	repo.tasks[taskID] = sqlc.CheckTimeEntryTaskProjectAccessRow{
		TaskID: taskID, ProjectID: uuid.New(), IsPm: true,
	}

	h := NewHandler(NewService(repo))
	currentID := userID
	router := Routes(h, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(ctxhelper.WithUserID(r.Context(), currentID.String())))
		})
	})

	// Start timer -> 201
	startBody, _ := json.Marshal(StartTimerRequest{TaskID: taskID.String()})
	reqStart := httptest.NewRequest(http.MethodPost, "/timer/start", bytes.NewReader(startBody))
	rrStart := httptest.NewRecorder()
	router.ServeHTTP(rrStart, reqStart)
	require.Equal(t, http.StatusCreated, rrStart.Code)

	// Duplicate start timer -> 409
	reqDup := httptest.NewRequest(http.MethodPost, "/timer/start", bytes.NewReader(startBody))
	rrDup := httptest.NewRecorder()
	router.ServeHTTP(rrDup, reqDup)
	assert.Equal(t, http.StatusConflict, rrDup.Code)

	// Other user cannot view active timer -> 403
	currentID = otherID
	reqActive := httptest.NewRequest(http.MethodGet, "/timer/active/"+userID.String(), nil)
	rrActive := httptest.NewRecorder()
	router.ServeHTTP(rrActive, reqActive)
	assert.Equal(t, http.StatusForbidden, rrActive.Code)

	// Owner stops timer -> 200
	currentID = userID
	reqStop := httptest.NewRequest(http.MethodPost, "/timer/stop", bytes.NewReader([]byte(`{}`)))
	rrStop := httptest.NewRecorder()
	router.ServeHTTP(rrStop, reqStop)
	assert.Equal(t, http.StatusOK, rrStop.Code)
}
