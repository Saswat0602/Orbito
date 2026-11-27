package search

import (
	"context"
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// Service implements tenant-scoped global, advanced, and autocomplete search.
type Service interface {
	GlobalSearch(ctx context.Context, req GlobalRequest, userID uuid.UUID) (*Response, error)
	AdvancedSearch(ctx context.Context, req AdvancedRequest, userID uuid.UUID) (*Response, error)
	GetSuggestions(ctx context.Context, query string, limit int, userID uuid.UUID) ([]string, error)
}

type service struct {
	repo Repository
}

// NewService constructs a search service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GlobalSearch(ctx context.Context, req GlobalRequest, userID uuid.UUID) (*Response, error) {
	start := time.Now()
	page, limit, offset := normalizePage(req.Page, req.Limit)

	qTrimmed := strings.TrimSpace(req.Query)
	if qTrimmed == "" {
		return emptyResponse(page, limit, start), nil
	}

	entityType := strings.ToLower(strings.TrimSpace(req.EntityType))
	if entityType == "" {
		entityType = "all"
	}

	filterOrg, filterWs, filterProj, err := parseScopeFilters(req.OrganizationID, req.WorkspaceID, req.ProjectID)
	if err != nil {
		return nil, err
	}

	scopes, err := s.repo.GetUserScopes(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load accessible scopes: %w", err)
	}
	if scopes.IsEmpty() || !requestedScopesAllowed(scopes, filterOrg, filterWs, filterProj) {
		return emptyResponse(page, limit, start), nil
	}

	escaped := escapeLikePattern(qTrimmed)
	subOffset := offset
	if entityType == "all" {
		subOffset = 0
	}

	results, collectErr := s.collectGlobalResults(ctx, escaped, qTrimmed, entityType, scopes, filterOrg, filterWs, filterProj, limit, subOffset)
	if collectErr != nil {
		return nil, collectErr
	}

	sortResults(results, req.SortBy, req.SortOrder)
	total := int64(len(results))
	if entityType == "all" {
		results = slicePage(results, int(offset), int(limit))
	}

	return buildResponse(results, total, page, limit, start), nil
}

func (s *service) collectGlobalResults(
	ctx context.Context,
	escaped, rawQuery, entityType string,
	scopes AccessibleScopes,
	filterOrg, filterWs, filterProj pgtype.UUID,
	limit, offset int32,
) ([]ResultItem, error) {
	var all []ResultItem

	if entityType == "all" || entityType == "tasks" {
		rows, err := s.repo.SearchTasksGlobal(ctx, sqlc.SearchTasksGlobalQueryParams{
			QueryText: escaped, OrgIds: scopes.OrganizationIDs, WsIds: scopes.WorkspaceIDs, ProjIds: scopes.ProjectIDs,
			FilterOrgID: filterOrg, FilterWsID: filterWs, FilterProjID: filterProj, LimitVal: limit, OffsetVal: offset,
		})
		if err != nil {
			return nil, fmt.Errorf("search tasks: %w", err)
		}
		all = append(all, mapGlobalTasks(rows, rawQuery)...)
	}

	if entityType == "all" || entityType == "projects" {
		rows, err := s.repo.SearchProjectsGlobal(ctx, sqlc.SearchProjectsGlobalQueryParams{
			QueryText: escaped, OrgIds: scopes.OrganizationIDs, WsIds: scopes.WorkspaceIDs, ProjIds: scopes.ProjectIDs,
			FilterOrgID: filterOrg, FilterWsID: filterWs, FilterProjID: filterProj, LimitVal: limit, OffsetVal: offset,
		})
		if err != nil {
			return nil, fmt.Errorf("search projects: %w", err)
		}
		all = append(all, mapGlobalProjects(rows, rawQuery)...)
	}

	if entityType == "all" || entityType == "users" {
		rows, err := s.repo.SearchUsersGlobal(ctx, sqlc.SearchUsersGlobalQueryParams{
			QueryText: escaped, OrgIds: scopes.OrganizationIDs, WsIds: scopes.WorkspaceIDs, ProjIds: scopes.ProjectIDs,
			LimitVal: limit, OffsetVal: offset,
		})
		if err != nil {
			return nil, fmt.Errorf("search users: %w", err)
		}
		all = append(all, mapGlobalUsers(rows, rawQuery)...)
	}

	if entityType == "all" || entityType == "comments" {
		rows, err := s.repo.SearchCommentsGlobal(ctx, sqlc.SearchCommentsGlobalQueryParams{
			QueryText: escaped, OrgIds: scopes.OrganizationIDs, WsIds: scopes.WorkspaceIDs, ProjIds: scopes.ProjectIDs,
			FilterOrgID: filterOrg, FilterWsID: filterWs, FilterProjID: filterProj, LimitVal: limit, OffsetVal: offset,
		})
		if err != nil {
			return nil, fmt.Errorf("search comments: %w", err)
		}
		all = append(all, mapGlobalComments(rows, rawQuery)...)
	}

	if entityType == "all" || entityType == "attachments" {
		rows, err := s.repo.SearchAttachmentsGlobal(ctx, sqlc.SearchAttachmentsGlobalQueryParams{
			QueryText: escaped, OrgIds: scopes.OrganizationIDs, WsIds: scopes.WorkspaceIDs, ProjIds: scopes.ProjectIDs,
			FilterOrgID: filterOrg, FilterWsID: filterWs, FilterProjID: filterProj, LimitVal: limit, OffsetVal: offset,
		})
		if err != nil {
			return nil, fmt.Errorf("search attachments: %w", err)
		}
		all = append(all, mapGlobalAttachments(rows, rawQuery)...)
	}

	if entityType == "all" || entityType == "sprints" {
		rows, err := s.repo.SearchSprintsGlobal(ctx, sqlc.SearchSprintsGlobalQueryParams{
			QueryText: escaped, OrgIds: scopes.OrganizationIDs, WsIds: scopes.WorkspaceIDs, ProjIds: scopes.ProjectIDs,
			FilterOrgID: filterOrg, FilterWsID: filterWs, FilterProjID: filterProj, LimitVal: limit, OffsetVal: offset,
		})
		if err != nil {
			return nil, fmt.Errorf("search sprints: %w", err)
		}
		all = append(all, mapGlobalSprints(rows, rawQuery)...)
	}

	return all, nil
}

func (s *service) AdvancedSearch(ctx context.Context, req AdvancedRequest, userID uuid.UUID) (*Response, error) {
	start := time.Now()
	page, limit, offset := normalizePage(req.Page, req.Limit)

	filterOrg, filterWs, filterProj, err := parseScopeFilters(req.OrganizationID, req.WorkspaceID, req.ProjectID)
	if err != nil {
		return nil, err
	}
	scopes, err := s.repo.GetUserScopes(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load accessible scopes: %w", err)
	}
	if scopes.IsEmpty() || !requestedScopesAllowed(scopes, filterOrg, filterWs, filterProj) {
		return emptyResponse(page, limit, start), nil
	}

	filters, parseErr := buildAdvancedParams(req, scopes, filterOrg, filterWs, filterProj, limit, offset)
	if parseErr != nil {
		return nil, parseErr
	}

	total, err := s.repo.CountTasksAdvanced(ctx, filters.count)
	if err != nil {
		return nil, fmt.Errorf("count advanced search tasks: %w", err)
	}
	rows, err := s.repo.SearchTasksAdvanced(ctx, filters.query)
	if err != nil {
		return nil, fmt.Errorf("advanced search tasks: %w", err)
	}

	results := s.enrichAdvancedResults(ctx, rows, req.Query)
	if req.SortBy != "" {
		sortResults(results, req.SortBy, req.SortOrder)
	}
	return buildResponse(results, total, page, limit, start), nil
}

func (s *service) GetSuggestions(ctx context.Context, query string, limit int, userID uuid.UUID) ([]string, error) {
	trimmed := strings.TrimSpace(query)
	if len(trimmed) < 2 {
		return []string{}, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	scopes, err := s.repo.GetUserScopes(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load accessible scopes: %w", err)
	}
	if scopes.IsEmpty() {
		return []string{}, nil
	}

	escaped := escapeLikePattern(trimmed)
	limitVal := clampInt32(limit)
	titles, err := s.repo.SearchTaskTitleSuggestions(ctx, sqlc.SearchTaskTitleSuggestionsParams{
		QueryText: escaped, OrgIds: scopes.OrganizationIDs, WsIds: scopes.WorkspaceIDs, ProjIds: scopes.ProjectIDs, LimitVal: limitVal,
	})
	if err != nil {
		return nil, fmt.Errorf("task title suggestions: %w", err)
	}
	projNames, err := s.repo.SearchProjectNameSuggestions(ctx, sqlc.SearchProjectNameSuggestionsParams{
		QueryText: escaped, OrgIds: scopes.OrganizationIDs, WsIds: scopes.WorkspaceIDs, ProjIds: scopes.ProjectIDs, LimitVal: limitVal,
	})
	if err != nil {
		return nil, fmt.Errorf("project name suggestions: %w", err)
	}

	seen := make(map[string]struct{}, limit)
	out := make([]string, 0, limit)
	for _, sItem := range append(titles, projNames...) {
		if _, dup := seen[sItem]; dup {
			continue
		}
		seen[sItem] = struct{}{}
		out = append(out, sItem)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (s *service) enrichAdvancedResults(ctx context.Context, rows []sqlc.SearchTasksAdvancedQueryRow, rawQuery string) []ResultItem {
	taskIDs := make([]uuid.UUID, 0, len(rows))
	for i := range rows {
		taskIDs = append(taskIDs, rows[i].ID)
	}
	assignees, _ := s.repo.ListAssigneesBatch(ctx, taskIDs)
	reporters, _ := s.repo.ListReportersBatch(ctx, taskIDs)

	assigneeByTask := make(map[uuid.UUID][]map[string]any)
	for _, a := range assignees {
		assigneeByTask[a.TaskID] = append(assigneeByTask[a.TaskID], map[string]any{
			"id": a.ID.String(), "firstName": a.FirstName, "lastName": a.LastName, "email": a.Email,
		})
	}
	reporterByTask := make(map[uuid.UUID][]map[string]any)
	for _, r := range reporters {
		reporterByTask[r.TaskID] = append(reporterByTask[r.TaskID], map[string]any{
			"id": r.ID.String(), "firstName": r.FirstName, "lastName": r.LastName, "email": r.Email,
		})
	}

	results := make([]ResultItem, 0, len(rows))
	for i := range rows {
		r := &rows[i]
		results = append(results, mapAdvancedTaskRow(r, rawQuery, assigneeByTask[r.ID], reporterByTask[r.ID]))
	}
	return results
}

func calculateRelevanceScore(title, description, query string) int {
	qLower := strings.ToLower(strings.TrimSpace(query))
	if qLower == "" {
		return 0
	}
	tLower := strings.ToLower(title)
	dLower := strings.ToLower(description)

	score := 0
	if tLower == qLower {
		score += 100
	}
	if strings.Contains(tLower, qLower) {
		score += 50
	}
	if strings.Contains(dLower, qLower) {
		score += 25
	}
	titleWords := strings.Fields(tLower)
	for _, qw := range strings.Fields(qLower) {
		if slices.Contains(titleWords, qw) {
			score += 30
		}
	}
	return score
}

func sortResults(results []ResultItem, sortBy, sortOrder string) {
	asc := strings.EqualFold(sortOrder, "asc")
	sort.SliceStable(results, func(i, j int) bool {
		a, b := results[i], results[j]
		var less bool
		switch sortBy {
		case "title":
			less = strings.ToLower(a.Title) < strings.ToLower(b.Title)
		case "createdAt":
			less = a.CreatedAt < b.CreatedAt
		case "updatedAt":
			less = a.UpdatedAt < b.UpdatedAt
		default:
			less = a.RelevanceScore < b.RelevanceScore
		}
		if asc {
			return less
		}
		return !less
	})
}

func parseScopeFilters(orgRaw, wsRaw, projRaw string) (pgtype.UUID, pgtype.UUID, pgtype.UUID, error) {
	org, err := parseOptUUID(orgRaw, "organizationId")
	if err != nil {
		return pgtype.UUID{}, pgtype.UUID{}, pgtype.UUID{}, err
	}
	ws, err := parseOptUUID(wsRaw, "workspaceId")
	if err != nil {
		return pgtype.UUID{}, pgtype.UUID{}, pgtype.UUID{}, err
	}
	proj, err := parseOptUUID(projRaw, "projectId")
	if err != nil {
		return pgtype.UUID{}, pgtype.UUID{}, pgtype.UUID{}, err
	}
	return org, ws, proj, nil
}

func requestedScopesAllowed(scopes AccessibleScopes, filterOrg, filterWs, filterProj pgtype.UUID) bool {
	if filterOrg.Valid && !slices.Contains(scopes.OrganizationIDs, uuid.UUID(filterOrg.Bytes)) {
		return false
	}
	if filterWs.Valid && !slices.Contains(scopes.WorkspaceIDs, uuid.UUID(filterWs.Bytes)) {
		return false
	}
	if filterProj.Valid && !slices.Contains(scopes.ProjectIDs, uuid.UUID(filterProj.Bytes)) {
		return false
	}
	return true
}

func parseOptUUID(raw, field string) (pgtype.UUID, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return pgtype.UUID{}, nil
	}
	id, err := uuid.Parse(trimmed)
	if err != nil {
		return pgtype.UUID{}, apperror.BadRequest(field + " must be a valid UUID")
	}
	return pgtype.UUID{Bytes: id, Valid: true}, nil
}

func escapeLikePattern(q string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(q)
}

func normalizePage(page, limit int) (int, int32, int32) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit
	return page, clampInt32(limit), clampInt32(offset)
}

func clampInt32(v int) int32 {
	if v < 0 {
		return 0
	}
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(v)
}

func slicePage(items []ResultItem, offset, limit int) []ResultItem {
	if offset >= len(items) {
		return []ResultItem{}
	}
	end := offset + limit
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end]
}

func emptyResponse(page int, limit int32, start time.Time) *Response {
	return &Response{
		Results:    []ResultItem{},
		Total:      0,
		Page:       page,
		Limit:      int(limit),
		TotalPages: 0,
		SearchTime: time.Since(start).Milliseconds(),
	}
}

func buildResponse(results []ResultItem, total int64, page int, limit int32, start time.Time) *Response {
	totalPages := 0
	if limit > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}
	if results == nil {
		results = []ResultItem{}
	}
	return &Response{
		Results:    results,
		Total:      total,
		Page:       page,
		Limit:      int(limit),
		TotalPages: totalPages,
		SearchTime: time.Since(start).Milliseconds(),
	}
}
