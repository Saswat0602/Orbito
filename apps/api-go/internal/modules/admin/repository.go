package admin

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository abstracts database operations for the admin module.
type Repository interface {
	GetDashboardStats(ctx context.Context, since time.Time) (*sqlc.GetDashboardStatsCountsRow, error)
	ListUsers(ctx context.Context, search string, status sqlc.UserStatus, role sqlc.Role, limit, offset int32) ([]sqlc.AdminListUsersRow, error)
	CountUsers(ctx context.Context, search string, status sqlc.UserStatus, role sqlc.Role) (int64, error)
	GetUserDetail(ctx context.Context, id uuid.UUID) (*sqlc.AdminGetUserDetailRow, error)
	GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]sqlc.AdminGetUserOrganizationsRow, error)
	CountSuperAdmins(ctx context.Context) (int64, error)
	UpdateUserRole(ctx context.Context, id uuid.UUID, role sqlc.Role) (*sqlc.AdminUpdateUserRoleRow, error)
	UpdateUserStatus(ctx context.Context, id uuid.UUID, status sqlc.UserStatus) (*sqlc.AdminUpdateUserStatusRow, error)
	SoftDeleteUser(ctx context.Context, id, deletedBy uuid.UUID) error
	RemoveUserMemberships(ctx context.Context, userID uuid.UUID) error
	GetUserOwnedOrganizations(ctx context.Context, ownerID uuid.UUID) ([]sqlc.AdminGetUserOwnedOrganizationsRow, error)
	SetUserResetToken(ctx context.Context, id uuid.UUID, token pgtype.Text, expiry pgtype.Timestamp) error
	ListOrganizations(ctx context.Context, search string, limit, offset int32) ([]sqlc.AdminListOrganizationsRow, error)
	CountOrganizations(ctx context.Context, search string) (int64, error)
	GetOrganizationDetail(ctx context.Context, id uuid.UUID) (*sqlc.AdminGetOrganizationDetailRow, error)
	GetOrganizationMembers(ctx context.Context, orgID uuid.UUID) ([]sqlc.AdminGetOrganizationMembersRow, error)
	GetOrganizationWorkspaces(ctx context.Context, orgID uuid.UUID) ([]sqlc.AdminGetOrganizationWorkspacesRow, error)
	ToggleOrganizationArchive(ctx context.Context, id uuid.UUID) (*sqlc.AdminToggleOrganizationArchiveRow, error)
	DeleteOrganization(ctx context.Context, id uuid.UUID) error
	ClearUsersDefaultOrg(ctx context.Context, orgID uuid.UUID) error
	TransferOrgOwnership(ctx context.Context, orgID, newOwnerID uuid.UUID) (*sqlc.AdminTransferOrgOwnershipRow, error)
	GetOrganizationMember(ctx context.Context, userID, orgID uuid.UUID) (*sqlc.AdminGetOrganizationMemberRow, error)
	CreateOrganizationMember(ctx context.Context, id uuid.UUID, role sqlc.Role, userID, orgID, createdByID uuid.UUID) error
	UpdateOrganizationMemberRole(ctx context.Context, id uuid.UUID, role sqlc.Role) error
	GetUserByID(ctx context.Context, id uuid.UUID) (*sqlc.User, error)
}

type pgRepository struct {
	queries *sqlc.Queries
}

// NewRepository constructs a PostgreSQL backed repository for admin tasks.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{queries: db.Queries()}
}

func (r *pgRepository) GetDashboardStats(ctx context.Context, since time.Time) (*sqlc.GetDashboardStatsCountsRow, error) {
	row, err := r.queries.GetDashboardStatsCounts(ctx, pgtype.Timestamp{Time: since, Valid: true})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) ListUsers(ctx context.Context, search string, status sqlc.UserStatus, role sqlc.Role, limit, offset int32) ([]sqlc.AdminListUsersRow, error) {
	return r.queries.AdminListUsers(ctx, sqlc.AdminListUsersParams{
		Search:    search,
		Status:    string(status),
		Role:      string(role),
		LimitVal:  limit,
		OffsetVal: offset,
	})
}

func (r *pgRepository) CountUsers(ctx context.Context, search string, status sqlc.UserStatus, role sqlc.Role) (int64, error) {
	return r.queries.AdminCountUsers(ctx, sqlc.AdminCountUsersParams{
		Search: search,
		Status: string(status),
		Role:   string(role),
	})
}

func (r *pgRepository) GetUserDetail(ctx context.Context, id uuid.UUID) (*sqlc.AdminGetUserDetailRow, error) {
	row, err := r.queries.AdminGetUserDetail(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetUserOrganizations(ctx context.Context, userID uuid.UUID) ([]sqlc.AdminGetUserOrganizationsRow, error) {
	return r.queries.AdminGetUserOrganizations(ctx, userID)
}

func (r *pgRepository) CountSuperAdmins(ctx context.Context) (int64, error) {
	return r.queries.AdminCountSuperAdmins(ctx)
}

func (r *pgRepository) UpdateUserRole(ctx context.Context, id uuid.UUID, role sqlc.Role) (*sqlc.AdminUpdateUserRoleRow, error) {
	row, err := r.queries.AdminUpdateUserRole(ctx, sqlc.AdminUpdateUserRoleParams{
		Role: role,
		ID:   id,
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) UpdateUserStatus(ctx context.Context, id uuid.UUID, status sqlc.UserStatus) (*sqlc.AdminUpdateUserStatusRow, error) {
	row, err := r.queries.AdminUpdateUserStatus(ctx, sqlc.AdminUpdateUserStatusParams{
		Status: status,
		ID:     id,
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) SoftDeleteUser(ctx context.Context, id, deletedBy uuid.UUID) error {
	return r.queries.AdminSoftDeleteUser(ctx, sqlc.AdminSoftDeleteUserParams{
		ID:        id,
		DeletedBy: pgtype.UUID{Bytes: deletedBy, Valid: true},
	})
}

func (r *pgRepository) RemoveUserMemberships(ctx context.Context, userID uuid.UUID) error {
	if err := r.queries.AdminDeleteUserMemberships(ctx, userID); err != nil {
		return err
	}
	if err := r.queries.AdminDeleteUserWorkspaceMemberships(ctx, userID); err != nil {
		return err
	}
	return r.queries.AdminDeleteUserProjectMemberships(ctx, userID)
}

func (r *pgRepository) GetUserOwnedOrganizations(ctx context.Context, ownerID uuid.UUID) ([]sqlc.AdminGetUserOwnedOrganizationsRow, error) {
	return r.queries.AdminGetUserOwnedOrganizations(ctx, ownerID)
}

func (r *pgRepository) SetUserResetToken(ctx context.Context, id uuid.UUID, token pgtype.Text, expiry pgtype.Timestamp) error {
	return r.queries.AdminSetUserResetToken(ctx, sqlc.AdminSetUserResetTokenParams{
		ResetToken:       token,
		ResetTokenExpiry: expiry,
		ID:               id,
	})
}

func (r *pgRepository) ListOrganizations(ctx context.Context, search string, limit, offset int32) ([]sqlc.AdminListOrganizationsRow, error) {
	return r.queries.AdminListOrganizations(ctx, sqlc.AdminListOrganizationsParams{
		Search:    search,
		LimitVal:  limit,
		OffsetVal: offset,
	})
}

func (r *pgRepository) CountOrganizations(ctx context.Context, search string) (int64, error) {
	return r.queries.AdminCountOrganizations(ctx, search)
}

func (r *pgRepository) GetOrganizationDetail(ctx context.Context, id uuid.UUID) (*sqlc.AdminGetOrganizationDetailRow, error) {
	row, err := r.queries.AdminGetOrganizationDetail(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetOrganizationMembers(ctx context.Context, orgID uuid.UUID) ([]sqlc.AdminGetOrganizationMembersRow, error) {
	return r.queries.AdminGetOrganizationMembers(ctx, orgID)
}

func (r *pgRepository) GetOrganizationWorkspaces(ctx context.Context, orgID uuid.UUID) ([]sqlc.AdminGetOrganizationWorkspacesRow, error) {
	return r.queries.AdminGetOrganizationWorkspaces(ctx, orgID)
}

func (r *pgRepository) ToggleOrganizationArchive(ctx context.Context, id uuid.UUID) (*sqlc.AdminToggleOrganizationArchiveRow, error) {
	row, err := r.queries.AdminToggleOrganizationArchive(ctx, id)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) DeleteOrganization(ctx context.Context, id uuid.UUID) error {
	return r.queries.AdminDeleteOrganization(ctx, id)
}

func (r *pgRepository) ClearUsersDefaultOrg(ctx context.Context, orgID uuid.UUID) error {
	return r.queries.AdminClearUsersDefaultOrg(ctx, pgtype.UUID{Bytes: orgID, Valid: true})
}

func (r *pgRepository) TransferOrgOwnership(ctx context.Context, orgID, newOwnerID uuid.UUID) (*sqlc.AdminTransferOrgOwnershipRow, error) {
	row, err := r.queries.AdminTransferOrgOwnership(ctx, sqlc.AdminTransferOrgOwnershipParams{
		NewOwnerID: newOwnerID,
		OrgID:      orgID,
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetOrganizationMember(ctx context.Context, userID, orgID uuid.UUID) (*sqlc.AdminGetOrganizationMemberRow, error) {
	row, err := r.queries.AdminGetOrganizationMember(ctx, sqlc.AdminGetOrganizationMemberParams{
		UserID:         userID,
		OrganizationID: orgID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) CreateOrganizationMember(ctx context.Context, id uuid.UUID, role sqlc.Role, userID, orgID, createdByID uuid.UUID) error {
	return r.queries.AdminCreateOrganizationMember(ctx, sqlc.AdminCreateOrganizationMemberParams{
		ID:             id,
		Role:           role,
		UserID:         userID,
		OrganizationID: orgID,
		CreatedByID:    pgtype.UUID{Bytes: createdByID, Valid: true},
	})
}

func (r *pgRepository) UpdateOrganizationMemberRole(ctx context.Context, id uuid.UUID, role sqlc.Role) error {
	return r.queries.AdminUpdateOrganizationMemberRole(ctx, sqlc.AdminUpdateOrganizationMemberRoleParams{
		ID:   id,
		Role: role,
	})
}

func (r *pgRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*sqlc.User, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &sqlc.User{
		ID:        row.ID,
		Email:     row.Email,
		FirstName: row.FirstName,
		LastName:  row.LastName,
		Role:      row.Role,
		Status:    row.Status,
		DeletedAt: row.DeletedAt,
	}, nil
}
