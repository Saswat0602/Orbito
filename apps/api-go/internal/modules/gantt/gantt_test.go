package gantt

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	projID    uuid.UUID
	sprintID  uuid.UUID
	tasks     []sqlc.ListTasksForProjectGanttRow
	assignees []sqlc.ListTaskAssigneesForGanttRow
	deps      []sqlc.ListTaskDependenciesForGanttRow
}

func (m *mockRepo) GetProjectBasic(_ context.Context, projectID uuid.UUID) (*sqlc.GetProjectBasicForGanttRow, error) {
	if projectID != m.projID {
		return nil, nil
	}
	return &sqlc.GetProjectBasicForGanttRow{ID: projectID, Name: "Orbito"}, nil
}

func (m *mockRepo) GetSprintBasic(_ context.Context, sprintID uuid.UUID) (*sqlc.GetSprintBasicForGanttRow, error) {
	if sprintID != m.sprintID {
		return nil, nil
	}
	return &sqlc.GetSprintBasicForGanttRow{
		ID:        sprintID,
		Name:      "Sprint 1",
		ProjectID: m.projID,
		StartDate: pgtype.Timestamp{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true},
		EndDate:   pgtype.Timestamp{Time: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), Valid: true},
	}, nil
}

func (m *mockRepo) ListSprintsByProject(_ context.Context, _ uuid.UUID) ([]sqlc.ListSprintsForGanttByProjectRow, error) {
	return []sqlc.ListSprintsForGanttByProjectRow{{
		ID:        m.sprintID,
		Name:      "Sprint 1",
		StartDate: pgtype.Timestamp{Time: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), Valid: true},
		EndDate:   pgtype.Timestamp{Time: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC), Valid: true},
	}}, nil
}

func (m *mockRepo) ListProjectTasks(_ context.Context, _ uuid.UUID) ([]sqlc.ListTasksForProjectGanttRow, error) {
	return m.tasks, nil
}

func (m *mockRepo) ListSprintTasks(_ context.Context, _ uuid.UUID) ([]sqlc.ListTasksForSprintGanttRow, error) {
	out := make([]sqlc.ListTasksForSprintGanttRow, len(m.tasks))
	for i, t := range m.tasks {
		out[i] = sqlc.ListTasksForSprintGanttRow(t)
	}
	return out, nil
}

func (m *mockRepo) ListAssigneesByTaskIDs(_ context.Context, _ []uuid.UUID) ([]sqlc.ListTaskAssigneesForGanttRow, error) {
	return m.assignees, nil
}

func (m *mockRepo) ListDependenciesByProject(_ context.Context, _ uuid.UUID) ([]sqlc.ListTaskDependenciesForGanttRow, error) {
	return m.deps, nil
}

func (m *mockRepo) ListResourceTasks(_ context.Context, _ uuid.UUID) ([]sqlc.ListResourceTasksForProjectGanttRow, error) {
	var out []sqlc.ListResourceTasksForProjectGanttRow
	for _, t := range m.tasks {
		if t.StartDate.Valid && t.DueDate.Valid {
			out = append(out, sqlc.ListResourceTasksForProjectGanttRow{
				ID: t.ID, Title: t.Title, StartDate: t.StartDate, DueDate: t.DueDate, StoryPoints: t.StoryPoints,
			})
		}
	}
	return out, nil
}

func TestGantt_ProjectCriticalPathAndResources(t *testing.T) {
	projID, sprintID, t1, t2, u1, u2 := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	d1 := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	d2 := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	repo := &mockRepo{
		projID:   projID,
		sprintID: sprintID,
		tasks: []sqlc.ListTasksForProjectGanttRow{
			{
				ID: t1, Title: "Design", Slug: "OR-1", Priority: sqlc.TaskPriorityHIGH, Type: sqlc.TaskTypeTASK,
				StatusName: "In Progress", StatusColor: "#ff0",
				StartDate: pgtype.Timestamp{Time: d1, Valid: true}, DueDate: pgtype.Timestamp{Time: d2, Valid: true},
				StoryPoints: pgtype.Int4{Int32: 4, Valid: true},
			},
			{
				ID: t2, Title: "Build", Slug: "OR-2", Priority: sqlc.TaskPriorityMEDIUM, Type: sqlc.TaskTypeTASK,
				StatusName: "Done", StatusColor: "#0f0",
				StartDate: pgtype.Timestamp{Time: d1, Valid: true}, DueDate: pgtype.Timestamp{Time: d2, Valid: true},
				StoryPoints: pgtype.Int4{Int32: 2, Valid: true},
			},
		},
		assignees: []sqlc.ListTaskAssigneesForGanttRow{
			{TaskID: t1, UserID: u1, FirstName: "Alice", LastName: "A"},
			{TaskID: t1, UserID: u2, FirstName: "Bob", LastName: "B"},
		},
		deps: []sqlc.ListTaskDependenciesForGanttRow{
			{BlockingTaskID: t1, DependentTaskID: t2, Type: sqlc.DependencyTypeBLOCKS},
		},
	}

	router := Routes(NewHandler(NewService(repo)), func(next http.Handler) http.Handler { return next })

	// Project Gantt chart
	req := httptest.NewRequest(http.MethodGet, "/project/"+projID.String(), nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	require.Equal(t, http.StatusOK, rr.Code)

	var chart ChartData
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &chart))
	assert.Len(t, chart.Tasks, 2)
	assert.Equal(t, []string{t1.String(), t2.String()}, chart.CriticalPath)
	assert.Len(t, chart.Milestones, 2)

	// Resource allocation divides 4 points evenly across 2 assignees (= 2.0 each)
	reqRes := httptest.NewRequest(http.MethodGet, "/project/"+projID.String()+"/resources", nil)
	rrRes := httptest.NewRecorder()
	router.ServeHTTP(rrRes, reqRes)
	require.Equal(t, http.StatusOK, rrRes.Code)

	var alloc []ResourceAllocation
	require.NoError(t, json.Unmarshal(rrRes.Body.Bytes(), &alloc))
	require.Len(t, alloc, 2)
	assert.Equal(t, 2.0, alloc[0].Workload)
	assert.Equal(t, 2.0, alloc[1].Workload)
}

