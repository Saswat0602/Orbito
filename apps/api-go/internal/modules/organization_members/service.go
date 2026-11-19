package organization_members

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// Service defines business operations on organization members.
type Service interface {
	AddMember(ctx context.Context, req AddMemberRequest, currentUserID uuid.UUID) (*OrganizationMemberResponse, error)
	FindAll(ctx context.Context, orgID *uuid.UUID, search string, currentUserID uuid.UUID) ([]OrganizationMemberResponse, error)
	FindAllBySlug(ctx context.Context, slug string, page, limit int, search string, currentUserID uuid.UUID) (*PaginatedMembersResponse, error)
	FindOne(ctx context.Context, id, currentUserID uuid.UUID) (*OrganizationMemberResponse, error)
	FindByUserAndOrg(ctx context.Context, userID, orgID, currentUserID uuid.UUID) (*OrganizationMemberResponse, error)
	UpdateRole(ctx context.Context, id uuid.UUID, newRole sqlc.Role, currentUserID uuid.UUID) (*OrganizationMemberResponse, error)
	Remove(ctx context.Context, id, currentUserID uuid.UUID) error
	SetDefaultOrganization(ctx context.Context, orgID, currentUserID uuid.UUID) error
}

type service struct {
	repo Repository
}

// NewService constructs an organization members domain service.
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

func (s *service) AddMember(ctx context.Context, req AddMemberRequest, currentUserID uuid.UUID) (*OrganizationMemberResponse, error) {
	existing, err := s.repo.GetByUserAndOrg(ctx, req.UserID, req.OrganizationID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperror.Conflict("Member already exists")
	}

	globalRole, _ := s.repo.GetUserGlobalRole(ctx, currentUserID)
	if globalRole != sqlc.RoleSUPERADMIN {
		requester, rErr := s.repo.GetByUserAndOrg(ctx, currentUserID, req.OrganizationID)
		if rErr != nil || requester == nil {
			return nil, apperror.Forbidden("You are not a member of this organization")
		}
		if middleware.RoleRank[middleware.Role(requester.Role)] < middleware.RoleRank[middleware.RoleManager] {
			return nil, apperror.Forbidden("Only organization owners and admins can add members")
		}
		if req.Role == sqlc.RoleOWNER && requester.Role != sqlc.RoleOWNER {
			return nil, apperror.Forbidden("Only organization owners can assign the OWNER role")
		}
	}

	m, err := s.repo.Create(ctx, sqlc.CreateOrganizationMemberParams{
		ID:             uuid.New(),
		Role:           req.Role,
		UserID:         req.UserID,
		OrganizationID: req.OrganizationID,
		CreatedByID:    pgtype.UUID{Bytes: currentUserID, Valid: true},
		UpdatedByID:    pgtype.UUID{Bytes: currentUserID, Valid: true},
		IsDefault:      false,
	})
	if err != nil {
		return nil, err
	}

	return s.FindOne(ctx, m.ID, currentUserID)
}

func (s *service) FindAll(ctx context.Context, orgID *uuid.UUID, search string, _ uuid.UUID) ([]OrganizationMemberResponse, error) {
	if orgID == nil || *orgID == uuid.Nil {
		return []OrganizationMemberResponse{}, nil
	}

	rows, err := s.repo.List(ctx, *orgID, search)
	if err != nil {
		return nil, err
	}

	res := make([]OrganizationMemberResponse, 0, len(rows))
	for _, r := range rows {
		res = append(res, OrganizationMemberResponse{
			ID:             r.ID,
			Role:           string(r.Role),
			JoinedAt:       tsToTime(r.JoinedAt),
			UserID:         r.UserID,
			OrganizationID: r.OrganizationID,
			IsDefault:      r.IsDefault,
			User: &MemberUserDTO{
				ID:        r.UserUserID,
				Email:     r.UserEmail,
				FirstName: r.UserFirstName,
				LastName:  r.UserLastName,
				Avatar:    textToPtr(r.UserAvatar),
				Status:    string(r.UserStatus),
			},
			Organization: &MemberOrgDTO{
				ID:      r.OrgID,
				Name:    r.OrgName,
				Slug:    r.OrgSlug,
				OwnerID: r.OrgOwnerID,
			},
			CreatedAt: tsToTime(r.CreatedAt),
			UpdatedAt: tsToTime(r.UpdatedAt),
		})
	}

	return res, nil
}

func (s *service) FindAllBySlug(ctx context.Context, slug string, page, limit int, search string, _ uuid.UUID) (*PaginatedMembersResponse, error) {
	var p int32 = 1
	if page > 1 && page <= 1000000 {
		p = int32(page)
	}
	var l int32 = 20
	if limit >= 1 && limit <= 100 {
		l = int32(limit)
	}
	offset := (p - 1) * l

	total, err := s.repo.CountBySlug(ctx, slug, search)
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.ListBySlug(ctx, slug, search, l, offset)
	if err != nil {
		return nil, err
	}

	items := make([]OrganizationMemberResponse, 0, len(rows))
	for _, r := range rows {
		items = append(items, OrganizationMemberResponse{
			ID:             r.ID,
			Role:           string(r.Role),
			JoinedAt:       tsToTime(r.JoinedAt),
			UserID:         r.UserID,
			OrganizationID: r.OrganizationID,
			IsDefault:      r.IsDefault,
			User: &MemberUserDTO{
				ID:        r.UserUserID,
				Email:     r.UserEmail,
				FirstName: r.UserFirstName,
				LastName:  r.UserLastName,
				Avatar:    textToPtr(r.UserAvatar),
				Status:    string(r.UserStatus),
			},
			Organization: &MemberOrgDTO{
				ID:      r.OrgID,
				Name:    r.OrgName,
				Slug:    r.OrgSlug,
				OwnerID: r.OrgOwnerID,
			},
			CreatedAt: tsToTime(r.CreatedAt),
			UpdatedAt: tsToTime(r.UpdatedAt),
		})
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))
	return &PaginatedMembersResponse{
		Items:      items,
		TotalItems: total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

func (s *service) FindOne(ctx context.Context, id, _ uuid.UUID) (*OrganizationMemberResponse, error) {
	r, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, apperror.NotFound("Organization member not found")
	}

	return &OrganizationMemberResponse{
		ID:             r.ID,
		Role:           string(r.Role),
		JoinedAt:       tsToTime(r.JoinedAt),
		UserID:         r.UserID,
		OrganizationID: r.OrganizationID,
		IsDefault:      r.IsDefault,
		User: &MemberUserDTO{
			ID:        r.UserUserID,
			Email:     r.UserEmail,
			FirstName: r.UserFirstName,
			LastName:  r.UserLastName,
			Avatar:    textToPtr(r.UserAvatar),
			Status:    string(r.UserStatus),
		},
		Organization: &MemberOrgDTO{
			ID:      r.OrgID,
			Name:    r.OrgName,
			Slug:    r.OrgSlug,
			OwnerID: r.OrgOwnerID,
		},
		CreatedAt: tsToTime(r.CreatedAt),
		UpdatedAt: tsToTime(r.UpdatedAt),
	}, nil
}

func (s *service) FindByUserAndOrg(ctx context.Context, userID, orgID, _ uuid.UUID) (*OrganizationMemberResponse, error) {
	r, err := s.repo.GetByUserAndOrg(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, apperror.NotFound("Membership not found")
	}

	return &OrganizationMemberResponse{
		ID:             r.ID,
		Role:           string(r.Role),
		JoinedAt:       tsToTime(r.JoinedAt),
		UserID:         r.UserID,
		OrganizationID: r.OrganizationID,
		IsDefault:      r.IsDefault,
		User: &MemberUserDTO{
			ID:        r.UserUserID,
			Email:     r.UserEmail,
			FirstName: r.UserFirstName,
			LastName:  r.UserLastName,
			Avatar:    textToPtr(r.UserAvatar),
			Status:    string(r.UserStatus),
		},
		Organization: &MemberOrgDTO{
			ID:      r.OrgID,
			Name:    r.OrgName,
			Slug:    r.OrgSlug,
			OwnerID: r.OrgOwnerID,
		},
		CreatedAt: tsToTime(r.CreatedAt),
		UpdatedAt: tsToTime(r.UpdatedAt),
	}, nil
}

func (s *service) UpdateRole(ctx context.Context, id uuid.UUID, newRole sqlc.Role, currentUserID uuid.UUID) (*OrganizationMemberResponse, error) {
	member, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, apperror.NotFound("Organization member not found")
	}

	globalRole, _ := s.repo.GetUserGlobalRole(ctx, currentUserID)
	if globalRole != sqlc.RoleSUPERADMIN {
		requester, rErr := s.repo.GetByUserAndOrg(ctx, currentUserID, member.OrganizationID)
		if rErr != nil || requester == nil {
			return nil, apperror.Forbidden("You are not a member of this organization")
		}

		isOwner := member.OrgOwnerID == currentUserID
		isAdmin := middleware.RoleRank[middleware.Role(requester.Role)] >= middleware.RoleRank[middleware.RoleManager]

		if !isOwner && !isAdmin {
			return nil, apperror.Forbidden("Only organization owners and admins can update member roles")
		}

		if newRole == sqlc.RoleOWNER && !isOwner && requester.Role != sqlc.RoleOWNER {
			return nil, apperror.Forbidden("Only organization owners can assign the OWNER role")
		}
	}

	if member.OrgOwnerID == member.UserID && newRole != sqlc.RoleOWNER {
		return nil, apperror.BadRequest("Cannot change the role of organization owner")
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
		return apperror.NotFound("Organization member not found")
	}

	if member.OrgOwnerID == member.UserID {
		return apperror.BadRequest("Cannot remove organization owner from organization")
	}

	globalRole, _ := s.repo.GetUserGlobalRole(ctx, currentUserID)
	if globalRole != sqlc.RoleSUPERADMIN {
		requester, rErr := s.repo.GetByUserAndOrg(ctx, currentUserID, member.OrganizationID)
		if rErr != nil || requester == nil {
			return apperror.Forbidden("You are not a member of this organization")
		}

		isSelf := member.UserID == currentUserID
		isOwner := member.OrgOwnerID == currentUserID
		isAdmin := middleware.RoleRank[middleware.Role(requester.Role)] >= middleware.RoleRank[middleware.RoleManager]

		if !isSelf && !isOwner && !isAdmin {
			return apperror.Forbidden("You can only remove yourself or you must be an admin/owner")
		}
	}

	return s.repo.WithTx(ctx, func(txRepo Repository) error {
		_ = txRepo.ClearDefaultOrg(ctx, member.UserID, member.OrganizationID)
		return txRepo.Delete(ctx, id)
	})
}

func (s *service) SetDefaultOrganization(ctx context.Context, orgID, currentUserID uuid.UUID) error {
	membership, err := s.repo.GetByUserAndOrg(ctx, currentUserID, orgID)
	if err != nil {
		return err
	}
	if membership == nil {
		return apperror.Forbidden("You are not a member of this organization")
	}

	return s.repo.SetDefaultOrg(ctx, currentUserID, orgID)
}
