package email

// SendEmailDTO represents the payload for enqueuing a single email.
type SendEmailDTO struct {
	To       string         `json:"to" validate:"required,email"`
	Subject  string         `json:"subject" validate:"required"`
	Template string         `json:"template" validate:"required"`
	Data     map[string]any `json:"data" validate:"required"`
	Priority string         `json:"priority,omitempty"`
	Delay    int            `json:"delay,omitempty"`
}

// BulkEmailDTO represents the payload for enqueuing multiple emails.
type BulkEmailDTO struct {
	Recipients []string       `json:"recipients" validate:"required,min=1,dive,email"`
	Subject    string         `json:"subject" validate:"required"`
	Template   string         `json:"template" validate:"required"`
	Data       map[string]any `json:"data" validate:"required"`
	Priority   string         `json:"priority,omitempty"`
}

// QueueStatsResponse matches the JSON shape returned by GET /api/email/queue/stats.
type QueueStatsResponse struct {
	Waiting   int `json:"waiting"`
	Active    int `json:"active"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
}
