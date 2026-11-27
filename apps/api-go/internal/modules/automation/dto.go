package automation

import (
	"time"

	"github.com/google/uuid"
)

const iso8601UTC = "2006-01-02T15:04:05.000Z"

// CreateRuleRequest matches NestJS CreateAutomationRuleDto.
type CreateRuleRequest struct {
	Name           string         `json:"name" validate:"required,min=1,max=255"`
	Description    *string        `json:"description"`
	TriggerType    string         `json:"triggerType" validate:"required"`
	TriggerConfig  map[string]any `json:"triggerConfig"`
	ActionType     string         `json:"actionType" validate:"required"`
	ActionConfig   map[string]any `json:"actionConfig"`
	OrganizationID *string        `json:"organizationId"`
	WorkspaceID    *string        `json:"workspaceId"`
	ProjectID      *string        `json:"projectId"`
	CreatedBy      *string        `json:"createdBy"`
	Status         *string        `json:"status"`
}

// UpdateRuleRequest supports partial updates to an automation rule.
type UpdateRuleRequest struct {
	Name          *string        `json:"name"`
	Description   *string        `json:"description"`
	Status        *string        `json:"status"`
	TriggerType   *string        `json:"triggerType"`
	TriggerConfig map[string]any `json:"triggerConfig"`
	ActionType    *string        `json:"actionType"`
	ActionConfig  map[string]any `json:"actionConfig"`
}

// UserBrief represents creator or trigger user summary inside automation payloads.
type UserBrief struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
}

// NamedEntityBrief represents organization or workspace summaries.
type NamedEntityBrief struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

// ProjectBrief represents project scope summary on a rule.
type ProjectBrief struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

// ExecutionResponse represents a rule_executions row.
type ExecutionResponse struct {
	ID            uuid.UUID      `json:"id"`
	RuleID        uuid.UUID      `json:"ruleId"`
	Success       bool           `json:"success"`
	ErrorMessage  *string        `json:"errorMessage"`
	ExecutionTime int32          `json:"executionTime"`
	TriggerData   map[string]any `json:"triggerData"`
	ActionResult  map[string]any `json:"actionResult"`
	TriggeredByID *uuid.UUID     `json:"triggeredById"`
	TriggeredBy   *UserBrief     `json:"triggeredBy,omitempty"`
	CreatedAt     string         `json:"createdAt"`
	UpdatedAt     string         `json:"updatedAt"`
}

// RuleResponse represents an automation rule with relations matching Prisma output.
type RuleResponse struct {
	ID             uuid.UUID           `json:"id"`
	Name           string              `json:"name"`
	Description    *string             `json:"description"`
	Status         string              `json:"status"`
	TriggerType    string              `json:"triggerType"`
	TriggerConfig  map[string]any      `json:"triggerConfig"`
	ActionType     string              `json:"actionType"`
	ActionConfig   map[string]any      `json:"actionConfig"`
	OrganizationID *uuid.UUID          `json:"organizationId"`
	WorkspaceID    *uuid.UUID          `json:"workspaceId"`
	ProjectID      *uuid.UUID          `json:"projectId"`
	ExecutionCount int32               `json:"executionCount"`
	LastExecuted   *string             `json:"lastExecuted"`
	CreatedBy      *uuid.UUID          `json:"createdBy"`
	CreatedByUser  *UserBrief          `json:"createdByUser,omitempty"`
	Organization   *NamedEntityBrief   `json:"organization,omitempty"`
	Workspace      *NamedEntityBrief   `json:"workspace,omitempty"`
	Project        *ProjectBrief       `json:"project,omitempty"`
	Executions     []ExecutionResponse `json:"executions,omitempty"`
	CreatedAt      string              `json:"createdAt"`
	UpdatedAt      string              `json:"updatedAt"`
}

// RuleStatsResponse matches GET /api/automation/rules/:id/stats.
type RuleStatsResponse struct {
	RuleID           uuid.UUID `json:"ruleId"`
	RuleName         string    `json:"ruleName"`
	TotalExecutions  int64     `json:"totalExecutions"`
	SuccessCount     int64     `json:"successCount"`
	FailureCount     int64     `json:"failureCount"`
	SuccessRate      float64   `json:"successRate"`
	AvgExecutionTime float64   `json:"avgExecutionTime"`
	LastExecuted     *string   `json:"lastExecuted"`
}

// ExecutionJobPayload is serialized onto the automation queue.
type ExecutionJobPayload struct {
	RuleID        uuid.UUID      `json:"ruleId"`
	TriggerType   string         `json:"triggerType"`
	TriggerData   map[string]any `json:"triggerData"`
	TriggeredByID *uuid.UUID     `json:"triggeredById,omitempty"`
}

func formatTime(t time.Time) string {
	return t.UTC().Format(iso8601UTC)
}
