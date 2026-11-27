package ai_chat

import "strings"

const appGuideText = `## Orbito Navigation & Workflows
- /settings -> Organization list; /settings/{org} -> Org detail (Settings, Workflows, Members)
- /{ws} -> Workspace; /{ws}/settings; /{ws}/members; /{ws}/tasks; /{ws}/tasks/new
- /{ws}/{proj} -> Project; /{ws}/{proj}/settings; /{ws}/{proj}/members; /{ws}/{proj}/tasks
- /{ws}/{proj}/tasks?view=gantt -> Project Gantt timeline view
- /admin, /admin/users, /admin/organizations, /admin/config -> System administration (SUPER_ADMIN only)

## Key Disambiguations
- Create Project vs Create Task: Create Project on /{ws} creates a project; tasks must be created on /tasks, /{ws}/tasks, /{ws}/{proj}/tasks, or /{ws}/tasks/new.
- Invite Members: navigate to /{ws}/members, /{ws}/{proj}/members, or /settings/{org} Members tab.
- Settings Buttons: use sidebar-org-settings, sidebar-workspace-settings, sidebar-project-settings, or dashboard-settings-button.`

// EnhancePromptWithContext appends Orbito navigation hints relevant to a user task prompt.
func EnhancePromptWithContext(task, currentURL string) string {
	return "Current URL: " + currentURL + "\nTask: " + task + "\n\n" + appGuideText
}

// GetAutomationPrompt returns step-by-step guidance if the user prompt targets task/workspace UI automation.
func GetAutomationPrompt(task string) string {
	lower := strings.ToLower(task)
	switch {
	case strings.Contains(lower, "create task") || strings.Contains(lower, "new task"):
		return "TASK CREATION RULE: Ensure both workspace and project are explicitly confirmed before submitting on /{ws}/tasks/new or /{ws}/{proj}/tasks."
	case strings.Contains(lower, "sprint"):
		return "SPRINT RULE: Manage sprints from the project Sprints tab (Create, Start, Complete)."
	default:
		return ""
	}
}

// BuildSystemPrompt combines personality tone, custom instructions, and Orbito guidance.
func BuildSystemPrompt(tone, customPrompt string) string {
	if strings.TrimSpace(tone) == "" {
		tone = "Professional"
	}
	base := "You are the Orbito AI Assistant. Respond concisely in a " + tone + " tone.\n\n" + appGuideText
	if strings.TrimSpace(customPrompt) != "" {
		base += "\n\nAdditional Instructions:\n" + strings.TrimSpace(customPrompt)
	}
	return base
}
