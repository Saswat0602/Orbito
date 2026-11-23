package project_members

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// Service defines business operations for project members.
type Service interface {
	Create(ctx context.Context, req CreateProjectMemberRequest, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectMemberResponse, error)
	ListMembers(ctx context.Context, projectID uuid.UUID, search *string, currentUserID uuid.UUID, isSuperAdmin bool) ([]ProjectMemberResponse, error)
	FindOne(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectMemberResponse, error)
	FindByUserAndProject(ctx context.Context, targetUserID, projectID uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectMemberResponse, error)
	UpdateRole(ctx context.Context, id uuid.UUID, req UpdateProjectMemberRoleRequest, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectMemberResponse, error)
	Remove(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) error
	GetStats(ctx context.Context, projectID uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectMemberStatsResponse, error)
	ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) ([]map[string]any, error)
	ListUserProjects(ctx context.Context, targetUserID uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) ([]map[string]any, error)
}

type service struct {
	repo         Repository
	rbacResolver middleware.RBACResolver
}

// NewService constructs a project members service instance.
func NewService(repo Repository, rbacResolver middleware.RBACResolver) Service {
	return &service{
		repo:         repo,
		rbacResolver: rbacResolver,
	}
}

func parseMemberRole(roleStr *string) (sqlc.Role, error) {
	if roleStr == nil || strings.TrimSpace(*roleStr) == "" {
		return sqlc.RoleMEMBER, nil
	}
	r := strings.ToUpper(strings.TrimSpace(*roleStr))
	switch r {
	case string(middleware.RoleOwner):
		return sqlc.RoleOWNER, nil
	case string(middleware.RoleManager):
		return sqlc.RoleMANAGER, nil
	case string(middleware.RoleMember):
		return sqlc.RoleMEMBER, nil
	case string(middleware.RoleViewer):
		return sqlc.RoleVIEWER, nil
	default:
		return "", apperror.BadRequest(fmt.Sprintf("Invalid project member role: %s", *roleStr))
	}
}

func (s *service) checkElevatedAuthority(ctx context.Context, projectID, userID uuid.UUID, isSuperAdmin bool) error {
	if isSuperAdmin {
		return nil
	}

	role, err := s.rbacResolver.GetProjectMemberRole(ctx, userID, projectID)
	if err == nil && (role == middleware.RoleOwner || role == middleware.RoleManager) {
		return nil
	}

	proj, pErr := s.repo.GetProjectByID(ctx, projectID)
	if pErr == nil && proj != nil {
		wsRole, wErr := s.rbacResolver.GetWorkspaceMemberRole(ctx, userID, proj.WorkspaceID)
		if wErr == nil && (wsRole == middleware.RoleOwner || wsRole == middleware.RoleManager) {
			return nil
		}
	}

	return apperror.Forbidden("Forbidden: insufficient role to manage project members")
}

func (s *service) Create(ctx context.Context, req CreateProjectMemberRequest, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectMemberResponse, error) {
	proj, err := s.repo.GetProjectByID(ctx, req.ProjectID)
	if err != nil {
		return nil, err
	}
	if proj == nil {
		return nil, apperror.NotFound("Project not found")
	}

	if authErr := s.checkElevatedAuthority(ctx, req.ProjectID, currentUserID, isSuperAdmin); authErr != nil {
		return nil, authErr
	}

	assignedRole, err := parseMemberRole(req.Role)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, apperror.NotFound("User not found")
	}

	existing, err := s.repo.GetByUserAndProject(ctx, req.UserID, req.ProjectID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperror.Conflict("Member already exists in this project")
	}

	memberID := uuid.New()
	_, err = s.repo.Create(ctx, sqlc.CreateProjectMemberRecordParams{
		ID:          memberID,
		UserID:      req.UserID,
		ProjectID:   req.ProjectID,
		Role:        assignedRole,
		CreatedByID: pgtype.UUID{Bytes: currentUserID, Valid: true},
		UpdatedByID: pgtype.UUID{Bytes: currentUserID, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return s.FindOne(ctx, memberID, currentUserID, isSuperAdmin)
}

func (s *service) ListMembers(ctx context.Context, projectID uuid.UUID, search *string, currentUserID uuid.UUID, isSuperAdmin bool) ([]ProjectMemberResponse, error) {
	proj, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if proj == nil {
		return nil, apperror.NotFound("Project not found")
	}

	rows, err := s.repo.ListMembers(ctx, projectID, search)
	if err != nil {
		return nil, err
	}

	res := make([]ProjectMemberResponse, 0, len(rows))
	for _, r := range rows {
		var av, un *string
		if r.UserAvatar.Valid {
			av = &r.UserAvatar.String
		}
		if r.UserUsername.Valid {
			un = &r.UserUsername.String
		}

		res = append(res, ProjectMemberResponse{
			ID:          r.ID,
			Role:        string(r.Role),
			JoinedAt:    r.JoinedAt.Time,
			UserID:      r.UserID,
			ProjectID:   r.ProjectID,
			ProjectName: r.ProjectName,
			ProjectSlug: r.ProjectSlug,
			CreatedAt:   r.CreatedAt.Time,
			UpdatedAt:   r.UpdatedAt.Time,
			User: ProjectMemberUserProfile{
				ID:        r.MemberUserID,
				Email:     r.UserEmail,
				FirstName: r.UserFirstName,
				LastName:  r.UserLastName,
				Avatar:    av,
				Username:  un,
			},
		})
	}

	return res, nil
}

func (s *service) FindOne(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectMemberResponse, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, apperror.NotFound("Project member not found")
	}

	var av, un *string
	if row.UserAvatar.Valid {
		av = &row.UserAvatar.String
	}
	if row.UserUsername.Valid {
		un = &row.UserUsername.String
	}

	return &ProjectMemberResponse{
		ID:          row.ID,
		Role:        string(row.Role),
		JoinedAt:    row.JoinedAt.Time,
		UserID:      row.UserID,
		ProjectID:   row.ProjectID,
		ProjectName: row.ProjectName,
		ProjectSlug: row.ProjectSlug,
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
		User: ProjectMemberUserProfile{
			ID:        row.MemberUserID,
			Email:     row.UserEmail,
			FirstName: row.UserFirstName,
			LastName:  row.UserLastName,
			Avatar:    av,
			Username:  un,
		},
	}, nil
}

func (s *service) FindByUserAndProject(ctx context.Context, targetUserID, projectID uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectMemberResponse, error) {
	row, err := s.repo.GetByUserAndProject(ctx, targetUserID, projectID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, apperror.NotFound("Membership not found")
	}

	var av, un *string
	if row.UserAvatar.Valid {
		av = &row.UserAvatar.String
	}
	if row.UserUsername.Valid {
		un = &row.UserUsername.String
	}

	return &ProjectMemberResponse{
		ID:        row.ID,
		Role:      string(row.Role),
		JoinedAt:  row.JoinedAt.Time,
		UserID:    row.UserID,
		ProjectID: row.ProjectID,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
		User: ProjectMemberUserProfile{
			ID:        row.MemberUserID,
			Email:     row.UserEmail,
			FirstName: row.UserFirstName,
			LastName:  row.UserLastName,
			Avatar:    av,
			Username:  un,
		},
	}, nil
}

func (s *service) UpdateRole(ctx context.Context, id uuid.UUID, req UpdateProjectMemberRoleRequest, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectMemberResponse, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, apperror.NotFound("Project member not found")
	}

	if err := s.checkElevatedAuthority(ctx, existing.ProjectID, currentUserID, isSuperAdmin); err != nil {
		return nil, err
	}

	newRole, err := parseMemberRole(&req.Role)
	if err != nil {
		return nil, err
	}

	if existing.Role == sqlc.RoleOWNER && newRole != sqlc.RoleOWNER {
		ownersCount, cErr := s.repo.CountOwners(ctx, existing.ProjectID)
		if cErr != nil {
			return nil, cErr
		}
		if ownersCount <= 1 {
			return nil, apperror.BadRequest("Cannot demote the sole project owner")
		}
	}

	_, err = s.repo.UpdateRole(ctx, id, newRole, currentUserID)
	if err != nil {
		return nil, err
	}

	return s.FindOne(ctx, id, currentUserID, isSuperAdmin)
}

func (s *service) Remove(ctx context.Context, id uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NotFound("Project member not found")
	}

	isSelf := existing.UserID == currentUserID
	if !isSelf && !isSuperAdmin {
		if authErr := s.checkElevatedAuthority(ctx, existing.ProjectID, currentUserID, isSuperAdmin); authErr != nil {
			return authErr
		}
	}

	if existing.Role == sqlc.RoleOWNER {
		ownersCount, cErr := s.repo.CountOwners(ctx, existing.ProjectID)
		if cErr != nil {
			return cErr
		}
		if ownersCount <= 1 {
			return apperror.BadRequest("Cannot remove the sole project owner")
		}
	}

	return s.repo.Delete(ctx, id)
}

func (s *service) GetStats(ctx context.Context, projectID uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) (*ProjectMemberStatsResponse, error) {
	proj, err := s.repo.GetProjectByID(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if proj == nil {
		return nil, apperror.NotFound("Project not found")
	}

	roleStats, err := s.repo.RoleStats(ctx, projectID)
	if err != nil {
		return nil, err
	}

	var total int64
	byRole := make(map[string]int64)
	for _, rs := range roleStats {
		byRole[string(rs.Role)] = rs.Count
		total += rs.Count
	}

	return &ProjectMemberStatsResponse{
		TotalMembers: total,
		ByRole:       byRole,
	}, nil
}

func (s *service) ListByWorkspace(ctx context.Context, workspaceID uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) ([]map[string]any, error) {
	rows, err := s.repo.ListByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, err
	}

	res := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		res = append(res, map[string]any{
			"id":          r.ID,
			"role":        string(r.Role),
			"joinedAt":    r.JoinedAt.Time,
			"userId":      r.UserID,
			"projectId":   r.ProjectID,
			"projectName": r.ProjectName,
			"projectSlug": r.ProjectSlug,
			"email":       r.UserEmail,
			"firstName":   r.UserFirstName,
			"lastName":    r.UserLastName,
		})
	}
	return res, nil
}

func (s *service) ListUserProjects(ctx context.Context, targetUserID uuid.UUID, currentUserID uuid.UUID, isSuperAdmin bool) ([]map[string]any, error) {
	rows, err := s.repo.ListUserProjects(ctx, targetUserID)
	if err != nil {
		return nil, err
	}

	res := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		var tp *string
		if r.TaskPrefix.Valid {
			tp = &r.TaskPrefix.String
		}
		var av *string
		if r.Avatar.Valid {
			av = &r.Avatar.String
		}

		res = append(res, map[string]any{
			"id":         r.ID,
			"name":       r.Name,
			"slug":       r.Slug,
			"taskPrefix": tp,
			"avatar":     av,
			"color":      r.Color,
			"status":     string(r.Status),
			"priority":   string(r.Priority),
			"role":       string(r.MemberRole),
			"joinedAt":   r.JoinedAt.Time,
		})
	}
	return res, nil
}
