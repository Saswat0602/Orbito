package sprints

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/slugify"
)

// Service defines business operations for sprints.
type Service interface {
	Create(ctx context.Context, req CreateSprintRequest, userID uuid.UUID, isSuperAdmin bool) (*SprintResponse, error)
	FindAll(ctx context.Context, projectID *uuid.UUID, status *string, userID uuid.UUID, isSuperAdmin bool) ([]SprintResponse, error)
	FindAllByProjectSlug(ctx context.Context, projectSlug string, status *string, userID uuid.UUID, isSuperAdmin bool) ([]SprintResponse, error)
	FindBySlug(ctx context.Context, projectSlug, sprintSlug string, userID uuid.UUID, isSuperAdmin bool) (*SprintResponse, error)
	FindOne(ctx context.Context, id, userID uuid.UUID, isSuperAdmin bool) (*SprintResponse, error)
	GetActiveSprint(ctx context.Context, projectID, userID uuid.UUID, isSuperAdmin bool) (*SprintResponse, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateSprintRequest, userID uuid.UUID, isSuperAdmin bool) (*SprintResponse, error)
	StartSprint(ctx context.Context, id, userID uuid.UUID, isSuperAdmin bool) (*SprintResponse, error)
	CompleteSprint(ctx context.Context, id, userID uuid.UUID, isSuperAdmin bool) (*SprintResponse, error)
	Remove(ctx context.Context, id, userID uuid.UUID, isSuperAdmin bool) error
	ArchiveSprint(ctx context.Context, id, userID uuid.UUID, isSuperAdmin bool) error
}

type service struct {
	repo    Repository
	emitter EventEmitter
}

// NewService constructs a sprints service. emitter is optional and may be nil.
func NewService(repo Repository, emitter EventEmitter) Service {
	return &service{repo: repo, emitter: emitter}
}

type projectAccess struct {
	projectID  uuid.UUID
	isElevated bool
}

func evaluateProjectAccess(
	projectID uuid.UUID,
	visibility sqlc.ProjectVisibility,
	orgOwnerID uuid.UUID,
	orgArchived bool,
	orgRole, wsRole, projectRole sqlc.NullRole,
	userID uuid.UUID,
	isSuperAdmin bool,
) (projectAccess, error) {
	if isSuperAdmin {
		return projectAccess{projectID: projectID, isElevated: true}, nil
	}
	if orgOwnerID == userID && !orgArchived {
		return projectAccess{projectID: projectID, isElevated: true}, nil
	}
	if orgRole.Valid && (orgRole.Role == sqlc.RoleMANAGER || orgRole.Role == sqlc.RoleOWNER) {
		return projectAccess{projectID: projectID, isElevated: true}, nil
	}
	if wsRole.Valid && (wsRole.Role == sqlc.RoleMANAGER || wsRole.Role == sqlc.RoleOWNER) {
		return projectAccess{projectID: projectID, isElevated: true}, nil
	}
	if projectRole.Valid {
		elevated := projectRole.Role == sqlc.RoleMANAGER || projectRole.Role == sqlc.RoleOWNER
		return projectAccess{projectID: projectID, isElevated: elevated}, nil
	}
	if visibility == sqlc.ProjectVisibilityINTERNAL && wsRole.Valid {
		return projectAccess{projectID: projectID, isElevated: false}, nil
	}
	if visibility == sqlc.ProjectVisibilityPUBLIC {
		return projectAccess{projectID: projectID, isElevated: false}, nil
	}
	if !wsRole.Valid && !orgRole.Valid {
		return projectAccess{}, apperror.Forbidden("Not a member of this project")
	}

	effective := middleware.RoleViewer
	if wsRole.Valid {
		effective = middleware.Role(wsRole.Role)
	} else if orgRole.Valid {
		effective = middleware.Role(orgRole.Role)
	}
	elevated := effective == middleware.RoleManager || effective == middleware.RoleOwner
	return projectAccess{projectID: projectID, isElevated: elevated}, nil
}

func (s *service) checkProjectAccess(ctx context.Context, projectID, userID uuid.UUID, isSuperAdmin bool) (projectAccess, *sqlc.GetProjectAccessInfoForSprintRow, error) {
	info, err := s.repo.GetProjectAccessByID(ctx, projectID, userID)
	if err != nil {
		return projectAccess{}, nil, apperror.NotFound("Project not found")
	}
	acc, err := evaluateProjectAccess(info.ID, info.Visibility, info.OrgOwnerID, info.OrgArchived, info.OrgRole, info.WsRole, info.ProjectRole, userID, isSuperAdmin)
	if err != nil {
		return projectAccess{}, nil, err
	}
	return acc, info, nil
}

func (s *service) checkProjectAccessBySlug(ctx context.Context, slug string, userID uuid.UUID, isSuperAdmin bool) (projectAccess, *sqlc.GetProjectAccessInfoForSprintBySlugRow, error) {
	info, err := s.repo.GetProjectAccessBySlug(ctx, slug, userID)
	if err != nil {
		return projectAccess{}, nil, apperror.NotFound("Project not found")
	}
	acc, err := evaluateProjectAccess(info.ID, info.Visibility, info.OrgOwnerID, info.OrgArchived, info.OrgRole, info.WsRole, info.ProjectRole, userID, isSuperAdmin)
	if err != nil {
		return projectAccess{}, nil, err
	}
	return acc, info, nil
}

func (s *service) generateUniqueSlug(ctx context.Context, name string, projectID uuid.UUID) (string, error) {
	base := slugify.Slugify(name)
	if base == "item" {
		base = "sprint"
	}
	candidate := base
	for counter := 1; ; counter++ {
		exists, err := s.repo.CheckSlugExists(ctx, projectID, candidate, nil)
		if err != nil {
			return "", fmt.Errorf("check sprint slug: %w", err)
		}
		if !exists {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s-%d", base, counter)
	}
}

func (s *service) ensureSlug(ctx context.Context, sprintID, projectID uuid.UUID, name string, current pgtype.Text) pgtype.Text {
	if current.Valid && current.String != "" {
		return current
	}
	if name == "" {
		return current
	}
	generated, err := s.generateUniqueSlug(ctx, name, projectID)
	if err != nil {
		return current
	}
	if err := s.repo.UpdateSlugByID(ctx, sprintID, generated); err != nil {
		return current
	}
	return pgtype.Text{String: generated, Valid: true}
}

func (s *service) recordActivity(ctx context.Context, actType sqlc.ActivityType, desc string, sprintID, userID, orgID uuid.UUID, oldVal, newVal any) {
	var oldBytes, newBytes []byte
	if oldVal != nil {
		oldBytes, _ = json.Marshal(oldVal)
	}
	if newVal != nil {
		newBytes, _ = json.Marshal(newVal)
	}
	_ = s.repo.LogActivity(ctx, sqlc.CreateSprintActivityLogParams{
		ID:             uuid.New(),
		Type:           actType,
		Description:    desc,
		EntityType:     "Sprint",
		EntityID:       sprintID,
		OldValue:       oldBytes,
		NewValue:       newBytes,
		UserID:         userID,
		OrganizationID: pgtype.UUID{Bytes: orgID, Valid: orgID != uuid.Nil},
		CreatedByID:    pgtype.UUID{Bytes: userID, Valid: true},
		UpdatedByID:    pgtype.UUID{Bytes: userID, Valid: true},
	})
}

func (s *service) Create(ctx context.Context, req CreateSprintRequest, userID uuid.UUID, isSuperAdmin bool) (*SprintResponse, error) {
	hasID := req.ProjectID != nil && *req.ProjectID != ""
	hasSlug := req.ProjectSlug != nil && *req.ProjectSlug != ""
	if !hasID && !hasSlug {
		return nil, apperror.BadRequest("Either projectSlug or projectId must be provided")
	}

	var acc projectAccess
	var projectID uuid.UUID
	var projectName string
	var err error

	if hasID {
		pid := uuid.MustParse(*req.ProjectID)
		var info *sqlc.GetProjectAccessInfoForSprintRow
		acc, info, err = s.checkProjectAccess(ctx, pid, userID, isSuperAdmin)
		if err != nil {
			return nil, err
		}
		projectID = info.ID
		projectName = info.Name
	} else {
		var info *sqlc.GetProjectAccessInfoForSprintBySlugRow
		acc, info, err = s.checkProjectAccessBySlug(ctx, *req.ProjectSlug, userID, isSuperAdmin)
		if err != nil {
			return nil, err
		}
		projectID = info.ID
		projectName = info.Name
	}

	if !acc.isElevated {
		return nil, apperror.Forbidden("Only managers and owners can create sprints")
	}

	status := sqlc.SprintStatusPLANNING
	if req.Status != nil && *req.Status != "" {
		status = sqlc.SprintStatus(*req.Status)
	}

	if status == sqlc.SprintStatusACTIVE {
		activeExists, aErr := s.repo.HasActiveSprint(ctx, projectID, nil)
		if aErr != nil {
			return nil, fmt.Errorf("check active sprint: %w", aErr)
		}
		if activeExists {
			return nil, apperror.Conflict("There is already an active sprint in this project")
		}
	}

	slug, err := s.generateUniqueSlug(ctx, req.Name, projectID)
	if err != nil {
		return nil, err
	}

	var goal pgtype.Text
	if req.Goal != nil {
		goal = pgtype.Text{String: *req.Goal, Valid: true}
	}
	var startTS, endTS pgtype.Timestamp
	if req.StartDate != nil && *req.StartDate != "" {
		parsed, ok := parseDateString(*req.StartDate)
		if !ok {
			return nil, apperror.BadRequest("startDate must be a valid ISO 8601 date string")
		}
		startTS = pgtype.Timestamp{Time: parsed, Valid: true}
	}
	if req.EndDate != nil && *req.EndDate != "" {
		parsed, ok := parseDateString(*req.EndDate)
		if !ok {
			return nil, apperror.BadRequest("endDate must be a valid ISO 8601 date string")
		}
		endTS = pgtype.Timestamp{Time: parsed, Valid: true}
	}

	sprintID := uuid.New()
	_, err = s.repo.Create(ctx, sqlc.CreateSprintRecordParams{
		ID:          sprintID,
		Name:        req.Name,
		Slug:        pgtype.Text{String: slug, Valid: true},
		Goal:        goal,
		Status:      status,
		StartDate:   startTS,
		EndDate:     endTS,
		ProjectID:   projectID,
		CreatedByID: pgtype.UUID{Bytes: userID, Valid: true},
		UpdatedByID: pgtype.UUID{Bytes: userID, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create sprint: %w", err)
	}

	details, err := s.repo.GetDetailsByID(ctx, sprintID)
	if err != nil {
		return nil, apperror.NotFound("Sprint not found")
	}

	s.recordActivity(
		ctx,
		sqlc.ActivityTypeSPRINTCREATED,
		fmt.Sprintf("Created sprint %q in project %q", details.Name, projectName),
		details.ID,
		userID,
		details.OrganizationID,
		nil,
		map[string]any{
			"name":      details.Name,
			"slug":      details.Slug.String,
			"status":    details.Status,
			"projectId": details.ProjectID.String(),
		},
	)

	return mapDetailsRow(details, nil), nil
}

func (s *service) FindAll(ctx context.Context, projectID *uuid.UUID, status *string, userID uuid.UUID, isSuperAdmin bool) ([]SprintResponse, error) {
	var pidParam pgtype.UUID
	if projectID == nil {
		if !isSuperAdmin {
			return nil, apperror.BadRequest("projectId is required for non-super-admins")
		}
	} else {
		if _, _, err := s.checkProjectAccess(ctx, *projectID, userID, isSuperAdmin); err != nil {
			return nil, err
		}
		pidParam = pgtype.UUID{Bytes: *projectID, Valid: true}
	}

	var statusParam pgtype.Text
	if status != nil && *status != "" {
		statusParam = pgtype.Text{String: *status, Valid: true}
	}

	rows, err := s.repo.ListFiltered(ctx, pidParam, pgtype.Text{}, statusParam)
	if err != nil {
		return nil, fmt.Errorf("list sprints: %w", err)
	}
	return s.mapFilteredRows(ctx, rows), nil
}

func (s *service) FindAllByProjectSlug(ctx context.Context, projectSlug string, status *string, userID uuid.UUID, isSuperAdmin bool) ([]SprintResponse, error) {
	var slugParam pgtype.Text
	if projectSlug == "" {
		if !isSuperAdmin {
			return nil, apperror.BadRequest("project slug is required for non-super-admins")
		}
	} else {
		if _, _, err := s.checkProjectAccessBySlug(ctx, projectSlug, userID, isSuperAdmin); err != nil {
			return nil, err
		}
		slugParam = pgtype.Text{String: projectSlug, Valid: true}
	}

	var statusParam pgtype.Text
	if status != nil && *status != "" {
		statusParam = pgtype.Text{String: *status, Valid: true}
	}

	rows, err := s.repo.ListFiltered(ctx, pgtype.UUID{}, slugParam, statusParam)
	if err != nil {
		return nil, fmt.Errorf("list sprints by slug: %w", err)
	}
	return s.mapFilteredRows(ctx, rows), nil
}

func (s *service) mapFilteredRows(ctx context.Context, rows []sqlc.ListSprintsFilteredRow) []SprintResponse {
	out := make([]SprintResponse, 0, len(rows))
	for _, r := range rows {
		slug := s.ensureSlug(ctx, r.ID, r.ProjectID, r.Name, r.Slug)
		out = append(out, SprintResponse{
			ID:          r.ID.String(),
			Name:        r.Name,
			Slug:        pgTextPtr(slug),
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
				},
			},
			Count: &SprintCount{Tasks: r.TaskCount},
		})
	}
	return out
}

func (s *service) FindBySlug(ctx context.Context, projectSlug, sprintSlug string, userID uuid.UUID, isSuperAdmin bool) (*SprintResponse, error) {
	_, proj, err := s.checkProjectAccessBySlug(ctx, projectSlug, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}
	r, err := s.repo.GetByProjectAndSlug(ctx, proj.Slug, sprintSlug)
	if err != nil {
		return nil, apperror.NotFound("Sprint not found")
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
			},
		},
		Count: &SprintCount{Tasks: r.TaskCount},
	}, nil
}

func (s *service) buildSprintTasks(ctx context.Context, sprintID uuid.UUID, includeReporters bool) ([]SprintTaskItem, error) {
	rows, err := s.repo.ListSprintTasks(ctx, sprintID)
	if err != nil {
		return nil, fmt.Errorf("list sprint tasks: %w", err)
	}
	taskIDs := make([]uuid.UUID, 0, len(rows))
	for _, r := range rows {
		taskIDs = append(taskIDs, r.ID)
	}

	assignees, err := s.repo.ListTaskAssignees(ctx, taskIDs)
	if err != nil {
		return nil, fmt.Errorf("list sprint task assignees: %w", err)
	}
	assigneeMap := make(map[uuid.UUID][]TaskUserWrapper)
	for _, a := range assignees {
		f, l := a.FirstName, a.LastName
		assigneeMap[a.TaskID] = append(assigneeMap[a.TaskID], TaskUserWrapper{
			User: TaskUserBrief{
				ID:        a.ID.String(),
				FirstName: &f,
				LastName:  &l,
				Avatar:    pgTextPtr(a.Avatar),
			},
		})
	}

	reporterMap := make(map[uuid.UUID][]TaskUserWrapper)
	if includeReporters {
		reporters, err := s.repo.ListTaskReporters(ctx, taskIDs)
		if err != nil {
			return nil, fmt.Errorf("list sprint task reporters: %w", err)
		}
		for _, rep := range reporters {
			f, l := rep.FirstName, rep.LastName
			reporterMap[rep.TaskID] = append(reporterMap[rep.TaskID], TaskUserWrapper{
				User: TaskUserBrief{
					ID:        rep.ID.String(),
					FirstName: &f,
					LastName:  &l,
					Avatar:    pgTextPtr(rep.Avatar),
				},
			})
		}
	}

	out := make([]SprintTaskItem, 0, len(rows))
	for _, t := range rows {
		aList := assigneeMap[t.ID]
		if aList == nil {
			aList = []TaskUserWrapper{}
		}
		item := SprintTaskItem{
			ID:                t.ID.String(),
			Title:             t.Title,
			Description:       pgTextPtr(t.Description),
			Type:              string(t.Type),
			Priority:          string(t.Priority),
			TaskNumber:        t.TaskNumber,
			Slug:              t.Slug,
			StartDate:         pgTSPtr(t.StartDate),
			DueDate:           pgTSPtr(t.DueDate),
			CompletedAt:       pgTSPtr(t.CompletedAt),
			StoryPoints:       pgInt4Ptr(t.StoryPoints),
			OriginalEstimate:  pgInt4Ptr(t.OriginalEstimate),
			RemainingEstimate: pgInt4Ptr(t.RemainingEstimate),
			ProjectID:         t.ProjectID.String(),
			StatusID:          t.StatusID.String(),
			SprintID:          pgUUIDPtr(t.SprintID),
			ParentTaskID:      pgUUIDPtr(t.ParentTaskID),
			CreatedAt:         formatTS(t.CreatedAt),
			UpdatedAt:         formatTS(t.UpdatedAt),
			IsArchived:        t.IsArchived,
			Status: &TaskStatusBrief{
				ID:       t.StatusID.String(),
				Name:     t.StatusName,
				Color:    t.StatusColor,
				Category: string(t.StatusCategory),
			},
			Assignees: aList,
		}
		if includeReporters {
			rList := reporterMap[t.ID]
			if rList == nil {
				rList = []TaskUserWrapper{}
			}
			item.Reporters = rList
			item.Count = &SprintTaskCount{
				ChildTasks: t.ChildTaskCount,
				Comments:   t.CommentCount,
			}
		}
		out = append(out, item)
	}
	return out, nil
}

func (s *service) FindOne(ctx context.Context, id, userID uuid.UUID, isSuperAdmin bool) (*SprintResponse, error) {
	details, err := s.repo.GetDetailsByID(ctx, id)
	if err != nil {
		return nil, apperror.NotFound("Sprint not found")
	}
	if _, _, authErr := s.checkProjectAccess(ctx, details.ProjectID, userID, isSuperAdmin); authErr != nil {
		return nil, authErr
	}
	details.Slug = s.ensureSlug(ctx, details.ID, details.ProjectID, details.Name, details.Slug)

	tasks, err := s.buildSprintTasks(ctx, details.ID, true)
	if err != nil {
		return nil, err
	}
	return mapDetailsRow(details, tasks), nil
}

func (s *service) GetActiveSprint(ctx context.Context, projectID, userID uuid.UUID, isSuperAdmin bool) (*SprintResponse, error) {
	if _, _, err := s.checkProjectAccess(ctx, projectID, userID, isSuperAdmin); err != nil {
		return nil, err
	}
	r, err := s.repo.GetActiveByProjectID(ctx, projectID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get active sprint: %w", err)
	}
	tasks, err := s.buildSprintTasks(ctx, r.ID, false)
	if err != nil {
		return nil, err
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
		},
		Tasks: tasks,
		Count: &SprintCount{Tasks: r.TaskCount},
	}, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req UpdateSprintRequest, userID uuid.UUID, isSuperAdmin bool) (*SprintResponse, error) {
	current, err := s.repo.GetDetailsByID(ctx, id)
	if err != nil {
		return nil, apperror.NotFound("Sprint not found")
	}

	acc, _, err := s.checkProjectAccess(ctx, current.ProjectID, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}
	if !acc.isElevated {
		return nil, apperror.Forbidden("Only managers and owners can update sprints")
	}

	if req.Status != nil && *req.Status == string(sqlc.SprintStatusACTIVE) && current.Status != sqlc.SprintStatusACTIVE {
		activeExists, aErr := s.repo.HasActiveSprint(ctx, current.ProjectID, &id)
		if aErr != nil {
			return nil, fmt.Errorf("check active sprint conflict: %w", aErr)
		}
		if activeExists {
			return nil, apperror.Conflict("There is already an active sprint in this project")
		}
	}

	params := sqlc.UpdateSprintByIDParams{
		ID:          id,
		UpdatedByID: pgtype.UUID{Bytes: userID, Valid: true},
	}
	if req.Name != nil {
		params.Name = pgtype.Text{String: *req.Name, Valid: true}
	}
	if req.Goal != nil {
		params.Goal = pgtype.Text{String: *req.Goal, Valid: true}
	}
	if req.Status != nil {
		params.Status = sqlc.NullSprintStatus{SprintStatus: sqlc.SprintStatus(*req.Status), Valid: true}
	}
	if req.StartDate != nil && *req.StartDate != "" {
		parsed, ok := parseDateString(*req.StartDate)
		if !ok {
			return nil, apperror.BadRequest("startDate must be a valid ISO 8601 date string")
		}
		params.StartDate = pgtype.Timestamp{Time: parsed, Valid: true}
	}
	if req.EndDate != nil && *req.EndDate != "" {
		parsed, ok := parseDateString(*req.EndDate)
		if !ok {
			return nil, apperror.BadRequest("endDate must be a valid ISO 8601 date string")
		}
		params.EndDate = pgtype.Timestamp{Time: parsed, Valid: true}
	}

	updatedID, err := s.repo.UpdateByID(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperror.NotFound("Sprint not found")
		}
		return nil, fmt.Errorf("update sprint: %w", err)
	}

	details, err := s.repo.GetDetailsByID(ctx, updatedID)
	if err != nil {
		return nil, apperror.NotFound("Sprint not found")
	}

	actType := sqlc.ActivityTypeSPRINTUPDATED
	desc := fmt.Sprintf("Updated sprint %q", details.Name)
	if req.Status != nil && *req.Status != string(current.Status) {
		switch details.Status {
		case sqlc.SprintStatusACTIVE:
			actType = sqlc.ActivityTypeSPRINTSTARTED
			desc = fmt.Sprintf("Started sprint %q", details.Name)
		case sqlc.SprintStatusCOMPLETED:
			actType = sqlc.ActivityTypeSPRINTCOMPLETED
			desc = fmt.Sprintf("Completed sprint %q", details.Name)
		case sqlc.SprintStatusCANCELLED:
			actType = sqlc.ActivityTypeSPRINTCANCELLED
			desc = fmt.Sprintf("Cancelled sprint %q", details.Name)
		}
	}
	s.recordActivity(
		ctx,
		actType,
		desc,
		details.ID,
		userID,
		details.OrganizationID,
		map[string]any{"status": current.Status, "projectId": current.ProjectID.String()},
		map[string]any{"name": details.Name, "status": details.Status, "projectId": current.ProjectID.String()},
	)

	return mapDetailsRow(details, nil), nil
}

func (s *service) StartSprint(ctx context.Context, id, userID uuid.UUID, isSuperAdmin bool) (*SprintResponse, error) {
	current, err := s.repo.GetDetailsByID(ctx, id)
	if err != nil {
		return nil, apperror.NotFound("Sprint not found")
	}
	if current.Status != sqlc.SprintStatusPLANNING {
		return nil, apperror.BadRequest("Only planning sprints can be started")
	}
	if !current.StartDate.Valid || !current.EndDate.Valid {
		return nil, apperror.BadRequest("Sprint must have start and end dates to be started")
	}
	activeExists, err := s.repo.HasActiveSprint(ctx, current.ProjectID, nil)
	if err != nil {
		return nil, fmt.Errorf("check active sprint: %w", err)
	}
	if activeExists {
		return nil, apperror.Conflict("There is already an active sprint in this project")
	}

	activeStatus := string(sqlc.SprintStatusACTIVE)
	updated, err := s.Update(ctx, id, UpdateSprintRequest{Status: &activeStatus}, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}
	if s.emitter != nil {
		s.emitter.EmitSprintStarted(current.ProjectID.String(), updated)
	}
	return updated, nil
}

func (s *service) CompleteSprint(ctx context.Context, id, userID uuid.UUID, isSuperAdmin bool) (*SprintResponse, error) {
	current, err := s.repo.GetDetailsByID(ctx, id)
	if err != nil {
		return nil, apperror.NotFound("Sprint not found")
	}
	if current.Status != sqlc.SprintStatusACTIVE {
		return nil, apperror.BadRequest("Only active sprints can be completed")
	}

	completedStatus := string(sqlc.SprintStatusCOMPLETED)
	updated, err := s.Update(ctx, id, UpdateSprintRequest{Status: &completedStatus}, userID, isSuperAdmin)
	if err != nil {
		return nil, err
	}
	if s.emitter != nil {
		s.emitter.EmitSprintCompleted(current.ProjectID.String(), updated)
	}
	return updated, nil
}

func (s *service) Remove(ctx context.Context, id, userID uuid.UUID, isSuperAdmin bool) error {
	current, err := s.repo.GetDetailsByID(ctx, id)
	if err != nil {
		return apperror.NotFound("Sprint not found")
	}
	acc, _, err := s.checkProjectAccess(ctx, current.ProjectID, userID, isSuperAdmin)
	if err != nil {
		return err
	}
	if !acc.isElevated {
		return apperror.Forbidden("Only managers and owners can delete sprints")
	}
	if current.Status == sqlc.SprintStatusACTIVE {
		return apperror.BadRequest("Cannot delete an active sprint")
	}

	affected, err := s.repo.DeleteByID(ctx, id)
	if err != nil {
		return fmt.Errorf("delete sprint: %w", err)
	}
	if affected == 0 {
		return apperror.NotFound("Sprint not found")
	}

	s.recordActivity(
		ctx,
		sqlc.ActivityTypeSPRINTDELETED,
		fmt.Sprintf("Deleted sprint %q", current.Name),
		id,
		userID,
		current.OrganizationID,
		map[string]any{
			"name":      current.Name,
			"slug":      pgTextPtr(current.Slug),
			"status":    current.Status,
			"projectId": current.ProjectID.String(),
		},
		nil,
	)
	return nil
}

func (s *service) ArchiveSprint(ctx context.Context, id, userID uuid.UUID, isSuperAdmin bool) error {
	current, err := s.repo.GetDetailsByID(ctx, id)
	if err != nil {
		return apperror.NotFound("Sprint not found")
	}
	acc, _, err := s.checkProjectAccess(ctx, current.ProjectID, userID, isSuperAdmin)
	if err != nil {
		return err
	}
	if !acc.isElevated {
		return apperror.Forbidden("Only managers and owners can archive sprints")
	}
	affected, err := s.repo.ArchiveByID(ctx, id)
	if err != nil {
		return fmt.Errorf("archive sprint: %w", err)
	}
	if affected == 0 {
		return apperror.NotFound("Sprint not found")
	}
	return nil
}
