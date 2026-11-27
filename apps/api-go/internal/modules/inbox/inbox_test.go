package inbox

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
	"github.com/orbito/api-go/internal/platform/crypto"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/ctxhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubRepo struct {
	projID    uuid.UUID
	inbox     *sqlc.GetProjectInboxByProjectIDRow
	acct      *sqlc.EmailAccount
	messages  map[uuid.UUID]sqlc.GetInboxMessageDetailByIDRow
	inserted  []sqlc.InsertInboxMessageRecordParams
	rules     []sqlc.InboxRule
	converted bool
}

func (s *stubRepo) GetProjectByID(_ context.Context, id uuid.UUID) (*sqlc.GetProjectByIDForInboxRow, error) {
	return &sqlc.GetProjectByIDForInboxRow{ID: id, Slug: "ORB", Name: "Orbito"}, nil
}
func (s *stubRepo) GetInboxByProjectID(_ context.Context, _ uuid.UUID) (*sqlc.GetProjectInboxByProjectIDRow, error) {
	return s.inbox, nil
}
func (s *stubRepo) FindInboxByEmail(_ context.Context, _ string) (*sqlc.GetProjectInboxByEmailAddressRow, error) {
	return nil, nil
}
func (s *stubRepo) CreateInbox(_ context.Context, arg sqlc.CreateProjectInboxRecordParams) (sqlc.ProjectInbox, error) {
	now := pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
	s.inbox = &sqlc.GetProjectInboxByProjectIDRow{
		ID: arg.ID, ProjectID: arg.ProjectID, Enabled: true, Name: arg.Name,
		EmailAddress: arg.EmailAddress, AutoCreateTask: arg.AutoCreateTask,
		DefaultTaskType: arg.DefaultTaskType, DefaultPriority: arg.DefaultPriority, DefaultStatusID: arg.DefaultStatusID,
		CreatedAt: now, UpdatedAt: now,
	}
	return sqlc.ProjectInbox{ID: arg.ID}, nil
}
func (s *stubRepo) UpdateInbox(_ context.Context, _ sqlc.UpdateProjectInboxRecordParams) (sqlc.ProjectInbox, error) {
	return sqlc.ProjectInbox{}, nil
}
func (s *stubRepo) UpdateInboxEmailAddress(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}
func (s *stubRepo) GetEmailAccountByInboxID(_ context.Context, _ uuid.UUID) (*sqlc.EmailAccount, error) {
	return s.acct, nil
}
func (s *stubRepo) GetEmailAccountByID(_ context.Context, _ uuid.UUID) (*sqlc.EmailAccount, error) {
	return s.acct, nil
}
func (s *stubRepo) UpsertEmailAccount(_ context.Context, arg sqlc.UpsertEmailAccountForInboxParams) (sqlc.EmailAccount, error) {
	s.acct = &sqlc.EmailAccount{
		ID: arg.ID, ProjectInboxID: arg.ProjectInboxID, EmailAddress: arg.EmailAddress,
		ImapHost: arg.ImapHost, ImapUsername: arg.ImapUsername, ImapPassword: arg.ImapPassword, ImapFolder: arg.ImapFolder,
		SmtpHost: arg.SmtpHost, SmtpUsername: arg.SmtpUsername, SmtpPassword: arg.SmtpPassword, SyncEnabled: true,
	}
	return *s.acct, nil
}
func (s *stubRepo) ListSyncEnabledAccounts(_ context.Context) ([]sqlc.ListSyncEnabledEmailAccountsRow, error) {
	return nil, nil
}
func (s *stubRepo) GetAccountSyncDetailsByProjectID(_ context.Context, projectID uuid.UUID) (*sqlc.GetEmailAccountSyncDetailsByProjectIDRow, error) {
	return &sqlc.GetEmailAccountSyncDetailsByProjectIDRow{
		ID: s.acct.ID, ProjectInboxID: s.acct.ProjectInboxID, EmailAddress: s.acct.EmailAddress,
		ImapHost: s.acct.ImapHost, ImapUsername: s.acct.ImapUsername, ImapPassword: s.acct.ImapPassword,
		ImapFolder: "INBOX", ProjectID: projectID,
	}, nil
}
func (s *stubRepo) UpdateEmailAccountSyncState(_ context.Context, _ uuid.UUID, _ time.Time, _ *string, _ *int32, _ *string) error {
	return nil
}
func (s *stubRepo) CreateSyncLog(_ context.Context, _ sqlc.CreateEmailSyncLogParams) error {
	return nil
}
func (s *stubRepo) ListMessages(_ context.Context, _ sqlc.ListInboxMessagesFilteredParams) ([]sqlc.ListInboxMessagesFilteredRow, error) {
	return nil, nil
}
func (s *stubRepo) ListMessageAttachmentsBatch(_ context.Context, _ []uuid.UUID) ([]sqlc.ListMessageAttachmentsBatchRow, error) {
	return nil, nil
}
func (s *stubRepo) GetMessageDetailByID(_ context.Context, id uuid.UUID) (*sqlc.GetInboxMessageDetailByIDRow, error) {
	r, ok := s.messages[id]
	if !ok {
		return nil, nil
	}
	return &r, nil
}
func (s *stubRepo) InboxMessageExists(_ context.Context, _ uuid.UUID, _ string) (bool, error) {
	return false, nil
}
func (s *stubRepo) InsertInboxMessage(_ context.Context, arg sqlc.InsertInboxMessageRecordParams) (sqlc.InboxMessage, error) {
	s.inserted = append(s.inserted, arg)
	return sqlc.InboxMessage{ID: arg.ID}, nil
}
func (s *stubRepo) FindTaskByThreadID(_ context.Context, _ string) (*sqlc.FindTaskByEmailThreadIDRow, error) {
	return nil, nil
}
func (s *stubRepo) CreateCommentFromMessage(_ context.Context, _, _ uuid.UUID, _, _ string) error {
	return nil
}
func (s *stubRepo) CreateTaskFromMessageTx(_ context.Context, msg *sqlc.GetInboxMessageDetailByIDRow, _ sqlc.TaskType, _ sqlc.TaskPriority, _ uuid.UUID) (*sqlc.InsertTaskFromInboxMessageRow, error) {
	s.converted = true
	return &sqlc.InsertTaskFromInboxMessageRow{ID: uuid.New(), Title: msg.Subject, Slug: "ORB-1", ProjectID: msg.ProjectID, TaskNumber: 1}, nil
}
func (s *stubRepo) MarkMessageConverted(_ context.Context, _, _ uuid.UUID) error { return nil }
func (s *stubRepo) UpdateMessageStatus(_ context.Context, id uuid.UUID, st sqlc.MessageStatus) (sqlc.InboxMessage, error) {
	now := pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
	return sqlc.InboxMessage{ID: id, Status: st, EmailDate: now, CreatedAt: now, UpdatedAt: now}, nil
}
func (s *stubRepo) CreateRule(_ context.Context, arg sqlc.CreateInboxRuleRecordParams) (sqlc.InboxRule, error) {
	now := pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
	r := sqlc.InboxRule{
		ID: arg.ID, ProjectInboxID: arg.ProjectInboxID, Name: arg.Name, Priority: arg.Priority,
		Enabled: arg.Enabled, Conditions: arg.Conditions, Actions: arg.Actions, StopOnMatch: arg.StopOnMatch,
		CreatedAt: now, UpdatedAt: now,
	}
	s.rules = append(s.rules, r)
	return r, nil
}
func (s *stubRepo) ListRulesByInboxID(_ context.Context, _ uuid.UUID) ([]sqlc.InboxRule, error) {
	return s.rules, nil
}
func (s *stubRepo) UpdateRule(_ context.Context, _ sqlc.UpdateInboxRuleByIDParams) (*sqlc.InboxRule, error) {
	return nil, nil
}
func (s *stubRepo) DeleteRule(_ context.Context, _ uuid.UUID) (int64, error) { return 1, nil }
func (s *stubRepo) GetCommentReplyContext(_ context.Context, _ uuid.UUID) (*sqlc.GetCommentEmailReplyContextRow, error) {
	return nil, nil
}
func (s *stubRepo) MarkCommentSentAsEmail(_ context.Context, _ uuid.UUID, _ string, _ []string) error {
	return nil
}

type stubIMAP struct{ msgs []ParsedIMAPMessage }

func (m *stubIMAP) FetchMessagesSinceUID(_ string, _ uint32) ([]ParsedIMAPMessage, uint32, error) {
	return m.msgs, 42, nil
}
func (m *stubIMAP) Close() error { return nil }

func TestInboxEncryptionRulesSyncAndConvert(t *testing.T) {
	fc, err := crypto.NewFieldCrypto("test-secret-encryption-key-for-inbox")
	require.NoError(t, err)

	projID := uuid.New()
	statusID := uuid.New()
	msgID := uuid.New()
	userID := uuid.New()

	repo := &stubRepo{
		projID: projID,
		messages: map[uuid.UUID]sqlc.GetInboxMessageDetailByIDRow{
			msgID: {
				ID: msgID, ProjectID: projID, Subject: "Production outage",
				ThreadID: "root-msg-1", DefaultTaskType: sqlc.TaskTypeBUG, DefaultPriority: sqlc.TaskPriorityHIGH,
				DefaultStatusID: statusID,
			},
		},
	}

	svc := NewService(repo, fc, nil, nil).(*service)
	svc.imapDial = func(string, int, string, string, string, bool, bool) (IMAPClient, error) {
		return &stubIMAP{msgs: []ParsedIMAPMessage{{
			UID: 42, MessageID: "<m42@ex.com>", Subject: "URGENT: Server down",
			FromEmail: "vip@customer.com", BodyText: "Help please", Date: time.Now().UTC(),
		}}}, nil
	}

	_, err = svc.CreateInbox(context.Background(), projID, CreateInboxRequest{
		Name: "Support", DefaultStatusID: statusID.String(),
	})
	require.NoError(t, err)

	acctResp, err := svc.SetupEmailAccount(context.Background(), projID, SetupEmailRequest{
		EmailAddress: "support@orbito.io", IMAPHost: "imap.ex.com", IMAPUsername: "u", IMAPPassword: "secret-imap-password",
		SMTPHost: "smtp.ex.com", SMTPUsername: "u", SMTPPassword: "secret-smtp-password",
	})
	require.NoError(t, err)
	assert.Equal(t, "support@orbito.io", acctResp.EmailAddress)
	assert.NotEqual(t, "secret-imap-password", repo.acct.ImapPassword.String, "IMAP password must be encrypted in DB")

	_, err = svc.CreateRule(context.Background(), projID, CreateRuleRequest{
		Name:       "VIP Escalation",
		Conditions: map[string]any{"any": []any{map[string]any{"subject": map[string]any{"matches": "urgent|critical"}}}},
		Actions:    map[string]any{"setPriority": "HIGHEST"},
	})
	require.NoError(t, err)

	outcome := EvaluateRules(repo.rules, "vip@customer.com", "URGENT: Server down", "Help")
	assert.True(t, outcome.Matched)
	require.NotNil(t, outcome.Priority)
	assert.Equal(t, sqlc.TaskPriorityHIGHEST, *outcome.Priority)

	count, err := svc.SyncProjectInbox(context.Background(), projID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
	require.Len(t, repo.inserted, 1)

	taskRes, err := svc.ConvertToTask(context.Background(), msgID, userID)
	require.NoError(t, err)
	assert.True(t, repo.converted)
	assert.Equal(t, "ORB-1", taskRes["slug"])

	assert.Equal(t, "root@thread", ExtractThreadID("child@t", "inreply@t", []string{"<root@thread>"}))

	// HTTP route smoke check for templates + inbox convert
	r := chi.NewRouter()
	h := NewHandler(svc)
	r.Mount("/projects/{projectId}/inbox", ProjectInboxRoutes(h, func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			next.ServeHTTP(w, req.WithContext(ctxhelper.WithUserID(req.Context(), userID.String())))
		})
	}))
	r.Mount("/email-templates", EmailTemplatesRoutes(func(next http.Handler) http.Handler { return next }))

	tplReq := httptest.NewRequest(http.MethodGet, "/email-templates/auto-reply-support", nil)
	tplRec := httptest.NewRecorder()
	r.ServeHTTP(tplRec, tplReq)
	assert.Equal(t, http.StatusOK, tplRec.Code)
	assert.Contains(t, tplRec.Body.String(), "Auto Reply - Support")

	body, _ := json.Marshal(map[string]string{})
	ignReq := httptest.NewRequest(http.MethodPost, "/projects/"+projID.String()+"/inbox/messages/"+msgID.String()+"/ignore", bytes.NewReader(body))
	ignRec := httptest.NewRecorder()
	r.ServeHTTP(ignRec, ignReq)
	assert.Equal(t, http.StatusOK, ignRec.Code)
}
