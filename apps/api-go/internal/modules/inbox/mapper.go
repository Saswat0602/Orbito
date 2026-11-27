package inbox

import (
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

func mapInboxRow(row *sqlc.GetProjectInboxByProjectIDRow, acct *sqlc.EmailAccount) ConfigResponse {
	resp := ConfigResponse{
		ID:                row.ID,
		ProjectID:         row.ProjectID,
		Enabled:           row.Enabled,
		Name:              row.Name,
		Description:       pgTextPtr(row.Description),
		EmailAddress:      pgTextPtr(row.EmailAddress),
		EmailSignature:    pgTextPtr(row.EmailSignature),
		AutoReplyEnabled:  row.AutoReplyEnabled,
		AutoReplyTemplate: pgTextPtr(row.AutoReplyTemplate),
		AutoCreateTask:    row.AutoCreateTask,
		DefaultTaskType:   string(row.DefaultTaskType),
		DefaultPriority:   string(row.DefaultPriority),
		DefaultStatusID:   row.DefaultStatusID,
		CreatedAt:         formatTime(row.CreatedAt.Time),
		UpdatedAt:         formatTime(row.UpdatedAt.Time),
		Project:           map[string]any{"id": row.ProjectID, "slug": row.ProjectSlug, "name": row.ProjectName},
		DefaultStatus:     map[string]any{"id": row.DefaultStatusID, "name": row.DefaultStatusName.String, "color": row.DefaultStatusColor.String},
		Count:             map[string]int64{"messages": row.MessagesCount, "rules": row.RulesCount},
	}
	if row.DefaultAssigneeID.Valid {
		id := uuid.UUID(row.DefaultAssigneeID.Bytes)
		resp.DefaultAssigneeID = &id
	}
	if row.SyncInterval.Valid {
		v := row.SyncInterval.Int32
		resp.SyncInterval = &v
	}
	if row.AssigneeID.Valid {
		resp.DefaultAssignee = map[string]any{
			"id":        uuid.UUID(row.AssigneeID.Bytes),
			"firstName": row.AssigneeFirstName.String,
			"lastName":  row.AssigneeLastName.String,
			"email":     row.AssigneeEmail.String,
		}
	}
	if acct != nil {
		mapped := mapEmailAccount(acct)
		resp.EmailAccount = &mapped
	}
	return resp
}

func mapEmailAccount(a *sqlc.EmailAccount) EmailAccountResponse {
	res := EmailAccountResponse{
		ID:             a.ID,
		EmailAddress:   a.EmailAddress,
		DisplayName:    pgTextPtr(a.DisplayName),
		SyncEnabled:    a.SyncEnabled,
		LastSyncError:  pgTextPtr(a.LastSyncError),
		IMAPHost:       pgTextPtr(a.ImapHost),
		IMAPUsername:   pgTextPtr(a.ImapUsername),
		IMAPUseSSL:     a.ImapUseSsl,
		IMAPServername: pgTextPtr(a.ImapServername),
		IMAPFolder:     a.ImapFolder,
		SMTPHost:       pgTextPtr(a.SmtpHost),
		SMTPUsername:   pgTextPtr(a.SmtpUsername),
		SMTPServername: pgTextPtr(a.SmtpServername),
		SMTPRequireTLS: a.SmtpRequireTls,
	}
	if a.LastSyncAt.Valid {
		s := formatTime(a.LastSyncAt.Time)
		res.LastSyncAt = &s
	}
	if a.ImapPort.Valid {
		v := a.ImapPort.Int32
		res.IMAPPort = &v
	}
	if a.SmtpPort.Valid {
		v := a.SmtpPort.Int32
		res.SMTPPort = &v
	}
	return res
}

func mapMessageListRow(r *sqlc.ListInboxMessagesFilteredRow, atts []map[string]any) MessageResponse {
	m := MessageResponse{
		ID:             r.ID,
		ProjectInboxID: r.ProjectInboxID,
		MessageID:      r.MessageID,
		ThreadID:       r.ThreadID,
		InReplyTo:      pgTextPtr(r.InReplyTo),
		References:     emptyIfNil(r.References),
		Subject:        r.Subject,
		FromEmail:      r.FromEmail,
		FromName:       pgTextPtr(r.FromName),
		ToEmails:       emptyIfNil(r.ToEmails),
		CcEmails:       emptyIfNil(r.CcEmails),
		BodyText:       pgTextPtr(r.BodyText),
		BodyHTML:       pgTextPtr(r.BodyHtml),
		Snippet:        pgTextPtr(r.Snippet),
		HasAttachments: r.HasAttachments,
		Converted:      r.Converted,
		Status:         string(r.Status),
		IsSpam:         r.IsSpam,
		EmailDate:      formatTime(r.EmailDate.Time),
		Attachments:    atts,
		CreatedAt:      formatTime(r.CreatedAt.Time),
		UpdatedAt:      formatTime(r.UpdatedAt.Time),
	}
	if r.ConvertedAt.Valid {
		s := formatTime(r.ConvertedAt.Time)
		m.ConvertedAt = &s
	}
	if r.TaskID.Valid {
		m.Task = map[string]any{
			"id": uuid.UUID(r.TaskID.Bytes), "title": r.TaskTitle.String, "taskNumber": r.TaskNumber.Int32,
			"status": map[string]any{"name": r.TaskStatusName.String, "color": r.TaskStatusColor.String},
		}
	}
	if r.ConvUserID.Valid {
		m.ConvertedByUser = map[string]any{
			"id": uuid.UUID(r.ConvUserID.Bytes), "firstName": r.ConvUserFirstName.String, "lastName": r.ConvUserLastName.String,
		}
	}
	return m
}

func mapMessageDetailRow(r *sqlc.GetInboxMessageDetailByIDRow, atts []map[string]any) MessageResponse {
	m := MessageResponse{
		ID:             r.ID,
		ProjectInboxID: r.ProjectInboxID,
		MessageID:      r.MessageID,
		ThreadID:       r.ThreadID,
		InReplyTo:      pgTextPtr(r.InReplyTo),
		References:     emptyIfNil(r.References),
		Subject:        r.Subject,
		FromEmail:      r.FromEmail,
		FromName:       pgTextPtr(r.FromName),
		ToEmails:       emptyIfNil(r.ToEmails),
		CcEmails:       emptyIfNil(r.CcEmails),
		BodyText:       pgTextPtr(r.BodyText),
		BodyHTML:       pgTextPtr(r.BodyHtml),
		Snippet:        pgTextPtr(r.Snippet),
		HasAttachments: r.HasAttachments,
		Converted:      r.Converted,
		Status:         string(r.Status),
		IsSpam:         r.IsSpam,
		EmailDate:      formatTime(r.EmailDate.Time),
		Attachments:    atts,
		CreatedAt:      formatTime(r.CreatedAt.Time),
		UpdatedAt:      formatTime(r.UpdatedAt.Time),
	}
	if r.ConvertedAt.Valid {
		s := formatTime(r.ConvertedAt.Time)
		m.ConvertedAt = &s
	}
	if r.TaskID.Valid {
		m.Task = map[string]any{"id": uuid.UUID(r.TaskID.Bytes), "title": r.TaskTitle.String, "slug": r.TaskSlug.String}
	}
	return m
}

func mapRuleRow(r *sqlc.InboxRule) RuleResponse {
	return RuleResponse{
		ID:             r.ID,
		ProjectInboxID: r.ProjectInboxID,
		Name:           r.Name,
		Description:    pgTextPtr(r.Description),
		Priority:       r.Priority,
		Enabled:        r.Enabled,
		Conditions:     decodeJSONMap(r.Conditions),
		Actions:        decodeJSONMap(r.Actions),
		StopOnMatch:    r.StopOnMatch,
		CreatedAt:      formatTime(r.CreatedAt.Time),
		UpdatedAt:      formatTime(r.UpdatedAt.Time),
	}
}
