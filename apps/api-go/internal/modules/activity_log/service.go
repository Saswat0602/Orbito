package activity_log

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/pagination"
)

// Service defines business operations for audit logging and activity queries.
type Service interface {
	LogActivity(ctx context.Context, in LogInput) error
	LogTaskCreated(ctx context.Context, taskID, title, key string, newValue any, userID uuid.UUID) error
	LogTaskUpdated(ctx context.Context, taskID, title string, oldValue, newValue any, userID uuid.UUID) error
	LogTaskStatusChanged(ctx context.Context, taskID, title, oldStatus, newStatus string, userID uuid.UUID) error
	LogTaskAssigned(ctx context.Context, taskID, title, assigneeID string, userID uuid.UUID) error
	LogTaskDeleted(ctx context.Context, taskID, title string, oldValue any, userID uuid.UUID) error
	GetOrganizationIDFromEntity(ctx context.Context, entityType, entityID string) (*uuid.UUID, error)

	GetTaskActivities(ctx context.Context, taskIdentifier string, page, limit int) (*PaginatedRawResponse, error)
	GetRecentActivityByWorkspaceOptimized(ctx context.Context, workspaceID uuid.UUID, limit, page int) (*PaginatedEnrichedResponse, error)
	GetRecentActivityByProject(ctx context.Context, projectIdentifier string, limit, page int) (*PaginatedProjectResponse, error)
	GetRecentActivityByOrganizationComprehensive(ctx context.Context, orgID uuid.UUID, limit, page int, filters *OrgFilters) (*PaginatedEnrichedResponse, error)
	GetOrganizationActivityStats(ctx context.Context, orgID uuid.UUID, days int) (*OrganizationStatsResponse, error)
	GetUserRecentActivity(ctx context.Context, userID uuid.UUID, limit, page int, orgID *uuid.UUID) (*PaginatedRawResponse, error)
}

type service struct {
	repo Repository
}

// NewService constructs an activity_log service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func safeInt32(n int) int32 {
	if n < 0 {
		return 0
	}
	if n > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(n)
}

func toPgUUID(u *uuid.UUID) pgtype.UUID {
	if u == nil {
		return pgtype.UUID{Valid: false}
	}
	return pgtype.UUID{Bytes: *u, Valid: true}
}

func toPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func toPgTS(t *time.Time) pgtype.Timestamp {
	if t == nil {
		return pgtype.Timestamp{Valid: false}
	}
	return pgtype.Timestamp{Time: t.UTC(), Valid: true}
}

func marshalOptionalJSON(v any) []byte {
	if v == nil {
		return nil
	}
	if raw, ok := v.(json.RawMessage); ok {
		return raw
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

func (s *service) LogActivity(ctx context.Context, in LogInput) error {
	if s.repo == nil {
		return nil
	}
	resolvedIDStr := in.EntityID
	if strings.EqualFold(in.EntityType, "task") {
		if _, parseErr := uuid.Parse(resolvedIDStr); parseErr != nil {
			taskRow, tErr := s.repo.ResolveTaskByIDOrSlug(ctx, resolvedIDStr)
			if tErr == nil && taskRow != nil {
				resolvedIDStr = taskRow.ID.String()
			}
		}
	}

	entityUUID, parseErr := uuid.Parse(resolvedIDStr)
	if parseErr != nil {
		// Match NestJS: prevent crash on non-UUID entityId
		return nil
	}

	orgID := in.OrganizationID
	if orgID == nil {
		orgID, _ = s.GetOrganizationIDFromEntity(ctx, in.EntityType, resolvedIDStr)
	}

	_, err := s.repo.Create(ctx, sqlc.CreateActivityLogRecordParams{
		ID:             uuid.New(),
		Type:           in.Type,
		Description:    in.Description,
		EntityType:     in.EntityType,
		EntityID:       entityUUID,
		OldValue:       marshalOptionalJSON(in.OldValue),
		NewValue:       marshalOptionalJSON(in.NewValue),
		UserID:         in.UserID,
		OrganizationID: toPgUUID(orgID),
		CreatedByID:    pgtype.UUID{Bytes: in.UserID, Valid: true},
	})
	return err
}

func (s *service) LogTaskCreated(ctx context.Context, taskID, title, key string, newValue any, userID uuid.UUID) error {
	return s.LogActivity(ctx, LogInput{
		Type:        sqlc.ActivityTypeTASKCREATED,
		Description: fmt.Sprintf(`Created task "%s" [%s]`, title, key),
		EntityType:  "Task",
		EntityID:    taskID,
		UserID:      userID,
		NewValue:    newValue,
	})
}

func (s *service) LogTaskUpdated(ctx context.Context, taskID, title string, oldValue, newValue any, userID uuid.UUID) error {
	return s.LogActivity(ctx, LogInput{
		Type:        sqlc.ActivityTypeTASKUPDATED,
		Description: fmt.Sprintf(`Updated task "%s"`, title),
		EntityType:  "Task",
		EntityID:    taskID,
		UserID:      userID,
		OldValue:    oldValue,
		NewValue:    newValue,
	})
}

func (s *service) LogTaskStatusChanged(ctx context.Context, taskID, title, oldStatus, newStatus string, userID uuid.UUID) error {
	return s.LogActivity(ctx, LogInput{
		Type:        sqlc.ActivityTypeTASKSTATUSCHANGED,
		Description: fmt.Sprintf(`Changed task "%s" status from "%s" to "%s"`, title, oldStatus, newStatus),
		EntityType:  "Task",
		EntityID:    taskID,
		UserID:      userID,
		OldValue:    map[string]string{"status": oldStatus},
		NewValue:    map[string]string{"status": newStatus},
	})
}

func (s *service) LogTaskAssigned(ctx context.Context, taskID, title, assigneeID string, userID uuid.UUID) error {
	return s.LogActivity(ctx, LogInput{
		Type:        sqlc.ActivityTypeTASKASSIGNED,
		Description: fmt.Sprintf(`Assigned task "%s" to user`, title),
		EntityType:  "Task",
		EntityID:    taskID,
		UserID:      userID,
		NewValue:    map[string]string{"assigneeId": assigneeID},
	})
}

func (s *service) LogTaskDeleted(ctx context.Context, taskID, title string, oldValue any, userID uuid.UUID) error {
	return s.LogActivity(ctx, LogInput{
		Type:        sqlc.ActivityTypeTASKDELETED,
		Description: fmt.Sprintf(`Deleted task "%s"`, title),
		EntityType:  "Task",
		EntityID:    taskID,
		UserID:      userID,
		OldValue:    oldValue,
	})
}

func (s *service) GetOrganizationIDFromEntity(ctx context.Context, entityType, entityID string) (*uuid.UUID, error) {
	if s.repo == nil || entityID == "" {
		return nil, nil
	}
	switch strings.ToLower(strings.TrimSpace(entityType)) {
	case "task":
		return s.repo.GetOrganizationIDForTask(ctx, entityID)
	case "project":
		pid, err := uuid.Parse(entityID)
		if err != nil {
			return nil, nil
		}
		return s.repo.GetOrganizationIDForProject(ctx, pid)
	case "workspace":
		wid, err := uuid.Parse(entityID)
		if err != nil {
			return nil, nil
		}
		return s.repo.GetOrganizationIDForWorkspace(ctx, wid)
	case "organization":
		oid, err := uuid.Parse(entityID)
		if err != nil {
			return nil, nil
		}
		return &oid, nil
	case "sprint":
		sid, err := uuid.Parse(entityID)
		if err != nil {
			return nil, nil
		}
		return s.repo.GetOrganizationIDForSprint(ctx, sid)
	default:
		return nil, nil
	}
}

func (s *service) GetTaskActivities(ctx context.Context, taskIdentifier string, page, limit int) (*PaginatedRawResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	if limit > 50 {
		limit = 50
	}
	if s.repo == nil {
		return &PaginatedRawResponse{
			Activities: []RawActivityItem{},
			Pagination: pagination.New(page, limit, 0),
		}, nil
	}

	taskRow, err := s.repo.ResolveTaskByIDOrSlug(ctx, taskIdentifier)
	if err != nil {
		return nil, fmt.Errorf("resolve task: %w", err)
	}
	if taskRow == nil {
		return nil, apperror.NotFound("Task not found")
	}

	total, err := s.repo.CountTaskActivities(ctx, taskRow.ID)
	if err != nil {
		return nil, fmt.Errorf("count task activities: %w", err)
	}

	offset := pagination.Offset(page, limit)
	rows, err := s.repo.ListTaskActivities(ctx, taskRow.ID, safeInt32(limit), safeInt32(offset))
	if err != nil {
		return nil, fmt.Errorf("list task activities: %w", err)
	}

	items := make([]RawActivityItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, RawActivityItem{
			ID:             r.ID.String(),
			Type:           string(r.Type),
			Description:    r.Description,
			EntityType:     r.EntityType,
			EntityID:       r.EntityID.String(),
			OldValue:       rawJSONOrNull(r.OldValue),
			NewValue:       rawJSONOrNull(r.NewValue),
			UserID:         r.UserID.String(),
			OrganizationID: pgUUIDPtr(r.OrganizationID),
			CreatedBy:      pgUUIDPtr(r.CreatedByID),
			UpdatedBy:      pgUUIDPtr(r.UpdatedByID),
			CreatedAt:      pgTSISO(r.CreatedAt),
			UpdatedAt:      pgTSISO(r.UpdatedAt),
			User: ActorUser{
				ID:        r.ActorID.String(),
				FirstName: r.ActorFirstName,
				LastName:  r.ActorLastName,
				Email:     r.ActorEmail,
				Avatar:    pgTextPtr(r.ActorAvatar),
			},
		})
	}

	return &PaginatedRawResponse{
		Activities: items,
		Pagination: pagination.New(page, limit, total),
	}, nil
}

func (s *service) GetRecentActivityByWorkspaceOptimized(ctx context.Context, workspaceID uuid.UUID, limit, page int) (*PaginatedEnrichedResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if s.repo == nil {
		return &PaginatedEnrichedResponse{
			Activities: []EnrichedActivityItem{},
			Pagination: pagination.New(page, limit, 0),
		}, nil
	}

	ws, err := s.repo.CheckWorkspaceExists(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("check workspace: %w", err)
	}
	if ws == nil {
		return nil, apperror.NotFound("Workspace not found")
	}

	total, err := s.repo.CountWorkspaceActivityOptimized(ctx, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("count workspace activity: %w", err)
	}

	offset := pagination.Offset(page, limit)
	rows, err := s.repo.ListWorkspaceActivityOptimized(ctx, workspaceID, safeInt32(limit), safeInt32(offset))
	if err != nil {
		return nil, fmt.Errorf("list workspace activity: %w", err)
	}

	items := make([]EnrichedActivityItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, EnrichedActivityItem{
			ID:            r.ID.String(),
			Type:          string(r.Type),
			Description:   r.Description,
			EntityType:    r.EntityType,
			EntityID:      r.EntityID.String(),
			CreatedAt:     pgTSISO(r.CreatedAt),
			TaskSlug:      nonEmptyPtr(r.TaskSlug),
			ProjectSlug:   firstNonEmpty(r.TaskProjectSlug, r.ProjectSlug, r.SprintProjectSlug),
			WorkspaceSlug: firstNonEmpty(r.TaskWorkspaceSlug, r.ProjectWorkspaceSlug, r.WorkspaceSlug, r.SprintWorkspaceSlug),
			SprintSlug:    nonEmptyPtr(r.SprintSlug),
			User: ActorSummary{
				ID:     r.UserID.String(),
				Name:   fullName(r.FirstName, r.LastName),
				Email:  r.Email,
				Avatar: pgTextPtr(r.Avatar),
			},
		})
	}

	return &PaginatedEnrichedResponse{
		Activities: items,
		Pagination: pagination.New(page, limit, total),
	}, nil
}

func (s *service) GetRecentActivityByProject(ctx context.Context, projectIdentifier string, limit, page int) (*PaginatedProjectResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if s.repo == nil {
		return &PaginatedProjectResponse{
			Activities: []ProjectActivityItem{},
			Pagination: pagination.New(page, limit, 0),
		}, nil
	}

	proj, err := s.repo.ResolveProjectByIDOrSlug(ctx, projectIdentifier)
	if err != nil {
		return nil, fmt.Errorf("resolve project: %w", err)
	}
	if proj == nil {
		return nil, apperror.NotFound("Project not found")
	}

	total, err := s.repo.CountProjectActivities(ctx, proj.ID)
	if err != nil {
		return nil, fmt.Errorf("count project activities: %w", err)
	}

	offset := pagination.Offset(page, limit)
	rows, err := s.repo.ListProjectActivities(ctx, proj.ID, safeInt32(limit), safeInt32(offset))
	if err != nil {
		return nil, fmt.Errorf("list project activities: %w", err)
	}

	items := make([]ProjectActivityItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, ProjectActivityItem{
			ID:            r.ID.String(),
			Type:          string(r.Type),
			Description:   r.Description,
			EntityType:    r.EntityType,
			EntityID:      r.EntityID.String(),
			CreatedAt:     pgTSISO(r.CreatedAt),
			ProjectSlug:   proj.Slug,
			WorkspaceSlug: pgTextPtr(proj.WorkspaceSlug),
			User: ActorSummary{
				ID:     r.UserID.String(),
				Name:   fullName(r.FirstName, r.LastName),
				Email:  r.Email,
				Avatar: pgTextPtr(r.Avatar),
			},
		})
	}

	return &PaginatedProjectResponse{
		Activities: items,
		Pagination: pagination.New(page, limit, total),
	}, nil
}

func (s *service) GetRecentActivityByOrganizationComprehensive(ctx context.Context, orgID uuid.UUID, limit, page int, filters *OrgFilters) (*PaginatedEnrichedResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if s.repo == nil {
		return &PaginatedEnrichedResponse{
			Activities: []EnrichedActivityItem{},
			Pagination: pagination.New(page, limit, 0),
		}, nil
	}

	var f OrgFilters
	if filters != nil {
		f = *filters
	}

	total, err := s.repo.CountOrganizationActivityComprehensive(ctx, sqlc.CountOrganizationActivityComprehensiveParams{
		OrganizationID: orgID,
		EntityType:     toPgText(f.EntityType),
		UserID:         toPgUUID(f.UserID),
		StartDate:      toPgTS(f.StartDate),
		EndDate:        toPgTS(f.EndDate),
	})
	if err != nil {
		return nil, fmt.Errorf("count org activity: %w", err)
	}

	offset := pagination.Offset(page, limit)
	rows, err := s.repo.ListOrganizationActivityComprehensive(ctx, sqlc.ListOrganizationActivityComprehensiveParams{
		OrganizationID: orgID,
		EntityType:     toPgText(f.EntityType),
		UserID:         toPgUUID(f.UserID),
		StartDate:      toPgTS(f.StartDate),
		EndDate:        toPgTS(f.EndDate),
		LimitVal:       safeInt32(limit),
		OffsetVal:      safeInt32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("list org activity: %w", err)
	}

	items := make([]EnrichedActivityItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, EnrichedActivityItem{
			ID:            r.ID.String(),
			Type:          string(r.Type),
			Description:   r.Description,
			EntityType:    r.EntityType,
			EntityID:      r.EntityID.String(),
			NewValue:      rawJSONOrNull(r.NewValue),
			CreatedAt:     pgTSISO(r.CreatedAt),
			TaskSlug:      nonEmptyPtr(r.TaskSlug),
			ProjectSlug:   firstNonEmpty(r.TaskProjectSlug, r.ProjectSlug, r.SprintProjectSlug),
			WorkspaceSlug: firstNonEmpty(r.TaskWorkspaceSlug, r.ProjectWorkspaceSlug, r.WorkspaceSlug, r.SprintWorkspaceSlug),
			SprintSlug:    nonEmptyPtr(r.SprintSlug),
			User: ActorSummary{
				ID:     r.UserID.String(),
				Name:   fullName(r.FirstName, r.LastName),
				Email:  r.Email,
				Avatar: pgTextPtr(r.Avatar),
			},
		})
	}

	return &PaginatedEnrichedResponse{
		Activities: items,
		Pagination: pagination.New(page, limit, total),
	}, nil
}

func (s *service) GetOrganizationActivityStats(ctx context.Context, orgID uuid.UUID, days int) (*OrganizationStatsResponse, error) {
	if days < 1 {
		days = 1
	}
	if days > 365 {
		days = 365
	}
	if s.repo == nil {
		return &OrganizationStatsResponse{
			ActivitiesByType: map[string]int64{},
			ActivitiesByUser: []UserActivityCount{},
			ActivitiesByDate: []DateActivityCount{},
		}, nil
	}

	since := time.Now().UTC().AddDate(0, 0, -days)
	rows, err := s.repo.ListOrganizationActivitiesForStats(ctx, orgID, pgtype.Timestamp{Time: since, Valid: true})
	if err != nil {
		return nil, fmt.Errorf("list org activity stats: %w", err)
	}

	byType := make(map[string]int64)
	userCounts := make(map[string]*UserActivityCount)
	dateCounts := make(map[string]int64)

	for _, r := range rows {
		byType[string(r.Type)]++

		uid := r.UserID.String()
		if _, ok := userCounts[uid]; !ok {
			userCounts[uid] = &UserActivityCount{
				UserID:   uid,
				UserName: fullName(r.FirstName, r.LastName),
			}
		}
		userCounts[uid].Count++

		if r.CreatedAt.Valid {
			day := r.CreatedAt.Time.UTC().Format("2006-01-02")
			dateCounts[day]++
		}
	}

	byUser := make([]UserActivityCount, 0, len(userCounts))
	for _, uc := range userCounts {
		byUser = append(byUser, *uc)
	}
	sort.Slice(byUser, func(i, j int) bool {
		return byUser[i].UserID < byUser[j].UserID
	})

	byDate := make([]DateActivityCount, 0, len(dateCounts))
	for d, c := range dateCounts {
		byDate = append(byDate, DateActivityCount{Date: d, Count: c})
	}
	sort.Slice(byDate, func(i, j int) bool {
		return byDate[i].Date < byDate[j].Date
	})

	return &OrganizationStatsResponse{
		TotalActivities:  int64(len(rows)),
		ActivitiesByType: byType,
		ActivitiesByUser: byUser,
		ActivitiesByDate: byDate,
	}, nil
}

func (s *service) GetUserRecentActivity(ctx context.Context, userID uuid.UUID, limit, page int, orgID *uuid.UUID) (*PaginatedRawResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if s.repo == nil {
		return &PaginatedRawResponse{
			Activities: []RawActivityItem{},
			Pagination: pagination.New(page, limit, 0),
		}, nil
	}

	total, err := s.repo.CountUserRecentActivities(ctx, sqlc.CountUserRecentActivitiesParams{
		UserID:         userID,
		OrganizationID: toPgUUID(orgID),
	})
	if err != nil {
		return nil, fmt.Errorf("count user recent activities: %w", err)
	}

	offset := pagination.Offset(page, limit)
	rows, err := s.repo.ListUserRecentActivities(ctx, sqlc.ListUserRecentActivitiesParams{
		UserID:         userID,
		OrganizationID: toPgUUID(orgID),
		LimitVal:       safeInt32(limit),
		OffsetVal:      safeInt32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("list user recent activities: %w", err)
	}

	items := make([]RawActivityItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, RawActivityItem{
			ID:             r.ID.String(),
			Type:           string(r.Type),
			Description:    r.Description,
			EntityType:     r.EntityType,
			EntityID:       r.EntityID.String(),
			OldValue:       rawJSONOrNull(r.OldValue),
			NewValue:       rawJSONOrNull(r.NewValue),
			UserID:         r.UserID.String(),
			OrganizationID: pgUUIDPtr(r.OrganizationID),
			CreatedBy:      pgUUIDPtr(r.CreatedByID),
			UpdatedBy:      pgUUIDPtr(r.UpdatedByID),
			CreatedAt:      pgTSISO(r.CreatedAt),
			UpdatedAt:      pgTSISO(r.UpdatedAt),
			User: ActorUser{
				ID:        r.ActorID.String(),
				FirstName: r.ActorFirstName,
				LastName:  r.ActorLastName,
				Email:     r.ActorEmail,
				Avatar:    pgTextPtr(r.ActorAvatar),
			},
		})
	}

	return &PaginatedRawResponse{
		Activities: items,
		Pagination: pagination.New(page, limit, total),
	}, nil
}
