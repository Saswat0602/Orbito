package workspaces

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/slugify"
)

// Service defines business workflows on workspaces.
type Service interface {
	Create(ctx context.Context, req CreateWorkspaceRequest, currentUserID uuid.UUID) (*WorkspaceResponse, error)
	FindAll(ctx context.Context, currentUserID uuid.UUID, orgID *uuid.UUID, search string) ([]WorkspaceResponse, error)
	FindArchived(ctx context.Context, orgID, currentUserID uuid.UUID) ([]WorkspaceResponse, error)
	FindOne(ctx context.Context, id, currentUserID uuid.UUID) (*WorkspaceResponse, error)
	FindBySlug(ctx context.Context, orgID uuid.UUID, slug string, currentUserID uuid.UUID) (*WorkspaceResponse, error)
	FindBySlugOnly(ctx context.Context, slug string, currentUserID uuid.UUID) (*WorkspaceResponse, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateWorkspaceRequest, currentUserID uuid.UUID) (*WorkspaceResponse, error)
	Archive(ctx context.Context, id, currentUserID uuid.UUID) error
	Unarchive(ctx context.Context, id, currentUserID uuid.UUID) error
	Delete(ctx context.Context, id, currentUserID uuid.UUID) error
	GetMembers(ctx context.Context, id, currentUserID uuid.UUID) ([]WorkspaceMemberDTO, error)
	GetTree(ctx context.Context, orgID, currentUserID uuid.UUID) ([]map[string]any, error)
	GetCharts(ctx context.Context, id, currentUserID uuid.UUID) (*WorkspaceChartDataResponse, error)
}

type service struct {
	repo          Repository
	chartsService ChartsService
}

// NewService constructs a workspaces service instance.
func NewService(repo Repository, chartsService ChartsService) Service {
	return &service{
		repo:          repo,
		chartsService: chartsService,
	}
}

func textToPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func ptrToText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func tsToTime(t pgtype.Timestamp) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func (s *service) Create(ctx context.Context, req CreateWorkspaceRequest, currentUserID uuid.UUID) (*WorkspaceResponse, error) {
	slug := req.Slug
	if slug == "" {
		slug = slugify.Slugify(req.Name)
	}

	exists, err := s.repo.CheckSlugExists(ctx, req.OrganizationID, slug, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.Conflict(fmt.Sprintf("Workspace with slug '%s' already exists in this organization", slug))
	}

	wsID := uuid.New()
	var createdWs *sqlc.Workspace

	err = s.repo.WithTx(ctx, func(txRepo Repository) error {
		var parentUUID pgtype.UUID
		if req.ParentWorkspaceID != nil && *req.ParentWorkspaceID != uuid.Nil {
			parentUUID = pgtype.UUID{Bytes: *req.ParentWorkspaceID, Valid: true}
		}

		ws, cErr := txRepo.Create(ctx, sqlc.CreateWorkspaceParams{
			ID:                wsID,
			Name:              req.Name,
			Slug:              slug,
			Description:       ptrToText(req.Description),
			Avatar:            ptrToText(req.Avatar),
			Color:             ptrToText(req.Color),
			Settings:          req.Settings,
			OrganizationID:    req.OrganizationID,
			ParentWorkspaceID: parentUUID,
			Path:              pgtype.Text{Valid: false},
			CreatedByID:       pgtype.UUID{Bytes: currentUserID, Valid: true},
			UpdatedByID:       pgtype.UUID{Bytes: currentUserID, Valid: true},
		})
		if cErr != nil {
			return cErr
		}
		createdWs = ws

		// Creator is automatically assigned as OWNER of the workspace
		_, mErr := txRepo.CreateMember(ctx, sqlc.CreateWorkspaceMemberDirectParams{
			ID:          uuid.New(),
			Role:        sqlc.RoleOWNER,
			UserID:      currentUserID,
			WorkspaceID: wsID,
			CreatedByID: pgtype.UUID{Bytes: currentUserID, Valid: true},
			UpdatedByID: pgtype.UUID{Bytes: currentUserID, Valid: true},
		})
		if mErr != nil {
			return mErr
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	var parentID *uuid.UUID
	if createdWs.ParentWorkspaceID.Valid {
		id := uuid.UUID(createdWs.ParentWorkspaceID.Bytes)
		parentID = &id
	}

	return &WorkspaceResponse{
		ID:                createdWs.ID,
		Name:              createdWs.Name,
		Slug:              createdWs.Slug,
		Description:       textToPtr(createdWs.Description),
		Avatar:            textToPtr(createdWs.Avatar),
		Color:             textToPtr(createdWs.Color),
		Settings:          createdWs.Settings,
		OrganizationID:    createdWs.OrganizationID,
		ParentWorkspaceID: parentID,
		Path:              textToPtr(createdWs.Path),
		Archive:           createdWs.Archive,
		CreatedAt:         tsToTime(createdWs.CreatedAt),
		UpdatedAt:         tsToTime(createdWs.UpdatedAt),
	}, nil
}

func (s *service) FindAll(ctx context.Context, currentUserID uuid.UUID, orgID *uuid.UUID, search string) ([]WorkspaceResponse, error) {
	rows, err := s.repo.ListForUser(ctx, currentUserID, orgID, search)
	if err != nil {
		return nil, err
	}

	res := make([]WorkspaceResponse, 0, len(rows))
	for _, r := range rows {
		roleStr := string(r.UserRole)
		var parentID *uuid.UUID
		if r.ParentWorkspaceID.Valid {
			id := uuid.UUID(r.ParentWorkspaceID.Bytes)
			parentID = &id
		}
		res = append(res, WorkspaceResponse{
			ID:                r.ID,
			Name:              r.Name,
			Slug:              r.Slug,
			Description:       textToPtr(r.Description),
			Avatar:            textToPtr(r.Avatar),
			Color:             textToPtr(r.Color),
			Settings:          r.Settings,
			OrganizationID:    r.OrganizationID,
			ParentWorkspaceID: parentID,
			Path:              textToPtr(r.Path),
			Role:              &roleStr,
			Archive:           r.Archive,
			CreatedAt:         tsToTime(r.CreatedAt),
			UpdatedAt:         tsToTime(r.UpdatedAt),
		})
	}

	return res, nil
}

func (s *service) FindArchived(ctx context.Context, orgID, _ uuid.UUID) ([]WorkspaceResponse, error) {
	rows, err := s.repo.ListArchived(ctx, orgID)
	if err != nil {
		return nil, err
	}

	res := make([]WorkspaceResponse, 0, len(rows))
	for _, r := range rows {
		var parentID *uuid.UUID
		if r.ParentWorkspaceID.Valid {
			id := uuid.UUID(r.ParentWorkspaceID.Bytes)
			parentID = &id
		}
		res = append(res, WorkspaceResponse{
			ID:                r.ID,
			Name:              r.Name,
			Slug:              r.Slug,
			Description:       textToPtr(r.Description),
			Avatar:            textToPtr(r.Avatar),
			Color:             textToPtr(r.Color),
			Settings:          r.Settings,
			OrganizationID:    r.OrganizationID,
			ParentWorkspaceID: parentID,
			Path:              textToPtr(r.Path),
			Archive:           r.Archive,
			CreatedAt:         tsToTime(r.CreatedAt),
			UpdatedAt:         tsToTime(r.UpdatedAt),
		})
	}

	return res, nil
}

func (s *service) FindOne(ctx context.Context, id, _ uuid.UUID) (*WorkspaceResponse, error) {
	row, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, apperror.NotFound("Workspace not found")
	}

	var parentID *uuid.UUID
	if row.ParentWorkspaceID.Valid {
		pID := uuid.UUID(row.ParentWorkspaceID.Bytes)
		parentID = &pID
	}

	return &WorkspaceResponse{
		ID:                row.ID,
		Name:              row.Name,
		Slug:              row.Slug,
		Description:       textToPtr(row.Description),
		Avatar:            textToPtr(row.Avatar),
		Color:             textToPtr(row.Color),
		Settings:          row.Settings,
		OrganizationID:    row.OrganizationID,
		ParentWorkspaceID: parentID,
		Path:              textToPtr(row.Path),
		Archive:           row.Archive,
		CreatedAt:         tsToTime(row.CreatedAt),
		UpdatedAt:         tsToTime(row.UpdatedAt),
	}, nil
}

func (s *service) FindBySlug(ctx context.Context, orgID uuid.UUID, slug string, _ uuid.UUID) (*WorkspaceResponse, error) {
	row, err := s.repo.GetByOrgAndSlug(ctx, orgID, slug)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, apperror.NotFound("Workspace not found")
	}

	var parentID *uuid.UUID
	if row.ParentWorkspaceID.Valid {
		pID := uuid.UUID(row.ParentWorkspaceID.Bytes)
		parentID = &pID
	}

	return &WorkspaceResponse{
		ID:                row.ID,
		Name:              row.Name,
		Slug:              row.Slug,
		Description:       textToPtr(row.Description),
		Avatar:            textToPtr(row.Avatar),
		Color:             textToPtr(row.Color),
		Settings:          row.Settings,
		OrganizationID:    row.OrganizationID,
		ParentWorkspaceID: parentID,
		Path:              textToPtr(row.Path),
		Archive:           row.Archive,
		CreatedAt:         tsToTime(row.CreatedAt),
		UpdatedAt:         tsToTime(row.UpdatedAt),
	}, nil
}

func (s *service) FindBySlugOnly(ctx context.Context, slug string, _ uuid.UUID) (*WorkspaceResponse, error) {
	row, err := s.repo.GetBySlugOnly(ctx, slug)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, apperror.NotFound("Workspace not found")
	}

	var parentID *uuid.UUID
	if row.ParentWorkspaceID.Valid {
		pID := uuid.UUID(row.ParentWorkspaceID.Bytes)
		parentID = &pID
	}

	return &WorkspaceResponse{
		ID:                row.ID,
		Name:              row.Name,
		Slug:              row.Slug,
		Description:       textToPtr(row.Description),
		Avatar:            textToPtr(row.Avatar),
		Color:             textToPtr(row.Color),
		Settings:          row.Settings,
		OrganizationID:    row.OrganizationID,
		ParentWorkspaceID: parentID,
		Path:              textToPtr(row.Path),
		Archive:           row.Archive,
		CreatedAt:         tsToTime(row.CreatedAt),
		UpdatedAt:         tsToTime(row.UpdatedAt),
	}, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req UpdateWorkspaceRequest, currentUserID uuid.UUID) (*WorkspaceResponse, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, apperror.NotFound("Workspace not found")
	}

	var finalSlug string
	if req.Slug != nil && *req.Slug != "" && *req.Slug != existing.Slug {
		exists, cErr := s.repo.CheckSlugExists(ctx, existing.OrganizationID, *req.Slug, &id)
		if cErr != nil {
			return nil, cErr
		}
		if exists {
			return nil, apperror.Conflict(fmt.Sprintf("Slug '%s' is already taken in this organization", *req.Slug))
		}
		finalSlug = *req.Slug
	}

	var nameStr string
	if req.Name != nil {
		nameStr = *req.Name
	}

	updated, err := s.repo.Update(ctx, sqlc.UpdateWorkspaceParams{
		ID:          id,
		Column2:     nameStr,
		Column3:     finalSlug,
		Description: ptrToText(req.Description),
		Avatar:      ptrToText(req.Avatar),
		Color:       ptrToText(req.Color),
		Settings:    req.Settings,
		UpdatedByID: pgtype.UUID{Bytes: currentUserID, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	var parentID *uuid.UUID
	if updated.ParentWorkspaceID.Valid {
		pID := uuid.UUID(updated.ParentWorkspaceID.Bytes)
		parentID = &pID
	}

	return &WorkspaceResponse{
		ID:                updated.ID,
		Name:              updated.Name,
		Slug:              updated.Slug,
		Description:       textToPtr(updated.Description),
		Avatar:            textToPtr(updated.Avatar),
		Color:             textToPtr(updated.Color),
		Settings:          updated.Settings,
		OrganizationID:    updated.OrganizationID,
		ParentWorkspaceID: parentID,
		Path:              textToPtr(updated.Path),
		Archive:           updated.Archive,
		CreatedAt:         tsToTime(updated.CreatedAt),
		UpdatedAt:         tsToTime(updated.UpdatedAt),
	}, nil
}

func (s *service) Archive(ctx context.Context, id, currentUserID uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NotFound("Workspace not found")
	}
	if existing.Archive {
		return apperror.Conflict("Workspace is already archived")
	}

	return s.repo.Archive(ctx, id, currentUserID)
}

func (s *service) Unarchive(ctx context.Context, id, currentUserID uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NotFound("Workspace not found")
	}
	if !existing.Archive {
		return apperror.Conflict("Workspace is not archived")
	}

	return s.repo.Unarchive(ctx, id, currentUserID)
}

func (s *service) Delete(ctx context.Context, id, _ uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NotFound("Workspace not found")
	}

	return s.repo.Delete(ctx, id)
}

func (s *service) GetMembers(ctx context.Context, id, _ uuid.UUID) ([]WorkspaceMemberDTO, error) {
	rows, err := s.repo.ListMembersDetails(ctx, id)
	if err != nil {
		return nil, err
	}

	res := make([]WorkspaceMemberDTO, 0, len(rows))
	for _, r := range rows {
		res = append(res, WorkspaceMemberDTO{
			ID:          r.ID,
			Role:        string(r.Role),
			JoinedAt:    tsToTime(r.JoinedAt),
			UserID:      r.UserID,
			WorkspaceID: r.WorkspaceID,
			UserEmail:   r.UserEmail,
			FirstName:   r.UserFirstName,
			LastName:    r.UserLastName,
			Avatar:      textToPtr(r.UserAvatar),
		})
	}

	return res, nil
}

func (s *service) GetTree(ctx context.Context, orgID, _ uuid.UUID) ([]map[string]any, error) {
	rows, err := s.repo.ListTree(ctx, orgID)
	if err != nil {
		return nil, err
	}

	res := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		var parentID any
		if r.ParentWorkspaceID.Valid {
			parentID = uuid.UUID(r.ParentWorkspaceID.Bytes).String()
		}
		res = append(res, map[string]any{
			"id":                r.ID.String(),
			"name":              r.Name,
			"slug":              r.Slug,
			"parentWorkspaceId": parentID,
			"path":              textToPtr(r.Path),
			"archive":           r.Archive,
		})
	}

	return res, nil
}

func (s *service) GetCharts(ctx context.Context, id, _ uuid.UUID) (*WorkspaceChartDataResponse, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, apperror.NotFound("Workspace not found")
	}

	return s.chartsService.GetChartData(ctx, id)
}
