package organizations

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

// Service defines business operations on organizations.
type Service interface {
	Create(ctx context.Context, req CreateOrganizationRequest, currentUserID uuid.UUID) (*OrganizationResponse, error)
	FindAll(ctx context.Context, currentUserID uuid.UUID) ([]OrganizationResponse, error)
	FindOne(ctx context.Context, id uuid.UUID) (*OrganizationResponse, error)
	FindBySlug(ctx context.Context, slug string) (*OrganizationResponse, error)
	Update(ctx context.Context, id uuid.UUID, req UpdateOrganizationRequest, currentUserID uuid.UUID) (*OrganizationResponse, error)
	Archive(ctx context.Context, id, currentUserID uuid.UUID) error
	Delete(ctx context.Context, id, currentUserID uuid.UUID) error
	TransferOwnership(ctx context.Context, id, newOwnerID, currentUserID uuid.UUID) (*OrganizationResponse, error)
	GetStats(ctx context.Context, id uuid.UUID) (*OrganizationStatsResponse, error)
}

type service struct {
	repo Repository
}

// NewService constructs a new organizations domain service.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(ctx context.Context, req CreateOrganizationRequest, currentUserID uuid.UUID) (*OrganizationResponse, error) {
	ownerID := currentUserID
	if req.OwnerID != nil && *req.OwnerID != uuid.Nil {
		ownerID = *req.OwnerID
	}

	slug, err := s.generateUniqueSlug(ctx, req.Name, nil)
	if err != nil {
		return nil, err
	}

	orgID := uuid.New()
	var createdOrg *sqlc.Organization

	err = s.repo.WithTx(ctx, func(txRepo Repository) error {
		org, cErr := txRepo.Create(ctx, sqlc.CreateOrganizationParams{
			ID:          orgID,
			Name:        req.Name,
			Slug:        slug,
			Description: ptrToText(req.Description),
			Avatar:      ptrToText(req.Avatar),
			Website:     ptrToText(req.Website),
			Settings:    req.Settings,
			OwnerID:     ownerID,
			CreatedByID: pgtype.UUID{Bytes: currentUserID, Valid: true},
			UpdatedByID: pgtype.UUID{Bytes: currentUserID, Valid: true},
		})
		if cErr != nil {
			return cErr
		}
		createdOrg = org

		// Add owner to organization members
		_, mErr := txRepo.CreateMember(ctx, sqlc.CreateOrganizationMemberParams{
			ID:             uuid.New(),
			Role:           sqlc.RoleOWNER,
			UserID:         ownerID,
			OrganizationID: orgID,
			CreatedByID:    pgtype.UUID{Bytes: currentUserID, Valid: true},
			UpdatedByID:    pgtype.UUID{Bytes: currentUserID, Valid: true},
			IsDefault:      true,
		})
		if mErr != nil {
			return mErr
		}

		// Set default organization for user if not yet configured
		_ = txRepo.SetUserDefaultOrg(ctx, ownerID, orgID)

		// Create default workflow, statuses, and transitions
		_, wfErr := txRepo.CreateDefaultWorkflow(ctx, orgID, currentUserID)
		if wfErr != nil {
			return wfErr
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &OrganizationResponse{
		ID:          createdOrg.ID,
		Name:        createdOrg.Name,
		Slug:        createdOrg.Slug,
		Description: textToPtr(createdOrg.Description),
		Avatar:      textToPtr(createdOrg.Avatar),
		Website:     textToPtr(createdOrg.Website),
		Settings:    createdOrg.Settings,
		OwnerID:     createdOrg.OwnerID,
		Archive:     createdOrg.Archive,
		CreatedAt:   tsToTime(createdOrg.CreatedAt),
		UpdatedAt:   tsToTime(createdOrg.UpdatedAt),
	}, nil
}

func (s *service) FindAll(ctx context.Context, currentUserID uuid.UUID) ([]OrganizationResponse, error) {
	rows, err := s.repo.ListForUser(ctx, currentUserID)
	if err != nil {
		return nil, err
	}

	res := make([]OrganizationResponse, 0, len(rows))
	for _, r := range rows {
		roleStr := string(r.UserRole)
		res = append(res, OrganizationResponse{
			ID:          r.ID,
			Name:        r.Name,
			Slug:        r.Slug,
			Description: textToPtr(r.Description),
			Avatar:      textToPtr(r.Avatar),
			Website:     textToPtr(r.Website),
			Settings:    r.Settings,
			OwnerID:     r.OwnerID,
			Owner: &OrganizationOwnerDTO{
				ID:        r.OwnerUserID,
				FirstName: r.OwnerFirstName,
				LastName:  r.OwnerLastName,
				Email:     r.OwnerEmail,
			},
			Role:      &roleStr,
			Archive:   r.Archive,
			CreatedAt: tsToTime(r.CreatedAt),
			UpdatedAt: tsToTime(r.UpdatedAt),
		})
	}

	return res, nil
}

func (s *service) FindOne(ctx context.Context, id uuid.UUID) (*OrganizationResponse, error) {
	row, err := s.repo.GetDetailsByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, apperror.NotFound("Organization not found")
	}

	return &OrganizationResponse{
		ID:          row.ID,
		Name:        row.Name,
		Slug:        row.Slug,
		Description: textToPtr(row.Description),
		Avatar:      textToPtr(row.Avatar),
		Website:     textToPtr(row.Website),
		Settings:    row.Settings,
		OwnerID:     row.OwnerID,
		Owner: &OrganizationOwnerDTO{
			ID:        row.OwnerUserID,
			FirstName: row.OwnerFirstName,
			LastName:  row.OwnerLastName,
			Email:     row.OwnerEmail,
			Avatar:    textToPtr(row.OwnerAvatar),
		},
		Archive:   row.Archive,
		CreatedAt: tsToTime(row.CreatedAt),
		UpdatedAt: tsToTime(row.UpdatedAt),
	}, nil
}

func (s *service) FindBySlug(ctx context.Context, slug string) (*OrganizationResponse, error) {
	row, err := s.repo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, apperror.NotFound("Organization not found")
	}

	return &OrganizationResponse{
		ID:          row.ID,
		Name:        row.Name,
		Slug:        row.Slug,
		Description: textToPtr(row.Description),
		Avatar:      textToPtr(row.Avatar),
		Website:     textToPtr(row.Website),
		Settings:    row.Settings,
		OwnerID:     row.OwnerID,
		Owner: &OrganizationOwnerDTO{
			ID:        row.OwnerUserID,
			FirstName: row.OwnerFirstName,
			LastName:  row.OwnerLastName,
			Email:     row.OwnerEmail,
			Avatar:    textToPtr(row.OwnerAvatar),
		},
		Archive:   row.Archive,
		CreatedAt: tsToTime(row.CreatedAt),
		UpdatedAt: tsToTime(row.UpdatedAt),
	}, nil
}

func (s *service) Update(ctx context.Context, id uuid.UUID, req UpdateOrganizationRequest, currentUserID uuid.UUID) (*OrganizationResponse, error) {
	existing, err := s.repo.GetDetailsByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, apperror.NotFound("Organization not found")
	}

	var finalSlug string
	if req.Slug != nil && *req.Slug != "" && *req.Slug != existing.Slug {
		exists, cErr := s.repo.CheckSlugExists(ctx, *req.Slug, &id)
		if cErr != nil {
			return nil, cErr
		}
		if exists {
			return nil, apperror.Conflict(fmt.Sprintf("Slug \"%s\" is already taken. Please choose a different slug.", *req.Slug))
		}
		finalSlug = *req.Slug
	} else if req.Name != nil && *req.Name != existing.Name && (req.Slug == nil || *req.Slug == "") {
		genSlug, gErr := s.generateUniqueSlug(ctx, *req.Name, &id)
		if gErr != nil {
			return nil, gErr
		}
		finalSlug = genSlug
	}

	var nameStr string
	if req.Name != nil {
		nameStr = *req.Name
	}

	updated, err := s.repo.Update(ctx, sqlc.UpdateOrganizationParams{
		ID:          id,
		Column2:     nameStr,
		Column3:     finalSlug,
		Description: ptrToText(req.Description),
		Avatar:      ptrToText(req.Avatar),
		Website:     ptrToText(req.Website),
		Settings:    req.Settings,
		UpdatedByID: pgtype.UUID{Bytes: currentUserID, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	return &OrganizationResponse{
		ID:          updated.ID,
		Name:        updated.Name,
		Slug:        updated.Slug,
		Description: textToPtr(updated.Description),
		Avatar:      textToPtr(updated.Avatar),
		Website:     textToPtr(updated.Website),
		Settings:    updated.Settings,
		OwnerID:     updated.OwnerID,
		Archive:     updated.Archive,
		CreatedAt:   tsToTime(updated.CreatedAt),
		UpdatedAt:   tsToTime(updated.UpdatedAt),
	}, nil
}

func (s *service) Archive(ctx context.Context, id, currentUserID uuid.UUID) error {
	existing, err := s.repo.GetDetailsByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NotFound("Organization not found")
	}
	if existing.Archive {
		return apperror.Conflict("Organization is already archived")
	}

	return s.repo.Archive(ctx, id, currentUserID)
}

func (s *service) Delete(ctx context.Context, id, _ uuid.UUID) error {
	existing, err := s.repo.GetDetailsByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return apperror.NotFound("Organization not found")
	}

	return s.repo.WithTx(ctx, func(txRepo Repository) error {
		_ = txRepo.ClearUserDefaultOrg(ctx, id)
		return txRepo.Delete(ctx, id)
	})
}

func (s *service) TransferOwnership(ctx context.Context, id, newOwnerID, currentUserID uuid.UUID) (*OrganizationResponse, error) {
	org, err := s.repo.GetDetailsByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if org == nil {
		return nil, apperror.NotFound("Organization not found")
	}

	if org.OwnerID == newOwnerID {
		return nil, apperror.Forbidden("User is already the owner of this organization")
	}

	newOwner, err := s.repo.GetUserByID(ctx, newOwnerID)
	if err != nil {
		return nil, err
	}
	if newOwner == nil || newOwner.DeletedAt.Valid {
		return nil, apperror.NotFound("New owner user not found")
	}
	if newOwner.Status != sqlc.UserStatusACTIVE {
		return nil, apperror.Forbidden("Cannot transfer ownership to an inactive or suspended user")
	}

	err = s.repo.WithTx(ctx, func(txRepo Repository) error {
		// Check target user's current membership
		membership, mErr := txRepo.GetMember(ctx, newOwnerID, id)
		if mErr != nil {
			return mErr
		}

		if membership == nil {
			_, cErr := txRepo.CreateMember(ctx, sqlc.CreateOrganizationMemberParams{
				ID:             uuid.New(),
				Role:           sqlc.RoleOWNER,
				UserID:         newOwnerID,
				OrganizationID: id,
				CreatedByID:    pgtype.UUID{Bytes: currentUserID, Valid: true},
				UpdatedByID:    pgtype.UUID{Bytes: currentUserID, Valid: true},
				IsDefault:      false,
			})
			if cErr != nil {
				return cErr
			}
		} else if membership.Role != sqlc.RoleOWNER {
			uErr := txRepo.UpdateMemberRole(ctx, id, newOwnerID, sqlc.RoleOWNER, currentUserID)
			if uErr != nil {
				return uErr
			}
		}

		// Demote previous owner to MANAGER
		prevOwnerMember, pErr := txRepo.GetMember(ctx, org.OwnerID, id)
		if pErr == nil && prevOwnerMember != nil && prevOwnerMember.Role == sqlc.RoleOWNER {
			_ = txRepo.UpdateMemberRole(ctx, id, org.OwnerID, sqlc.RoleMANAGER, currentUserID)
		}

		return txRepo.TransferOwner(ctx, id, newOwnerID, currentUserID)
	})
	if err != nil {
		return nil, err
	}

	return s.FindOne(ctx, id)
}

func (s *service) GetStats(ctx context.Context, id uuid.UUID) (*OrganizationStatsResponse, error) {
	existing, err := s.repo.GetDetailsByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, apperror.NotFound("Organization not found")
	}

	return s.repo.GetStats(ctx, id)
}

func (s *service) generateUniqueSlug(ctx context.Context, name string, excludeID *uuid.UUID) (string, error) {
	baseSlug := slugify.Slugify(name)
	slug := baseSlug
	suffix := 1

	for {
		exists, err := s.repo.CheckSlugExists(ctx, slug, excludeID)
		if err != nil {
			return "", err
		}
		if !exists {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, suffix)
		suffix++
		if suffix > 100 {
			slug = fmt.Sprintf("%s-%d", baseSlug, time.Now().Unix())
			return slug, nil
		}
	}
}
