package workspace_members

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// Service defines business operations for workspace members.
type Service interface {
	AddMember(ctx context.Context, req AddWorkspaceMemberRequest, currentUserID uuid.UUID) (*WorkspaceMemberResponse, error)
	FindAll(ctx context.Context, wsID *uuid.UUID, search string, page, limit int, currentUserID uuid.UUID) (*PaginatedWorkspaceMembersResponse, error)
	FindOne(ctx context.Context, id, currentUserID uuid.UUID) (*WorkspaceMemberResponse, error)
	FindByUserAndWorkspace(ctx context.Context, userID, wsID, currentUserID uuid.UUID) (*WorkspaceMemberResponse, error)
	UpdateRole(ctx context.Context, id uuid.UUID, newRole sqlc.Role, currentUserID uuid.UUID) (*WorkspaceMemberResponse, error)
	Remove(ctx context.Context, id, currentUserID uuid.UUID) error
}

type service struct {
	repo Repository
}

// NewService creates a new workspace members service instance.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func textToPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func tsToTime(t pgtype.Timestamp) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func (s *service) AddMember(ctx context.Context, req AddWorkspaceMemberRequest, currentUserID uuid.UUID) (*WorkspaceMemberResponse, error) {
	existing, err := s.repo.GetByUserAndWorkspace(ctx, req.UserID, req.WorkspaceID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperror.Conflict("Member already exists")
	}

	globalRole, _ := s.repo.GetUserGlobalRole(ctx, currentUserID)
	if globalRole != sqlc.RoleSUPERADMIN {
		requester, rErr := s.repo.GetByUserAndWorkspace(ctx, currentUserID, req.WorkspaceID)
		if rErr != nil || requester == nil {
			return nil, apperror.Forbidden("You are not a member of this workspace")
		}
		if middleware.RoleRank[middleware.Role(requester.Role)] < middleware.RoleRank[middleware.RoleManager] {
			return nil, apperror.Forbidden("Only workspace owners and admins can add members")
		}
		if req.Role == sqlc.RoleOWNER && requester.Role != sqlc.RoleOWNER {
			return nil, apperror.Forbidden("Only workspace owners can assign the OWNER role")
		}
	}

	m, err := s.repo.Create(ctx, sqlc.CreateWorkspaceMemberDirectParams{
		ID:          uuid.New(),
		Role:        req.Role,
		UserID:      req.UserID,
		WorkspaceID: req.WorkspaceID,
		CreatedByID: pgtype.UUID{Bytes: currentUserID, Valid: true},
		UpdatedByID: pgtype.UUID{Bytes: currentUserID, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return s.FindOne(ctx, m.ID, currentUserID)
}

func (s *service) FindAll(ctx context.Context, wsID *uuid.UUID, search string, page, limit int, _ uuid.UUID) (*PaginatedWorkspaceMembersResponse, error) {
	var p int32 = 1
	if page > 1 && page <= 1000000 {
		p = int32(page)
	}
	var l int32 = 20
	if limit >= 1 && limit <= 100 {
		l = int32(limit)
	}
	offset := (p - 1) * l

	total, err := s.repo.Count(ctx, wsID, search)
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.List(ctx, wsID, search, l, offset)
	if err != nil {
		return nil, err
	}

	items := make([]WorkspaceMemberResponse, 0, len(rows))
	for _, r := range rows {
		items = append(items, WorkspaceMemberResponse{
			ID:          r.ID,
			Role:        string(r.Role),
			JoinedAt:    tsToTime(r.JoinedAt),
			UserID:      r.UserID,
			WorkspaceID: r.WorkspaceID,
			User: &MemberUserDTO{
				ID:        r.UserUserID,
				Email:     r.UserEmail,
				FirstName: r.UserFirstName,
				LastName:  r.UserLastName,
				Avatar:    textToPtr(r.UserAvatar),
				Status:    string(r.UserStatus),
			},
			Workspace: &MemberWorkspaceDTO{
				ID:             r.WsID,
				Name:           r.WsName,
				Slug:           r.WsSlug,
				OrganizationID: r.WsOrgID,
			},
			CreatedAt: tsToTime(r.CreatedAt),
			UpdatedAt: tsToTime(r.UpdatedAt),
		})
	}

	totalPages := int((total + int64(l) - 1) / int64(l))
	return &PaginatedWorkspaceMembersResponse{
		Items:      items,
		TotalItems: total,
		Page:       int(p),
		Limit:      int(l),
		TotalPages: totalPages,
	}, nil
}

func (s *service) FindOne(ctx context.Context, id, _ uuid.UUID) (*WorkspaceMemberResponse, error) {
	r, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, apperror.NotFound("Workspace member not found")
	}

	return &WorkspaceMemberResponse{
		ID:          r.ID,
		Role:        string(r.Role),
		JoinedAt:    tsToTime(r.JoinedAt),
		UserID:      r.UserID,
		WorkspaceID: r.WorkspaceID,
		User: &MemberUserDTO{
			ID:        r.UserUserID,
			Email:     r.UserEmail,
			FirstName: r.UserFirstName,
			LastName:  r.UserLastName,
			Avatar:    textToPtr(r.UserAvatar),
			Status:    string(r.UserStatus),
		},
		Workspace: &MemberWorkspaceDTO{
			ID:             r.WsID,
			Name:           r.WsName,
			Slug:           r.WsSlug,
			OrganizationID: r.WsOrgID,
		},
		CreatedAt: tsToTime(r.CreatedAt),
		UpdatedAt: tsToTime(r.UpdatedAt),
	}, nil
}

func (s *service) FindByUserAndWorkspace(ctx context.Context, userID, wsID, _ uuid.UUID) (*WorkspaceMemberResponse, error) {
	r, err := s.repo.GetByUserAndWorkspace(ctx, userID, wsID)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, apperror.NotFound("Membership not found")
	}

	return &WorkspaceMemberResponse{
		ID:          r.ID,
		Role:        string(r.Role),
		JoinedAt:    tsToTime(r.JoinedAt),
		UserID:      r.UserID,
		WorkspaceID: r.WorkspaceID,
		User: &MemberUserDTO{
			ID:        r.UserUserID,
			Email:     r.UserEmail,
			FirstName: r.UserFirstName,
			LastName:  r.UserLastName,
			Avatar:    textToPtr(r.UserAvatar),
			Status:    string(r.UserStatus),
		},
		Workspace: &MemberWorkspaceDTO{
			ID:             r.WsID,
			Name:           r.WsName,
			Slug:           r.WsSlug,
			OrganizationID: r.WsOrgID,
		},
		CreatedAt: tsToTime(r.CreatedAt),
		UpdatedAt: tsToTime(r.UpdatedAt),
	}, nil
}

func (s *service) UpdateRole(ctx context.Context, id uuid.UUID, newRole sqlc.Role, currentUserID uuid.UUID) (*WorkspaceMemberResponse, error) {
	member, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, apperror.NotFound("Workspace member not found")
	}

	globalRole, _ := s.repo.GetUserGlobalRole(ctx, currentUserID)
	if globalRole != sqlc.RoleSUPERADMIN {
		requester, rErr := s.repo.GetByUserAndWorkspace(ctx, currentUserID, member.WorkspaceID)
		if rErr != nil || requester == nil {
			return nil, apperror.Forbidden("You are not a member of this workspace")
		}

		isAdmin := middleware.RoleRank[middleware.Role(requester.Role)] >= middleware.RoleRank[middleware.RoleManager]
		if !isAdmin {
			return nil, apperror.Forbidden("Only workspace owners and admins can update member roles")
		}

		if newRole == sqlc.RoleOWNER && requester.Role != sqlc.RoleOWNER {
			return nil, apperror.Forbidden("Only workspace owners can assign the OWNER role")
		}
	}

	_, err = s.repo.UpdateRole(ctx, id, newRole, currentUserID)
	if err != nil {
		return nil, err
	}

	return s.FindOne(ctx, id, currentUserID)
}

func (s *service) Remove(ctx context.Context, id, currentUserID uuid.UUID) error {
	member, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if member == nil {
		return apperror.NotFound("Workspace member not found")
	}

	globalRole, _ := s.repo.GetUserGlobalRole(ctx, currentUserID)
	if globalRole != sqlc.RoleSUPERADMIN {
		requester, rErr := s.repo.GetByUserAndWorkspace(ctx, currentUserID, member.WorkspaceID)
		if rErr != nil || requester == nil {
			return apperror.Forbidden("You are not a member of this workspace")
		}

		isSelf := member.UserID == currentUserID
		isAdmin := middleware.RoleRank[middleware.Role(requester.Role)] >= middleware.RoleRank[middleware.RoleManager]

		if !isSelf && !isAdmin {
			return apperror.Forbidden("You can only remove yourself or you must be an admin/owner")
		}
	}

	return s.repo.Delete(ctx, id)
}
