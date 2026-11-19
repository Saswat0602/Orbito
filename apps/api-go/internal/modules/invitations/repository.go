package invitations

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository defines data access operations for invitations.
type Repository interface {
	Create(ctx context.Context, params sqlc.CreateInvitationParams) (*sqlc.Invitation, error)
	GetByID(ctx context.Context, id uuid.UUID) (*sqlc.GetInvitationByIDRow, error)
	GetByToken(ctx context.Context, token string) (*sqlc.GetInvitationByTokenRow, error)
	FindPending(ctx context.Context, params sqlc.FindPendingInvitationParams) (*sqlc.Invitation, error)
	GetUserInvitations(ctx context.Context, email string) ([]sqlc.GetUserInvitationsRow, error)
	GetEntityInvitations(ctx context.Context, entityType string, entityID uuid.UUID) ([]sqlc.GetEntityInvitationsRow, error)
	ExpirePendingEntityInvitations(ctx context.Context, entityType string, entityID uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.InvitationStatus) error
	Resend(ctx context.Context, params sqlc.ResendInvitationParams) (*sqlc.Invitation, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListRelatedProjectInvitations(ctx context.Context, email string, workspaceID uuid.UUID) ([]sqlc.Invitation, error)
	GetProjectWithWorkspaceAndOrg(ctx context.Context, id uuid.UUID) (*sqlc.GetProjectWithWorkspaceAndOrgRow, error)
	CheckProjectBelongsToWorkspace(ctx context.Context, projectID, workspaceID uuid.UUID) (bool, error)
	CheckOrgMemberExists(ctx context.Context, userID, orgID uuid.UUID) (bool, error)
	CheckWorkspaceMemberExists(ctx context.Context, userID, wsID uuid.UUID) (bool, error)
	CheckProjectMemberExists(ctx context.Context, userID, projID uuid.UUID) (bool, error)
	GetUserByEmail(ctx context.Context, email string) (*sqlc.GetUserByEmailRow, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*sqlc.GetUserByIDRow, error)
	GetWorkspaceByID(ctx context.Context, id uuid.UUID) (*sqlc.GetWorkspaceByIDRow, error)
	GetOrganizationByID(ctx context.Context, id uuid.UUID) (*sqlc.GetOrganizationDetailsByIDRow, error)
	CreateOrganizationMember(ctx context.Context, params sqlc.CreateOrganizationMemberParams) (*sqlc.OrganizationMember, error)
	CreateWorkspaceMember(ctx context.Context, params sqlc.CreateWorkspaceMemberDirectParams) (*sqlc.WorkspaceMember, error)
	CreateProjectMember(ctx context.Context, params sqlc.CreateProjectMemberDirectParams) (*sqlc.ProjectMember, error)
	WithTx(ctx context.Context, fn func(repo Repository) error) error
}

type pgRepository struct {
	db      *database.DB
	queries *sqlc.Queries
}

// NewRepository creates a new PostgreSQL invitations repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{
		db:      db,
		queries: db.Queries(),
	}
}

func (r *pgRepository) Create(ctx context.Context, params sqlc.CreateInvitationParams) (*sqlc.Invitation, error) {
	inv, err := r.queries.CreateInvitation(ctx, params)
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

func (r *pgRepository) GetByID(ctx context.Context, id uuid.UUID) (*sqlc.GetInvitationByIDRow, error) {
	row, err := r.queries.GetInvitationByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetByToken(ctx context.Context, token string) (*sqlc.GetInvitationByTokenRow, error) {
	row, err := r.queries.GetInvitationByToken(ctx, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) FindPending(ctx context.Context, params sqlc.FindPendingInvitationParams) (*sqlc.Invitation, error) {
	inv, err := r.queries.FindPendingInvitation(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &inv, nil
}

func (r *pgRepository) GetUserInvitations(ctx context.Context, email string) ([]sqlc.GetUserInvitationsRow, error) {
	return r.queries.GetUserInvitations(ctx, email)
}

func (r *pgRepository) GetEntityInvitations(ctx context.Context, entityType string, entityID uuid.UUID) ([]sqlc.GetEntityInvitationsRow, error) {
	return r.queries.GetEntityInvitations(ctx, sqlc.GetEntityInvitationsParams{
		EntityType: entityType,
		EntityID:   pgtype.UUID{Bytes: entityID, Valid: true},
	})
}

func (r *pgRepository) ExpirePendingEntityInvitations(ctx context.Context, entityType string, entityID uuid.UUID) error {
	return r.queries.ExpirePendingEntityInvitations(ctx, sqlc.ExpirePendingEntityInvitationsParams{
		EntityType: entityType,
		EntityID:   pgtype.UUID{Bytes: entityID, Valid: true},
	})
}

func (r *pgRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.InvitationStatus) error {
	return r.queries.UpdateInvitationStatus(ctx, sqlc.UpdateInvitationStatusParams{
		ID:     id,
		Status: status,
	})
}

func (r *pgRepository) Resend(ctx context.Context, params sqlc.ResendInvitationParams) (*sqlc.Invitation, error) {
	inv, err := r.queries.ResendInvitation(ctx, params)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &inv, nil
}

func (r *pgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteInvitation(ctx, id)
}

func (r *pgRepository) ListRelatedProjectInvitations(ctx context.Context, email string, workspaceID uuid.UUID) ([]sqlc.Invitation, error) {
	return r.queries.ListRelatedProjectInvitations(ctx, sqlc.ListRelatedProjectInvitationsParams{
		InviteeEmail: email,
		WorkspaceID:  workspaceID,
	})
}

func (r *pgRepository) GetProjectWithWorkspaceAndOrg(ctx context.Context, id uuid.UUID) (*sqlc.GetProjectWithWorkspaceAndOrgRow, error) {
	row, err := r.queries.GetProjectWithWorkspaceAndOrg(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) CheckProjectBelongsToWorkspace(ctx context.Context, projectID, workspaceID uuid.UUID) (bool, error) {
	count, err := r.queries.CheckProjectBelongsToWorkspace(ctx, sqlc.CheckProjectBelongsToWorkspaceParams{
		ID:          projectID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *pgRepository) CheckOrgMemberExists(ctx context.Context, userID, orgID uuid.UUID) (bool, error) {
	count, err := r.queries.CheckOrgMemberExists(ctx, sqlc.CheckOrgMemberExistsParams{
		UserID:         userID,
		OrganizationID: orgID,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *pgRepository) CheckWorkspaceMemberExists(ctx context.Context, userID, wsID uuid.UUID) (bool, error) {
	count, err := r.queries.CheckWorkspaceMemberExists(ctx, sqlc.CheckWorkspaceMemberExistsParams{
		UserID:      userID,
		WorkspaceID: wsID,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *pgRepository) CheckProjectMemberExists(ctx context.Context, userID, projID uuid.UUID) (bool, error) {
	count, err := r.queries.CheckProjectMemberExists(ctx, sqlc.CheckProjectMemberExistsParams{
		UserID:    userID,
		ProjectID: projID,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *pgRepository) GetUserByEmail(ctx context.Context, email string) (*sqlc.GetUserByEmailRow, error) {
	u, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *pgRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*sqlc.GetUserByIDRow, error) {
	u, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *pgRepository) GetWorkspaceByID(ctx context.Context, id uuid.UUID) (*sqlc.GetWorkspaceByIDRow, error) {
	ws, err := r.queries.GetWorkspaceByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &ws, nil
}

func (r *pgRepository) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*sqlc.GetOrganizationDetailsByIDRow, error) {
	org, err := r.queries.GetOrganizationDetailsByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &org, nil
}

func (r *pgRepository) CreateOrganizationMember(ctx context.Context, params sqlc.CreateOrganizationMemberParams) (*sqlc.OrganizationMember, error) {
	m, err := r.queries.CreateOrganizationMember(ctx, params)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *pgRepository) CreateWorkspaceMember(ctx context.Context, params sqlc.CreateWorkspaceMemberDirectParams) (*sqlc.WorkspaceMember, error) {
	m, err := r.queries.CreateWorkspaceMemberDirect(ctx, params)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *pgRepository) CreateProjectMember(ctx context.Context, params sqlc.CreateProjectMemberDirectParams) (*sqlc.ProjectMember, error) {
	m, err := r.queries.CreateProjectMemberDirect(ctx, params)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *pgRepository) WithTx(ctx context.Context, fn func(repo Repository) error) error {
	return r.db.WithTx(ctx, func(q *sqlc.Queries) error {
		txRepo := &pgRepository{
			db:      r.db,
			queries: q,
		}
		return fn(txRepo)
	})
}
