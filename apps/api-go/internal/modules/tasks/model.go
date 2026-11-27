package tasks

import (
	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// EventEmitter broadcasts real-time task domain events (optional, wired in Phase 5).
type EventEmitter interface {
	EmitTaskCreated(projectID string, payload any)
	EmitTaskUpdated(projectID, taskID string, payload any)
	EmitTaskDeleted(projectID, taskID string)
}

type effectiveAccess struct {
	role         middleware.Role
	isSuperAdmin bool
	isElevated   bool
	canChange    bool
}

func evaluateTaskAccess(row *sqlc.GetTaskAccessDetailsRow, userID uuid.UUID, isSuperAdmin bool) (effectiveAccess, error) {
	if isSuperAdmin {
		return effectiveAccess{role: middleware.RoleSuperAdmin, isSuperAdmin: true, isElevated: true, canChange: true}, nil
	}
	if row.OrgOwnerID == userID && !row.OrgArchived {
		return effectiveAccess{role: middleware.RoleOwner, isElevated: true, canChange: true}, nil
	}
	if row.OrgRole.Valid && (row.OrgRole.Role == sqlc.RoleMANAGER || row.OrgRole.Role == sqlc.RoleOWNER) {
		return effectiveAccess{role: middleware.Role(row.OrgRole.Role), isElevated: true, canChange: !row.OrgArchived}, nil
	}
	if row.WsRole.Valid && (row.WsRole.Role == sqlc.RoleMANAGER || row.WsRole.Role == sqlc.RoleOWNER) {
		return effectiveAccess{role: middleware.Role(row.WsRole.Role), isElevated: true, canChange: !row.OrgArchived}, nil
	}

	hasAnyMembership := row.ProjectRole.Valid || row.WsRole.Valid || row.OrgRole.Valid || row.IsAssignee || row.IsReporter
	if !hasAnyMembership {
		if row.Visibility == sqlc.ProjectVisibilityPUBLIC {
			return effectiveAccess{role: middleware.RoleViewer, isElevated: false, canChange: false}, nil
		}
		return effectiveAccess{}, apperror.Forbidden("Not a member of this task's project")
	}

	eff := middleware.RoleViewer
	if row.ProjectRole.Valid {
		eff = middleware.Role(row.ProjectRole.Role)
	} else if row.WsRole.Valid {
		eff = middleware.Role(row.WsRole.Role)
	} else if row.OrgRole.Valid {
		eff = middleware.Role(row.OrgRole.Role)
	}

	isCreator := row.CreatedByID.Valid && uuid.UUID(row.CreatedByID.Bytes) == userID
	isElevated := eff == middleware.RoleManager || eff == middleware.RoleOwner
	canChange := !row.OrgArchived && (isElevated || isCreator || row.IsAssignee || row.IsReporter)

	return effectiveAccess{
		role:       eff,
		isElevated: isElevated,
		canChange:  canChange,
	}, nil
}

func mapTaskRowBase(r *sqlc.GetTaskFullByIDRow) TaskResponse {
	var sprintBrief *SprintBrief
	if r.SprintID.Valid {
		sid := uuid.UUID(r.SprintID.Bytes).String()
		sprintBrief = &SprintBrief{
			ID:        sid,
			Name:      r.SprintName.String,
			Slug:      pgTextPtr(r.SprintSlug),
			Status:    string(r.SprintStatus.SprintStatus),
			StartDate: pgTSPtr(r.SprintStartDate),
			EndDate:   pgTSPtr(r.SprintEndDate),
		}
	}

	var parentBrief *ParentTaskBrief
	if r.ParentTaskID.Valid {
		pid := uuid.UUID(r.ParentTaskID.Bytes).String()
		parentBrief = &ParentTaskBrief{
			ID:    pid,
			Title: r.ParentTitle.String,
			Slug:  r.ParentSlug.String,
			Type:  string(r.ParentType.TaskType),
		}
	}

	var creatorBrief *UserBrief
	if r.CreatedByID.Valid {
		cid := uuid.UUID(r.CreatedByID.Bytes).String()
		creatorBrief = &UserBrief{
			ID:        cid,
			FirstName: pgTextPtr(r.CreatorFirstName),
			LastName:  pgTextPtr(r.CreatorLastName),
		}
	}

	return TaskResponse{
		ID:                r.ID.String(),
		Title:             r.Title,
		Description:       pgTextPtr(r.Description),
		Type:              string(r.Type),
		Priority:          string(r.Priority),
		TaskNumber:        r.TaskNumber,
		Slug:              r.Slug,
		StartDate:         pgTSPtr(r.StartDate),
		DueDate:           pgTSPtr(r.DueDate),
		CompletedAt:       pgTSPtr(r.CompletedAt),
		StoryPoints:       pgInt4Ptr(r.StoryPoints),
		OriginalEstimate:  pgInt4Ptr(r.OriginalEstimate),
		RemainingEstimate: pgInt4Ptr(r.RemainingEstimate),
		CustomFields:      decodeJSONField(r.CustomFields),
		ProjectID:         r.ProjectID.String(),
		StatusID:          r.StatusID.String(),
		SprintID:          pgUUIDPtr(r.SprintID),
		ParentTaskID:      pgUUIDPtr(r.ParentTaskID),
		CreatedBy:         pgUUIDPtr(r.CreatedByID),
		CreatedAt:         formatTS(r.CreatedAt),
		UpdatedAt:         formatTS(r.UpdatedAt),
		IsArchived:        r.IsArchived,
		AllowEmailReplies: r.AllowEmailReplies,
		IsRecurring:       r.IsRecurring,
		Project: &ProjectBrief{
			ID:   r.ProjectID.String(),
			Name: r.ProjectName,
			Slug: r.ProjectSlug,
			Workspace: &WorkspaceBrief{
				ID:             r.WorkspaceID.String(),
				Name:           r.WorkspaceName,
				Slug:           r.WorkspaceSlug,
				OrganizationID: r.OrganizationID.String(),
				Organization: &OrgBrief{
					ID:   r.OrganizationID.String(),
					Name: r.OrgName,
					Slug: r.OrgSlug,
				},
			},
		},
		Status: &StatusBrief{
			ID:       r.StatusID.String(),
			Name:     r.StatusName,
			Color:    r.StatusColor,
			Category: string(r.StatusCategory),
		},
		Sprint:        sprintBrief,
		ParentTask:    parentBrief,
		CreatedByUser: creatorBrief,
		Assignees:     []UserBrief{},
		Reporters:     []UserBrief{},
		Count: &TaskCounts{
			ChildTasks:  r.ChildTaskCount,
			Comments:    r.CommentCount,
			Attachments: r.AttachmentCount,
			Watchers:    r.WatcherCount,
			TimeEntries: r.TimeEntryCount,
		},
	}
}

func mapRecurrenceModel(rt *sqlc.RecurringTask) *RecurrenceResponse {
	if rt == nil {
		return nil
	}
	days := rt.DaysOfWeek
	if days == nil {
		days = []int32{}
	}
	return &RecurrenceResponse{
		ID:                rt.ID.String(),
		TaskID:            rt.TaskID.String(),
		RecurrenceType:    string(rt.RecurrenceType),
		Interval:          rt.Interval,
		DaysOfWeek:        days,
		DayOfMonth:        pgInt4Ptr(rt.DayOfMonth),
		MonthOfYear:       pgInt4Ptr(rt.MonthOfYear),
		EndType:           string(rt.EndType),
		EndDate:           pgTSPtr(rt.EndDate),
		OccurrenceCount:   pgInt4Ptr(rt.OccurrenceCount),
		CurrentOccurrence: rt.CurrentOccurrence,
		NextOccurrence:    formatTS(rt.NextOccurrence),
		IsActive:          rt.IsActive,
	}
}

