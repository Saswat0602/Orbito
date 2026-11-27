package search

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubRepo struct {
	scopes   AccessibleScopes
	tasks    []sqlc.SearchTasksGlobalQueryRow
	projects []sqlc.SearchProjectsGlobalQueryRow
	titles   []string
	pNames   []string
}

func (s *stubRepo) GetUserScopes(_ context.Context, _ uuid.UUID) (AccessibleScopes, error) {
	return s.scopes, nil
}
func (s *stubRepo) SearchTasksGlobal(_ context.Context, _ sqlc.SearchTasksGlobalQueryParams) ([]sqlc.SearchTasksGlobalQueryRow, error) {
	return s.tasks, nil
}
func (s *stubRepo) SearchProjectsGlobal(_ context.Context, _ sqlc.SearchProjectsGlobalQueryParams) ([]sqlc.SearchProjectsGlobalQueryRow, error) {
	return s.projects, nil
}
func (s *stubRepo) SearchUsersGlobal(_ context.Context, _ sqlc.SearchUsersGlobalQueryParams) ([]sqlc.SearchUsersGlobalQueryRow, error) {
	return nil, nil
}
func (s *stubRepo) SearchCommentsGlobal(_ context.Context, _ sqlc.SearchCommentsGlobalQueryParams) ([]sqlc.SearchCommentsGlobalQueryRow, error) {
	return nil, nil
}
func (s *stubRepo) SearchAttachmentsGlobal(_ context.Context, _ sqlc.SearchAttachmentsGlobalQueryParams) ([]sqlc.SearchAttachmentsGlobalQueryRow, error) {
	return nil, nil
}
func (s *stubRepo) SearchSprintsGlobal(_ context.Context, _ sqlc.SearchSprintsGlobalQueryParams) ([]sqlc.SearchSprintsGlobalQueryRow, error) {
	return nil, nil
}
func (s *stubRepo) CountTasksAdvanced(_ context.Context, _ sqlc.CountTasksAdvancedQueryParams) (int64, error) {
	return 0, nil
}
func (s *stubRepo) SearchTasksAdvanced(_ context.Context, _ sqlc.SearchTasksAdvancedQueryParams) ([]sqlc.SearchTasksAdvancedQueryRow, error) {
	return nil, nil
}
func (s *stubRepo) ListAssigneesBatch(_ context.Context, _ []uuid.UUID) ([]sqlc.ListTaskAssigneesBatchRow, error) {
	return nil, nil
}
func (s *stubRepo) ListReportersBatch(_ context.Context, _ []uuid.UUID) ([]sqlc.ListTaskReportersBatchRow, error) {
	return nil, nil
}
func (s *stubRepo) SearchTaskTitleSuggestions(_ context.Context, _ sqlc.SearchTaskTitleSuggestionsParams) ([]string, error) {
	return s.titles, nil
}
func (s *stubRepo) SearchProjectNameSuggestions(_ context.Context, _ sqlc.SearchProjectNameSuggestionsParams) ([]string, error) {
	return s.pNames, nil
}

func TestSearchGlobalScoringSuggestionsAndRateLimit(t *testing.T) {
	orgID := uuid.New()
	wsID := uuid.New()
	projID := uuid.New()
	userID := uuid.New()
	now := pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}

	repo := &stubRepo{
		scopes: AccessibleScopes{OrganizationIDs: []uuid.UUID{orgID}, WorkspaceIDs: []uuid.UUID{wsID}, ProjectIDs: []uuid.UUID{projID}},
		tasks: []sqlc.SearchTasksGlobalQueryRow{
			{
				ID: uuid.New(), Title: "auth", Slug: "ORB-1", Type: sqlc.TaskTypeTASK, Priority: sqlc.TaskPriorityHIGH,
				ProjectID: projID, ProjectSlug: "ORB", WorkspaceID: wsID, OrganizationID: orgID, CreatedAt: now, UpdatedAt: now,
			},
			{
				ID: uuid.New(), Title: "OAuth setup", Description: pgtype.Text{String: "auth flow", Valid: true},
				Slug: "ORB-2", Type: sqlc.TaskTypeBUG, Priority: sqlc.TaskPriorityMEDIUM,
				ProjectID: projID, ProjectSlug: "ORB", WorkspaceID: wsID, OrganizationID: orgID, CreatedAt: now, UpdatedAt: now,
			},
		},
		titles: []string{"auth", "OAuth setup"},
		pNames: []string{"auth", "Auth Portal"},
	}

	svc := NewService(repo)
	h := NewHandler(svc)
	r := chi.NewRouter()
	MountRoutes(r, h, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			next.ServeHTTP(w, req.WithContext(ctxhelper.WithUserID(req.Context(), userID.String())))
		})
	})

	body, _ := json.Marshal(GlobalRequest{Query: "auth"})
	req := httptest.NewRequest(http.MethodPost, "/search/global", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp Response
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.Len(t, resp.Results, 2)
	assert.Equal(t, "auth", resp.Results[0].Title)
	assert.Greater(t, resp.Results[0].RelevanceScore, resp.Results[1].RelevanceScore)

	// Short suggestion (<2 chars) returns empty array.
	sugShortReq := httptest.NewRequest(http.MethodGet, "/search/suggestions?q=a", nil)
	sugShortRec := httptest.NewRecorder()
	r.ServeHTTP(sugShortRec, sugShortReq)
	assert.Equal(t, http.StatusOK, sugShortRec.Code)
	assert.Equal(t, "[]\n", sugShortRec.Body.String())

	// Normal suggestion deduplicates.
	sugReq := httptest.NewRequest(http.MethodGet, "/search/suggestions?q=au", nil)
	sugRec := httptest.NewRecorder()
	r.ServeHTTP(sugRec, sugReq)
	assert.Equal(t, http.StatusOK, sugRec.Code)
	var suggestions []string
	require.NoError(t, json.Unmarshal(sugRec.Body.Bytes(), &suggestions))
	assert.Equal(t, []string{"auth", "OAuth setup", "Auth Portal"}, suggestions)
}
