package organizations

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository manages persistence for organizations and their related hierarchy.
type Repository interface {
	Create(ctx context.Context, params sqlc.CreateOrganizationParams) (*sqlc.Organization, error)
	GetDetailsByID(ctx context.Context, id uuid.UUID) (*sqlc.GetOrganizationDetailsByIDRow, error)
	GetBySlug(ctx context.Context, slug string) (*sqlc.GetOrganizationBySlugRow, error)
	ListForUser(ctx context.Context, userID uuid.UUID) ([]sqlc.ListOrganizationsForUserRow, error)
	Update(ctx context.Context, params sqlc.UpdateOrganizationParams) (*sqlc.Organization, error)
	Archive(ctx context.Context, id, userID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	TransferOwner(ctx context.Context, id, newOwnerID, updatedByID uuid.UUID) error
	CreateMember(ctx context.Context, params sqlc.CreateOrganizationMemberParams) (*sqlc.OrganizationMember, error)
	GetMember(ctx context.Context, userID, orgID uuid.UUID) (*sqlc.OrganizationMember, error)
	UpdateMemberRole(ctx context.Context, orgID, userID uuid.UUID, role sqlc.Role, updatedByID uuid.UUID) error
	CheckSlugExists(ctx context.Context, slug string, excludeID *uuid.UUID) (bool, error)
	SetUserDefaultOrg(ctx context.Context, userID, orgID uuid.UUID) error
	ClearUserDefaultOrg(ctx context.Context, orgID uuid.UUID) error
	CreateDefaultWorkflow(ctx context.Context, orgID, userID uuid.UUID) (*sqlc.Workflow, error)
	GetStats(ctx context.Context, orgID uuid.UUID) (*OrganizationStatsResponse, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*sqlc.GetUserByIDRow, error)
	WithTx(ctx context.Context, fn func(repo Repository) error) error
}

type pgRepository struct {
	db      *database.DB
	queries *sqlc.Queries
}

// NewRepository creates a PostgreSQL-backed repository for organizations.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{
		db:      db,
		queries: db.Queries(),
	}
}

func (r *pgRepository) Create(ctx context.Context, params sqlc.CreateOrganizationParams) (*sqlc.Organization, error) {
	org, err := r.queries.CreateOrganization(ctx, params)
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *pgRepository) GetDetailsByID(ctx context.Context, id uuid.UUID) (*sqlc.GetOrganizationDetailsByIDRow, error) {
	row, err := r.queries.GetOrganizationDetailsByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetBySlug(ctx context.Context, slug string) (*sqlc.GetOrganizationBySlugRow, error) {
	row, err := r.queries.GetOrganizationBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) ListForUser(ctx context.Context, userID uuid.UUID) ([]sqlc.ListOrganizationsForUserRow, error) {
	return r.queries.ListOrganizationsForUser(ctx, userID)
}

func (r *pgRepository) Update(ctx context.Context, params sqlc.UpdateOrganizationParams) (*sqlc.Organization, error) {
	org, err := r.queries.UpdateOrganization(ctx, params)
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *pgRepository) Archive(ctx context.Context, id, userID uuid.UUID) error {
	return r.queries.ArchiveOrganization(ctx, sqlc.ArchiveOrganizationParams{
		ID:          id,
		UpdatedByID: pgtype.UUID{Bytes: userID, Valid: true},
	})
}

func (r *pgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteOrganization(ctx, id)
}

func (r *pgRepository) TransferOwner(ctx context.Context, id, newOwnerID, updatedByID uuid.UUID) error {
	return r.queries.TransferOrganizationOwner(ctx, sqlc.TransferOrganizationOwnerParams{
		ID:          id,
		OwnerID:     newOwnerID,
		UpdatedByID: pgtype.UUID{Bytes: updatedByID, Valid: true},
	})
}

func (r *pgRepository) CreateMember(ctx context.Context, params sqlc.CreateOrganizationMemberParams) (*sqlc.OrganizationMember, error) {
	m, err := r.queries.CreateOrganizationMember(ctx, params)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *pgRepository) GetMember(ctx context.Context, userID, orgID uuid.UUID) (*sqlc.OrganizationMember, error) {
	m, err := r.queries.GetOrganizationMember(ctx, sqlc.GetOrganizationMemberParams{
		UserID:         userID,
		OrganizationID: orgID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *pgRepository) UpdateMemberRole(ctx context.Context, orgID, userID uuid.UUID, role sqlc.Role, updatedByID uuid.UUID) error {
	return r.queries.UpdateOrganizationMemberRole(ctx, sqlc.UpdateOrganizationMemberRoleParams{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           role,
		UpdatedByID:    pgtype.UUID{Bytes: updatedByID, Valid: true},
	})
}

func (r *pgRepository) CheckSlugExists(ctx context.Context, slug string, excludeID *uuid.UUID) (bool, error) {
	var exc uuid.UUID
	if excludeID != nil {
		exc = *excludeID
	}
	return r.queries.CheckOrganizationSlugExists(ctx, sqlc.CheckOrganizationSlugExistsParams{
		Slug:    slug,
		Column2: exc,
	})
}

func (r *pgRepository) SetUserDefaultOrg(ctx context.Context, userID, orgID uuid.UUID) error {
	return r.queries.SetUserDefaultOrganization(ctx, sqlc.SetUserDefaultOrganizationParams{
		ID:                    userID,
		DefaultOrganizationID: pgtype.UUID{Bytes: orgID, Valid: true},
	})
}

func (r *pgRepository) ClearUserDefaultOrg(ctx context.Context, orgID uuid.UUID) error {
	return r.queries.ClearUserDefaultOrganization(ctx, pgtype.UUID{Bytes: orgID, Valid: true})
}

func (r *pgRepository) CreateDefaultWorkflow(ctx context.Context, orgID, userID uuid.UUID) (*sqlc.Workflow, error) {
	desc := "Auto-generated default workflow for new organizations"
	wfID := uuid.New()

	wf, err := r.queries.CreateWorkflow(ctx, sqlc.CreateWorkflowParams{
		ID:             wfID,
		Name:           "Default Workflow",
		Description:    pgtype.Text{String: desc, Valid: true},
		IsDefault:      true,
		OrganizationID: orgID,
		CreatedByID:    pgtype.UUID{Bytes: userID, Valid: true},
		UpdatedByID:    pgtype.UUID{Bytes: userID, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	// Statuses
	statuses := []struct {
		name      string
		color     string
		category  sqlc.StatusCategory
		position  int32
		isDefault bool
	}{
		{"To Do", "#6366f1", sqlc.StatusCategoryTODO, 1, true},
		{"In Progress", "#f59e0b", sqlc.StatusCategoryINPROGRESS, 2, false},
		{"In Review", "#8b5cf6", sqlc.StatusCategoryINPROGRESS, 3, false},
		{"Done", "#10b981", sqlc.StatusCategoryDONE, 4, false},
	}

	statusMap := make(map[string]uuid.UUID)
	for _, s := range statuses {
		stID := uuid.New()
		_, sErr := r.queries.CreateTaskStatus(ctx, sqlc.CreateTaskStatusParams{
			ID:          stID,
			Name:        s.name,
			Color:       s.color,
			Category:    s.category,
			Position:    s.position,
			IsDefault:   s.isDefault,
			WorkflowID:  wfID,
			CreatedByID: pgtype.UUID{Bytes: userID, Valid: true},
			UpdatedByID: pgtype.UUID{Bytes: userID, Valid: true},
		})
		if sErr != nil {
			return nil, sErr
		}
		statusMap[s.name] = stID
	}

	// Transitions
	transitions := []struct {
		from string
		to   string
	}{
		{"To Do", "In Progress"},
		{"In Progress", "In Review"},
		{"In Progress", "Done"},
		{"In Review", "To Do"},
		{"In Review", "Done"},
	}

	for _, tr := range transitions {
		fromID, ok1 := statusMap[tr.from]
		toID, ok2 := statusMap[tr.to]
		if ok1 && ok2 {
			_, tErr := r.queries.CreateStatusTransition(ctx, sqlc.CreateStatusTransitionParams{
				ID:           uuid.New(),
				WorkflowID:   wfID,
				FromStatusID: fromID,
				ToStatusID:   toID,
				CreatedByID:  pgtype.UUID{Bytes: userID, Valid: true},
				UpdatedByID:  pgtype.UUID{Bytes: userID, Valid: true},
			})
			if tErr != nil {
				return nil, tErr
			}
		}
	}

	return &wf, nil
}

func (r *pgRepository) GetStats(ctx context.Context, orgID uuid.UUID) (*OrganizationStatsResponse, error) {
	members, err := r.queries.CountOrganizationMembers(ctx, orgID)
	if err != nil {
		return nil, err
	}

	workspaces, err := r.queries.CountOrganizationWorkspaces(ctx, orgID)
	if err != nil {
		return nil, err
	}

	projects, err := r.queries.CountOrganizationProjects(ctx, orgID)
	if err != nil {
		return nil, err
	}

	tasks, err := r.queries.CountOrganizationTasks(ctx, orgID)
	if err != nil {
		return nil, err
	}

	tasksDone, err := r.queries.CountOrganizationTasksCompleted(ctx, orgID)
	if err != nil {
		return nil, err
	}

	return &OrganizationStatsResponse{
		TotalMembers:    members,
		TotalWorkspaces: workspaces,
		TotalProjects:   projects,
		TotalTasks:      tasks,
		CompletedTasks:  tasksDone,
	}, nil
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

func (r *pgRepository) WithTx(ctx context.Context, fn func(repo Repository) error) error {
	return r.db.WithTx(ctx, func(q *sqlc.Queries) error {
		txRepo := &pgRepository{
			db:      r.db,
			queries: q,
		}
		return fn(txRepo)
	})
}
