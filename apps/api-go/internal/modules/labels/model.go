package labels

import (
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

func mapDetailsRow(r *sqlc.GetLabelDetailsByIDRow, includeOrg bool, taskLabels []TaskLabelEntry) *LabelResponse {
	var createdUser, updatedUser *UserBrief
	if r.CreatedByUserID.Valid {
		createdUser = &UserBrief{
			ID:        uuid.UUID(r.CreatedByUserID.Bytes).String(),
			Email:     r.CreatedByEmail.String,
			FirstName: pgTextPtr(r.CreatedByFirstName),
			LastName:  pgTextPtr(r.CreatedByLastName),
		}
	}
	if r.UpdatedByUserID.Valid {
		updatedUser = &UserBrief{
			ID:        uuid.UUID(r.UpdatedByUserID.Bytes).String(),
			Email:     r.UpdatedByEmail.String,
			FirstName: pgTextPtr(r.UpdatedByFirstName),
			LastName:  pgTextPtr(r.UpdatedByLastName),
		}
	}

	ws := &WorkspaceBrief{
		ID:   r.WorkspaceID.String(),
		Name: r.WorkspaceName,
		Slug: r.WorkspaceSlug,
	}
	if includeOrg {
		ws.Organization = &OrgBrief{
			ID:   r.OrganizationID.String(),
			Name: r.OrgName,
			Slug: r.OrgSlug,
		}
	}

	return &LabelResponse{
		ID:          r.ID.String(),
		Name:        r.Name,
		Color:       r.Color,
		Description: pgTextPtr(r.Description),
		ProjectID:   r.ProjectID.String(),
		CreatedByID: pgUUIDPtr(r.CreatedByID),
		UpdatedByID: pgUUIDPtr(r.UpdatedByID),
		CreatedAt:   formatTS(r.CreatedAt),
		UpdatedAt:   formatTS(r.UpdatedAt),
		Project: &ProjectBrief{
			ID:        r.ProjectID.String(),
			Name:      r.ProjectName,
			Slug:      r.ProjectSlug,
			Workspace: ws,
		},
		CreatedByUser: createdUser,
		UpdatedByUser: updatedUser,
		TaskLabels:    taskLabels,
		Count:         &LabelCount{TaskLabels: r.TaskLabelCount},
	}
}

