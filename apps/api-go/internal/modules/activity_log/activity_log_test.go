package activity_log

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
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockActivityRepo struct {
	logs        []sqlc.ActivityLog
	taskID      uuid.UUID
	taskSlug    string
	projectID   uuid.UUID
	projectSlug string
	workspaceID uuid.UUID
	orgID       uuid.UUID
}

func (m *mockActivityRepo) Create(_ context.Context, arg sqlc.CreateActivityLogRecordParams) (sqlc.ActivityLog, error) {
	now := pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
	rec := sqlc.ActivityLog{
		ID:             arg.ID,
		Type:           arg.Type,
		Description:    arg.Description,
		EntityType:     arg.EntityType,
		EntityID:       arg.EntityID,
		OldValue:       arg.OldValue,
		NewValue:       arg.NewValue,
		UserID:         arg.UserID,
		OrganizationID: arg.OrganizationID,
		CreatedByID:    arg.CreatedByID,
		UpdatedByID:    arg.CreatedByID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	m.logs = append([]sqlc.ActivityLog{rec}, m.logs...)
	return rec, nil
}

func (m *mockActivityRepo) ResolveTaskByIDOrSlug(_ context.Context, identifier string) (*sqlc.ResolveActivityTaskByIDOrSlugRow, error) {
	if identifier == m.taskID.String() || identifier == m.taskSlug {
		return &sqlc.ResolveActivityTaskByIDOrSlugRow{ID: m.taskID, Slug: m.taskSlug}, nil
	}
	return nil, nil
}

func (m *mockActivityRepo) ResolveProjectByIDOrSlug(_ context.Context, identifier string) (*sqlc.ResolveActivityProjectByIDOrSlugRow, error) {
	if identifier == m.projectID.String() || identifier == m.projectSlug {
		return &sqlc.ResolveActivityProjectByIDOrSlugRow{
			ID:            m.projectID,
			Slug:          m.projectSlug,
			WorkspaceSlug: pgtype.Text{String: "main-ws", Valid: true},
		}, nil
	}
	return nil, nil
}

func (m *mockActivityRepo) CheckWorkspaceExists(_ context.Context, workspaceID uuid.UUID) (*sqlc.CheckWorkspaceExistsForActivityRow, error) {
	if workspaceID == m.workspaceID {
		return &sqlc.CheckWorkspaceExistsForActivityRow{ID: m.workspaceID, OrganizationID: m.orgID}, nil
	}
	return nil, nil
}

func (m *mockActivityRepo) GetOrganizationIDForTask(_ context.Context, _ string) (*uuid.UUID, error) {
	return &m.orgID, nil
}
func (m *mockActivityRepo) GetOrganizationIDForProject(_ context.Context, _ uuid.UUID) (*uuid.UUID, error) {
	return &m.orgID, nil
}
func (m *mockActivityRepo) GetOrganizationIDForWorkspace(_ context.Context, _ uuid.UUID) (*uuid.UUID, error) {
	return &m.orgID, nil
}
func (m *mockActivityRepo) GetOrganizationIDForSprint(_ context.Context, _ uuid.UUID) (*uuid.UUID, error) {
	return &m.orgID, nil
}

func (m *mockActivityRepo) CountTaskActivities(_ context.Context, taskID uuid.UUID) (int64, error) {
	var c int64
	for _, l := range m.logs {
		if l.EntityID == taskID {
			c++
		}
	}
	return c, nil
}

func (m *mockActivityRepo) ListTaskActivities(_ context.Context, taskID uuid.UUID, _, _ int32) ([]sqlc.ListTaskActivitiesRow, error) {
	var out []sqlc.ListTaskActivitiesRow
	for _, l := range m.logs {
		if l.EntityID == taskID {
			out = append(out, sqlc.ListTaskActivitiesRow{
				ID:             l.ID,
				Type:           l.Type,
				Description:    l.Description,
				EntityType:     l.EntityType,
				EntityID:       l.EntityID,
				OldValue:       l.OldValue,
				NewValue:       l.NewValue,
				UserID:         l.UserID,
				OrganizationID: l.OrganizationID,
				CreatedAt:      l.CreatedAt,
				UpdatedAt:      l.UpdatedAt,
				ActorID:        l.UserID,
				ActorFirstName: "Saswat",
				ActorLastName:  "Mohanty",
				ActorEmail:     "saswat@orbito.io",
			})
		}
	}
	return out, nil
}

func (m *mockActivityRepo) CountWorkspaceActivityOptimized(_ context.Context, _ uuid.UUID) (int64, error) {
	return int64(len(m.logs)), nil
}

func (m *mockActivityRepo) ListWorkspaceActivityOptimized(_ context.Context, _ uuid.UUID, _, _ int32) ([]sqlc.ListWorkspaceActivityOptimizedRow, error) {
	var out []sqlc.ListWorkspaceActivityOptimizedRow
	for _, l := range m.logs {
		out = append(out, sqlc.ListWorkspaceActivityOptimizedRow{
			ID:                l.ID,
			Type:              l.Type,
			Description:       l.Description,
			EntityType:        l.EntityType,
			EntityID:          l.EntityID,
			CreatedAt:         l.CreatedAt,
			UserID:            l.UserID,
			FirstName:         "Saswat",
			LastName:          "Mohanty",
			Email:             "saswat@orbito.io",
			TaskSlug:          m.taskSlug,
			TaskProjectSlug:   m.projectSlug,
			TaskWorkspaceSlug: "main-ws",
		})
	}
	return out, nil
}

func (m *mockActivityRepo) CountProjectActivities(_ context.Context, _ uuid.UUID) (int64, error) {
	return int64(len(m.logs)), nil
}

func (m *mockActivityRepo) ListProjectActivities(_ context.Context, _ uuid.UUID, _, _ int32) ([]sqlc.ListProjectActivitiesRow, error) {
	var out []sqlc.ListProjectActivitiesRow
	for _, l := range m.logs {
		out = append(out, sqlc.ListProjectActivitiesRow{
			ID:          l.ID,
			Type:        l.Type,
			Description: l.Description,
			EntityType:  l.EntityType,
			EntityID:    l.EntityID,
			CreatedAt:   l.CreatedAt,
			UserID:      l.UserID,
			FirstName:   "Saswat",
			LastName:    "Mohanty",
			Email:       "saswat@orbito.io",
		})
	}
	return out, nil
}

func (m *mockActivityRepo) CountOrganizationActivityComprehensive(_ context.Context, _ sqlc.CountOrganizationActivityComprehensiveParams) (int64, error) {
	return int64(len(m.logs)), nil
}

func (m *mockActivityRepo) ListOrganizationActivityComprehensive(_ context.Context, _ sqlc.ListOrganizationActivityComprehensiveParams) ([]sqlc.ListOrganizationActivityComprehensiveRow, error) {
	var out []sqlc.ListOrganizationActivityComprehensiveRow
	for _, l := range m.logs {
		out = append(out, sqlc.ListOrganizationActivityComprehensiveRow{
			ID:                l.ID,
			Type:              l.Type,
			Description:       l.Description,
			EntityType:        l.EntityType,
			EntityID:          l.EntityID,
			NewValue:          l.NewValue,
			CreatedAt:         l.CreatedAt,
			UserID:            l.UserID,
			FirstName:         "Saswat",
			LastName:          "Mohanty",
			Email:             "saswat@orbito.io",
			TaskSlug:          m.taskSlug,
			TaskProjectSlug:   m.projectSlug,
			TaskWorkspaceSlug: "main-ws",
		})
	}
	return out, nil
}

func (m *mockActivityRepo) ListOrganizationActivitiesForStats(_ context.Context, _ uuid.UUID, _ pgtype.Timestamp) ([]sqlc.ListOrganizationActivitiesForStatsRow, error) {
	var out []sqlc.ListOrganizationActivitiesForStatsRow
	for _, l := range m.logs {
		out = append(out, sqlc.ListOrganizationActivitiesForStatsRow{
			Type:      l.Type,
			UserID:    l.UserID,
			CreatedAt: l.CreatedAt,
			FirstName: "Saswat",
			LastName:  "Mohanty",
		})
	}
	return out, nil
}

func (m *mockActivityRepo) CountUserRecentActivities(_ context.Context, _ sqlc.CountUserRecentActivitiesParams) (int64, error) {
	return int64(len(m.logs)), nil
}

func (m *mockActivityRepo) ListUserRecentActivities(_ context.Context, _ sqlc.ListUserRecentActivitiesParams) ([]sqlc.ListUserRecentActivitiesRow, error) {
	var out []sqlc.ListUserRecentActivitiesRow
	for _, l := range m.logs {
		out = append(out, sqlc.ListUserRecentActivitiesRow{
			ID:             l.ID,
			Type:           l.Type,
			Description:    l.Description,
			EntityType:     l.EntityType,
			EntityID:       l.EntityID,
			OldValue:       l.OldValue,
			NewValue:       l.NewValue,
			UserID:         l.UserID,
			OrganizationID: l.OrganizationID,
			CreatedAt:      l.CreatedAt,
			UpdatedAt:      l.UpdatedAt,
			ActorID:        l.UserID,
			ActorFirstName: "Saswat",
			ActorLastName:  "Mohanty",
			ActorEmail:     "saswat@orbito.io",
		})
	}
	return out, nil
}

func withUser(req *http.Request, userID uuid.UUID) *http.Request {
	return req.WithContext(ctxhelper.WithUserID(req.Context(), userID.String()))
}

func TestActivityLogEndpointsAndLogging(t *testing.T) {
	taskID := uuid.New()
	projID := uuid.New()
	wsID := uuid.New()
	orgID := uuid.New()
	userID := uuid.New()

	repo := &mockActivityRepo{
		taskID:      taskID,
		taskSlug:    "ORB-1",
		projectID:   projID,
		projectSlug: "orb",
		workspaceID: wsID,
		orgID:       orgID,
	}
	svc := NewService(repo)
	handler := NewHandler(svc)
	router := Routes(handler, nil)

	// Log via slug resolution
	require.NoError(t, svc.LogTaskCreated(context.Background(), "ORB-1", "Build Audit", "ORB-1", map[string]string{"title": "Build Audit"}, userID))
	require.NoError(t, svc.LogTaskStatusChanged(context.Background(), taskID.String(), "Build Audit", "TODO", "DONE", userID))
	require.Len(t, repo.logs, 2)

	// GET /?workspaceId= returns plain array with resolved slugs
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, withUser(httptest.NewRequest(http.MethodGet, "/?workspaceId="+wsID.String(), nil), userID))
	require.Equal(t, http.StatusOK, rr.Code)
	var wsArr []EnrichedActivityItem
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &wsArr))
	require.Len(t, wsArr, 2)
	require.NotNil(t, wsArr[0].TaskSlug)
	assert.Equal(t, "ORB-1", *wsArr[0].TaskSlug)

	// GET /project/{slug} returns plain array
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, withUser(httptest.NewRequest(http.MethodGet, "/project/orb", nil), userID))
	require.Equal(t, http.StatusOK, rr.Code)
	var projArr []ProjectActivityItem
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &projArr))
	require.Len(t, projArr, 2)
	assert.Equal(t, "orb", projArr[0].ProjectSlug)

	// GET /organization/{id}/recent validates entityType
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, withUser(httptest.NewRequest(http.MethodGet, "/organization/"+orgID.String()+"/recent?entityType=Bogus", nil), userID))
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, withUser(httptest.NewRequest(http.MethodGet, "/organization/"+orgID.String()+"/recent?entityType=Task", nil), userID))
	require.Equal(t, http.StatusOK, rr.Code)
	var orgRes PaginatedEnrichedResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &orgRes))
	assert.Equal(t, int64(2), orgRes.Pagination.TotalCount)

	// GET /organization/{id}/stats
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, withUser(httptest.NewRequest(http.MethodGet, "/organization/"+orgID.String()+"/stats?days=30", nil), userID))
	require.Equal(t, http.StatusOK, rr.Code)
	var stats OrganizationStatsResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &stats))
	assert.Equal(t, int64(2), stats.TotalActivities)
	assert.Equal(t, int64(1), stats.ActivitiesByType["TASK_CREATED"])
	assert.Equal(t, int64(1), stats.ActivitiesByType["TASK_STATUS_CHANGED"])

	// GET /task/{slug}/activities
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, withUser(httptest.NewRequest(http.MethodGet, "/task/ORB-1/activities", nil), userID))
	require.Equal(t, http.StatusOK, rr.Code)
	var taskRes PaginatedRawResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &taskRes))
	assert.Equal(t, int64(2), taskRes.Pagination.TotalCount)
}
