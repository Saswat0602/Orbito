package inbox

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/mail"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/platform/queue"
)

// ParsedIMAPMessage holds a fetched raw message normalized from an IMAP mailbox.
type ParsedIMAPMessage struct {
	UID        uint32
	MessageID  string
	InReplyTo  string
	References []string
	Subject    string
	FromEmail  string
	FromName   string
	ToEmails   []string
	CcEmails   []string
	BodyText   string
	Date       time.Time
}

// IMAPClient abstracts IMAP mailbox operations for production dialers and unit tests.
type IMAPClient interface {
	FetchMessagesSinceUID(folder string, minUID uint32) ([]ParsedIMAPMessage, uint32, error)
	Close() error
}

// IMAPDialer establishes an authenticated IMAP connection for a stored EmailAccount.
type IMAPDialer func(host string, port int, user, pass, servername string, useSSL, rejectUnauth bool) (IMAPClient, error)

// RegisterProcessor binds the background email-sync worker on the queue runtime.
func RegisterProcessor(rt *queue.Runtime, svc Service) {
	rt.RegisterQueue(queue.QueueEmailSync)
	rt.RegisterHandler(queue.TaskSyncProjectInbox, func(ctx context.Context, raw []byte) error {
		var job EmailSyncJobPayload
		if err := json.Unmarshal(raw, &job); err != nil {
			return fmt.Errorf("decode email sync payload: %w", err)
		}
		_, err := svc.SyncProjectInbox(ctx, job.ProjectID)
		return err
	})
}

func (s *service) SyncAllInboxes(ctx context.Context) {
	accounts, err := s.repo.ListSyncEnabledAccounts(ctx)
	if err != nil {
		s.log.Error("ListSyncEnabledAccounts failed", slog.String("error", err.Error()))
		return
	}
	now := time.Now().UTC()
	for i := range accounts {
		acct := &accounts[i]
		intervalMins := int64(5)
		if acct.SyncInterval.Valid && acct.SyncInterval.Int32 > 0 {
			intervalMins = int64(acct.SyncInterval.Int32)
		}
		if acct.LastSyncAt.Valid && now.Sub(acct.LastSyncAt.Time.UTC()) < time.Duration(intervalMins)*time.Minute {
			continue
		}
		if _, syncErr := s.syncAccount(ctx, accountSyncTarget{
			account: &sqlc.EmailAccount{
				ID: acct.ID, ProjectInboxID: acct.ProjectInboxID, EmailAddress: acct.EmailAddress,
				ImapHost: acct.ImapHost, ImapPort: acct.ImapPort, ImapUsername: acct.ImapUsername,
				ImapPassword: acct.ImapPassword, ImapUseSsl: acct.ImapUseSsl,
				ImapTlsRejectUnauthorized: acct.ImapTlsRejectUnauthorized, ImapServername: acct.ImapServername,
				ImapFolder: acct.ImapFolder, UidNext: acct.UidNext,
			},
			projectID: acct.ProjectID,
		}); syncErr != nil {
			s.log.Warn("Scheduled inbox sync failed", slog.String("email", acct.EmailAddress), slog.String("error", syncErr.Error()))
		}
	}
}

func (s *service) SyncProjectInbox(ctx context.Context, projectID uuid.UUID) (int, error) {
	details, err := s.repo.GetAccountSyncDetailsByProjectID(ctx, projectID)
	if err != nil {
		return 0, fmt.Errorf("load project inbox sync details: %w", err)
	}
	if details == nil {
		return 0, fmt.Errorf("email account not configured for project %s", projectID)
	}
	return s.syncAccount(ctx, accountSyncTarget{
		account: &sqlc.EmailAccount{
			ID: details.ID, ProjectInboxID: details.ProjectInboxID, EmailAddress: details.EmailAddress,
			ImapHost: details.ImapHost, ImapPort: details.ImapPort, ImapUsername: details.ImapUsername,
			ImapPassword: details.ImapPassword, ImapUseSsl: details.ImapUseSsl,
			ImapTlsRejectUnauthorized: details.ImapTlsRejectUnauthorized, ImapServername: details.ImapServername,
			ImapFolder: details.ImapFolder, UidNext: details.UidNext,
		},
		projectID: details.ProjectID,
	})
}

type accountSyncTarget struct {
	account   *sqlc.EmailAccount
	projectID uuid.UUID
}

func (s *service) syncAccount(ctx context.Context, target accountSyncTarget) (int, error) {
	started := time.Now().UTC()
	acct := target.account
	pass := acct.ImapPassword.String
	if s.crypto != nil && pass != "" {
		pass = s.crypto.Decrypt(pass)
	}
	port := 993
	if acct.ImapPort.Valid && acct.ImapPort.Int32 > 0 {
		port = int(acct.ImapPort.Int32)
	}

	client, err := s.imapDial(acct.ImapHost.String, port, acct.ImapUsername.String, pass, acct.ImapServername.String, acct.ImapUseSsl, acct.ImapTlsRejectUnauthorized)
	if err != nil {
		errMsg := err.Error()
		_ = s.repo.UpdateEmailAccountSyncState(ctx, acct.ID, started, &errMsg, nil, nil)
		_ = s.repo.CreateSyncLog(ctx, sqlc.CreateEmailSyncLogParams{
			ID: uuid.New(), ProjectID: target.projectID, StartedAt: pgtype.Timestamp{Time: started, Valid: true},
			CompletedAt: pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}, Status: sqlc.SyncStatusFAILED,
			Error: pgtype.Text{String: errMsg, Valid: true}, MessagesProcessed: 0,
		})
		return 0, err
	}
	defer func() { _ = client.Close() }()

	minUID := uint32(1)
	if acct.UidNext.Valid && acct.UidNext.Int32 > 0 {
		minUID = uint32(acct.UidNext.Int32) // #nosec G115 -- guarded by Int32 > 0
	}
	msgs, maxUID, fetchErr := client.FetchMessagesSinceUID(acct.ImapFolder, minUID)
	if fetchErr != nil {
		errMsg := fetchErr.Error()
		_ = s.repo.UpdateEmailAccountSyncState(ctx, acct.ID, started, &errMsg, nil, nil)
		return 0, fetchErr
	}

	rules, _ := s.repo.ListRulesByInboxID(ctx, acct.ProjectInboxID)
	processed := 0
	for i := range msgs {
		m := &msgs[i]
		exists, _ := s.repo.InboxMessageExists(ctx, acct.ProjectInboxID, m.MessageID)
		if exists {
			continue
		}
		threadID := ExtractThreadID(m.MessageID, m.InReplyTo, m.References)
		outcome := EvaluateRules(rules, m.FromEmail, m.Subject, m.BodyText)
		snippet := m.BodyText
		if len(snippet) > 160 {
			snippet = snippet[:160]
		}
		headersBytes, _ := json.Marshal(map[string]any{"from": m.FromEmail, "subject": m.Subject})
		_, insErr := s.repo.InsertInboxMessage(ctx, sqlc.InsertInboxMessageRecordParams{
			ID:             uuid.New(),
			ProjectInboxID: acct.ProjectInboxID,
			MessageID:      m.MessageID,
			ImapUid:        pgtype.Int8{Int64: int64(m.UID), Valid: true},
			ThreadID:       threadID,
			InReplyTo:      pgtype.Text{String: m.InReplyTo, Valid: m.InReplyTo != ""},
			References:     emptyIfNil(m.References),
			Subject:        m.Subject,
			FromEmail:      m.FromEmail,
			FromName:       pgtype.Text{String: m.FromName, Valid: m.FromName != ""},
			ToEmails:       emptyIfNil(m.ToEmails),
			CcEmails:       emptyIfNil(m.CcEmails),
			BccEmails:      []string{},
			BodyText:       pgtype.Text{String: m.BodyText, Valid: m.BodyText != ""},
			Snippet:        pgtype.Text{String: snippet, Valid: snippet != ""},
			Headers:        headersBytes,
			HasAttachments: false,
			EmailDate:      pgtype.Timestamp{Time: m.Date.UTC(), Valid: true},
		})
		if insErr != nil {
			continue
		}
		if outcome.MarkAsSpam {
			s.log.Debug("Marked incoming message per inbox rule", slog.String("messageId", m.MessageID))
		}
		processed++
	}

	var nextUIDPtr *int32
	if maxUID >= minUID && len(msgs) > 0 {
		next := int32(maxUID + 1) // #nosec G115 -- IMAP UIDs bound well below MaxInt32 in sync ranges
		nextUIDPtr = &next
	}
	_ = s.repo.UpdateEmailAccountSyncState(ctx, acct.ID, started, nil, nextUIDPtr, nil)
	_ = s.repo.CreateSyncLog(ctx, sqlc.CreateEmailSyncLogParams{
		ID: uuid.New(), ProjectID: target.projectID, StartedAt: pgtype.Timestamp{Time: started, Valid: true},
		CompletedAt: pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}, Status: sqlc.SyncStatusSUCCESS,
		MessagesProcessed: int32(processed), // #nosec G115
	})
	return processed, nil
}

// ExtractThreadID resolves the RFC 5322 root thread ID from References, In-Reply-To, or Message-ID.
func ExtractThreadID(messageID, inReplyTo string, references []string) string {
	clean := func(id string) string { return strings.Trim(strings.TrimSpace(id), "<>") }
	for _, ref := range references {
		if c := clean(ref); c != "" {
			return c
		}
	}
	if c := clean(inReplyTo); c != "" {
		return c
	}
	if c := clean(messageID); c != "" {
		return c
	}
	return uuid.NewString()
}

type liveIMAPClient struct {
	c *imapclient.Client
}

func defaultIMAPDialer(host string, port int, user, pass, servername string, useSSL, rejectUnauth bool) (IMAPClient, error) {
	if host == "" {
		return nil, fmt.Errorf("IMAP host is not configured")
	}
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	sni := servername
	if sni == "" {
		sni = host
	}
	tlsCfg := &tls.Config{
		ServerName:         sni,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: !rejectUnauth, // #nosec G402 -- explicit per-account configuration flag
	}
	opts := &imapclient.Options{TLSConfig: tlsCfg}
	var c *imapclient.Client
	var err error
	if useSSL {
		c, err = imapclient.DialTLS(addr, opts)
	} else {
		c, err = imapclient.DialStartTLS(addr, opts)
	}
	if err != nil {
		return nil, fmt.Errorf("connect IMAP: %w", err)
	}
	if loginErr := c.Login(user, pass).Wait(); loginErr != nil {
		_ = c.Close()
		return nil, fmt.Errorf("IMAP login failed: %w", loginErr)
	}
	return &liveIMAPClient{c: c}, nil
}

func (l *liveIMAPClient) FetchMessagesSinceUID(folder string, minUID uint32) ([]ParsedIMAPMessage, uint32, error) {
	if folder == "" {
		folder = "INBOX"
	}
	if _, err := l.c.Select(folder, nil).Wait(); err != nil {
		return nil, 0, fmt.Errorf("select folder %s: %w", folder, err)
	}
	var uidSet imap.UIDSet
	uidSet.AddRange(imap.UID(minUID), 0)
	bodyItem := &imap.FetchItemBodySection{}
	msgs, err := l.c.Fetch(uidSet, &imap.FetchOptions{
		UID:         true,
		Envelope:    true,
		BodySection: []*imap.FetchItemBodySection{bodyItem},
	}).Collect()
	if err != nil {
		return nil, 0, fmt.Errorf("fetch IMAP messages: %w", err)
	}

	out := make([]ParsedIMAPMessage, 0, len(msgs))
	var maxUID uint32
	for _, buf := range msgs {
		parsed := parseFetchBuffer(buf, bodyItem)
		if parsed.UID > maxUID {
			maxUID = parsed.UID
		}
		out = append(out, parsed)
	}
	return out, maxUID, nil
}

func (l *liveIMAPClient) Close() error {
	return l.c.Close()
}

func parseFetchBuffer(buf *imapclient.FetchMessageBuffer, bodyItem *imap.FetchItemBodySection) ParsedIMAPMessage {
	res := ParsedIMAPMessage{
		UID:       uint32(buf.UID),
		MessageID: uuid.NewString(),
		Date:      time.Now().UTC(),
	}
	if buf.Envelope != nil {
		if buf.Envelope.MessageID != "" {
			res.MessageID = buf.Envelope.MessageID
		}
		if len(buf.Envelope.InReplyTo) > 0 {
			res.InReplyTo = buf.Envelope.InReplyTo[0]
		}
		res.Subject = buf.Envelope.Subject
		if !buf.Envelope.Date.IsZero() {
			res.Date = buf.Envelope.Date
		}
		if len(buf.Envelope.From) > 0 {
			res.FromEmail = buf.Envelope.From[0].Addr()
			res.FromName = buf.Envelope.From[0].Name
		}
		for _, t := range buf.Envelope.To {
			res.ToEmails = append(res.ToEmails, t.Addr())
		}
		for _, c := range buf.Envelope.Cc {
			res.CcEmails = append(res.CcEmails, c.Addr())
		}
	}
	if rawBody := buf.BodySection[bodyItem]; len(rawBody) > 0 {
		if mailMsg, err := mail.ReadMessage(bytes.NewReader(rawBody)); err == nil {
			if refs := mailMsg.Header.Get("References"); refs != "" {
				res.References = strings.Fields(refs)
			}
			var textBuf bytes.Buffer
			_, _ = textBuf.ReadFrom(mailMsg.Body)
			res.BodyText = strings.TrimSpace(textBuf.String())
		}
	}
	return res
}

func emptyIfNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
