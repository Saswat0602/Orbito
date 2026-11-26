package workflows

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/middleware"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	orgs        map[uuid.UUID]*sqlc.GetOrganizationOwnerAndUserRoleRow
	orgsBySlug  map[string]*sqlc.GetOrganizationBySlugForWorkflowRow
	workflows   map[uuid.UUID]*sqlc.GetWorkflowDetailsByIDRow
	wfStatuses  map[uuid.UUID][]sqlc.TaskStatus
	transitions map[uuid.UUID][]sqlc.GetWorkflowTransitionsByWorkflowIDRow
	projectDeps map[uuid.UUID]int64
	taskDeps    map[uuid.UUID]int64
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		orgs:        make(map[uuid.UUID]*sqlc.GetOrganizationOwnerAndUserRoleRow),
		orgsBySlug:  make(map[string]*sqlc.GetOrganizationBySlugForWorkflowRow),
		workflows:   make(map[uuid.UUID]*sqlc.GetWorkflowDetailsByIDRow),
		wfStatuses:  make(map[uuid.UUID][]sqlc.TaskStatus),
		transitions: make(map[uuid.UUID][]sqlc.GetWorkflowTransitionsByWorkflowIDRow),
		projectDeps: make(map[uuid.UUID]int64),
		taskDeps:    make(map[uuid.UUID]int64),
	}
}

func (m *mockRepo) Create(_ context.Context, workflow sqlc.CreateWorkflowParams, statuses []sqlc.CreateTaskStatusParams) (*sqlc.Workflow, error) {
	desc := pgtype.Text{}
	if workflow.Description.Valid {
		desc = workflow.Description
	}

	row := &sqlc.GetWorkflowDetailsByIDRow{
		ID:             workflow.ID,
		Name:           workflow.Name,
		Description:    desc,
		IsDefault:      workflow.IsDefault,
		OrganizationID: workflow.OrganizationID,
		CreatedByID:    workflow.CreatedByID,
		UpdatedByID:    workflow.UpdatedByID,
		CreatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
		UpdatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
		OrgID:          workflow.OrganizationID,
		OrgName:        "Test Org",
		OrgSlug:        "test-org",
	}
	m.workflows[workflow.ID] = row

	var stRows []sqlc.TaskStatus
	for _, s := range statuses {
		stRows = append(stRows, sqlc.TaskStatus{
			ID:          s.ID,
			Name:        s.Name,
			Color:       s.Color,
			Category:    s.Category,
			Position:    s.Position,
			IsDefault:   s.IsDefault,
			WorkflowID:  workflow.ID,
			CreatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
			UpdatedAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
			CreatedByID: s.CreatedByID,
			UpdatedByID: s.UpdatedByID,
		})
	}
	m.wfStatuses[workflow.ID] = stRows

	return &sqlc.Workflow{
		ID:             workflow.ID,
		Name:           workflow.Name,
		Description:    desc,
		IsDefault:      workflow.IsDefault,
		OrganizationID: workflow.OrganizationID,
		CreatedByID:    workflow.CreatedByID,
		UpdatedByID:    workflow.UpdatedByID,
		CreatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
		UpdatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
	}, nil
}

func (m *mockRepo) FindDetailsByID(_ context.Context, id uuid.UUID) (*sqlc.GetWorkflowDetailsByIDRow, error) {
	row, ok := m.workflows[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return row, nil
}

func (m *mockRepo) FindStatusesByWorkflowID(_ context.Context, workflowID uuid.UUID) ([]sqlc.TaskStatus, error) {
	return m.wfStatuses[workflowID], nil
}

func (m *mockRepo) FindTransitionsByWorkflowID(_ context.Context, workflowID uuid.UUID) ([]sqlc.GetWorkflowTransitionsByWorkflowIDRow, error) {
	return m.transitions[workflowID], nil
}

func (m *mockRepo) FindDefaultByOrgID(_ context.Context, orgID uuid.UUID) (*sqlc.GetDefaultWorkflowByOrgIDRow, error) {
	for _, wf := range m.workflows {
		if wf.OrganizationID == orgID && wf.IsDefault {
			return &sqlc.GetDefaultWorkflowByOrgIDRow{
				ID:             wf.ID,
				Name:           wf.Name,
				Description:    wf.Description,
				IsDefault:      wf.IsDefault,
				OrganizationID: wf.OrganizationID,
				CreatedByID:    wf.CreatedByID,
				UpdatedByID:    wf.UpdatedByID,
				CreatedAt:      wf.CreatedAt,
				UpdatedAt:      wf.UpdatedAt,
				OrgID:          wf.OrgID,
				OrgName:        wf.OrgName,
				OrgSlug:        wf.OrgSlug,
			}, nil
		}
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockRepo) ListByOrgID(_ context.Context, orgID uuid.UUID) ([]sqlc.ListWorkflowsByOrganizationIDRow, error) {
	var rows []sqlc.ListWorkflowsByOrganizationIDRow
	for _, wf := range m.workflows {
		if wf.OrganizationID == orgID {
			rows = append(rows, sqlc.ListWorkflowsByOrganizationIDRow{
				ID:             wf.ID,
				Name:           wf.Name,
				Description:    wf.Description,
				IsDefault:      wf.IsDefault,
				OrganizationID: wf.OrganizationID,
				CreatedByID:    wf.CreatedByID,
				UpdatedByID:    wf.UpdatedByID,
				CreatedAt:      wf.CreatedAt,
				UpdatedAt:      wf.UpdatedAt,
				OrgID:          wf.OrgID,
				OrgName:        wf.OrgName,
				OrgSlug:        wf.OrgSlug,
			})
		}
	}
	return rows, nil
}

func (m *mockRepo) ListByOrgIDs(_ context.Context, orgIDs []uuid.UUID) ([]sqlc.ListWorkflowsByOrganizationIDsRow, error) {
	orgMap := make(map[uuid.UUID]bool)
	for _, id := range orgIDs {
		orgMap[id] = true
	}
	var rows []sqlc.ListWorkflowsByOrganizationIDsRow
	for _, wf := range m.workflows {
		if orgMap[wf.OrganizationID] {
			rows = append(rows, sqlc.ListWorkflowsByOrganizationIDsRow{
				ID:             wf.ID,
				Name:           wf.Name,
				Description:    wf.Description,
				IsDefault:      wf.IsDefault,
				OrganizationID: wf.OrganizationID,
				CreatedByID:    wf.CreatedByID,
				UpdatedByID:    wf.UpdatedByID,
				CreatedAt:      wf.CreatedAt,
				UpdatedAt:      wf.UpdatedAt,
				OrgID:          wf.OrgID,
				OrgName:        wf.OrgName,
				OrgSlug:        wf.OrgSlug,
			})
		}
	}
	return rows, nil
}

func (m *mockRepo) ListAll(_ context.Context) ([]sqlc.ListAllWorkflowsRow, error) {
	var rows []sqlc.ListAllWorkflowsRow
	for _, wf := range m.workflows {
		rows = append(rows, sqlc.ListAllWorkflowsRow{
			ID:             wf.ID,
			Name:           wf.Name,
			Description:    wf.Description,
			IsDefault:      wf.IsDefault,
			OrganizationID: wf.OrganizationID,
			CreatedByID:    wf.CreatedByID,
			UpdatedByID:    wf.UpdatedByID,
			CreatedAt:      wf.CreatedAt,
			UpdatedAt:      wf.UpdatedAt,
			OrgID:          wf.OrgID,
			OrgName:        wf.OrgName,
			OrgSlug:        wf.OrgSlug,
		})
	}
	return rows, nil
}

func (m *mockRepo) Update(_ context.Context, id uuid.UUID, _ uuid.UUID, name *string, desc *string, isDefault *bool, _ uuid.UUID) (*sqlc.Workflow, error) {
	wf, ok := m.workflows[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	if name != nil {
		wf.Name = *name
	}
	if desc != nil {
		wf.Description = pgtype.Text{String: *desc, Valid: true}
	}
	if isDefault != nil {
		wf.IsDefault = *isDefault
	}
	return &sqlc.Workflow{
		ID:             wf.ID,
		Name:           wf.Name,
		Description:    wf.Description,
		IsDefault:      wf.IsDefault,
		OrganizationID: wf.OrganizationID,
	}, nil
}

func (m *mockRepo) SetDefault(_ context.Context, id uuid.UUID, orgID uuid.UUID, _ uuid.UUID) (*sqlc.Workflow, error) {
	wf, ok := m.workflows[id]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	for _, other := range m.workflows {
		if other.OrganizationID == orgID {
			other.IsDefault = false
		}
	}
	wf.IsDefault = true
	return &sqlc.Workflow{
		ID:             wf.ID,
		Name:           wf.Name,
		IsDefault:      true,
		OrganizationID: wf.OrganizationID,
	}, nil
}

func (m *mockRepo) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := m.workflows[id]; !ok {
		return fmt.Errorf("not found")
	}
	delete(m.workflows, id)
	delete(m.wfStatuses, id)
	return nil
}

func (m *mockRepo) CountProjectsUsingWorkflow(_ context.Context, id uuid.UUID) (int64, error) {
	return m.projectDeps[id], nil
}

func (m *mockRepo) CountTasksUsingWorkflow(_ context.Context, id uuid.UUID) (int64, error) {
	return m.taskDeps[id], nil
}

func (m *mockRepo) CountActiveStatuses(_ context.Context, id uuid.UUID) (int64, error) {
	return int64(len(m.wfStatuses[id])), nil
}

func (m *mockRepo) CountTransitions(_ context.Context, id uuid.UUID) (int64, error) {
	return int64(len(m.transitions[id])), nil
}

func (m *mockRepo) GetUserAccessibleOrgIDs(_ context.Context, _ uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	for id := range m.orgs {
		ids = append(ids, id)
	}
	return ids, nil
}

func (m *mockRepo) GetOrgOwnerAndUserRole(_ context.Context, orgID uuid.UUID, _ uuid.UUID) (*sqlc.GetOrganizationOwnerAndUserRoleRow, error) {
	row, ok := m.orgs[orgID]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return row, nil
}

func (m *mockRepo) GetWorkflowOrgOwnerAndUserRole(_ context.Context, workflowID uuid.UUID, userID uuid.UUID) (*sqlc.GetWorkflowOrgOwnerAndUserRoleRow, error) {
	wf, ok := m.workflows[workflowID]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	orgRow, ok := m.orgs[wf.OrganizationID]
	if !ok {
		return nil, fmt.Errorf("org not found")
	}
	return &sqlc.GetWorkflowOrgOwnerAndUserRoleRow{
		WorkflowID:     wf.ID,
		OrganizationID: wf.OrganizationID,
		OwnerID:        orgRow.OwnerID,
		MemberRole:     orgRow.MemberRole,
	}, nil
}

func (m *mockRepo) GetOrgBySlug(_ context.Context, slug string, _ uuid.UUID) (*sqlc.GetOrganizationBySlugForWorkflowRow, error) {
	row, ok := m.orgsBySlug[slug]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return row, nil
}

func TestWorkflowsService(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()
	managerID := uuid.New()
	memberID := uuid.New()
	outsiderID := uuid.New()
	orgID := uuid.New()

	managerRole := sqlc.Role(middleware.RoleManager)
	memberRole := sqlc.Role(middleware.RoleMember)

	repo := newMockRepo()
	repo.orgs[orgID] = &sqlc.GetOrganizationOwnerAndUserRoleRow{
		ID:         orgID,
		OwnerID:    ownerID,
		MemberRole: sqlc.NullRole{Valid: false},
	}
	repo.orgsBySlug["test-org"] = &sqlc.GetOrganizationBySlugForWorkflowRow{
		ID:         orgID,
		OwnerID:    ownerID,
		MemberRole: sqlc.NullRole{Valid: false},
	}

	svc := NewService(repo)

	t.Run("Create workflow - validates organization uuid", func(t *testing.T) {
		_, err := svc.Create(ctx, CreateWorkflowRequest{
			Name:           "Invalid Org",
			OrganizationID: "invalid-uuid",
		}, ownerID, false)
		require.Error(t, err)
		var appErr *apperror.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, 400, appErr.StatusCode)
	})

	t.Run("Create workflow - denies non-member", func(t *testing.T) {
		repo.orgs[orgID].MemberRole = sqlc.NullRole{Valid: false}
		_, err := svc.Create(ctx, CreateWorkflowRequest{
			Name:           "Dev Workflow",
			OrganizationID: orgID.String(),
		}, outsiderID, false)
		require.Error(t, err)
		var appErr *apperror.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, 403, appErr.StatusCode)
	})

	var createdWfID uuid.UUID
	t.Run("Create workflow - success for owner with 4 default statuses", func(t *testing.T) {
		desc := "Standard engineering workflow"
		isDefault := true
		res, err := svc.Create(ctx, CreateWorkflowRequest{
			Name:           "Engineering Workflow",
			Description:    &desc,
			IsDefault:      &isDefault,
			OrganizationID: orgID.String(),
		}, ownerID, false)
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, "Engineering Workflow", res.Name)
		assert.Equal(t, true, res.IsDefault)
		assert.Equal(t, 4, len(res.Statuses))
		assert.Equal(t, "To Do", res.Statuses[0].Name)
		assert.Equal(t, "In Progress", res.Statuses[1].Name)
		assert.Equal(t, "In Review", res.Statuses[2].Name)
		assert.Equal(t, "Done", res.Statuses[3].Name)
		assert.Equal(t, int64(4), res.Count.Statuses)

		parsed, err := uuid.Parse(res.ID)
		require.NoError(t, err)
		createdWfID = parsed
	})

	t.Run("FindAll by Organization ID", func(t *testing.T) {
		res, err := svc.FindAll(ctx, &orgID, ownerID, false)
		require.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, "Engineering Workflow", res[0].Name)
	})

	t.Run("FindByOrganizationSlug", func(t *testing.T) {
		t.Run("fails on empty slug", func(t *testing.T) {
			_, err := svc.FindByOrganizationSlug(ctx, "  ", ownerID, false)
			require.Error(t, err)
			var appErr *apperror.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, 400, appErr.StatusCode)
		})

		t.Run("fails on unknown slug", func(t *testing.T) {
			_, err := svc.FindByOrganizationSlug(ctx, "nonexistent", ownerID, false)
			require.Error(t, err)
			var appErr *apperror.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, 404, appErr.StatusCode)
		})

		t.Run("succeeds for valid slug", func(t *testing.T) {
			res, err := svc.FindByOrganizationSlug(ctx, "test-org", ownerID, false)
			require.NoError(t, err)
			assert.Len(t, res, 1)
		})
	})

	t.Run("FindOne", func(t *testing.T) {
		t.Run("fails on unknown id", func(t *testing.T) {
			_, err := svc.FindOne(ctx, uuid.New(), ownerID, false)
			require.Error(t, err)
			var appErr *apperror.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, 404, appErr.StatusCode)
		})

		t.Run("succeeds for existing workflow", func(t *testing.T) {
			res, err := svc.FindOne(ctx, createdWfID, ownerID, false)
			require.NoError(t, err)
			assert.Equal(t, "Engineering Workflow", res.Name)
			assert.Equal(t, "test-org", res.Organization.Slug)
			assert.Equal(t, 4, len(res.Statuses))
		})
	})

	t.Run("GetDefaultWorkflow", func(t *testing.T) {
		res, err := svc.GetDefaultWorkflow(ctx, orgID, ownerID, false)
		require.NoError(t, err)
		assert.Equal(t, createdWfID.String(), res.ID)
		assert.True(t, res.IsDefault)
	})

	t.Run("Update workflow", func(t *testing.T) {
		t.Run("denies member role (requires MANAGER)", func(t *testing.T) {
			repo.orgs[orgID].MemberRole = sqlc.NullRole{Role: memberRole, Valid: true}
			newName := "Member Hack"
			_, err := svc.Update(ctx, createdWfID, UpdateWorkflowRequest{
				Name: &newName,
			}, memberID, false)
			require.Error(t, err)
			var appErr *apperror.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, 403, appErr.StatusCode)
		})

		t.Run("allows manager to update workflow details", func(t *testing.T) {
			repo.orgs[orgID].MemberRole = sqlc.NullRole{Role: managerRole, Valid: true}
			newName := "Updated Workflow Name"
			res, err := svc.Update(ctx, createdWfID, UpdateWorkflowRequest{
				Name: &newName,
			}, managerID, false)
			require.NoError(t, err)
			assert.Equal(t, "Updated Workflow Name", res.Name)
		})
	})

	t.Run("MakeDefault workflow", func(t *testing.T) {
		t.Run("fails if workflow does not belong to organization", func(t *testing.T) {
			otherOrgID := uuid.New()
			repo.orgs[otherOrgID] = &sqlc.GetOrganizationOwnerAndUserRoleRow{
				ID:         otherOrgID,
				OwnerID:    ownerID,
				MemberRole: sqlc.NullRole{Valid: false},
			}
			_, err := svc.MakeDefault(ctx, createdWfID, otherOrgID, ownerID, false)
			require.Error(t, err)
			var appErr *apperror.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, 400, appErr.StatusCode)
			assert.Contains(t, err.Error(), "Workflow does not belong to this organization")
		})

		t.Run("succeeds setting workflow as default", func(t *testing.T) {
			res, err := svc.MakeDefault(ctx, createdWfID, orgID, ownerID, false)
			require.NoError(t, err)
			assert.True(t, res.IsDefault)
		})
	})

	t.Run("Delete workflow", func(t *testing.T) {
		t.Run("fails with ConflictException when projects are using it", func(t *testing.T) {
			repo.projectDeps[createdWfID] = 2
			repo.taskDeps[createdWfID] = 0

			err := svc.Delete(ctx, createdWfID, ownerID, false)
			require.Error(t, err)
			var appErr *apperror.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, 409, appErr.StatusCode)
			assert.Contains(t, err.Error(), "This workflow is still in use by 2 projects. Move them to another workflow before deleting it")
		})

		t.Run("fails with ConflictException when tasks and projects are using it", func(t *testing.T) {
			repo.projectDeps[createdWfID] = 1
			repo.taskDeps[createdWfID] = 5

			err := svc.Delete(ctx, createdWfID, ownerID, false)
			require.Error(t, err)
			var appErr *apperror.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, 409, appErr.StatusCode)
			assert.Contains(t, err.Error(), "This workflow is still in use by 5 tasks and 1 project. Move them to another workflow before deleting it")
		})

		t.Run("succeeds when no tasks or projects depend on it", func(t *testing.T) {
			repo.projectDeps[createdWfID] = 0
			repo.taskDeps[createdWfID] = 0

			err := svc.Delete(ctx, createdWfID, ownerID, false)
			require.NoError(t, err)

			_, err = svc.FindOne(ctx, createdWfID, ownerID, false)
			require.Error(t, err)
			var appErr *apperror.AppError
			require.ErrorAs(t, err, &appErr)
			assert.Equal(t, 404, appErr.StatusCode)
		})
	})
}
