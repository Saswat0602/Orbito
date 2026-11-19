package projects

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/slugify"
)

// Service defines business operations for projects.
type Service interface {
	Create(ctx context.Context, req CreateProjectRequest, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectResponse, error)
	FindAll(ctx context.Context, workspaceID *uuid.UUID, status, priority, search *string, page, pageSize int, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectListResponse, error)
	FindByOrganizationID(ctx context.Context, orgID uuid.UUID, workspaceID *uuid.UUID, status, priority, search *string, page, pageSize int, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectListResponse, error)
	FindOne(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectResponse, error)
	GetBySlug(ctx context.Context, slug string, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectResponse, error)
	FindByKey(ctx context.Context, workspaceID uuid.UUID, key string, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectResponse, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateProjectRequest, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectResponse, error)
	Archive(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) error
	Unarchive(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) error
	Delete(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) error
	GetStats(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectStatsResponse, error)
	GetCharts(ctx context.Context, slug string, currentUserID uuid.UUID, isSuperAdmin bool) (map[string]any, error)
}

type service struct {
	repo         Repository
	rbacResolver middleware.RBACResolver
}

// NewService constructs a projects service instance.
func NewService(repo Repository, rbacResolver middleware.RBACResolver) Service {
	return &service{
		repo:         repo,
		rbacResolver: rbacResolver,
	}
}

func generateTaskPrefix(name string) string {
	words := regexp.MustCompile(`[\s-]+`).Split(strings.TrimSpace(name), -1)
	prefix := ""
	if len(words) > 1 {
		for _, w := range words {
			if len(w) > 0 {
				prefix += string(w[0])
			}
		}
	} else {
		vowels := regexp.MustCompile(`(?i)[aeiou\s-]`)
		consonants := vowels.ReplaceAllString(name, "")
		if len(consonants) >= 2 {
			prefix = consonants
		} else {
			prefix = name
		}
	}
	prefix = strings.ToUpper(prefix)
	if len(prefix) > 4 {
		prefix = prefix[:4]
	}
	if !regexp.MustCompile(`^[A-Z0-9]+$`).MatchString(prefix) {
		prefix = "PROJ"
	}
	return prefix
}

func (s *service) generateUniqueSlug(ctx context.Context, rawName string) (string, error) {
	baseSlug := slugify.Slugify(rawName)
	if baseSlug == "" {
		baseSlug = "project"
	}

	existing, err := s.repo.ListSlugsStartingWith(ctx, baseSlug)
	if err != nil {
		return "", err
	}

	if len(existing) == 0 {
		return baseSlug, nil
	}

	hasExact := false
	maxSuffix := 0
	re := regexp.MustCompile(fmt.Sprintf(`^%s-(\d+)$`, regexp.QuoteMeta(baseSlug)))

	for _, s := range existing {
		if s == baseSlug {
			hasExact = true
		}
		matches := re.FindStringSubmatch(s)
		if len(matches) == 2 {
			num, _ := strconv.Atoi(matches[1])
			if num > maxSuffix {
				maxSuffix = num
			}
		}
	}

	if !hasExact {
		return baseSlug, nil
	}
	return fmt.Sprintf("%s-%d", baseSlug, maxSuffix+1), nil
}

func (s *service) Create(ctx context.Context, req CreateProjectRequest, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectResponse, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, apperror.BadRequest("Project name is required")
	}

	ws, err := s.repo.GetWorkspaceByID(ctx, req.WorkspaceID)
	if err != nil {
		return nil, err
	}
	if ws == nil {
		return nil, apperror.NotFound("Workspace not found")
	}

	org, err := s.repo.GetOrganizationByID(ctx, ws.OrganizationID)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, apperror.NotFound("Organization not found")
	}
	if org.Archive {
		return nil, apperror.Forbidden("Cannot create project in an archived organization")
	}

	isOrgOwner := org.OwnerID == currentUserID
	if !isOrgOwner && !isSuperAdmin {
		wsRole, rErr := s.rbacResolver.GetWorkspaceMemberRole(ctx, currentUserID, req.WorkspaceID)
		if rErr != nil || (wsRole != middleware.RoleOwner && wsRole != middleware.RoleManager && wsRole != middleware.RoleMember) {
			return nil, apperror.Forbidden("Insufficient permissions to create projects in this workspace")
		}
	}

	slugToUse := req.Slug
	if strings.TrimSpace(slugToUse) == "" {
		slugToUse = req.Name
	}
	uniqueSlug, sErr := s.generateUniqueSlug(ctx, slugToUse)
	if sErr != nil {
		return nil, sErr
	}

	var workflowIDToUse uuid.UUID
	if req.WorkflowID != nil {
		customWf, wErr := s.repo.GetWorkflowByID(ctx, *req.WorkflowID)
		if wErr != nil {
			return nil, wErr
		}
		if customWf == nil {
			return nil, apperror.NotFound("Custom workflow not found")
		}
		if customWf.OrganizationID != ws.OrganizationID {
			return nil, apperror.Forbidden("Custom workflow must belong to the same organization as the workspace")
		}
		workflowIDToUse = customWf.ID
	} else {
		defaultWf, wErr := s.repo.GetDefaultWorkflowForOrg(ctx, ws.OrganizationID)
		if wErr != nil {
			return nil, wErr
		}
		if defaultWf == nil {
			return nil, apperror.NotFound("Default workflow not found for organization")
		}
		workflowIDToUse = defaultWf.ID
	}

	taskPrefix := ""
	if req.TaskPrefix != nil && *req.TaskPrefix != "" {
		taskPrefix = strings.ToUpper(strings.TrimSpace(*req.TaskPrefix))
	} else if req.Key != nil && *req.Key != "" {
		taskPrefix = strings.ToUpper(strings.TrimSpace(*req.Key))
	} else {
		taskPrefix = generateTaskPrefix(req.Name)
	}
	if len(taskPrefix) > 8 {
		taskPrefix = taskPrefix[:8]
	}
	if !regexp.MustCompile(`^[A-Z0-9]+$`).MatchString(taskPrefix) {
		taskPrefix = "PROJ"
	}

	status := sqlc.ProjectStatusPLANNING
	if req.Status != nil && *req.Status != "" {
		status = sqlc.ProjectStatus(strings.ToUpper(*req.Status))
	}
	priority := sqlc.ProjectPriorityMEDIUM
	if req.Priority != nil && *req.Priority != "" {
		priority = sqlc.ProjectPriority(strings.ToUpper(*req.Priority))
	}
	visibility := sqlc.ProjectVisibilityPRIVATE
	if req.Visibility != nil && *req.Visibility != "" {
		visibility = sqlc.ProjectVisibility(strings.ToUpper(*req.Visibility))
	}

	color := req.Color
	if color == "" {
		color = "#3b82f6"
	}

	projectID := uuid.New()
	var createdProject *sqlc.Project

	txErr := s.repo.WithTx(ctx, func(txRepo Repository) error {
		p, cErr := txRepo.Create(ctx, sqlc.CreateProjectParams{
			ID:          projectID,
			Name:        req.Name,
			Slug:        uniqueSlug,
			TaskPrefix:  pgtype.Text{String: taskPrefix, Valid: true},
			Description: toPgText(req.Description),
			Avatar:      toPgText(req.Avatar),
			Color:       color,
			Status:      status,
			Priority:    priority,
			Visibility:  visibility,
			StartDate:   toPgTimestampPtr(req.StartDate),
			EndDate:     toPgTimestampPtr(req.EndDate),
			Settings:    req.Settings,
			WorkflowID:  workflowIDToUse,
			WorkspaceID: req.WorkspaceID,
			CreatedByID: pgtype.UUID{Bytes: currentUserID, Valid: true},
			UpdatedByID: pgtype.UUID{Bytes: currentUserID, Valid: true},
		})
		if cErr != nil {
			return cErr
		}
		createdProject = p

		// Create default initial sprint
		now := time.Now()
		_, _ = txRepo.CreateInitialSprint(ctx, sqlc.CreateInitialSprintParams{
			ID:          uuid.New(),
			Name:        "Sprint 1",
			Goal:        pgtype.Text{String: "Initial sprint", Valid: true},
			Status:      sqlc.SprintStatusPLANNING,
			IsDefault:   true,
			StartDate:   toPgTimestamp(now),
			EndDate:     toPgTimestamp(now.Add(30 * 24 * time.Hour)),
			ProjectID:   projectID,
			CreatedByID: pgtype.UUID{Bytes: currentUserID, Valid: true},
			UpdatedByID: pgtype.UUID{Bytes: currentUserID, Valid: true},
		})

		// Add creator as OWNER of project
		_, _ = txRepo.CreateProjectMember(ctx, sqlc.CreateProjectMemberDirectParams{
			ID:          uuid.New(),
			UserID:      currentUserID,
			ProjectID:   projectID,
			Role:        sqlc.RoleOWNER,
			CreatedByID: pgtype.UUID{Bytes: currentUserID, Valid: true},
		})

		return nil
	})

	if txErr != nil {
		return nil, txErr
	}

	return s.FindOne(ctx, createdProject.ID, currentUserID, isSuperAdmin)
}

func (s *service) FindAll(
	ctx context.Context,
	workspaceID *uuid.UUID,
	status, priority, search *string,
	page, pageSize int,
	currentUserID uuid.UUID,
	isSuperAdmin bool,
) (*ProjectListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var p32 int32 = 1
	if page > 1 && page <= 1000000 {
		p32 = int32(page)
	}
	var ps32 int32 = 10
	if pageSize >= 1 && pageSize <= 100 {
		ps32 = int32(pageSize)
	}
	offset := (p32 - 1) * ps32

	rows, err := s.repo.ListForUser(ctx, sqlc.ListProjectsForUserParams{
		WorkspaceID: toPgUUID(workspaceID),
		Status:      toPgText(status),
		Priority:    toPgText(priority),
		Search:      toPgText(search),
		IsAdmin:     isSuperAdmin,
		UserID:      currentUserID,
		Lim:         ps32,
		Offs:        offset,
	})
	if err != nil {
		return nil, err
	}

	total, err := s.repo.CountForUser(ctx, sqlc.CountProjectsForUserParams{
		WorkspaceID: toPgUUID(workspaceID),
		Status:      toPgText(status),
		Priority:    toPgText(priority),
		Search:      toPgText(search),
		IsAdmin:     isSuperAdmin,
		UserID:      currentUserID,
	})
	if err != nil {
		return nil, err
	}

	data := make([]ProjectResponse, 0, len(rows))
	for _, r := range rows {
		var tp *string
		if r.TaskPrefix.Valid {
			tp = &r.TaskPrefix.String
		}
		var desc *string
		if r.Description.Valid {
			desc = &r.Description.String
		}
		var av *string
		if r.Avatar.Valid {
			av = &r.Avatar.String
		}
		var sDate, eDate *time.Time
		if r.StartDate.Valid {
			sDate = &r.StartDate.Time
		}
		if r.EndDate.Valid {
			eDate = &r.EndDate.Time
		}

		data = append(data, ProjectResponse{
			ID:            r.ID,
			Name:          r.Name,
			Slug:          r.Slug,
			TaskPrefix:    tp,
			Description:   desc,
			Avatar:        av,
			Color:         r.Color,
			Status:        string(r.Status),
			Priority:      string(r.Priority),
			Visibility:    string(r.Visibility),
			StartDate:     sDate,
			EndDate:       eDate,
			WorkflowID:    r.WorkflowID,
			WorkspaceID:   r.WorkspaceID,
			CreatedAt:     r.CreatedAt.Time,
			UpdatedAt:     r.UpdatedAt.Time,
			Archive:       r.Archive,
			WorkspaceName: r.WorkspaceName,
			WorkspaceSlug: r.WorkspaceSlug,
			TaskCount:     r.TaskCount,
			MemberCount:   r.MemberCount,
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	return &ProjectListResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *service) FindByOrganizationID(
	ctx context.Context,
	orgID uuid.UUID,
	workspaceID *uuid.UUID,
	status, priority, search *string,
	page, pageSize int,
	currentUserID uuid.UUID,
	isSuperAdmin bool,
) (*ProjectListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var p32 int32 = 1
	if page > 1 && page <= 1000000 {
		p32 = int32(page)
	}
	var ps32 int32 = 10
	if pageSize >= 1 && pageSize <= 100 {
		ps32 = int32(pageSize)
	}
	offset := (p32 - 1) * ps32

	rows, err := s.repo.ListByOrganization(ctx, sqlc.ListProjectsByOrganizationParams{
		OrganizationID: orgID,
		WorkspaceID:    toPgUUID(workspaceID),
		Status:         toPgText(status),
		Priority:       toPgText(priority),
		Search:         toPgText(search),
		IsAdmin:        isSuperAdmin,
		UserID:         currentUserID,
		Lim:            ps32,
		Offs:           offset,
	})
	if err != nil {
		return nil, err
	}

	total, err := s.repo.CountByOrganization(ctx, sqlc.CountProjectsByOrganizationParams{
		OrganizationID: orgID,
		WorkspaceID:    toPgUUID(workspaceID),
		Status:         toPgText(status),
		Priority:       toPgText(priority),
		Search:         toPgText(search),
		IsAdmin:        isSuperAdmin,
		UserID:         currentUserID,
	})
	if err != nil {
		return nil, err
	}

	data := make([]ProjectResponse, 0, len(rows))
	for _, r := range rows {
		var tp *string
		if r.TaskPrefix.Valid {
			tp = &r.TaskPrefix.String
		}
		var desc *string
		if r.Description.Valid {
			desc = &r.Description.String
		}
		var av *string
		if r.Avatar.Valid {
			av = &r.Avatar.String
		}
		var sDate, eDate *time.Time
		if r.StartDate.Valid {
			sDate = &r.StartDate.Time
		}
		if r.EndDate.Valid {
			eDate = &r.EndDate.Time
		}

		data = append(data, ProjectResponse{
			ID:            r.ID,
			Name:          r.Name,
			Slug:          r.Slug,
			TaskPrefix:    tp,
			Description:   desc,
			Avatar:        av,
			Color:         r.Color,
			Status:        string(r.Status),
			Priority:      string(r.Priority),
			Visibility:    string(r.Visibility),
			StartDate:     sDate,
			EndDate:       eDate,
			WorkflowID:    r.WorkflowID,
			WorkspaceID:   r.WorkspaceID,
			CreatedAt:     r.CreatedAt.Time,
			UpdatedAt:     r.UpdatedAt.Time,
			Archive:       r.Archive,
			WorkspaceName: r.WorkspaceName,
			WorkspaceSlug: r.WorkspaceSlug,
			TaskCount:     r.TaskCount,
			MemberCount:   r.MemberCount,
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	return &ProjectListResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *service) FindOne(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectResponse, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, apperror.NotFound("Project not found")
	}

	taskCount, _ := s.repo.CountTasks(ctx, id)
	memberCount, _ := s.repo.CountMembers(ctx, id)

	var tp *string
	if row.TaskPrefix.Valid {
		tp = &row.TaskPrefix.String
	}
	var desc *string
	if row.Description.Valid {
		desc = &row.Description.String
	}
	var av *string
	if row.Avatar.Valid {
		av = &row.Avatar.String
	}
	var sDate, eDate *time.Time
	if row.StartDate.Valid {
		sDate = &row.StartDate.Time
	}
	if row.EndDate.Valid {
		eDate = &row.EndDate.Time
	}
	var cID, uID *uuid.UUID
	if row.CreatedByID.Valid {
		idVal := uuid.UUID(row.CreatedByID.Bytes)
		cID = &idVal
	}
	if row.UpdatedByID.Valid {
		idVal := uuid.UUID(row.UpdatedByID.Bytes)
		uID = &idVal
	}
	orgID := row.OrganizationID

	return &ProjectResponse{
		ID:             row.ID,
		Name:           row.Name,
		Slug:           row.Slug,
		TaskPrefix:     tp,
		Description:    desc,
		Avatar:         av,
		Color:          row.Color,
		Status:         string(row.Status),
		Priority:       string(row.Priority),
		Visibility:     string(row.Visibility),
		StartDate:      sDate,
		EndDate:        eDate,
		Settings:       row.Settings,
		WorkflowID:     row.WorkflowID,
		WorkspaceID:    row.WorkspaceID,
		CreatedByID:    cID,
		UpdatedByID:    uID,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
		Archive:        row.Archive,
		WorkspaceName:  row.WorkspaceName,
		WorkspaceSlug:  row.WorkspaceSlug,
		OrganizationID: &orgID,
		TaskCount:      taskCount,
		MemberCount:    memberCount,
	}, nil
}

func (s *service) GetBySlug(ctx context.Context, slug string, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectResponse, error) {
	row, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, apperror.NotFound("Project not found")
	}

	taskCount, _ := s.repo.CountTasks(ctx, row.ID)
	memberCount, _ := s.repo.CountMembers(ctx, row.ID)

	var tp *string
	if row.TaskPrefix.Valid {
		tp = &row.TaskPrefix.String
	}
	var desc *string
	if row.Description.Valid {
		desc = &row.Description.String
	}
	var av *string
	if row.Avatar.Valid {
		av = &row.Avatar.String
	}
	var sDate, eDate *time.Time
	if row.StartDate.Valid {
		sDate = &row.StartDate.Time
	}
	if row.EndDate.Valid {
		eDate = &row.EndDate.Time
	}
	var cID, uID *uuid.UUID
	if row.CreatedByID.Valid {
		idVal := uuid.UUID(row.CreatedByID.Bytes)
		cID = &idVal
	}
	if row.UpdatedByID.Valid {
		idVal := uuid.UUID(row.UpdatedByID.Bytes)
		uID = &idVal
	}
	orgID := row.OrganizationID

	return &ProjectResponse{
		ID:             row.ID,
		Name:           row.Name,
		Slug:           row.Slug,
		TaskPrefix:     tp,
		Description:    desc,
		Avatar:         av,
		Color:          row.Color,
		Status:         string(row.Status),
		Priority:       string(row.Priority),
		Visibility:     string(row.Visibility),
		StartDate:      sDate,
		EndDate:        eDate,
		Settings:       row.Settings,
		WorkflowID:     row.WorkflowID,
		WorkspaceID:    row.WorkspaceID,
		CreatedByID:    cID,
		UpdatedByID:    uID,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
		Archive:        row.Archive,
		WorkspaceName:  row.WorkspaceName,
		WorkspaceSlug:  row.WorkspaceSlug,
		OrganizationID: &orgID,
		TaskCount:      taskCount,
		MemberCount:    memberCount,
	}, nil
}

func (s *service) FindByKey(ctx context.Context, workspaceID uuid.UUID, key string, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectResponse, error) {
	row, err := s.repo.GetByKey(ctx, workspaceID, key)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, apperror.NotFound("Project not found")
	}

	taskCount, _ := s.repo.CountTasks(ctx, row.ID)
	memberCount, _ := s.repo.CountMembers(ctx, row.ID)

	var tp *string
	if row.TaskPrefix.Valid {
		tp = &row.TaskPrefix.String
	}
	var desc *string
	if row.Description.Valid {
		desc = &row.Description.String
	}
	var av *string
	if row.Avatar.Valid {
		av = &row.Avatar.String
	}
	var sDate, eDate *time.Time
	if row.StartDate.Valid {
		sDate = &row.StartDate.Time
	}
	if row.EndDate.Valid {
		eDate = &row.EndDate.Time
	}
	var cID, uID *uuid.UUID
	if row.CreatedByID.Valid {
		idVal := uuid.UUID(row.CreatedByID.Bytes)
		cID = &idVal
	}
	if row.UpdatedByID.Valid {
		idVal := uuid.UUID(row.UpdatedByID.Bytes)
		uID = &idVal
	}
	orgID := row.OrganizationID

	return &ProjectResponse{
		ID:             row.ID,
		Name:           row.Name,
		Slug:           row.Slug,
		TaskPrefix:     tp,
		Description:    desc,
		Avatar:         av,
		Color:          row.Color,
		Status:         string(row.Status),
		Priority:       string(row.Priority),
		Visibility:     string(row.Visibility),
		StartDate:      sDate,
		EndDate:        eDate,
		Settings:       row.Settings,
		WorkflowID:     row.WorkflowID,
		WorkspaceID:    row.WorkspaceID,
		CreatedByID:    cID,
		UpdatedByID:    uID,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
		Archive:        row.Archive,
		WorkspaceName:  row.WorkspaceName,
		WorkspaceSlug:  row.WorkspaceSlug,
		OrganizationID: &orgID,
		TaskCount:      taskCount,
		MemberCount:    memberCount,
	}, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req UpdateProjectRequest, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectResponse, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, apperror.NotFound("Project not found")
	}

	if !isSuperAdmin {
		role, rErr := s.rbacResolver.GetProjectMemberRole(ctx, currentUserID, id)
		if rErr != nil || (role != middleware.RoleOwner && role != middleware.RoleManager) {
			wsRole, wErr := s.rbacResolver.GetWorkspaceMemberRole(ctx, currentUserID, existing.WorkspaceID)
			if wErr != nil || (wsRole != middleware.RoleOwner && wsRole != middleware.RoleManager) {
				return nil, apperror.Forbidden("Forbidden: insufficient role to update project")
			}
		}
	}

	var statusVal pgtype.Text
	if req.Status != nil && *req.Status != "" {
		statusVal = pgtype.Text{String: strings.ToUpper(*req.Status), Valid: true}
	}
	var priorityVal pgtype.Text
	if req.Priority != nil && *req.Priority != "" {
		priorityVal = pgtype.Text{String: strings.ToUpper(*req.Priority), Valid: true}
	}
	var visibilityVal pgtype.Text
	if req.Visibility != nil && *req.Visibility != "" {
		visibilityVal = pgtype.Text{String: strings.ToUpper(*req.Visibility), Valid: true}
	}

	_, err = s.repo.Update(ctx, sqlc.UpdateProjectParams{
		ID:          id,
		Name:        toPgText(req.Name),
		Slug:        toPgText(req.Slug),
		TaskPrefix:  toPgText(req.TaskPrefix),
		Description: toPgText(req.Description),
		Avatar:      toPgText(req.Avatar),
		Color:       toPgText(req.Color),
		Status:      sqlc.NullProjectStatus{ProjectStatus: sqlc.ProjectStatus(statusVal.String), Valid: statusVal.Valid},
		Priority:    sqlc.NullProjectPriority{ProjectPriority: sqlc.ProjectPriority(priorityVal.String), Valid: priorityVal.Valid},
		Visibility:  sqlc.NullProjectVisibility{ProjectVisibility: sqlc.ProjectVisibility(visibilityVal.String), Valid: visibilityVal.Valid},
		StartDate:   toPgTimestampPtr(req.StartDate),
		EndDate:     toPgTimestampPtr(req.EndDate),
		Settings:    req.Settings,
		WorkflowID:  toPgUUID(req.WorkflowID),
		UpdatedByID: pgtype.UUID{Bytes: currentUserID, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return s.FindOne(ctx, id, currentUserID, isSuperAdmin)
}

func (s *service) Archive(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NotFound("Project not found")
	}

	if !isSuperAdmin {
		role, rErr := s.rbacResolver.GetProjectMemberRole(ctx, currentUserID, id)
		if rErr != nil || (role != middleware.RoleOwner && role != middleware.RoleManager) {
			return apperror.Forbidden("Forbidden: insufficient permissions to archive project")
		}
	}

	return s.repo.Archive(ctx, id, currentUserID)
}

func (s *service) Unarchive(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NotFound("Project not found")
	}

	ws, err := s.repo.GetWorkspaceByID(ctx, existing.WorkspaceID)
	if err != nil {
		return err
	}
	if ws != nil && ws.Archive {
		return apperror.Forbidden("Parent workspace is still archived")
	}

	if !isSuperAdmin {
		role, rErr := s.rbacResolver.GetProjectMemberRole(ctx, currentUserID, id)
		if rErr != nil || (role != middleware.RoleOwner && role != middleware.RoleManager) {
			return apperror.Forbidden("Forbidden: insufficient permissions to unarchive project")
		}
	}

	return s.repo.Unarchive(ctx, id, currentUserID)
}

func (s *service) Delete(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NotFound("Project not found")
	}

	if !isSuperAdmin {
		role, rErr := s.rbacResolver.GetProjectMemberRole(ctx, currentUserID, id)
		if rErr != nil || role != middleware.RoleOwner {
			return apperror.Forbidden("Forbidden: only project OWNER can delete project")
		}
	}

	return s.repo.Delete(ctx, id)
}

func (s *service) GetStats(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectStatsResponse, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, apperror.NotFound("Project not found")
	}

	totalTasks, err := s.repo.CountTasks(ctx, id)
	if err != nil {
		return nil, err
	}
	completedTasks, err := s.repo.CountCompletedTasks(ctx, id)
	if err != nil {
		return nil, err
	}
	memberCount, err := s.repo.CountMembers(ctx, id)
	if err != nil {
		return nil, err
	}

	rate := float64(0)
	if totalTasks > 0 {
		rate = math.Round((float64(completedTasks)/float64(totalTasks)*100)*100) / 100
	}

	return &ProjectStatsResponse{
		TotalTasks:     totalTasks,
		CompletedTasks: completedTasks,
		CompletionRate: rate,
		MemberCount:    memberCount,
	}, nil
}

func (s *service) GetCharts(ctx context.Context, slug string, currentUserID uuid.UUID, isSuperAdmin bool) (map[string]any, error) {
	proj, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if proj == nil {
		return nil, apperror.NotFound("Project not found")
	}

	totalTasks, _ := s.repo.CountTasks(ctx, proj.ID)
	completedTasks, _ := s.repo.CountCompletedTasks(ctx, proj.ID)
	rate := float64(0)
	if totalTasks > 0 {
		rate = math.Round((float64(completedTasks)/float64(totalTasks)*100)*100) / 100
	}

	statusRows, _ := s.repo.TaskStatusStats(ctx, proj.ID)
	statusData := make([]map[string]any, 0, len(statusRows))
	for _, sr := range statusRows {
		statusData = append(statusData, map[string]any{
			"statusId":   sr.StatusID,
			"statusName": sr.StatusName,
			"category":   string(sr.Category),
			"count":      sr.TaskCount,
		})
	}

	return map[string]any{
		"kpi-metrics": map[string]any{
			"totalTasks":     totalTasks,
			"completedTasks": completedTasks,
			"completionRate": rate,
		},
		"task-status": statusData,
	}, nil
}

// Helpers

func toPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func toPgTimestamp(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: t, Valid: true}
}

func toPgTimestampPtr(t *time.Time) pgtype.Timestamp {
	if t == nil {
		return pgtype.Timestamp{Valid: false}
	}
	return pgtype.Timestamp{Time: *t, Valid: true}
}
