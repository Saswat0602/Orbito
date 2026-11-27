package inbox

import "strings"

// EmailTemplateData represents a reusable email template entry.
type EmailTemplateData struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Subject     string   `json:"subject"`
	Content     string   `json:"content"`
	Category    string   `json:"category"`
	Variables   []string `json:"variables"`
	IsDefault   bool     `json:"isDefault"`
	Description string   `json:"description,omitempty"`
}

// TemplateCategory describes a template grouping category.
type TemplateCategory struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

// TemplateVariable describes a placeholder variable available in email templates.
type TemplateVariable struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

var defaultTemplates = []EmailTemplateData{
	{
		ID:          "auto-reply-support",
		Name:        "Auto Reply - Support",
		Subject:     "Re: {{subject}}",
		Content:     "Thank you for contacting our support team.\n\nWe have received your message and created ticket #{{taskNumber}} to track your request.\n\nBest regards,\n{{projectName}} Support Team",
		Category:    "auto-reply",
		Variables:   []string{"subject", "taskNumber", "projectName"},
		IsDefault:   true,
		Description: "Standard auto-reply for support requests",
	},
	{
		ID:          "auto-reply-general",
		Name:        "Auto Reply - General Inquiry",
		Subject:     "Thank you for your message",
		Content:     "Hello,\n\nThank you for reaching out to us. Ticket Reference: #{{taskNumber}}\nProject: {{projectName}}\n\nBest regards,\n{{projectName}} Team",
		Category:    "auto-reply",
		Variables:   []string{"taskNumber", "projectName", "websiteUrl"},
		IsDefault:   true,
		Description: "General auto-reply for inquiries",
	},
	{
		ID:          "auto-reply-out-of-hours",
		Name:        "Auto Reply - Out of Hours",
		Subject:     "Out of Hours - We'll respond soon",
		Content:     "Thank you for your email.\n\nWe have received your message outside of our normal business hours ({{timezone}}).\nYour ticket: #{{taskNumber}}\n\nBest regards,\n{{projectName}} Team",
		Category:    "auto-reply",
		Variables:   []string{"taskNumber", "timezone", "emergencyPhone", "projectName"},
		IsDefault:   true,
		Description: "Auto-reply for messages received outside business hours",
	},
	{
		ID:          "welcome-new-customer",
		Name:        "Welcome - New Customer",
		Subject:     "Welcome to {{projectName}}!",
		Content:     "Welcome to {{projectName}}!\n\nTicket Reference: #{{taskNumber}}\nAccount Manager: {{assigneeName}}\n\n{{projectName}} Team",
		Category:    "welcome",
		Variables:   []string{"projectName", "supportEmail", "taskNumber", "assigneeName", "websiteUrl"},
		IsDefault:   true,
		Description: "Welcome email for new customers",
	},
	{
		ID:          "support-issue-resolved",
		Name:        "Support - Issue Resolved",
		Subject:     "Issue Resolved - {{subject}}",
		Content:     "Good news! We've resolved your support request.\n\nTicket: #{{taskNumber}}\nIssue: {{subject}}\nResolution: {{resolution}}\n\nBest regards,\n{{assigneeName}}\n{{projectName}} Support Team",
		Category:    "support",
		Variables:   []string{"subject", "taskNumber", "resolution", "projectName", "assigneeName"},
		IsDefault:   true,
		Description: "Template for resolved support tickets",
	},
	{
		ID:          "support-need-info",
		Name:        "Support - Need More Information",
		Subject:     "Re: {{subject}} - Additional Information Needed",
		Content:     "Hi {{customerName}},\n\nTicket: #{{taskNumber}}\n\n{{additionalInfoNeeded}}\n\nBest regards,\n{{assigneeName}}\n{{projectName}} Support Team",
		Category:    "support",
		Variables:   []string{"subject", "customerName", "taskNumber", "additionalInfoNeeded", "assigneeName", "projectName"},
		IsDefault:   true,
		Description: "Request additional information from customer",
	},
	{
		ID:          "support-escalated",
		Name:        "Support - Issue Escalated",
		Subject:     "Escalated: {{subject}}",
		Content:     "Your support request #{{taskNumber}} has been escalated by {{escalatedBy}}.\nReason: {{escalationReason}}\nSLA: {{escalationSla}} hours\nTrack progress: {{taskUrl}}",
		Category:    "support",
		Variables:   []string{"taskNumber", "subject", "escalatedBy", "escalationReason", "escalationSla", "taskUrl", "projectName"},
		IsDefault:   true,
		Description: "Notification for escalated support tickets",
	},
	{
		ID:          "notification-task-assigned",
		Name:        "Notification - Task Assignment",
		Subject:     "Task Assigned: {{subject}}",
		Content:     "A new task has been assigned to you.\n\nTitle: {{subject}}\nTask #: {{taskNumber}}\nPriority: {{priority}}\nProject: {{projectName}}\n{{taskUrl}}",
		Category:    "notification",
		Variables:   []string{"subject", "taskNumber", "priority", "dueDate", "projectName", "description", "taskUrl"},
		IsDefault:   true,
		Description: "Notification for task assignments",
	},
	{
		ID:          "notification-status-update",
		Name:        "Notification - Status Update",
		Subject:     "Status Update: {{subject}}",
		Content:     "Task #{{taskNumber}} status changed from {{previousStatus}} to {{currentStatus}} by {{updatedBy}}.\n{{statusComment}}\n{{taskUrl}}",
		Category:    "notification",
		Variables:   []string{"subject", "taskNumber", "previousStatus", "currentStatus", "updatedBy", "statusComment", "taskUrl", "projectName"},
		IsDefault:   true,
		Description: "Notification for status changes",
	},
	{
		ID:          "custom-follow-up",
		Name:        "Custom - Follow Up",
		Subject:     "Following up on {{subject}}",
		Content:     "Hi {{customerName}},\n\nTicket: #{{taskNumber}}\n{{followUpMessage}}\n\nBest regards,\n{{assigneeName}}",
		Category:    "custom",
		Variables:   []string{"customerName", "subject", "taskNumber", "followUpMessage", "assigneeName", "projectName"},
		IsDefault:   false,
		Description: "Follow-up email template",
	},
	{
		ID:          "custom-meeting-scheduled",
		Name:        "Custom - Meeting Scheduled",
		Subject:     "Meeting Scheduled - {{subject}}",
		Content:     "Hi {{customerName}},\n\nMeeting Date: {{meetingDate}} {{meetingTime}} ({{meetingDuration}})\nLink: {{meetingLink}}\nAgenda: {{meetingAgenda}}",
		Category:    "custom",
		Variables:   []string{"customerName", "subject", "meetingDate", "meetingTime", "meetingDuration", "meetingLink", "meetingAgenda", "assigneeName", "projectName"},
		IsDefault:   false,
		Description: "Meeting scheduling template",
	},
	{
		ID:          "custom-feedback-request",
		Name:        "Custom - Feedback Request",
		Subject:     "How was your experience? - {{subject}}",
		Content:     "Hi {{customerName}},\n\nWe recently helped you with: {{subject}} (Ticket #{{taskNumber}}).\n\nThank you for choosing {{projectName}}!",
		Category:    "custom",
		Variables:   []string{"customerName", "subject", "taskNumber", "positivePoints", "improvementAreas", "assigneeName", "projectName"},
		IsDefault:   false,
		Description: "Customer feedback request template",
	},
}

var templateCategories = []TemplateCategory{
	{Value: "auto-reply", Label: "Auto Reply", Description: "Automated responses to incoming emails"},
	{Value: "welcome", Label: "Welcome", Description: "Welcome messages for new customers"},
	{Value: "support", Label: "Support", Description: "Customer support communications"},
	{Value: "notification", Label: "Notification", Description: "Status updates and notifications"},
	{Value: "custom", Label: "Custom", Description: "Custom templates for specific needs"},
}

var commonVariables = []TemplateVariable{
	{Name: "subject", Description: "Original email subject"},
	{Name: "taskNumber", Description: "Generated task number"},
	{Name: "projectName", Description: "Name of the project"},
	{Name: "customerName", Description: "Customer name from email"},
	{Name: "assigneeName", Description: "Name of assigned team member"},
	{Name: "supportEmail", Description: "Support email address"},
	{Name: "websiteUrl", Description: "Company website URL"},
	{Name: "taskUrl", Description: "Direct link to the task"},
	{Name: "priority", Description: "Task priority level"},
	{Name: "dueDate", Description: "Task due date"},
	{Name: "description", Description: "Task description"},
	{Name: "currentStatus", Description: "Current task status"},
	{Name: "previousStatus", Description: "Previous task status"},
	{Name: "updatedBy", Description: "Who updated the status"},
	{Name: "statusComment", Description: "Comment about status change"},
	{Name: "resolution", Description: "How the issue was resolved"},
	{Name: "escalatedBy", Description: "Who escalated the ticket"},
	{Name: "escalationReason", Description: "Reason for escalation"},
	{Name: "escalationSla", Description: "SLA for escalated tickets"},
	{Name: "timezone", Description: "Business timezone"},
	{Name: "emergencyPhone", Description: "Emergency contact number"},
	{Name: "additionalInfoNeeded", Description: "List of needed information"},
	{Name: "followUpMessage", Description: "Custom follow-up message"},
	{Name: "meetingDate", Description: "Scheduled meeting date"},
	{Name: "meetingTime", Description: "Scheduled meeting time"},
	{Name: "meetingDuration", Description: "Meeting duration"},
	{Name: "meetingLink", Description: "Meeting link or location"},
	{Name: "meetingAgenda", Description: "Meeting agenda"},
	{Name: "positivePoints", Description: "What went well"},
	{Name: "improvementAreas", Description: "Areas for improvement"},
}

// GetTemplates returns default templates optionally filtered by category or search substring.
func GetTemplates(category, search string) []EmailTemplateData {
	if search != "" {
		q := strings.ToLower(search)
		out := make([]EmailTemplateData, 0)
		for _, t := range defaultTemplates {
			if strings.Contains(strings.ToLower(t.Name), q) ||
				strings.Contains(strings.ToLower(t.Description), q) ||
				strings.Contains(strings.ToLower(t.Subject), q) {
				out = append(out, t)
			}
		}
		return out
	}
	if category != "" {
		out := make([]EmailTemplateData, 0)
		for _, t := range defaultTemplates {
			if t.Category == category {
				out = append(out, t)
			}
		}
		return out
	}
	return append([]EmailTemplateData(nil), defaultTemplates...)
}

// GetTemplateCategories returns all supported template categories.
func GetTemplateCategories() []TemplateCategory {
	return append([]TemplateCategory(nil), templateCategories...)
}

// GetCommonVariables returns all available template variables.
func GetCommonVariables() []TemplateVariable {
	return append([]TemplateVariable(nil), commonVariables...)
}

// GetTemplateByID finds a single template by ID or returns nil.
func GetTemplateByID(id string) *EmailTemplateData {
	for i := range defaultTemplates {
		if defaultTemplates[i].ID == id {
			tpl := defaultTemplates[i]
			return &tpl
		}
	}
	return nil
}
