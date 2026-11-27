package inbox

import (
	"context"
	"crypto/tls"
	"fmt"
	"html"
	"net"
	"net/smtp"
	"strings"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/sanitize"
)

func (s *service) SendCommentAsEmail(ctx context.Context, commentID uuid.UUID) (map[string]any, error) {
	rc, err := s.repo.GetCommentReplyContext(ctx, commentID)
	if err != nil {
		return nil, fmt.Errorf("load comment reply context: %w", err)
	}
	if rc == nil {
		return nil, apperror.NotFound("Comment not found")
	}
	if !rc.AllowEmailReplies {
		return nil, apperror.BadRequest("Email replies are not enabled for this task")
	}
	if rc.SentAsEmail {
		return nil, apperror.BadRequest("Comment has already been sent as email")
	}
	if !rc.AccountID.Valid || !rc.SmtpHost.Valid {
		return nil, apperror.BadRequest("No email account configured for this inbox")
	}
	if !rc.OrigFromEmail.Valid || rc.OrigFromEmail.String == "" {
		return nil, apperror.BadRequest("Task has no originating email recipient to reply to")
	}

	smtpPass := rc.SmtpPassword.String
	if s.crypto != nil && smtpPass != "" {
		smtpPass = s.crypto.Decrypt(smtpPass)
	}

	msgID := fmt.Sprintf("<%s@orbito>", uuid.NewString())
	recipients := []string{rc.OrigFromEmail.String}
	subject := "Re: " + rc.OrigSubject.String
	safeBody := sanitize.HTML(rc.Content)
	if rc.EmailSignature.Valid && rc.EmailSignature.String != "" {
		safeBody += "<br><br>" + html.EscapeString(rc.EmailSignature.String)
	}

	mime := buildReplyMIME(rc, msgID, recipients, subject, safeBody)
	port := int32(587)
	if rc.SmtpPort.Valid && rc.SmtpPort.Int32 > 0 {
		port = rc.SmtpPort.Int32
	}
	if sendErr := sendSMTPMessage(
		rc.SmtpHost.String, port, rc.SmtpUsername.String, smtpPass,
		rc.SmtpServername.String, rc.SmtpRequireTls.Bool,
		rc.AccountEmail.String, recipients, mime,
	); sendErr != nil {
		return nil, apperror.BadRequest("Failed to send email: " + sendErr.Error())
	}

	_ = s.repo.MarkCommentSentAsEmail(ctx, commentID, msgID, recipients)
	return map[string]any{
		"message":    "Comment sent as email successfully",
		"success":    true,
		"messageId":  msgID,
		"recipients": recipients,
	}, nil
}

func (s *service) TestEmailConfiguration(ctx context.Context, accountID uuid.UUID) (map[string]any, error) {
	acct, err := s.repo.GetEmailAccountByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("load email account: %w", err)
	}
	if acct == nil {
		return nil, apperror.NotFound("Email account not found")
	}

	imapPass := acct.ImapPassword.String
	if s.crypto != nil && imapPass != "" {
		imapPass = s.crypto.Decrypt(imapPass)
	}
	imapPort := 993
	if acct.ImapPort.Valid && acct.ImapPort.Int32 > 0 {
		imapPort = int(acct.ImapPort.Int32)
	}

	imapOK := false
	var imapErrStr string
	client, iErr := s.imapDial(acct.ImapHost.String, imapPort, acct.ImapUsername.String, imapPass, acct.ImapServername.String, acct.ImapUseSsl, acct.ImapTlsRejectUnauthorized)
	if iErr != nil {
		imapErrStr = iErr.Error()
	} else {
		imapOK = true
		_ = client.Close()
	}

	smtpOK := acct.SmtpHost.Valid && acct.SmtpHost.String != ""
	var smtpErrStr string
	if !smtpOK {
		smtpErrStr = "SMTP host is not configured"
	}
	return map[string]any{
		"imap": map[string]any{"success": imapOK, "error": imapErrStr},
		"smtp": map[string]any{"success": smtpOK, "error": smtpErrStr},
	}, nil
}

func buildReplyMIME(rc *sqlc.GetCommentEmailReplyContextRow, msgID string, recipients []string, subject, bodyHTML string) []byte {
	from := rc.AccountEmail.String
	if rc.AccountDisplayName.Valid && rc.AccountDisplayName.String != "" {
		from = fmt.Sprintf(`"%s" <%s>`, rc.AccountDisplayName.String, from)
	}
	headers := []string{
		"From: " + from,
		"To: " + strings.Join(recipients, ", "),
		"Subject: " + subject,
		"Message-ID: " + msgID,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=utf-8",
	}
	if rc.OrigMessageID.Valid && rc.OrigMessageID.String != "" {
		headers = append(headers, "In-Reply-To: <"+strings.Trim(rc.OrigMessageID.String, "<>")+">")
	}
	return []byte(strings.Join(headers, "\r\n") + "\r\n\r\n" + bodyHTML)
}

func sendSMTPMessage(host string, port int32, user, pass, servername string, requireTLS bool, from string, to []string, msg []byte) error {
	addr := net.JoinHostPort(host, fmt.Sprintf("%d", port))
	sni := servername
	if sni == "" {
		sni = host
	}
	tlsCfg := &tls.Config{ServerName: sni, MinVersion: tls.VersionTLS12}

	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer func() { _ = c.Close() }()

	if ok, _ := c.Extension("STARTTLS"); ok || requireTLS {
		if tErr := c.StartTLS(tlsCfg); tErr != nil {
			return tErr
		}
	}
	if user != "" && pass != "" {
		if aErr := c.Auth(smtp.PlainAuth("", user, pass, host)); aErr != nil {
			return aErr
		}
	}
	if mErr := c.Mail(from); mErr != nil {
		return mErr
	}
	for _, rcpt := range to {
		if rErr := c.Rcpt(rcpt); rErr != nil {
			return rErr
		}
	}
	wc, err := c.Data()
	if err != nil {
		return err
	}
	if _, wErr := wc.Write(msg); wErr != nil {
		return wErr
	}
	if cErr := wc.Close(); cErr != nil {
		return cErr
	}
	return c.Quit()
}
