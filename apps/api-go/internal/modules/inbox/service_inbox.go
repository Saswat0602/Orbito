package inbox

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/crypto"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/platform/queue"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// QueueEnqueuer abstracts enqueueing async inbox sync jobs.
type QueueEnqueuer interface {
	Enqueue(ctx context.Context, queueName, taskType string, payload any) error
}

// Service defines project inbox, rules, email sync, reply, and template operations.
type Service interface {
	CreateInbox(ctx context.Context, projectID uuid.UUID, req CreateInboxRequest) (*ConfigResponse, error)
	GetInbox(ctx context.Context, projectID uuid.UUID) (*ConfigResponse, error)
	UpdateInbox(ctx context.Context, projectID uuid.UUID, req UpdateInboxRequest) (*ConfigResponse, error)
	SetupEmailAccount(ctx context.Context, projectID uuid.UUID, req SetupEmailRequest) (*EmailAccountResponse, error)
	TriggerSyncJob(ctx context.Context, projectID uuid.UUID, userID *uuid.UUID) (string, error)
	SyncProjectInbox(ctx context.Context, projectID uuid.UUID) (int, error)
	SyncAllInboxes(ctx context.Context)
	GetMessages(ctx context.Context, projectID uuid.UUID, f MessageFilters) ([]MessageResponse, error)
	GetMessage(ctx context.Context, messageID uuid.UUID) (*MessageResponse, error)
	ConvertToTask(ctx context.Context, messageID, userID uuid.UUID) (map[string]any, error)
	UpdateMessageStatus(ctx context.Context, messageID uuid.UUID, status string) (*MessageResponse, error)

	GetRules(ctx context.Context, projectID uuid.UUID) ([]RuleResponse, error)
	CreateRule(ctx context.Context, projectID uuid.UUID, req CreateRuleRequest) (*RuleResponse, error)
	UpdateRule(ctx context.Context, ruleID uuid.UUID, req CreateRuleRequest) (*RuleResponse, error)
	DeleteRule(ctx context.Context, ruleID uuid.UUID) error

	SendCommentAsEmail(ctx context.Context, commentID uuid.UUID) (map[string]any, error)
	TestEmailConfiguration(ctx context.Context, accountID uuid.UUID) (map[string]any, error)
}

type service struct {
	repo     Repository
	crypto   *crypto.FieldCrypto
	enqueuer QueueEnqueuer
	imapDial IMAPDialer
	log      *slog.Logger
}

// NewService constructs an inbox service.
func NewService(repo Repository, fc *crypto.FieldCrypto, enqueuer QueueEnqueuer, log *slog.Logger) Service {
	if log == nil {
		log = slog.Default()
	}
	return &service{
		repo:     repo,
		crypto:   fc,
		enqueuer: enqueuer,
		imapDial: defaultIMAPDialer,
		log:      log,
	}
}

func (s *service) CreateInbox(ctx context.Context, projectID uuid.UUID, req CreateInboxRequest) (*ConfigResponse, error) {
	proj, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("check project: %w", err)
	}
	if proj == nil {
		return nil, apperror.NotFound("Project not found")
	}
	existing, err := s.repo.GetInboxByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("check existing inbox: %w", err)
	}
	if existing != nil {
		return s.hydrateInbox(ctx, existing)
	}

	if req.EmailAddress != nil && strings.TrimSpace(*req.EmailAddress) != "" {
		inUse, eErr := s.repo.FindInboxByEmail(ctx, strings.TrimSpace(*req.EmailAddress))
		if eErr != nil {
			return nil, fmt.Errorf("check inbox email: %w", eErr)
		}
		if inUse != nil {
			return nil, apperror.BadRequest("Email address " + *req.EmailAddress + " is already in use by another project inbox")
		}
	}

	statusID, err := uuid.Parse(strings.TrimSpace(req.DefaultStatusID))
	if err != nil {
		return nil, apperror.BadRequest("defaultStatusId must be a valid UUID")
	}
	assigneeUUID, err := parseOptUUID(req.DefaultAssigneeID, "defaultAssigneeId")
	if err != nil {
		return nil, err
	}

	syncInt := int32(5)
	if req.SyncInterval != nil && strings.TrimSpace(*req.SyncInterval) != "" {
		if parsed, convErr := strconv.ParseInt(strings.TrimSpace(*req.SyncInterval), 10, 32); convErr == nil && parsed > 0 {
			syncInt = int32(parsed)
		}
	}

	taskType := sqlc.TaskTypeTASK
	if req.DefaultTaskType != nil && *req.DefaultTaskType != "" {
		taskType = sqlc.TaskType(strings.ToUpper(*req.DefaultTaskType))
	}
	priority := sqlc.TaskPriorityMEDIUM
	if req.DefaultPriority != nil && *req.DefaultPriority != "" {
		priority = sqlc.TaskPriority(strings.ToUpper(*req.DefaultPriority))
	}

	autoReply := req.AutoReplyEnabled != nil && *req.AutoReplyEnabled
	autoCreate := true
	if req.AutoCreateTask != nil {
		autoCreate = *req.AutoCreateTask
	}

	_, err = s.repo.CreateInbox(ctx, sqlc.CreateProjectInboxRecordParams{
		ID:                uuid.New(),
		ProjectID:         projectID,
		Name:              strings.TrimSpace(req.Name),
		Description:       toPgText(req.Description),
		EmailAddress:      toPgText(req.EmailAddress),
		EmailSignature:    toPgText(req.EmailSignature),
		AutoReplyEnabled:  autoReply,
		AutoReplyTemplate: toPgText(req.AutoReplyTemplate),
		AutoCreateTask:    autoCreate,
		DefaultTaskType:   taskType,
		DefaultPriority:   priority,
		DefaultStatusID:   statusID,
		DefaultAssigneeID: assigneeUUID,
		SyncInterval:      pgtype.Int4{Int32: syncInt, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return s.GetInbox(ctx, projectID)
}

func (s *service) GetInbox(ctx context.Context, projectID uuid.UUID) (*ConfigResponse, error) {
	row, err := s.repo.GetInboxByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("get inbox: %w", err)
	}
	if row == nil {
		return nil, nil
	}
	return s.hydrateInbox(ctx, row)
}

func (s *service) hydrateInbox(ctx context.Context, row *sqlc.GetProjectInboxByProjectIDRow) (*ConfigResponse, error) {
	acct, err := s.repo.GetEmailAccountByInboxID(ctx, row.ID)
	if err != nil {
		return nil, fmt.Errorf("load inbox email account: %w", err)
	}
	resp := mapInboxRow(row, acct)
	return &resp, nil
}

func (s *service) UpdateInbox(ctx context.Context, projectID uuid.UUID, req UpdateInboxRequest) (*ConfigResponse, error) {
	current, err := s.repo.GetInboxByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("get inbox: %w", err)
	}
	if current == nil {
		return nil, apperror.NotFound("Inbox not found for this project")
	}

	if req.EmailAddress != nil && *req.EmailAddress != current.EmailAddress.String {
		inUse, eErr := s.repo.FindInboxByEmail(ctx, strings.TrimSpace(*req.EmailAddress))
		if eErr != nil {
			return nil, fmt.Errorf("check inbox email: %w", eErr)
		}
		if inUse != nil && inUse.ID != current.ID {
			return nil, apperror.BadRequest("Email address " + *req.EmailAddress + " is already in use by another project inbox")
		}
	}

	params := sqlc.UpdateProjectInboxRecordParams{
		ID:                current.ID,
		Name:              toPgText(req.Name),
		Description:       toPgText(req.Description),
		EmailAddress:      toPgText(req.EmailAddress),
		EmailSignature:    toPgText(req.EmailSignature),
		AutoReplyTemplate: toPgText(req.AutoReplyTemplate),
	}
	if req.AutoReplyEnabled != nil {
		params.AutoReplyEnabled = pgtype.Bool{Bool: *req.AutoReplyEnabled, Valid: true}
	}
	if req.AutoCreateTask != nil {
		params.AutoCreateTask = pgtype.Bool{Bool: *req.AutoCreateTask, Valid: true}
	}
	if req.DefaultTaskType != nil {
		params.DefaultTaskType = sqlc.NullTaskType{TaskType: sqlc.TaskType(strings.ToUpper(*req.DefaultTaskType)), Valid: true}
	}
	if req.DefaultPriority != nil {
		params.DefaultPriority = sqlc.NullTaskPriority{TaskPriority: sqlc.TaskPriority(strings.ToUpper(*req.DefaultPriority)), Valid: true}
	}
	if st, sErr := parseOptUUID(req.DefaultStatusID, "defaultStatusId"); sErr != nil {
		return nil, sErr
	} else if st.Valid {
		params.DefaultStatusID = st
	}
	if asg, aErr := parseOptUUID(req.DefaultAssigneeID, "defaultAssigneeId"); aErr != nil {
		return nil, aErr
	} else if asg.Valid {
		params.DefaultAssigneeID = asg
	}

	if _, uErr := s.repo.UpdateInbox(ctx, params); uErr != nil {
		return nil, uErr
	}
	return s.GetInbox(ctx, projectID)
}

func (s *service) SetupEmailAccount(ctx context.Context, projectID uuid.UUID, req SetupEmailRequest) (*EmailAccountResponse, error) {
	current, err := s.repo.GetInboxByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("get inbox: %w", err)
	}
	if current == nil {
		return nil, apperror.NotFound("Inbox not found for this project")
	}

	if req.EmailAddress != current.EmailAddress.String {
		inUse, eErr := s.repo.FindInboxByEmail(ctx, req.EmailAddress)
		if eErr != nil {
			return nil, fmt.Errorf("check email uniqueness: %w", eErr)
		}
		if inUse != nil && inUse.ID != current.ID {
			return nil, apperror.BadRequest("Email address " + req.EmailAddress + " is already in use by another project inbox")
		}
	}

	encIMAP, encErr := s.encryptSecret(req.IMAPPassword)
	if encErr != nil {
		return nil, encErr
	}
	encSMTP, encErr := s.encryptSecret(req.SMTPPassword)
	if encErr != nil {
		return nil, encErr
	}

	if uErr := s.repo.UpdateInboxEmailAddress(ctx, current.ID, req.EmailAddress); uErr != nil {
		return nil, uErr
	}

	imapPort := int32(993)
	if req.IMAPPort != nil {
		imapPort = *req.IMAPPort
	}
	smtpPort := int32(587)
	if req.SMTPPort != nil {
		smtpPort = *req.SMTPPort
	}
	imapFolder := "INBOX"
	if req.IMAPFolder != nil && strings.TrimSpace(*req.IMAPFolder) != "" {
		imapFolder = strings.TrimSpace(*req.IMAPFolder)
	}
	imapSSL := req.IMAPUseSSL == nil || *req.IMAPUseSSL
	imapTLSReject := req.IMAPTlsRejectUnauth == nil || *req.IMAPTlsRejectUnauth
	smtpTLSReject := req.SMTPTlsRejectUnauth == nil || *req.SMTPTlsRejectUnauth
	smtpReqTLS := req.SMTPRequireTLS != nil && *req.SMTPRequireTLS

	acct, err := s.repo.UpsertEmailAccount(ctx, sqlc.UpsertEmailAccountForInboxParams{
		ID:                        uuid.New(),
		ProjectInboxID:            current.ID,
		EmailAddress:              strings.TrimSpace(req.EmailAddress),
		DisplayName:               toPgText(req.DisplayName),
		ImapHost:                  pgtype.Text{String: req.IMAPHost, Valid: true},
		ImapPort:                  pgtype.Int4{Int32: imapPort, Valid: true},
		ImapUsername:              pgtype.Text{String: req.IMAPUsername, Valid: true},
		ImapPassword:              pgtype.Text{String: encIMAP, Valid: encIMAP != ""},
		ImapUseSsl:                imapSSL,
		ImapTlsRejectUnauthorized: imapTLSReject,
		ImapTlsMinVersion:         toPgTextDefault(req.IMAPTlsMinVersion, "TLSv1.2"),
		ImapServername:            toPgText(req.IMAPServername),
		ImapFolder:                imapFolder,
		SmtpHost:                  pgtype.Text{String: req.SMTPHost, Valid: true},
		SmtpPort:                  pgtype.Int4{Int32: smtpPort, Valid: true},
		SmtpUsername:              pgtype.Text{String: req.SMTPUsername, Valid: true},
		SmtpPassword:              pgtype.Text{String: encSMTP, Valid: encSMTP != ""},
		SmtpTlsRejectUnauthorized: smtpTLSReject,
		SmtpTlsMinVersion:         toPgTextDefault(req.SMTPTlsMinVersion, "TLSv1.2"),
		SmtpServername:            toPgText(req.SMTPServername),
		SmtpRequireTls:            smtpReqTLS,
	})
	if err != nil {
		return nil, err
	}
	res := mapEmailAccount(&acct)
	return &res, nil
}

func (s *service) TriggerSyncJob(ctx context.Context, projectID uuid.UUID, userID *uuid.UUID) (string, error) {
	jobID := uuid.NewString()
	if s.enqueuer != nil {
		err := s.enqueuer.Enqueue(ctx, queue.QueueEmailSync, queue.TaskSyncProjectInbox, EmailSyncJobPayload{
			ProjectID: projectID,
			UserID:    userID,
			Type:      "manual",
		})
		if err != nil {
			return "", fmt.Errorf("enqueue inbox sync job: %w", err)
		}
		return jobID, nil
	}
	_, err := s.SyncProjectInbox(ctx, projectID)
	return jobID, err
}

func (s *service) GetMessages(ctx context.Context, projectID uuid.UUID, f MessageFilters) ([]MessageResponse, error) {
	current, err := s.repo.GetInboxByProjectID(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("get inbox: %w", err)
	}
	if current == nil {
		return nil, apperror.NotFound("Inbox not found for this project")
	}

	params := sqlc.ListInboxMessagesFilteredParams{
		ProjectInboxID: current.ID,
		IncludeSpam:    f.IncludeSpam,
	}
	if f.ConvertedOnly {
		params.Status = sqlc.NullMessageStatus{MessageStatus: sqlc.MessageStatusCONVERTED, Valid: true}
	} else if f.Status != "" {
		st := sqlc.MessageStatus(strings.ToUpper(strings.TrimSpace(f.Status)))
		params.Status = sqlc.NullMessageStatus{MessageStatus: st, Valid: true}
	}
	if f.FromEmail != "" {
		params.FromEmail = pgtype.Text{String: escapeLike(f.FromEmail), Valid: true}
	}
	if f.Search != "" {
		params.SearchText = pgtype.Text{String: escapeLike(f.Search), Valid: true}
	}
	if f.FromDate != "" {
		if t, pErr := time.Parse(time.RFC3339, f.FromDate); pErr == nil {
			params.FromDate = pgtype.Timestamp{Time: t.UTC(), Valid: true}
		}
	}
	if f.ToDate != "" {
		if t, pErr := time.Parse(time.RFC3339, f.ToDate); pErr == nil {
			params.ToDate = pgtype.Timestamp{Time: t.UTC(), Valid: true}
		}
	}

	rows, err := s.repo.ListMessages(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list inbox messages: %w", err)
	}
	msgIDs := make([]uuid.UUID, 0, len(rows))
	for i := range rows {
		msgIDs = append(msgIDs, rows[i].ID)
	}
	atts, _ := s.repo.ListMessageAttachmentsBatch(ctx, msgIDs)
	attByMsg := make(map[uuid.UUID][]map[string]any)
	for _, a := range atts {
		attByMsg[a.MessageID] = append(attByMsg[a.MessageID], map[string]any{
			"id": a.ID, "filename": a.Filename, "mimeType": a.MimeType, "size": a.Size, "storageUrl": pgTextPtr(a.StorageUrl),
		})
	}

	out := make([]MessageResponse, 0, len(rows))
	for i := range rows {
		out = append(out, mapMessageListRow(&rows[i], attByMsg[rows[i].ID]))
	}
	return out, nil
}

func (s *service) GetMessage(ctx context.Context, messageID uuid.UUID) (*MessageResponse, error) {
	detail, err := s.repo.GetMessageDetailByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("get inbox message: %w", err)
	}
	if detail == nil {
		return nil, apperror.NotFound("Message not found")
	}
	atts, _ := s.repo.ListMessageAttachmentsBatch(ctx, []uuid.UUID{messageID})
	attList := make([]map[string]any, 0, len(atts))
	for _, a := range atts {
		attList = append(attList, map[string]any{
			"id": a.ID, "filename": a.Filename, "mimeType": a.MimeType, "size": a.Size, "storageUrl": pgTextPtr(a.StorageUrl),
		})
	}
	res := mapMessageDetailRow(detail, attList)
	return &res, nil
}

func (s *service) ConvertToTask(ctx context.Context, messageID, userID uuid.UUID) (map[string]any, error) {
	detail, err := s.repo.GetMessageDetailByID(ctx, messageID)
	if err != nil {
		return nil, fmt.Errorf("load message: %w", err)
	}
	if detail == nil {
		return nil, apperror.NotFound("Message not found")
	}
	if detail.Converted {
		return nil, apperror.BadRequest("Message already converted to task")
	}

	if detail.ThreadID != "" {
		existingTask, tErr := s.repo.FindTaskByThreadID(ctx, detail.ThreadID)
		if tErr == nil && existingTask != nil {
			content := detail.BodyText.String
			if content == "" {
				content = detail.BodyHtml.String
			}
			if content == "" {
				content = detail.Subject
			}
			if cErr := s.repo.CreateCommentFromMessage(ctx, existingTask.ID, userID, content, detail.MessageID); cErr != nil {
				return nil, fmt.Errorf("create reply comment: %w", cErr)
			}
			_ = s.repo.MarkMessageConverted(ctx, messageID, userID)
			return map[string]any{
				"id": existingTask.ID, "projectId": existingTask.ProjectID, "title": existingTask.Title, "slug": existingTask.Slug,
			}, nil
		}
	}

	taskRow, err := s.repo.CreateTaskFromMessageTx(ctx, detail, detail.DefaultTaskType, detail.DefaultPriority, userID)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"id": taskRow.ID, "projectId": taskRow.ProjectID, "title": taskRow.Title, "slug": taskRow.Slug,
		"taskNumber": taskRow.TaskNumber, "statusId": taskRow.StatusID,
	}, nil
}

func (s *service) UpdateMessageStatus(ctx context.Context, messageID uuid.UUID, status string) (*MessageResponse, error) {
	st := sqlc.MessageStatus(strings.ToUpper(strings.TrimSpace(status)))
	row, err := s.repo.UpdateMessageStatus(ctx, messageID, st)
	if err != nil {
		return nil, err
	}
	return &MessageResponse{
		ID: row.ID, ProjectInboxID: row.ProjectInboxID, MessageID: row.MessageID, ThreadID: row.ThreadID,
		Subject: row.Subject, FromEmail: row.FromEmail, Status: string(row.Status), Converted: row.Converted,
		EmailDate: formatTime(row.EmailDate.Time), CreatedAt: formatTime(row.CreatedAt.Time), UpdatedAt: formatTime(row.UpdatedAt.Time),
	}, nil
}

func (s *service) encryptSecret(plain string) (string, error) {
	if plain == "" || s.crypto == nil {
		return plain, nil
	}
	enc, err := s.crypto.Encrypt(plain)
	if err != nil {
		return "", fmt.Errorf("encrypt credential: %w", err)
	}
	return enc, nil
}

func parseOptUUID(raw *string, field string) (pgtype.UUID, error) {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return pgtype.UUID{}, nil
	}
	id, err := uuid.Parse(strings.TrimSpace(*raw))
	if err != nil {
		return pgtype.UUID{}, apperror.BadRequest(field + " must be a valid UUID")
	}
	return pgtype.UUID{Bytes: id, Valid: true}, nil
}

func toPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func toPgTextDefault(s *string, def string) pgtype.Text {
	if s == nil || strings.TrimSpace(*s) == "" {
		return pgtype.Text{String: def, Valid: true}
	}
	return pgtype.Text{String: strings.TrimSpace(*s), Valid: true}
}

func pgTextPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	v := t.String
	return &v
}

func escapeLike(q string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(q)
}
