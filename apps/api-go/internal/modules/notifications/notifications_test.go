package notifications

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
	items        []sqlc.Notification
	entityTask   *sqlc.GetNotificationEntityTaskRow
	taskCtx      *sqlc.GetTaskNotificationContextRow
	participants []sqlc.ListTaskNotificationParticipantsRow
	mentioned    []sqlc.FindMentionedUsersRow
	authorBrief  *sqlc.GetUserBriefForNotificationRow
}

func (m *mockRepo) Create(_ context.Context, arg sqlc.CreateNotificationRecordParams) (sqlc.Notification, error) {
	now := pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
	n := sqlc.Notification{
		ID:             arg.ID,
		Title:          arg.Title,
		Message:        arg.Message,
		Type:           arg.Type,
		Priority:       arg.Priority,
		IsRead:         false,
		EntityType:     arg.EntityType,
		EntityID:       arg.EntityID,
		ActionUrl:      arg.ActionUrl,
		UserID:         arg.UserID,
		OrganizationID: arg.OrganizationID,
		CreatedByID:    arg.CreatedByID,
		UpdatedByID:    arg.CreatedByID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	m.items = append([]sqlc.Notification{n}, m.items...)
	return n, nil
}

func (m *mockRepo) filterItems(userID uuid.UUID, isRead pgtype.Bool, notifType sqlc.NullNotificationType, orgID pgtype.UUID) []sqlc.Notification {
	var out []sqlc.Notification
	for _, it := range m.items {
		if it.UserID != userID {
			continue
		}
		if isRead.Valid && it.IsRead != isRead.Bool {
			continue
		}
		if notifType.Valid && it.Type != notifType.NotificationType {
			continue
		}
		if orgID.Valid && (!it.OrganizationID.Valid || it.OrganizationID.Bytes != orgID.Bytes) {
			continue
		}
		out = append(out, it)
	}
	return out
}

func (m *mockRepo) ListUser(_ context.Context, arg sqlc.ListUserNotificationsParams) ([]sqlc.ListUserNotificationsRow, error) {
	filtered := m.filterItems(arg.UserID, arg.IsRead, arg.NotifType, arg.OrganizationID)
	out := make([]sqlc.ListUserNotificationsRow, 0, len(filtered))
	for _, it := range filtered {
		out = append(out, sqlc.ListUserNotificationsRow{
			ID:             it.ID,
			Title:          it.Title,
			Message:        it.Message,
			Type:           it.Type,
			Priority:       it.Priority,
			IsRead:         it.IsRead,
			EntityType:     it.EntityType,
			EntityID:       it.EntityID,
			ActionUrl:      it.ActionUrl,
			UserID:         it.UserID,
			OrganizationID: it.OrganizationID,
			CreatedByID:    it.CreatedByID,
			UpdatedByID:    it.UpdatedByID,
			CreatedAt:      it.CreatedAt,
			UpdatedAt:      it.UpdatedAt,
			ReadAt:         it.ReadAt,
		})
	}
	return out, nil
}

func (m *mockRepo) CountUser(_ context.Context, arg sqlc.CountUserNotificationsParams) (int64, error) {
	return int64(len(m.filterItems(arg.UserID, arg.IsRead, arg.NotifType, arg.OrganizationID))), nil
}

func (m *mockRepo) ListUserOrgDetailed(_ context.Context, arg sqlc.ListNotificationsByUserOrgDetailedParams) ([]sqlc.ListNotificationsByUserOrgDetailedRow, error) {
	filtered := m.filterItems(arg.UserID, arg.IsRead, arg.NotifType, pgtype.UUID{Bytes: arg.OrganizationID, Valid: true})
	out := make([]sqlc.ListNotificationsByUserOrgDetailedRow, 0, len(filtered))
	for _, it := range filtered {
		out = append(out, sqlc.ListNotificationsByUserOrgDetailedRow{
			ID:                 it.ID,
			Title:              it.Title,
			Message:            it.Message,
			Type:               it.Type,
			Priority:           it.Priority,
			IsRead:             it.IsRead,
			EntityType:         it.EntityType,
			EntityID:           it.EntityID,
			ActionUrl:          it.ActionUrl,
			UserID:             it.UserID,
			OrganizationID:     it.OrganizationID,
			CreatedAt:          it.CreatedAt,
			UpdatedAt:          it.UpdatedAt,
			RecipientID:        it.UserID,
			RecipientFirstName: "Ada",
			RecipientLastName:  "Lovelace",
			RecipientEmail:     "ada@orbito.io",
		})
	}
	return out, nil
}

func (m *mockRepo) CountUserOrgDetailed(_ context.Context, arg sqlc.CountNotificationsByUserOrgDetailedParams) (int64, error) {
	return int64(len(m.filterItems(arg.UserID, arg.IsRead, arg.NotifType, pgtype.UUID{Bytes: arg.OrganizationID, Valid: true}))), nil
}

func (m *mockRepo) GroupByTypeForUserOrg(_ context.Context, arg sqlc.GroupNotificationsByTypeForUserOrgParams) ([]sqlc.GroupNotificationsByTypeForUserOrgRow, error) {
	counts := map[sqlc.NotificationType]int64{}
	for _, it := range m.filterItems(arg.UserID, arg.IsRead, arg.NotifType, pgtype.UUID{Bytes: arg.OrganizationID, Valid: true}) {
		counts[it.Type]++
	}
	var out []sqlc.GroupNotificationsByTypeForUserOrgRow
	for k, v := range counts {
		out = append(out, sqlc.GroupNotificationsByTypeForUserOrgRow{Type: k, Cnt: v})
	}
	return out, nil
}

func (m *mockRepo) GroupByPriorityForUserOrg(_ context.Context, arg sqlc.GroupNotificationsByPriorityForUserOrgParams) ([]sqlc.GroupNotificationsByPriorityForUserOrgRow, error) {
	counts := map[sqlc.NotificationPriority]int64{}
	for _, it := range m.filterItems(arg.UserID, arg.IsRead, arg.NotifType, pgtype.UUID{Bytes: arg.OrganizationID, Valid: true}) {
		counts[it.Priority]++
	}
	var out []sqlc.GroupNotificationsByPriorityForUserOrgRow
	for k, v := range counts {
		out = append(out, sqlc.GroupNotificationsByPriorityForUserOrgRow{Priority: k, Cnt: v})
	}
	return out, nil
}

func (m *mockRepo) GetByIDForUser(_ context.Context, id, userID uuid.UUID) (*sqlc.GetNotificationByIDForUserRow, error) {
	for _, it := range m.items {
		if it.ID == id && it.UserID == userID {
			return &sqlc.GetNotificationByIDForUserRow{
				ID:        it.ID,
				Title:     it.Title,
				Message:   it.Message,
				Type:      it.Type,
				Priority:  it.Priority,
				IsRead:    it.IsRead,
				UserID:    it.UserID,
				CreatedAt: it.CreatedAt,
				UpdatedAt: it.UpdatedAt,
			}, nil
		}
	}
	return nil, nil
}

func (m *mockRepo) MarkAsRead(_ context.Context, id, userID uuid.UUID) (int64, error) {
	var count int64
	for i := range m.items {
		if m.items[i].ID == id && m.items[i].UserID == userID {
			m.items[i].IsRead = true
			count++
		}
	}
	return count, nil
}

func (m *mockRepo) MarkAllAsRead(_ context.Context, arg sqlc.MarkAllNotificationsAsReadParams) (int64, error) {
	var count int64
	for i := range m.items {
		if m.items[i].UserID == arg.UserID && !m.items[i].IsRead {
			if arg.OrganizationID.Valid && (!m.items[i].OrganizationID.Valid || m.items[i].OrganizationID.Bytes != arg.OrganizationID.Bytes) {
				continue
			}
			m.items[i].IsRead = true
			count++
		}
	}
	return count, nil
}

func (m *mockRepo) CountUnread(_ context.Context, arg sqlc.CountUnreadNotificationsParams) (int64, error) {
	return int64(len(m.filterItems(arg.UserID, pgtype.Bool{Bool: false, Valid: true}, sqlc.NullNotificationType{}, arg.OrganizationID))), nil
}

func (m *mockRepo) CountUnreadByOrg(_ context.Context, userID uuid.UUID) ([]sqlc.CountUnreadNotificationsByOrganizationRow, error) {
	counts := map[uuid.UUID]int64{}
	for _, it := range m.items {
		if it.UserID == userID && !it.IsRead && it.OrganizationID.Valid {
			counts[uuid.UUID(it.OrganizationID.Bytes)]++
		}
	}
	var out []sqlc.CountUnreadNotificationsByOrganizationRow
	for oid, c := range counts {
		out = append(out, sqlc.CountUnreadNotificationsByOrganizationRow{
			OrganizationID:   pgtype.UUID{Bytes: oid, Valid: true},
			OrganizationName: "Acme Org",
			UnreadCount:      c,
		})
	}
	return out, nil
}

func (m *mockRepo) GroupByTypeForUser(_ context.Context, arg sqlc.GroupNotificationsByTypeForUserParams) ([]sqlc.GroupNotificationsByTypeForUserRow, error) {
	counts := map[sqlc.NotificationType]int64{}
	for _, it := range m.filterItems(arg.UserID, pgtype.Bool{}, sqlc.NullNotificationType{}, arg.OrganizationID) {
		counts[it.Type]++
	}
	var out []sqlc.GroupNotificationsByTypeForUserRow
	for k, v := range counts {
		out = append(out, sqlc.GroupNotificationsByTypeForUserRow{Type: k, Cnt: v})
	}
	return out, nil
}

func (m *mockRepo) DeleteForUser(_ context.Context, id, userID uuid.UUID) (int64, error) {
	var remaining []sqlc.Notification
	var removed int64
	for _, it := range m.items {
		if it.ID == id && it.UserID == userID {
			removed++
			continue
		}
		remaining = append(remaining, it)
	}
	m.items = remaining
	return removed, nil
}

func (m *mockRepo) DeleteMultipleForUser(_ context.Context, ids []uuid.UUID, userID uuid.UUID) (int64, error) {
	idSet := map[uuid.UUID]struct{}{}
	for _, id := range ids {
		idSet[id] = struct{}{}
	}
	var remaining []sqlc.Notification
	var removed int64
	for _, it := range m.items {
		if _, ok := idSet[it.ID]; ok && it.UserID == userID {
			removed++
			continue
		}
		remaining = append(remaining, it)
	}
	m.items = remaining
	return removed, nil
}

func (m *mockRepo) GetEntityTask(_ context.Context, _ uuid.UUID) (*sqlc.GetNotificationEntityTaskRow, error) {
	return m.entityTask, nil
}
func (m *mockRepo) GetEntityTaskComment(_ context.Context, _ uuid.UUID) (*sqlc.GetNotificationEntityTaskCommentRow, error) {
	return nil, nil
}
func (m *mockRepo) GetEntityProject(_ context.Context, _ uuid.UUID) (*sqlc.GetNotificationEntityProjectRow, error) {
	return nil, nil
}
func (m *mockRepo) GetEntityWorkspace(_ context.Context, _ uuid.UUID) (*sqlc.GetNotificationEntityWorkspaceRow, error) {
	return nil, nil
}
func (m *mockRepo) GetEntityOrganization(_ context.Context, _ uuid.UUID) (*sqlc.GetNotificationEntityOrganizationRow, error) {
	return nil, nil
}
func (m *mockRepo) GetEntitySprint(_ context.Context, _ uuid.UUID) (*sqlc.GetNotificationEntitySprintRow, error) {
	return nil, nil
}
func (m *mockRepo) GetEntityInvitation(_ context.Context, _ uuid.UUID) (*sqlc.GetNotificationEntityInvitationRow, error) {
	return nil, nil
}
func (m *mockRepo) GetEntityTaskAttachment(_ context.Context, _ uuid.UUID) (*sqlc.GetNotificationEntityTaskAttachmentRow, error) {
	return nil, nil
}
func (m *mockRepo) GetTaskContext(_ context.Context, _ uuid.UUID) (*sqlc.GetTaskNotificationContextRow, error) {
	return m.taskCtx, nil
}
func (m *mockRepo) ListTaskParticipants(_ context.Context, _ uuid.UUID) ([]sqlc.ListTaskNotificationParticipantsRow, error) {
	return m.participants, nil
}
func (m *mockRepo) FindMentionedUsers(_ context.Context, _ []string, _ []uuid.UUID) ([]sqlc.FindMentionedUsersRow, error) {
	return m.mentioned, nil
}
func (m *mockRepo) GetUserBrief(_ context.Context, _ uuid.UUID) (*sqlc.GetUserBriefForNotificationRow, error) {
	return m.authorBrief, nil
}

type mockEmitter struct {
	emitted []string
}

func (e *mockEmitter) EmitNotification(userID string, _ any) {
	e.emitted = append(e.emitted, userID)
}

type mockEmailSender struct {
	sent []string
}

func (e *mockEmailSender) SendAsync(to, _, _ string, _ map[string]any) {
	e.sent = append(e.sent, to)
}

func withAuthUser(req *http.Request, userID, role string) *http.Request {
	ctx := ctxhelper.WithUserID(req.Context(), userID)
	ctx = ctxhelper.WithUserRole(ctx, role)
	return req.WithContext(ctx)
}

func TestNotificationsLifecycleAndEndpoints(t *testing.T) {
	userID := uuid.New()
	otherUserID := uuid.New()
	orgID := uuid.New()
	taskID := uuid.New()

	repo := &mockRepo{
		entityTask: &sqlc.GetNotificationEntityTaskRow{
			ID:         taskID,
			Title:      "Fix Gateway",
			Slug:       "ORB-12",
			Priority:   sqlc.TaskPriorityHIGH,
			TaskNumber: 12,
		},
	}
	emitter := &mockEmitter{}
	svc := NewService(repo, emitter)
	handler := NewHandler(svc)
	router := Routes(handler, nil)

	entType := "Task"
	created, err := svc.CreateNotification(context.Background(), CreateInput{
		Title:          "Assigned",
		Message:        "Task assigned",
		Type:           sqlc.NotificationTypeTASKASSIGNED,
		UserID:         userID,
		OrganizationID: &orgID,
		EntityType:     &entType,
		EntityID:       &taskID,
	})
	require.NoError(t, err)
	require.Equal(t, []string{userID.String()}, emitter.emitted)

	// GET /unread-count
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, withAuthUser(httptest.NewRequest(http.MethodGet, "/unread-count?organizationId="+orgID.String(), nil), userID.String(), "MEMBER"))
	require.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `{"count":1}`, rr.Body.String())

	// GET /user/{userId}/organization/{orgId} IDOR check
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, withAuthUser(httptest.NewRequest(http.MethodGet, "/user/"+userID.String()+"/organization/"+orgID.String(), nil), otherUserID.String(), "MEMBER"))
	assert.Equal(t, http.StatusForbidden, rr.Code)

	// Allowed for self with hydrated entity
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, withAuthUser(httptest.NewRequest(http.MethodGet, "/user/"+userID.String()+"/organization/"+orgID.String(), nil), userID.String(), "MEMBER"))
	require.Equal(t, http.StatusOK, rr.Code)
	var detailed UserOrgNotificationsResponse
	require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &detailed))
	require.Len(t, detailed.Notifications, 1)
	require.NotNil(t, detailed.Notifications[0].Entity)
	assert.Equal(t, "Fix Gateway", detailed.Notifications[0].Entity.Name)

	// PATCH /{id}/read
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, withAuthUser(httptest.NewRequest(http.MethodPatch, "/"+created.ID+"/read", nil), userID.String(), "MEMBER"))
	require.Equal(t, http.StatusOK, rr.Code)

	// GET /unread-count now 0
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, withAuthUser(httptest.NewRequest(http.MethodGet, "/unread-count", nil), userID.String(), "MEMBER"))
	assert.JSONEq(t, `{"count":0}`, rr.Body.String())

	// DELETE /bulk
	delBody, _ := json.Marshal(DeleteBulkRequest{NotificationIDs: []string{created.ID}})
	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, withAuthUser(httptest.NewRequest(http.MethodDelete, "/bulk", bytes.NewReader(delBody)), userID.String(), "MEMBER"))
	assert.Equal(t, http.StatusNoContent, rr.Code)
	assert.Empty(t, repo.items)
}

func TestCommentHooks_MentionsAndParticipants(t *testing.T) {
	authorID := uuid.New()
	mentionedID := uuid.New()
	watcherID := uuid.New()
	taskID := uuid.New()
	orgID := uuid.New()

	repo := &mockRepo{
		taskCtx: &sqlc.GetTaskNotificationContextRow{
			ID:             taskID,
			Title:          "Ship Phase 5",
			Slug:           "ORB-43",
			OrganizationID: orgID,
		},
		authorBrief: &sqlc.GetUserBriefForNotificationRow{
			ID:        authorID,
			FirstName: "Saswat",
			LastName:  "Mohanty",
		},
		mentioned: []sqlc.FindMentionedUsersRow{
			{
				ID:        mentionedID,
				Username:  pgtype.Text{String: "alice", Valid: true},
				Email:     "alice@orbito.io",
				FirstName: "Alice",
				LastName:  "Dev",
			},
		},
		participants: []sqlc.ListTaskNotificationParticipantsRow{
			{UserID: watcherID, RoleKind: "watcher"},
			{UserID: authorID, RoleKind: "assignee"},
		},
	}

	svc := NewService(repo, nil)
	mail := &mockEmailSender{}
	hooks := NewCommentHooks(svc, repo, mail, "http://localhost:3001")

	hooks.OnCommentCreated(context.Background(), uuid.New(), taskID, authorID, "Hey @alice please review @[mention:"+mentionedID.String()+"]")
	require.Len(t, repo.items, 2)
	assert.Equal(t, []string{"alice@orbito.io"}, mail.sent)

	// OnCommentUpdated with identical content should NOT duplicate mention notifications
	hooks.OnCommentUpdated(context.Background(), uuid.New(), taskID, authorID, "Hey @alice please review", "Hey @alice please review")
	assert.Len(t, repo.items, 2)
}

