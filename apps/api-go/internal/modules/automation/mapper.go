package automation

import (
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

func mapListRow(row *sqlc.ListAutomationRulesFilteredRow) RuleResponse {
	resp := RuleResponse{
		ID:             row.ID,
		Name:           row.Name,
		Status:         string(row.Status),
		TriggerType:    string(row.TriggerType),
		TriggerConfig:  decodeJSONMap(row.TriggerConfig),
		ActionType:     string(row.ActionType),
		ActionConfig:   decodeJSONMap(row.ActionConfig),
		ExecutionCount: row.ExecutionCount,
		CreatedAt:      formatTime(row.CreatedAt.Time),
		UpdatedAt:      formatTime(row.UpdatedAt.Time),
	}
	if row.Description.Valid {
		v := row.Description.String
		resp.Description = &v
	}
	if row.OrganizationID.Valid {
		id := uuid.UUID(row.OrganizationID.Bytes)
		resp.OrganizationID = &id
		if row.OrgName.Valid {
			resp.Organization = &NamedEntityBrief{ID: id, Name: row.OrgName.String}
		}
	}
	if row.WorkspaceID.Valid {
		id := uuid.UUID(row.WorkspaceID.Bytes)
		resp.WorkspaceID = &id
		if row.WorkspaceName.Valid {
			resp.Workspace = &NamedEntityBrief{ID: id, Name: row.WorkspaceName.String}
		}
	}
	if row.ProjectID.Valid {
		id := uuid.UUID(row.ProjectID.Bytes)
		resp.ProjectID = &id
		if row.ProjectName.Valid {
			resp.Project = &ProjectBrief{ID: id, Name: row.ProjectName.String, Slug: row.ProjectSlug.String}
		}
	}
	if row.LastExecuted.Valid {
		s := formatTime(row.LastExecuted.Time)
		resp.LastExecuted = &s
	}
	if row.CreatedByID.Valid {
		id := uuid.UUID(row.CreatedByID.Bytes)
		resp.CreatedBy = &id
	}
	if row.CreatorID.Valid {
		resp.CreatedByUser = &UserBrief{
			ID:        uuid.UUID(row.CreatorID.Bytes),
			FirstName: row.CreatorFirstName.String,
			LastName:  row.CreatorLastName.String,
			Email:     row.CreatorEmail.String,
		}
	}
	return resp
}

func mapDetailRow(row *sqlc.GetAutomationRuleDetailByIDRow, execRows []sqlc.ListRuleExecutionsByRuleIDRow) RuleResponse {
	resp := RuleResponse{
		ID:             row.ID,
		Name:           row.Name,
		Status:         string(row.Status),
		TriggerType:    string(row.TriggerType),
		TriggerConfig:  decodeJSONMap(row.TriggerConfig),
		ActionType:     string(row.ActionType),
		ActionConfig:   decodeJSONMap(row.ActionConfig),
		ExecutionCount: row.ExecutionCount,
		CreatedAt:      formatTime(row.CreatedAt.Time),
		UpdatedAt:      formatTime(row.UpdatedAt.Time),
		Executions:     make([]ExecutionResponse, 0, len(execRows)),
	}
	if row.Description.Valid {
		v := row.Description.String
		resp.Description = &v
	}
	if row.OrganizationID.Valid {
		id := uuid.UUID(row.OrganizationID.Bytes)
		resp.OrganizationID = &id
		if row.OrgName.Valid {
			resp.Organization = &NamedEntityBrief{ID: id, Name: row.OrgName.String}
		}
	}
	if row.WorkspaceID.Valid {
		id := uuid.UUID(row.WorkspaceID.Bytes)
		resp.WorkspaceID = &id
		if row.WorkspaceName.Valid {
			resp.Workspace = &NamedEntityBrief{ID: id, Name: row.WorkspaceName.String}
		}
	}
	if row.ProjectID.Valid {
		id := uuid.UUID(row.ProjectID.Bytes)
		resp.ProjectID = &id
		if row.ProjectName.Valid {
			resp.Project = &ProjectBrief{ID: id, Name: row.ProjectName.String, Slug: row.ProjectSlug.String}
		}
	}
	if row.LastExecuted.Valid {
		s := formatTime(row.LastExecuted.Time)
		resp.LastExecuted = &s
	}
	if row.CreatedByID.Valid {
		id := uuid.UUID(row.CreatedByID.Bytes)
		resp.CreatedBy = &id
	}
	if row.CreatorID.Valid {
		resp.CreatedByUser = &UserBrief{
			ID:        uuid.UUID(row.CreatorID.Bytes),
			FirstName: row.CreatorFirstName.String,
			LastName:  row.CreatorLastName.String,
			Email:     row.CreatorEmail.String,
		}
	}
	for i := range execRows {
		resp.Executions = append(resp.Executions, mapExecutionRow(&execRows[i]))
	}
	return resp
}

func mapExecutionRow(row *sqlc.ListRuleExecutionsByRuleIDRow) ExecutionResponse {
	exec := ExecutionResponse{
		ID:            row.ID,
		RuleID:        row.RuleID,
		Success:       row.Success,
		ExecutionTime: row.ExecutionTime,
		TriggerData:   decodeJSONMap(row.TriggerData),
		ActionResult:  decodeJSONMap(row.ActionResult),
		CreatedAt:     formatTime(row.CreatedAt.Time),
		UpdatedAt:     formatTime(row.UpdatedAt.Time),
	}
	if row.ErrorMessage.Valid {
		v := row.ErrorMessage.String
		exec.ErrorMessage = &v
	}
	if row.TriggeredByID.Valid {
		id := uuid.UUID(row.TriggeredByID.Bytes)
		exec.TriggeredByID = &id
	}
	if row.TrigUserID.Valid {
		exec.TriggeredBy = &UserBrief{
			ID:        uuid.UUID(row.TrigUserID.Bytes),
			FirstName: row.TrigFirstName.String,
			LastName:  row.TrigLastName.String,
			Email:     row.TrigEmail.String,
		}
	}
	return exec
}
