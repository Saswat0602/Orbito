package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// Config holds SMTP connection and credential parameters.
type Config struct {
	Host     string
	Port     int
	User     string
	Pass     string
	From     string
	Insecure bool
}

// Message represents an email message to be sent.
type Message struct {
	To      []string
	Subject string
	HTML    string
	Text    string
}

// Mailer defines the interface for verifying and sending emails.
type Mailer interface {
	Send(ctx context.Context, msg Message) error
	Verify(ctx context.Context) error
}

type smtpMailer struct {
	cfg Config
}

// New constructs a new Mailer backed by SMTP.
func New(cfg Config) Mailer {
	if cfg.From == "" {
		cfg.From = "noreply@orbito.io"
	}
	if cfg.Port == 0 {
		cfg.Port = 587
	}
	return &smtpMailer{cfg: cfg}
}

func (m *smtpMailer) isSimulated() bool {
	return m.cfg.Host == "" || m.cfg.Host == "smtp.example.com" || m.cfg.Pass == ""
}

// Verify checks SMTP host reachability and credentials handshake.
func (m *smtpMailer) Verify(ctx context.Context) error {
	if m.isSimulated() {
		return nil
	}

	client, err := m.dial(ctx)
	if err != nil {
		return fmt.Errorf("failed to dial SMTP server: %w", err)
	}
	defer func() {
		_ = client.Close()
	}()

	err = m.auth(client)
	if err != nil {
		return fmt.Errorf("failed SMTP authentication: %w", err)
	}

	return client.Quit()
}

// Send sends an email message via SMTP.
func (m *smtpMailer) Send(ctx context.Context, msg Message) error {
	if m.isSimulated() {
		return nil
	}

	if len(msg.To) == 0 {
		return fmt.Errorf("no recipients provided")
	}

	client, err := m.dial(ctx)
	if err != nil {
		return fmt.Errorf("failed to dial SMTP server: %w", err)
	}
	defer func() {
		_ = client.Close()
	}()

	err = m.auth(client)
	if err != nil {
		return fmt.Errorf("failed SMTP authentication: %w", err)
	}

	err = client.Mail(m.cfg.From)
	if err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	for _, recipient := range msg.To {
		err = client.Rcpt(recipient)
		if err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to initiate data transfer: %w", err)
	}

	rawMsg := m.buildMimeMessage(msg)
	if _, writeErr := w.Write([]byte(rawMsg)); writeErr != nil {
		_ = w.Close()
		return fmt.Errorf("failed to write message body: %w", writeErr)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}

func (m *smtpMailer) dial(ctx context.Context) (*smtp.Client, error) {
	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)
	dialer := &net.Dialer{Timeout: 10 * time.Second}

	tlsConfig := &tls.Config{
		ServerName:         m.cfg.Host,
		InsecureSkipVerify: m.cfg.Insecure, //nolint:gosec
	}

	if m.cfg.Port == 465 {
		conn, err := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
		if err != nil {
			return nil, err
		}
		return smtp.NewClient(conn, m.cfg.Host)
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}

	client, err := smtp.NewClient(conn, m.cfg.Host)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(tlsConfig); err != nil {
			_ = client.Close()
			return nil, fmt.Errorf("failed to start TLS: %w", err)
		}
	}

	return client, nil
}

func (m *smtpMailer) auth(client *smtp.Client) error {
	if m.cfg.User != "" && m.cfg.Pass != "" {
		if ok, _ := client.Extension("AUTH"); ok {
			auth := smtp.PlainAuth("", m.cfg.User, m.cfg.Pass, m.cfg.Host)
			return client.Auth(auth)
		}
	}
	return nil
}

func (m *smtpMailer) buildMimeMessage(msg Message) string {
	toHeader := strings.Join(msg.To, ", ")
	body := msg.HTML
	contentType := "text/html; charset=UTF-8"
	if body == "" {
		body = msg.Text
		contentType = "text/plain; charset=UTF-8"
	}

	return fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: %s\r\n\r\n%s",
		m.cfg.From,
		toHeader,
		msg.Subject,
		contentType,
		body,
	)
}
