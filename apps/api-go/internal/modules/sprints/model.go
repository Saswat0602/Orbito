package sprints

import (
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// EventEmitter broadcasts sprint lifecycle events to connected WebSocket clients.
// It may be nil before the real-time gateway is initialized.
type EventEmitter interface {
	EmitSprintStarted(projectID string, payload any)
	EmitSprintCompleted(projectID string, payload any)
}

func mapDetailsRow(r *sqlc.GetSprintDetailsByIDRow, tasks []SprintTaskItem) *SprintResponse {
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

	return &SprintResponse{
		ID:          r.ID.String(),
		Name:        r.Name,
		Slug:        pgTextPtr(r.Slug),
		Goal:        pgTextPtr(r.Goal),
		Status:      string(r.Status),
		IsDefault:   r.IsDefault,
		Archive:     r.Archive,
		StartDate:   pgTSPtr(r.StartDate),
		EndDate:     pgTSPtr(r.EndDate),
		ProjectID:   r.ProjectID.String(),
		CreatedByID: pgUUIDPtr(r.CreatedByID),
		UpdatedByID: pgUUIDPtr(r.UpdatedByID),
		CreatedAt:   formatTS(r.CreatedAt),
		UpdatedAt:   formatTS(r.UpdatedAt),
		Project: &ProjectBrief{
			ID:   r.ProjectID.String(),
			Name: r.ProjectName,
			Slug: r.ProjectSlug,
			Workspace: &WorkspaceBrief{
				ID:   r.WorkspaceID.String(),
				Name: r.WorkspaceName,
				Slug: r.WorkspaceSlug,
				Organization: &OrgBrief{
					ID:   r.OrganizationID.String(),
					Name: r.OrgName,
					Slug: r.OrgSlug,
				},
			},
		},
		CreatedByUser: createdUser,
		UpdatedByUser: updatedUser,
		Tasks:         tasks,
		Count:         &SprintCount{Tasks: r.TaskCount},
	}
}

