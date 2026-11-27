package mailer

import (
	"fmt"
	"html"
	"strings"
)

// RenderTemplate generates an HTML email document for the given template and payload.
func RenderTemplate(template string, data map[string]any) string {
	title := "Orbito Notification"
	content := ""

	switch template {
	case "task-assigned":
		taskTitle := getString(data, "task.title", "a task")
		assigner := getString(data, "assigner.name", "A team member")
		title = fmt.Sprintf("Task Assigned: %s", taskTitle)
		content = fmt.Sprintf("<p>Hello,</p><p><strong>%s</strong> assigned you to task <strong>%s</strong>.</p>",
			html.EscapeString(assigner), html.EscapeString(taskTitle))

	case "task-status-changed":
		taskTitle := getString(data, "task.title", "a task")
		oldStatus := getString(data, "oldStatus", "")
		newStatus := getString(data, "newStatus", "")
		title = fmt.Sprintf("Status Changed: %s", taskTitle)
		content = fmt.Sprintf("<p>The status of task <strong>%s</strong> changed from <em>%s</em> to <strong>%s</strong>.</p>",
			html.EscapeString(taskTitle), html.EscapeString(oldStatus), html.EscapeString(newStatus))

	case "password-reset":
		resetLink := getString(data, "resetLink", "#")
		title = "Password Reset Request"
		content = fmt.Sprintf("<p>You requested a password reset. Click below to reset your password:</p><p><a href=\"%s\" style=\"display:inline-block;padding:10px 20px;background:#18181b;color:#fff;text-decoration:none;border-radius:4px;\">Reset Password</a></p><p>This link expires in 24 hours. If you did not request this, please ignore this email.</p>",
			html.EscapeString(resetLink))

	case "password-reset-confirmation":
		title = "Password Successfully Reset"
		content = "<p>Your Orbito account password has been successfully reset. If you did not make this change, please contact your administrator immediately.</p>"

	case "send-invitation", "workspace-invited":
		workspaceName := getString(data, "workspace.name", "an Orbito workspace")
		inviteLink := getString(data, "inviteLink", "#")
		title = fmt.Sprintf("You've been invited to %s", workspaceName)
		content = fmt.Sprintf("<p>You have been invited to join <strong>%s</strong> on Orbito.</p><p><a href=\"%s\" style=\"display:inline-block;padding:10px 20px;background:#18181b;color:#fff;text-decoration:none;border-radius:4px;\">Accept Invitation</a></p>",
			html.EscapeString(workspaceName), html.EscapeString(inviteLink))

	case "mention":
		mentioner := getString(data, "mentioner.name", "A teammate")
		taskTitle := getString(data, "entity.title", "a task")
		textSnippet := getString(data, "textContent", "")
		entityURL := getString(data, "entityUrl", "#")
		title = fmt.Sprintf("You were mentioned in task: %s", taskTitle)
		content = fmt.Sprintf("<p><strong>%s</strong> mentioned you in a comment on <strong>%s</strong>.</p>",
			html.EscapeString(mentioner), html.EscapeString(taskTitle))
		if textSnippet != "" {
			content += fmt.Sprintf("<blockquote style=\"border-left:2px solid #3f3f46;padding-left:12px;color:#d4d4d8;\">%s</blockquote>",
				html.EscapeString(textSnippet))
		}
		content += fmt.Sprintf("<p><a href=\"%s\" style=\"display:inline-block;padding:10px 20px;background:#18181b;color:#fff;text-decoration:none;border-radius:4px;\">View Task</a></p>",
			html.EscapeString(entityURL))

	default:
		msg := getString(data, "message", "You have a new notification from Orbito.")
		content = fmt.Sprintf("<p>%s</p>", html.EscapeString(msg))
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>%s</title>
</head>
<body style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background-color: #09090b; color: #f4f4f5; padding: 24px; margin: 0;">
  <div style="max-width: 600px; margin: 0 auto; background-color: #18181b; border: 1px solid #27272a; border-radius: 6px; padding: 24px;">
    <h2 style="margin-top: 0; color: #ffffff;">%s</h2>
    <div style="font-size: 14px; line-height: 1.6; color: #a1a1aa;">
      %s
    </div>
    <hr style="border: none; border-top: 1px solid #27272a; margin: 24px 0;" />
    <p style="font-size: 12px; color: #71717a; margin-bottom: 0;">
      Orbito Enterprise Work Management • Automated Notification
    </p>
  </div>
</body>
</html>`, html.EscapeString(title), html.EscapeString(title), content)
}

func getString(data map[string]any, path string, fallback string) string {
	if data == nil {
		return fallback
	}
	parts := strings.Split(path, ".")
	var cur any = data
	for _, p := range parts {
		if m, ok := cur.(map[string]any); ok {
			cur = m[p]
		} else {
			return fallback
		}
	}
	if s, ok := cur.(string); ok && s != "" {
		return s
	}
	return fallback
}
