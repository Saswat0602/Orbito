package workflows

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Repository defines data access operations for workflows.
type Repository interface {
	Create(ctx context.Context, workflow sqlc.CreateWorkflowParams, statuses []sqlc.CreateTaskStatusParams) (*sqlc.Workflow, error)
	FindDetailsByID(ctx context.Context, id uuid.UUID) (*sqlc.GetWorkflowDetailsByIDRow, error)
	FindStatusesByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]sqlc.GetWorkflowStatusesByWorkflowIDRow, error)
	FindTransitionsByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]sqlc.GetWorkflowTransitionsByWorkflowIDRow, error)
	FindDefaultByOrgID(ctx context.Context, orgID uuid.UUID) (*sqlc.GetDefaultWorkflowByOrgIDRow, error)
	ListByOrgID(ctx context.Context, orgID uuid.UUID) ([]sqlc.ListWorkflowsByOrganizationIDRow, error)
	ListByOrgIDs(ctx context.Context, orgIDs []uuid.UUID) ([]sqlc.ListWorkflowsByOrganizationIDsRow, error)
	ListAll(ctx context.Context) ([]sqlc.ListAllWorkflowsRow, error)
	Update(ctx context.Context, id uuid.UUID, orgID uuid.UUID, name *string, desc *string, isDefault *bool, updatedByID uuid.UUID) (*sqlc.Workflow, error)
	SetDefault(ctx context.Context, id uuid.UUID, orgID uuid.UUID, updatedByID uuid.UUID) (*sqlc.Workflow, error)
	Delete(ctx context.Context, id uuid.UUID) error
	CountProjectsUsingWorkflow(ctx context.Context, id uuid.UUID) (int64, error)
	CountTasksUsingWorkflow(ctx context.Context, id uuid.UUID) (int64, error)
	CountActiveStatuses(ctx context.Context, id uuid.UUID) (int64, error)
	CountTransitions(ctx context.Context, id uuid.UUID) (int64, error)
	GetUserAccessibleOrgIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	GetOrgOwnerAndUserRole(ctx context.Context, orgID uuid.UUID, userID uuid.UUID) (*sqlc.GetOrganizationOwnerAndUserRoleRow, error)
	GetWorkflowOrgOwnerAndUserRole(ctx context.Context, workflowID uuid.UUID, userID uuid.UUID) (*sqlc.GetWorkflowOrgOwnerAndUserRoleRow, error)
	GetOrgBySlug(ctx context.Context, slug string, userID uuid.UUID) (*sqlc.GetOrganizationBySlugForWorkflowRow, error)
}

type pgRepository struct {
	db      *database.DB
	queries *sqlc.Queries
}

// NewRepository constructs a new PostgreSQL workflows repository.
func NewRepository(db *database.DB) Repository {
	return &pgRepository{
		db:      db,
		queries: db.Queries(),
	}
}

func (r *pgRepository) Create(ctx context.Context, workflow sqlc.CreateWorkflowParams, statuses []sqlc.CreateTaskStatusParams) (*sqlc.Workflow, error) {
	var createdWorkflow sqlc.Workflow
	err := database.WithTx(ctx, r.db, func(txQueries *sqlc.Queries) error {
		if workflow.IsDefault {
			if err := txQueries.UnsetDefaultWorkflowsInOrg(ctx, sqlc.UnsetDefaultWorkflowsInOrgParams{
				OrganizationID: workflow.OrganizationID,
				UpdatedByID:    workflow.UpdatedByID,
			}); err != nil {
				return err
			}
		}

		wf, err := txQueries.CreateWorkflow(ctx, workflow)
		if err != nil {
			return err
		}
		createdWorkflow = wf

		for _, s := range statuses {
			s.WorkflowID = wf.ID
			if _, err := txQueries.CreateTaskStatus(ctx, s); err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return &createdWorkflow, nil
}

func (r *pgRepository) FindDetailsByID(ctx context.Context, id uuid.UUID) (*sqlc.GetWorkflowDetailsByIDRow, error) {
	row, err := r.queries.GetWorkflowDetailsByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) FindStatusesByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]sqlc.GetWorkflowStatusesByWorkflowIDRow, error) {
	return r.queries.GetWorkflowStatusesByWorkflowID(ctx, workflowID)
}

func (r *pgRepository) FindTransitionsByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]sqlc.GetWorkflowTransitionsByWorkflowIDRow, error) {
	return r.queries.GetWorkflowTransitionsByWorkflowID(ctx, workflowID)
}

func (r *pgRepository) FindDefaultByOrgID(ctx context.Context, orgID uuid.UUID) (*sqlc.GetDefaultWorkflowByOrgIDRow, error) {
	row, err := r.queries.GetDefaultWorkflowByOrgID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) ListByOrgID(ctx context.Context, orgID uuid.UUID) ([]sqlc.ListWorkflowsByOrganizationIDRow, error) {
	return r.queries.ListWorkflowsByOrganizationID(ctx, orgID)
}

func (r *pgRepository) ListByOrgIDs(ctx context.Context, orgIDs []uuid.UUID) ([]sqlc.ListWorkflowsByOrganizationIDsRow, error) {
	return r.queries.ListWorkflowsByOrganizationIDs(ctx, orgIDs)
}

func (r *pgRepository) ListAll(ctx context.Context) ([]sqlc.ListAllWorkflowsRow, error) {
	return r.queries.ListAllWorkflows(ctx)
}

func (r *pgRepository) Update(ctx context.Context, id uuid.UUID, orgID uuid.UUID, name *string, desc *string, isDefault *bool, updatedByID uuid.UUID) (*sqlc.Workflow, error) {
	var updatedWorkflow sqlc.Workflow
	err := database.WithTx(ctx, r.db, func(txQueries *sqlc.Queries) error {
		if isDefault != nil && *isDefault {
			if err := txQueries.UnsetDefaultWorkflowsExcept(ctx, sqlc.UnsetDefaultWorkflowsExceptParams{
				OrganizationID: orgID,
				UpdatedByID:    pgtype.UUID{Bytes: updatedByID, Valid: true},
				ID:             id,
			}); err != nil {
				return err
			}
		}

		params := sqlc.UpdateWorkflowDetailsParams{
			ID:          id,
			UpdatedByID: pgtype.UUID{Bytes: updatedByID, Valid: true},
		}
		if name != nil {
			params.Name = pgtype.Text{String: *name, Valid: true}
		}
		if desc != nil {
			params.Description = pgtype.Text{String: *desc, Valid: true}
		}
		if isDefault != nil {
			params.IsDefault = pgtype.Bool{Bool: *isDefault, Valid: true}
		}

		wf, err := txQueries.UpdateWorkflowDetails(ctx, params)
		if err != nil {
			return err
		}
		updatedWorkflow = wf
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &updatedWorkflow, nil
}

func (r *pgRepository) SetDefault(ctx context.Context, id uuid.UUID, orgID uuid.UUID, updatedByID uuid.UUID) (*sqlc.Workflow, error) {
	var updatedWorkflow sqlc.Workflow
	err := database.WithTx(ctx, r.db, func(txQueries *sqlc.Queries) error {
		if err := txQueries.UnsetDefaultWorkflowsInOrg(ctx, sqlc.UnsetDefaultWorkflowsInOrgParams{
			OrganizationID: orgID,
			UpdatedByID:    pgtype.UUID{Bytes: updatedByID, Valid: true},
		}); err != nil {
			return err
		}

		wf, err := txQueries.SetWorkflowDefault(ctx, sqlc.SetWorkflowDefaultParams{
			ID:          id,
			UpdatedByID: pgtype.UUID{Bytes: updatedByID, Valid: true},
		})
		if err != nil {
			return err
		}
		updatedWorkflow = wf
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &updatedWorkflow, nil
}

func (r *pgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteWorkflow(ctx, id)
}

func (r *pgRepository) CountProjectsUsingWorkflow(ctx context.Context, id uuid.UUID) (int64, error) {
	return r.queries.CountProjectsUsingWorkflow(ctx, id)
}

func (r *pgRepository) CountTasksUsingWorkflow(ctx context.Context, id uuid.UUID) (int64, error) {
	return r.queries.CountTasksUsingWorkflowStatuses(ctx, id)
}

func (r *pgRepository) CountActiveStatuses(ctx context.Context, id uuid.UUID) (int64, error) {
	return r.queries.CountWorkflowActiveStatuses(ctx, id)
}

func (r *pgRepository) CountTransitions(ctx context.Context, id uuid.UUID) (int64, error) {
	return r.queries.CountWorkflowTransitions(ctx, id)
}

func (r *pgRepository) GetUserAccessibleOrgIDs(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return r.queries.GetUserAccessibleOrgIDs(ctx, userID)
}

func (r *pgRepository) GetOrgOwnerAndUserRole(ctx context.Context, orgID uuid.UUID, userID uuid.UUID) (*sqlc.GetOrganizationOwnerAndUserRoleRow, error) {
	row, err := r.queries.GetOrganizationOwnerAndUserRole(ctx, sqlc.GetOrganizationOwnerAndUserRoleParams{
		ID:     orgID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetWorkflowOrgOwnerAndUserRole(ctx context.Context, workflowID uuid.UUID, userID uuid.UUID) (*sqlc.GetWorkflowOrgOwnerAndUserRoleRow, error) {
	row, err := r.queries.GetWorkflowOrgOwnerAndUserRole(ctx, sqlc.GetWorkflowOrgOwnerAndUserRoleParams{
		ID:     workflowID,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *pgRepository) GetOrgBySlug(ctx context.Context, slug string, userID uuid.UUID) (*sqlc.GetOrganizationBySlugForWorkflowRow, error) {
	row, err := r.queries.GetOrganizationBySlugForWorkflow(ctx, sqlc.GetOrganizationBySlugForWorkflowParams{
		Slug:   slug,
		UserID: userID,
	})
	if err != nil {
		return nil, err
	}
	return &row, nil
}
