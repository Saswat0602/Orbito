package notifications

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/pagination"
)

// EventEmitter broadcasts newly created notifications to connected WebSocket clients.
type EventEmitter interface {
	EmitNotification(userID string, notification any)
}

// Service defines notification business operations.
type Service interface {
	CreateNotification(ctx context.Context, in CreateInput) (*NotificationDTO, error)
	GetUserNotifications(ctx context.Context, userID uuid.UUID, filters ListFilters, page, limit int) (*ListResponse, error)
	GetUnreadCount(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID) (int64, error)
	GetUnreadCountsByOrganization(ctx context.Context, userID uuid.UUID) ([]OrgUnreadCountDTO, error)
	GetNotificationsByUserAndOrganization(ctx context.Context, userID, orgID uuid.UUID, filters UserOrgFilters, page, limit int) (*UserOrgNotificationsResponse, error)
	GetUserNotificationStats(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID) (*StatsSummaryDTO, error)
	GetNotificationByID(ctx context.Context, id, userID uuid.UUID) (*NotificationDTO, error)
	MarkAsRead(ctx context.Context, id, userID uuid.UUID) error
	MarkAllAsRead(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID) error
	DeleteMultipleNotifications(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) error
	DeleteNotification(ctx context.Context, id, userID uuid.UUID) error
	NotifyTaskAssigned(ctx context.Context, taskID uuid.UUID, assigneeIDs []uuid.UUID, assignedBy uuid.UUID) error
	NotifyTaskStatusChanged(ctx context.Context, taskID uuid.UUID, oldStatus, newStatus string, changedBy uuid.UUID) error
	NotifyTaskDueSoon(ctx context.Context, taskID uuid.UUID) error
}

type service struct {
	repo    Repository
	emitter EventEmitter
}

// NewService constructs a notifications service.
func NewService(repo Repository, emitter EventEmitter) Service {
	return &service{repo: repo, emitter: emitter}
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

func toPgBool(b *bool) pgtype.Bool {
	if b == nil {
		return pgtype.Bool{Valid: false}
	}
	return pgtype.Bool{Bool: *b, Valid: true}
}

func toPgTS(t *time.Time) pgtype.Timestamp {
	if t == nil {
		return pgtype.Timestamp{Valid: false}
	}
	return pgtype.Timestamp{Time: t.UTC(), Valid: true}
}

func toNullNotifType(t *sqlc.NotificationType) sqlc.NullNotificationType {
	if t == nil {
		return sqlc.NullNotificationType{Valid: false}
	}
	return sqlc.NullNotificationType{NotificationType: *t, Valid: true}
}

func toNullNotifPriority(p *sqlc.NotificationPriority) sqlc.NullNotificationPriority {
	if p == nil {
		return sqlc.NullNotificationPriority{Valid: false}
	}
	return sqlc.NullNotificationPriority{NotificationPriority: *p, Valid: true}
}

func (s *service) CreateNotification(ctx context.Context, in CreateInput) (*NotificationDTO, error) {
	if s.repo == nil {
		return nil, apperror.Internal("Notification repository unavailable")
	}

	prio := sqlc.NotificationPriorityMEDIUM
	if in.Priority != nil && *in.Priority != "" {
		prio = *in.Priority
	}

	row, err := s.repo.Create(ctx, sqlc.CreateNotificationRecordParams{
		ID:             uuid.New(),
		Title:          in.Title,
		Message:        in.Message,
		Type:           in.Type,
		Priority:       prio,
		EntityType:     toPgText(in.EntityType),
		EntityID:       toPgUUID(in.EntityID),
		ActionUrl:      toPgText(in.ActionURL),
		UserID:         in.UserID,
		OrganizationID: toPgUUID(in.OrganizationID),
		CreatedByID:    toPgUUID(in.CreatedBy),
	})
	if err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}

	dto := mapNotificationModel(row)
	if s.emitter != nil {
		s.emitter.EmitNotification(dto.UserID, dto)
	}
	return &dto, nil
}

func (s *service) GetUserNotifications(ctx context.Context, userID uuid.UUID, filters ListFilters, page, limit int) (*ListResponse, error) {
	if page < 1 {
		return nil, apperror.BadRequest("Page must be a positive number")
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if s.repo == nil {
		return &ListResponse{
			Notifications: []NotificationDTO{},
			Pagination:    pagination.New(page, limit, 0),
		}, nil
	}

	offset := pagination.Offset(page, limit)
	rows, err := s.repo.ListUser(ctx, sqlc.ListUserNotificationsParams{
		UserID:         userID,
		IsRead:         toPgBool(filters.IsRead),
		NotifType:      toNullNotifType(filters.Type),
		OrganizationID: toPgUUID(filters.OrganizationID),
		Since:          toPgTS(filters.Since),
		LimitVal:       safeInt32(limit),
		OffsetVal:      safeInt32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("list user notifications: %w", err)
	}

	total, err := s.repo.CountUser(ctx, sqlc.CountUserNotificationsParams{
		UserID:         userID,
		IsRead:         toPgBool(filters.IsRead),
		NotifType:      toNullNotifType(filters.Type),
		OrganizationID: toPgUUID(filters.OrganizationID),
		Since:          toPgTS(filters.Since),
	})
	if err != nil {
		return nil, fmt.Errorf("count user notifications: %w", err)
	}

	items := make([]NotificationDTO, 0, len(rows))
	for _, r := range rows {
		dto := NotificationDTO{
			ID:             r.ID.String(),
			Title:          r.Title,
			Message:        r.Message,
			Type:           string(r.Type),
			Priority:       string(r.Priority),
			IsRead:         r.IsRead,
			EntityType:     pgTextPtr(r.EntityType),
			EntityID:       pgUUIDPtr(r.EntityID),
			ActionURL:      pgTextPtr(r.ActionUrl),
			UserID:         r.UserID.String(),
			OrganizationID: pgUUIDPtr(r.OrganizationID),
			CreatedBy:      pgUUIDPtr(r.CreatedByID),
			UpdatedBy:      pgUUIDPtr(r.UpdatedByID),
			CreatedAt:      pgTSISO(r.CreatedAt),
			UpdatedAt:      pgTSISO(r.UpdatedAt),
			ReadAt:         pgTSISOPtr(r.ReadAt),
		}
		if r.CreatorID.Valid {
			dto.CreatedByUser = &CreatorBrief{
				ID:        uuid.UUID(r.CreatorID.Bytes).String(),
				FirstName: r.CreatorFirstName.String,
				LastName:  r.CreatorLastName.String,
				Avatar:    pgTextPtr(r.CreatorAvatar),
			}
		}
		items = append(items, dto)
	}

	return &ListResponse{
		Notifications: items,
		Pagination:    pagination.New(page, limit, total),
	}, nil
}

func (s *service) GetUnreadCount(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID) (int64, error) {
	if s.repo == nil {
		return 0, nil
	}
	return s.repo.CountUnread(ctx, sqlc.CountUnreadNotificationsParams{
		UserID:         userID,
		OrganizationID: toPgUUID(orgID),
	})
}

func (s *service) GetUnreadCountsByOrganization(ctx context.Context, userID uuid.UUID) ([]OrgUnreadCountDTO, error) {
	if s.repo == nil {
		return []OrgUnreadCountDTO{}, nil
	}
	rows, err := s.repo.CountUnreadByOrg(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("unread by organization: %w", err)
	}
	out := make([]OrgUnreadCountDTO, 0, len(rows))
	for _, r := range rows {
		if !r.OrganizationID.Valid {
			continue
		}
		orgName := r.OrganizationName
		if orgName == "" {
			orgName = "Unknown"
		}
		out = append(out, OrgUnreadCountDTO{
			OrganizationID:   uuid.UUID(r.OrganizationID.Bytes).String(),
			OrganizationName: orgName,
			UnreadCount:      r.UnreadCount,
		})
	}
	return out, nil
}

func (s *service) GetUserNotificationStats(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID) (*StatsSummaryDTO, error) {
	if s.repo == nil {
		return &StatsSummaryDTO{ByType: map[string]int64{}}, nil
	}

	total, err := s.repo.CountUser(ctx, sqlc.CountUserNotificationsParams{
		UserID:         userID,
		OrganizationID: toPgUUID(orgID),
	})
	if err != nil {
		return nil, fmt.Errorf("count total notifications: %w", err)
	}

	unread, err := s.repo.CountUnread(ctx, sqlc.CountUnreadNotificationsParams{
		UserID:         userID,
		OrganizationID: toPgUUID(orgID),
	})
	if err != nil {
		return nil, fmt.Errorf("count unread notifications: %w", err)
	}

	sevenDaysAgo := time.Now().UTC().AddDate(0, 0, -7)
	recent, err := s.repo.CountUser(ctx, sqlc.CountUserNotificationsParams{
		UserID:         userID,
		OrganizationID: toPgUUID(orgID),
		Since:          toPgTS(&sevenDaysAgo),
	})
	if err != nil {
		return nil, fmt.Errorf("count recent notifications: %w", err)
	}

	typeRows, err := s.repo.GroupByTypeForUser(ctx, sqlc.GroupNotificationsByTypeForUserParams{
		UserID:         userID,
		OrganizationID: toPgUUID(orgID),
	})
	if err != nil {
		return nil, fmt.Errorf("group notifications by type: %w", err)
	}

	byType := make(map[string]int64, len(typeRows))
	for _, tr := range typeRows {
		byType[string(tr.Type)] = tr.Cnt
	}

	return &StatsSummaryDTO{
		Total:  total,
		Unread: unread,
		Read:   total - unread,
		Recent: recent,
		ByType: byType,
	}, nil
}

func (s *service) GetNotificationByID(ctx context.Context, id, userID uuid.UUID) (*NotificationDTO, error) {
	if s.repo == nil {
		return nil, apperror.NotFound("Notification not found")
	}
	row, err := s.repo.GetByIDForUser(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("get notification: %w", err)
	}
	if row == nil {
		return nil, apperror.NotFound("Notification not found")
	}

	dto := NotificationDTO{
		ID:             row.ID.String(),
		Title:          row.Title,
		Message:        row.Message,
		Type:           string(row.Type),
		Priority:       string(row.Priority),
		IsRead:         row.IsRead,
		EntityType:     pgTextPtr(row.EntityType),
		EntityID:       pgUUIDPtr(row.EntityID),
		ActionURL:      pgTextPtr(row.ActionUrl),
		UserID:         row.UserID.String(),
		OrganizationID: pgUUIDPtr(row.OrganizationID),
		CreatedBy:      pgUUIDPtr(row.CreatedByID),
		UpdatedBy:      pgUUIDPtr(row.UpdatedByID),
		CreatedAt:      pgTSISO(row.CreatedAt),
		UpdatedAt:      pgTSISO(row.UpdatedAt),
		ReadAt:         pgTSISOPtr(row.ReadAt),
	}
	if row.CreatorID.Valid {
		dto.CreatedByUser = &CreatorBrief{
			ID:        uuid.UUID(row.CreatorID.Bytes).String(),
			FirstName: row.CreatorFirstName.String,
			LastName:  row.CreatorLastName.String,
			Avatar:    pgTextPtr(row.CreatorAvatar),
		}
	}
	return &dto, nil
}

func (s *service) MarkAsRead(ctx context.Context, id, userID uuid.UUID) error {
	if s.repo == nil {
		return nil
	}
	_, err := s.repo.MarkAsRead(ctx, id, userID)
	return err
}

func (s *service) MarkAllAsRead(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID) error {
	if s.repo == nil {
		return nil
	}
	_, err := s.repo.MarkAllAsRead(ctx, sqlc.MarkAllNotificationsAsReadParams{
		UserID:         userID,
		OrganizationID: toPgUUID(orgID),
	})
	return err
}

func (s *service) DeleteNotification(ctx context.Context, id, userID uuid.UUID) error {
	if s.repo == nil {
		return nil
	}
	_, err := s.repo.DeleteForUser(ctx, id, userID)
	return err
}

func (s *service) DeleteMultipleNotifications(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) error {
	if s.repo == nil || len(ids) == 0 {
		return nil
	}
	_, err := s.repo.DeleteMultipleForUser(ctx, ids, userID)
	return err
}

func (s *service) GetNotificationsByUserAndOrganization(ctx context.Context, userID, orgID uuid.UUID, filters UserOrgFilters, page, limit int) (*UserOrgNotificationsResponse, error) {
	if page < 1 {
		return nil, apperror.BadRequest("Page must be a positive number")
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if s.repo == nil {
		return &UserOrgNotificationsResponse{
			Notifications: []NotificationDTO{},
			Pagination:    pagination.New(page, limit, 0),
			Summary: UserOrgSummaryDTO{
				ByType:     map[string]int64{},
				ByPriority: map[string]int64{},
			},
		}, nil
	}

	offset := pagination.Offset(page, limit)
	rows, err := s.repo.ListUserOrgDetailed(ctx, sqlc.ListNotificationsByUserOrgDetailedParams{
		UserID:         userID,
		OrganizationID: orgID,
		IsRead:         toPgBool(filters.IsRead),
		NotifType:      toNullNotifType(filters.Type),
		Priority:       toNullNotifPriority(filters.Priority),
		StartDate:      toPgTS(filters.StartDate),
		EndDate:        toPgTS(filters.EndDate),
		OffsetVal:      safeInt32(offset),
		LimitVal:       safeInt32(limit),
	})
	if err != nil {
		return nil, fmt.Errorf("list user org notifications: %w", err)
	}

	total, err := s.repo.CountUserOrgDetailed(ctx, sqlc.CountNotificationsByUserOrgDetailedParams{
		UserID:         userID,
		OrganizationID: orgID,
		IsRead:         toPgBool(filters.IsRead),
		NotifType:      toNullNotifType(filters.Type),
		Priority:       toNullNotifPriority(filters.Priority),
		StartDate:      toPgTS(filters.StartDate),
		EndDate:        toPgTS(filters.EndDate),
	})
	if err != nil {
		return nil, fmt.Errorf("count user org notifications: %w", err)
	}

	entityCache := make(map[string]*EntityDetails)
	for _, r := range rows {
		if r.EntityType.Valid && r.EntityID.Valid {
			eid := uuid.UUID(r.EntityID.Bytes)
			key := r.EntityType.String + ":" + eid.String()
			if _, seen := entityCache[key]; !seen {
				entityCache[key] = s.getEntityDetails(ctx, r.EntityType.String, eid)
			}
		}
	}

	items := make([]NotificationDTO, 0, len(rows))
	for _, r := range rows {
		dto := NotificationDTO{
			ID:             r.ID.String(),
			Title:          r.Title,
			Message:        r.Message,
			Type:           string(r.Type),
			Priority:       string(r.Priority),
			IsRead:         r.IsRead,
			EntityType:     pgTextPtr(r.EntityType),
			EntityID:       pgUUIDPtr(r.EntityID),
			ActionURL:      pgTextPtr(r.ActionUrl),
			UserID:         r.UserID.String(),
			OrganizationID: pgUUIDPtr(r.OrganizationID),
			CreatedBy:      pgUUIDPtr(r.CreatedByID),
			UpdatedBy:      pgUUIDPtr(r.UpdatedByID),
			CreatedAt:      pgTSISO(r.CreatedAt),
			UpdatedAt:      pgTSISO(r.UpdatedAt),
			ReadAt:         pgTSISOPtr(r.ReadAt),
			User: &RecipientBrief{
				ID:        r.RecipientID.String(),
				FirstName: r.RecipientFirstName,
				LastName:  r.RecipientLastName,
				Email:     r.RecipientEmail,
				Username:  pgTextPtr(r.RecipientUsername),
				Avatar:    pgTextPtr(r.RecipientAvatar),
			},
		}
		if r.OrgID.Valid {
			dto.Organization = &OrgBrief{
				ID:     uuid.UUID(r.OrgID.Bytes).String(),
				Name:   r.OrgName.String,
				Slug:   r.OrgSlug.String,
				Avatar: pgTextPtr(r.OrgAvatar),
			}
		}
		if r.CreatorID.Valid {
			dto.CreatedByUser = &CreatorBrief{
				ID:        uuid.UUID(r.CreatorID.Bytes).String(),
				FirstName: r.CreatorFirstName.String,
				LastName:  r.CreatorLastName.String,
				Email:     pgTextPtr(r.CreatorEmail),
				Avatar:    pgTextPtr(r.CreatorAvatar),
			}
		}
		if r.EntityType.Valid && r.EntityID.Valid {
			key := r.EntityType.String + ":" + uuid.UUID(r.EntityID.Bytes).String()
			dto.Entity = entityCache[key]
		}
		items = append(items, dto)
	}

	falseVal := false
	unread, err := s.repo.CountUserOrgDetailed(ctx, sqlc.CountNotificationsByUserOrgDetailedParams{
		UserID:         userID,
		OrganizationID: orgID,
		IsRead:         toPgBool(&falseVal),
		NotifType:      toNullNotifType(filters.Type),
		Priority:       toNullNotifPriority(filters.Priority),
		StartDate:      toPgTS(filters.StartDate),
		EndDate:        toPgTS(filters.EndDate),
	})
	if err != nil {
		return nil, fmt.Errorf("count unread user org notifications: %w", err)
	}

	typeRows, err := s.repo.GroupByTypeForUserOrg(ctx, sqlc.GroupNotificationsByTypeForUserOrgParams{
		UserID:         userID,
		OrganizationID: orgID,
		IsRead:         toPgBool(filters.IsRead),
		NotifType:      toNullNotifType(filters.Type),
		Priority:       toNullNotifPriority(filters.Priority),
		StartDate:      toPgTS(filters.StartDate),
		EndDate:        toPgTS(filters.EndDate),
	})
	if err != nil {
		return nil, fmt.Errorf("group user org by type: %w", err)
	}
	byType := make(map[string]int64, len(typeRows))
	for _, tr := range typeRows {
		byType[string(tr.Type)] = tr.Cnt
	}

	prioRows, err := s.repo.GroupByPriorityForUserOrg(ctx, sqlc.GroupNotificationsByPriorityForUserOrgParams{
		UserID:         userID,
		OrganizationID: orgID,
		IsRead:         toPgBool(filters.IsRead),
		NotifType:      toNullNotifType(filters.Type),
		Priority:       toNullNotifPriority(filters.Priority),
		StartDate:      toPgTS(filters.StartDate),
		EndDate:        toPgTS(filters.EndDate),
	})
	if err != nil {
		return nil, fmt.Errorf("group user org by priority: %w", err)
	}
	byPriority := make(map[string]int64, len(prioRows))
	for _, pr := range prioRows {
		byPriority[string(pr.Priority)] = pr.Cnt
	}

	return &UserOrgNotificationsResponse{
		Notifications: items,
		Pagination:    pagination.New(page, limit, total),
		Summary: UserOrgSummaryDTO{
			Total:      total,
			Unread:     unread,
			ByType:     byType,
			ByPriority: byPriority,
		},
	}, nil
}

func (s *service) getEntityDetails(ctx context.Context, entityType string, entityID uuid.UUID) *EntityDetails {
	switch strings.ToLower(strings.TrimSpace(entityType)) {
	case "task":
		t, err := s.repo.GetEntityTask(ctx, entityID)
		if err != nil || t == nil {
			return nil
		}
		var parent map[string]any
		if t.ProjectID.Valid {
			parent = map[string]any{
				"id":   uuid.UUID(t.ProjectID.Bytes).String(),
				"name": t.ProjectName.String,
				"slug": t.ProjectSlug.String,
			}
		}
		var statusObj any
		if t.StatusID.Valid {
			statusObj = map[string]any{
				"id":       uuid.UUID(t.StatusID.Bytes).String(),
				"name":     t.StatusName.String,
				"color":    t.StatusColor.String,
				"category": string(t.StatusCategory.StatusCategory),
			}
		}
		slug := t.Slug
		return &EntityDetails{
			ID:     t.ID.String(),
			Type:   "task",
			Name:   t.Title,
			Slug:   &slug,
			Parent: parent,
			Extra: map[string]any{
				"priority":   string(t.Priority),
				"taskNumber": t.TaskNumber,
				"status":     statusObj,
			},
		}

	case "taskcomment", "task comment":
		c, err := s.repo.GetEntityTaskComment(ctx, entityID)
		if err != nil || c == nil {
			return nil
		}
		var parent map[string]any
		taskTitle := "Task"
		if c.TaskID.Valid {
			if c.TaskTitle.Valid {
				taskTitle = c.TaskTitle.String
			}
			parent = map[string]any{
				"id":         uuid.UUID(c.TaskID.Bytes).String(),
				"title":      c.TaskTitle.String,
				"slug":       c.TaskSlug.String,
				"taskNumber": c.TaskNumber.Int32,
			}
		}
		return &EntityDetails{
			ID:     c.ID.String(),
			Type:   "task_comment",
			Name:   "Comment on " + taskTitle,
			Parent: parent,
			Extra: map[string]any{
				"content":   c.Content,
				"createdAt": pgTSISO(c.CreatedAt),
			},
		}

	case "project":
		p, err := s.repo.GetEntityProject(ctx, entityID)
		if err != nil || p == nil {
			return nil
		}
		var parent map[string]any
		if p.WorkspaceID.Valid {
			parent = map[string]any{
				"id":   uuid.UUID(p.WorkspaceID.Bytes).String(),
				"name": p.WorkspaceName.String,
				"slug": p.WorkspaceSlug.String,
			}
		}
		slug := p.Slug
		return &EntityDetails{
			ID:     p.ID.String(),
			Type:   "project",
			Name:   p.Name,
			Slug:   &slug,
			Parent: parent,
			Extra: map[string]any{
				"avatar":      pgTextPtr(p.Avatar),
				"description": pgTextPtr(p.Description),
			},
		}

	case "workspace":
		w, err := s.repo.GetEntityWorkspace(ctx, entityID)
		if err != nil || w == nil {
			return nil
		}
		var parent map[string]any
		if w.OrgID.Valid {
			parent = map[string]any{
				"id":   uuid.UUID(w.OrgID.Bytes).String(),
				"name": w.OrgName.String,
				"slug": w.OrgSlug.String,
			}
		}
		slug := w.Slug
		return &EntityDetails{
			ID:     w.ID.String(),
			Type:   "workspace",
			Name:   w.Name,
			Slug:   &slug,
			Parent: parent,
			Extra: map[string]any{
				"avatar":      pgTextPtr(w.Avatar),
				"description": pgTextPtr(w.Description),
			},
		}

	case "organization":
		o, err := s.repo.GetEntityOrganization(ctx, entityID)
		if err != nil || o == nil {
			return nil
		}
		slug := o.Slug
		return &EntityDetails{
			ID:   o.ID.String(),
			Type: "organization",
			Name: o.Name,
			Slug: &slug,
			Extra: map[string]any{
				"avatar":      pgTextPtr(o.Avatar),
				"description": pgTextPtr(o.Description),
			},
		}

	case "sprint":
		sp, err := s.repo.GetEntitySprint(ctx, entityID)
		if err != nil || sp == nil {
			return nil
		}
		var parent map[string]any
		if sp.ProjectID.Valid {
			parent = map[string]any{
				"id":   uuid.UUID(sp.ProjectID.Bytes).String(),
				"name": sp.ProjectName.String,
				"slug": sp.ProjectSlug.String,
			}
		}
		return &EntityDetails{
			ID:     sp.ID.String(),
			Type:   "sprint",
			Name:   sp.Name,
			Parent: parent,
			Extra: map[string]any{
				"goal":      pgTextPtr(sp.Goal),
				"status":    string(sp.Status),
				"startDate": pgTSISOPtr(sp.StartDate),
				"endDate":   pgTSISOPtr(sp.EndDate),
			},
		}

	case "invitation":
		inv, err := s.repo.GetEntityInvitation(ctx, entityID)
		if err != nil || inv == nil {
			return nil
		}
		var parent map[string]any
		switch {
		case inv.ProjectID.Valid:
			parent = map[string]any{
				"id":   uuid.UUID(inv.ProjectID.Bytes).String(),
				"name": inv.ProjectName.String,
				"slug": inv.ProjectSlug.String,
			}
		case inv.WorkspaceID.Valid:
			parent = map[string]any{
				"id":   uuid.UUID(inv.WorkspaceID.Bytes).String(),
				"name": inv.WorkspaceName.String,
				"slug": inv.WorkspaceSlug.String,
			}
		case inv.OrgID.Valid:
			parent = map[string]any{
				"id":   uuid.UUID(inv.OrgID.Bytes).String(),
				"name": inv.OrgName.String,
				"slug": inv.OrgSlug.String,
			}
		}
		return &EntityDetails{
			ID:     inv.ID.String(),
			Type:   "invitation",
			Name:   inv.InviteeEmail,
			Parent: parent,
			Extra: map[string]any{
				"role":      inv.Role,
				"status":    string(inv.Status),
				"expiresAt": pgTSISO(inv.ExpiresAt),
			},
		}

	case "taskattachment", "task attachment":
		att, err := s.repo.GetEntityTaskAttachment(ctx, entityID)
		if err != nil || att == nil {
			return nil
		}
		var parent map[string]any
		if att.TaskID.Valid {
			parent = map[string]any{
				"id":    uuid.UUID(att.TaskID.Bytes).String(),
				"title": att.TaskTitle.String,
				"slug":  att.TaskSlug.String,
			}
		}
		return &EntityDetails{
			ID:     att.ID.String(),
			Type:   "task_attachment",
			Name:   att.FileName,
			Parent: parent,
			Extra: map[string]any{
				"fileSize": att.FileSize,
				"mimeType": att.MimeType,
			},
		}
	}
	return nil
}

func (s *service) NotifyTaskAssigned(ctx context.Context, taskID uuid.UUID, assigneeIDs []uuid.UUID, assignedBy uuid.UUID) error {
	if s.repo == nil || len(assigneeIDs) == 0 {
		return nil
	}
	task, err := s.repo.GetTaskContext(ctx, taskID)
	if err != nil || task == nil {
		return err
	}

	entityType := "Task"
	actionURL := "/tasks/" + taskID.String()
	prio := sqlc.NotificationPriorityHIGH
	orgID := task.OrganizationID

	for _, assigneeID := range assigneeIDs {
		if _, cErr := s.CreateNotification(ctx, CreateInput{
			Title:          "New Task Assignment",
			Message:        fmt.Sprintf(`You have been assigned to task "%s"`, task.Title),
			Type:           sqlc.NotificationTypeTASKASSIGNED,
			UserID:         assigneeID,
			OrganizationID: &orgID,
			Priority:       &prio,
			EntityType:     &entityType,
			EntityID:       &taskID,
			ActionURL:      &actionURL,
			CreatedBy:      &assignedBy,
		}); cErr != nil {
			return cErr
		}
	}
	return nil
}

func (s *service) NotifyTaskStatusChanged(ctx context.Context, taskID uuid.UUID, oldStatus, newStatus string, changedBy uuid.UUID) error {
	if s.repo == nil {
		return nil
	}
	task, err := s.repo.GetTaskContext(ctx, taskID)
	if err != nil || task == nil {
		return err
	}
	participants, err := s.repo.ListTaskParticipants(ctx, taskID)
	if err != nil {
		return err
	}

	seen := make(map[uuid.UUID]struct{})
	entityType := "Task"
	actionURL := "/tasks/" + taskID.String()
	orgID := task.OrganizationID

	for _, p := range participants {
		if p.RoleKind != "assignee" && p.RoleKind != "reporter" {
			continue
		}
		if p.UserID == changedBy {
			continue
		}
		if _, dup := seen[p.UserID]; dup {
			continue
		}
		seen[p.UserID] = struct{}{}

		if _, cErr := s.CreateNotification(ctx, CreateInput{
			Title:          "Task Status Updated",
			Message:        fmt.Sprintf(`Task "%s" status changed from %s to %s`, task.Title, oldStatus, newStatus),
			Type:           sqlc.NotificationTypeTASKSTATUSCHANGED,
			UserID:         p.UserID,
			OrganizationID: &orgID,
			EntityType:     &entityType,
			EntityID:       &taskID,
			ActionURL:      &actionURL,
			CreatedBy:      &changedBy,
		}); cErr != nil {
			return cErr
		}
	}
	return nil
}

func (s *service) NotifyTaskDueSoon(ctx context.Context, taskID uuid.UUID) error {
	if s.repo == nil {
		return nil
	}
	task, err := s.repo.GetTaskContext(ctx, taskID)
	if err != nil || task == nil || !task.DueDate.Valid {
		return err
	}
	participants, err := s.repo.ListTaskParticipants(ctx, taskID)
	if err != nil {
		return err
	}

	seen := make(map[uuid.UUID]struct{})
	entityType := "Task"
	actionURL := "/tasks/" + taskID.String()
	prio := sqlc.NotificationPriorityHIGH
	orgID := task.OrganizationID

	for _, p := range participants {
		if p.RoleKind != "assignee" {
			continue
		}
		if _, dup := seen[p.UserID]; dup {
			continue
		}
		seen[p.UserID] = struct{}{}

		if _, cErr := s.CreateNotification(ctx, CreateInput{
			Title:          "Task Due Soon",
			Message:        fmt.Sprintf(`Task "%s" is due soon`, task.Title),
			Type:           sqlc.NotificationTypeTASKDUESOON,
			UserID:         p.UserID,
			OrganizationID: &orgID,
			Priority:       &prio,
			EntityType:     &entityType,
			EntityID:       &taskID,
			ActionURL:      &actionURL,
		}); cErr != nil {
			return cErr
		}
	}
	return nil
}
