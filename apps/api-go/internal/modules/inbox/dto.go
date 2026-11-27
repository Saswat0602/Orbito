package inbox

import (
	"time"

	"github.com/google/uuid"
)

const iso8601UTC = "2006-01-02T15:04:05.000Z"

// CreateInboxRequest represents POST /api/projects/:projectId/inbox.
type CreateInboxRequest struct {
	Name              string  `json:"name" validate:"required,min=1,max=255"`
	Description       *string `json:"description"`
	EmailAddress      *string `json:"emailAddress" validate:"omitempty,email"`
	EmailSignature    *string `json:"emailSignature"`
	AutoReplyEnabled  *bool   `json:"autoReplyEnabled"`
	AutoReplyTemplate *string `json:"autoReplyTemplate"`
	SyncInterval      *string `json:"syncInterval"`
	AutoCreateTask    *bool   `json:"autoCreateTask"`
	DefaultTaskType   *string `json:"defaultTaskType"`
	DefaultPriority   *string `json:"defaultPriority"`
	DefaultStatusID   string  `json:"defaultStatusId" validate:"required,uuid"`
	DefaultAssigneeID *string `json:"defaultAssigneeId" validate:"omitempty,uuid"`
}

// UpdateInboxRequest represents PUT /api/projects/:projectId/inbox.
type UpdateInboxRequest struct {
	Name              *string `json:"name"`
	Description       *string `json:"description"`
	EmailAddress      *string `json:"emailAddress" validate:"omitempty,email"`
	EmailSignature    *string `json:"emailSignature"`
	AutoReplyEnabled  *bool   `json:"autoReplyEnabled"`
	AutoReplyTemplate *string `json:"autoReplyTemplate"`
	AutoCreateTask    *bool   `json:"autoCreateTask"`
	DefaultTaskType   *string `json:"defaultTaskType"`
	DefaultPriority   *string `json:"defaultPriority"`
	DefaultStatusID   *string `json:"defaultStatusId" validate:"omitempty,uuid"`
	DefaultAssigneeID *string `json:"defaultAssigneeId" validate:"omitempty,uuid"`
}

// SetupEmailRequest represents PUT /api/projects/:projectId/inbox/email-account.
type SetupEmailRequest struct {
	EmailAddress        string  `json:"emailAddress" validate:"required,email"`
	DisplayName         *string `json:"displayName"`
	IMAPHost            string  `json:"imapHost" validate:"required"`
	IMAPPort            *int32  `json:"imapPort"`
	IMAPUsername        string  `json:"imapUsername" validate:"required"`
	IMAPPassword        string  `json:"imapPassword" validate:"required"`
	IMAPUseSSL          *bool   `json:"imapUseSsl"`
	IMAPTlsRejectUnauth *bool   `json:"imapTlsRejectUnauth"`
	IMAPTlsMinVersion   *string `json:"imapTlsMinVersion"`
	IMAPServername      *string `json:"imapServername"`
	IMAPFolder          *string `json:"imapFolder"`
	SMTPHost            string  `json:"smtpHost" validate:"required"`
	SMTPPort            *int32  `json:"smtpPort"`
	SMTPUsername        string  `json:"smtpUsername" validate:"required"`
	SMTPPassword        string  `json:"smtpPassword" validate:"required"`
	SMTPTlsRejectUnauth *bool   `json:"smtpTlsRejectUnauth"`
	SMTPTlsMinVersion   *string `json:"smtpTlsMinVersion"`
	SMTPServername      *string `json:"smtpServername"`
	SMTPRequireTLS      *bool   `json:"smtpRequireTls"`
}

// MessageFilters represents GET /api/projects/:projectId/inbox/messages query params.
type MessageFilters struct {
	Status        string
	IncludeSpam   bool
	FromEmail     string
	FromDate      string
	ToDate        string
	Search        string
	ConvertedOnly bool
}

// CreateRuleRequest represents POST/PUT /api/projects/:projectId/inbox/rules.
type CreateRuleRequest struct {
	Name        string         `json:"name" validate:"required"`
	Description *string        `json:"description"`
	Priority    *int32         `json:"priority"`
	Enabled     *bool          `json:"enabled"`
	Conditions  map[string]any `json:"conditions" validate:"required"`
	Actions     map[string]any `json:"actions" validate:"required"`
	StopOnMatch *bool          `json:"stopOnMatch"`
}

// EmailAccountResponse strips encrypted credentials when returning account metadata.
type EmailAccountResponse struct {
	ID             uuid.UUID `json:"id"`
	EmailAddress   string    `json:"emailAddress"`
	DisplayName    *string   `json:"displayName"`
	SyncEnabled    bool      `json:"syncEnabled"`
	LastSyncAt     *string   `json:"lastSyncAt"`
	LastSyncError  *string   `json:"lastSyncError"`
	IMAPHost       *string   `json:"imapHost"`
	IMAPPort       *int32    `json:"imapPort"`
	IMAPUsername   *string   `json:"imapUsername"`
	IMAPUseSSL     bool      `json:"imapUseSsl"`
	IMAPServername *string   `json:"imapServername"`
	IMAPFolder     string    `json:"imapFolder"`
	SMTPHost       *string   `json:"smtpHost"`
	SMTPPort       *int32    `json:"smtpPort"`
	SMTPUsername   *string   `json:"smtpUsername"`
	SMTPServername *string   `json:"smtpServername"`
	SMTPRequireTLS bool      `json:"smtpRequireTls"`
}

// ConfigResponse represents GET / POST / PUT project inbox response.
type ConfigResponse struct {
	ID                uuid.UUID             `json:"id"`
	ProjectID         uuid.UUID             `json:"projectId"`
	Enabled           bool                  `json:"enabled"`
	Name              string                `json:"name"`
	Description       *string               `json:"description"`
	EmailAddress      *string               `json:"emailAddress"`
	EmailSignature    *string               `json:"emailSignature"`
	AutoReplyEnabled  bool                  `json:"autoReplyEnabled"`
	AutoReplyTemplate *string               `json:"autoReplyTemplate"`
	AutoCreateTask    bool                  `json:"autoCreateTask"`
	DefaultTaskType   string                `json:"defaultTaskType"`
	DefaultPriority   string                `json:"defaultPriority"`
	DefaultStatusID   uuid.UUID             `json:"defaultStatusId"`
	DefaultAssigneeID *uuid.UUID            `json:"defaultAssigneeId"`
	SyncInterval      *int32                `json:"syncInterval"`
	Project           map[string]any        `json:"project,omitempty"`
	DefaultStatus     map[string]any        `json:"defaultStatus,omitempty"`
	DefaultAssignee   map[string]any        `json:"defaultAssignee,omitempty"`
	EmailAccount      *EmailAccountResponse `json:"emailAccount,omitempty"`
	Count             map[string]int64      `json:"_count,omitempty"`
	CreatedAt         string                `json:"createdAt"`
	UpdatedAt         string                `json:"updatedAt"`
}

// MessageResponse represents an inbox email item with optional task link and attachments.
type MessageResponse struct {
	ID              uuid.UUID        `json:"id"`
	ProjectInboxID  uuid.UUID        `json:"projectInboxId"`
	MessageID       string           `json:"messageId"`
	ThreadID        string           `json:"threadId"`
	InReplyTo       *string          `json:"inReplyTo"`
	References      []string         `json:"references"`
	Subject         string           `json:"subject"`
	FromEmail       string           `json:"fromEmail"`
	FromName        *string          `json:"fromName"`
	ToEmails        []string         `json:"toEmails"`
	CcEmails        []string         `json:"ccEmails"`
	BodyText        *string          `json:"bodyText"`
	BodyHTML        *string          `json:"bodyHtml"`
	Snippet         *string          `json:"snippet"`
	HasAttachments  bool             `json:"hasAttachments"`
	Converted       bool             `json:"converted"`
	Status          string           `json:"status"`
	IsSpam          bool             `json:"isSpam"`
	EmailDate       string           `json:"emailDate"`
	ConvertedAt     *string          `json:"convertedAt"`
	Attachments     []map[string]any `json:"attachments,omitempty"`
	Task            map[string]any   `json:"task,omitempty"`
	ConvertedByUser map[string]any   `json:"convertedByUser,omitempty"`
	CreatedAt       string           `json:"createdAt"`
	UpdatedAt       string           `json:"updatedAt"`
}

// RuleResponse represents an inbox rule.
type RuleResponse struct {
	ID             uuid.UUID      `json:"id"`
	ProjectInboxID uuid.UUID      `json:"projectInboxId"`
	Name           string         `json:"name"`
	Description    *string        `json:"description"`
	Priority       int32          `json:"priority"`
	Enabled        bool           `json:"enabled"`
	Conditions     map[string]any `json:"conditions"`
	Actions        map[string]any `json:"actions"`
	StopOnMatch    bool           `json:"stopOnMatch"`
	CreatedAt      string         `json:"createdAt"`
	UpdatedAt      string         `json:"updatedAt"`
}

// EmailSyncJobPayload is queued on QueueEmailSync when triggering manual sync.
type EmailSyncJobPayload struct {
	ProjectID uuid.UUID  `json:"projectId"`
	UserID    *uuid.UUID `json:"userId,omitempty"`
	Type      string     `json:"type"`
}

func formatTime(t time.Time) string {
	return t.UTC().Format(iso8601UTC)
}
