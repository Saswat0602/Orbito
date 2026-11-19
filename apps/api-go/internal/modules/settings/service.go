package settings

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/crypto"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Service provides business logic for managing settings.
type Service struct {
	repo   Repository
	crypto *crypto.FieldCrypto
}

// NewService constructs a new settings service.
func NewService(repo Repository, crypto *crypto.FieldCrypto) *Service {
	return &Service{
		repo:   repo,
		crypto: crypto,
	}
}

// Get retrieves a setting value, prioritizing user-specific settings over global ones.
func (s *Service) Get(ctx context.Context, key string, userID *uuid.UUID, defaultValue *string) (*string, error) {
	if s.repo == nil {
		return defaultValue, nil
	}
	if userID != nil {
		userSetting, err := s.repo.GetByUserAndKey(ctx, *userID, key)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch user setting: %w", err)
		}
		if userSetting != nil {
			val := s.extractValue(userSetting.Value, userSetting.IsEncrypted)
			return val, nil
		}
	}

	globalSetting, err := s.repo.GetGlobalByKey(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch global setting: %w", err)
	}

	if globalSetting != nil {
		val := s.extractValue(globalSetting.Value, globalSetting.IsEncrypted)
		return val, nil
	}

	return defaultValue, nil
}

// Set creates or updates a setting for a specific user or globally.
func (s *Service) Set(ctx context.Context, dto SetSettingDTO, userID *uuid.UUID) error {
	if s.repo == nil {
		return fmt.Errorf("database repository is not initialized")
	}
	finalVal := dto.Value
	isEncrypted := dto.IsEncrypted != nil && *dto.IsEncrypted
	if isEncrypted && finalVal != "" && s.crypto != nil {
		encrypted, err := s.crypto.Encrypt(finalVal)
		if err != nil {
			return fmt.Errorf("encryption failed: %w", err)
		}
		finalVal = encrypted
	}

	category := "general"
	if dto.Category != nil && *dto.Category != "" {
		category = *dto.Category
	}

	var desc pgtype.Text
	if dto.Description != nil {
		desc = pgtype.Text{String: *dto.Description, Valid: true}
	}

	valText := pgtype.Text{String: finalVal, Valid: true}

	if userID != nil {
		_, err := s.repo.UpsertUserSetting(ctx, sqlc.UpsertUserSettingParams{
			ID:          uuid.New(),
			Key:         dto.Key,
			Value:       valText,
			Description: desc,
			Category:    category,
			IsEncrypted: isEncrypted,
			UserId:      pgtype.UUID{Bytes: *userID, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("failed to upsert user setting: %w", err)
		}
		return nil
	}

	// Global setting (userId is null)
	existing, err := s.repo.GetGlobalByKey(ctx, dto.Key)
	if err != nil {
		return fmt.Errorf("failed to check existing global setting: %w", err)
	}

	if existing != nil {
		_, err = s.repo.UpdateGlobalSetting(ctx, sqlc.UpdateGlobalSettingByIDParams{
			ID:          existing.ID,
			Value:       valText,
			Description: desc,
			Category:    category,
			IsEncrypted: isEncrypted,
		})
		if err != nil {
			return fmt.Errorf("failed to update global setting: %w", err)
		}
	} else {
		_, err = s.repo.CreateGlobalSetting(ctx, sqlc.CreateGlobalSettingParams{
			ID:          uuid.New(),
			Key:         dto.Key,
			Value:       valText,
			Description: desc,
			Category:    category,
			IsEncrypted: isEncrypted,
		})
		if err != nil {
			return fmt.Errorf("failed to create global setting: %w", err)
		}
	}

	return nil
}

// BulkSet sets multiple settings atomically for a user.
func (s *Service) BulkSet(ctx context.Context, settings []SetSettingDTO, userID uuid.UUID) error {
	if s.repo == nil {
		return fmt.Errorf("database repository is not initialized")
	}
	return s.repo.WithTx(ctx, func(txRepo Repository) error {
		txSvc := &Service{repo: txRepo, crypto: s.crypto}
		for _, setting := range settings {
			if err := txSvc.Set(ctx, setting, &userID); err != nil {
				return err
			}
		}
		return nil
	})
}

// GetAll returns user settings merged with non-conflicting global settings.
func (s *Service) GetAll(ctx context.Context, userID *uuid.UUID, category *string) ([]SettingResponse, error) {
	if s.repo == nil {
		return []SettingResponse{}, nil
	}
	catFilter := ""
	if category != nil {
		catFilter = *category
	}

	userKeys := make(map[string]struct{})
	result := make([]SettingResponse, 0)

	if userID != nil {
		userSettings, err := s.repo.ListUserSettings(ctx, *userID, catFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to list user settings: %w", err)
		}

		for _, row := range userSettings {
			userKeys[row.Key] = struct{}{}
			val := s.extractValue(row.Value, row.IsEncrypted)
			var desc *string
			if row.Description.Valid {
				d := row.Description.String
				desc = &d
			}
			result = append(result, SettingResponse{
				Key:            row.Key,
				Value:          val,
				Description:    desc,
				Category:       row.Category,
				IsUserSpecific: true,
				IsEncrypted:    row.IsEncrypted,
			})
		}
	}

	globalSettings, err := s.repo.ListGlobalSettings(ctx, catFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list global settings: %w", err)
	}

	for _, row := range globalSettings {
		if _, exists := userKeys[row.Key]; exists {
			continue // User setting overrides global setting
		}
		val := s.extractValue(row.Value, row.IsEncrypted)
		var desc *string
		if row.Description.Valid {
			d := row.Description.String
			desc = &d
		}
		result = append(result, SettingResponse{
			Key:            row.Key,
			Value:          val,
			Description:    desc,
			Category:       row.Category,
			IsUserSpecific: false,
			IsEncrypted:    row.IsEncrypted,
		})
	}

	return result, nil
}

// GetGlobalSettings returns all global settings (userId IS NULL).
func (s *Service) GetGlobalSettings(ctx context.Context, category *string) ([]SettingResponse, error) {
	if s.repo == nil {
		return []SettingResponse{}, nil
	}
	catFilter := ""
	if category != nil {
		catFilter = *category
	}

	settings, err := s.repo.ListGlobalSettings(ctx, catFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to list global settings: %w", err)
	}

	result := make([]SettingResponse, len(settings))
	for i, row := range settings {
		val := s.extractValue(row.Value, row.IsEncrypted)
		var desc *string
		if row.Description.Valid {
			d := row.Description.String
			desc = &d
		}
		result[i] = SettingResponse{
			Key:            row.Key,
			Value:          val,
			Description:    desc,
			Category:       row.Category,
			IsUserSpecific: false,
			IsEncrypted:    row.IsEncrypted,
		}
	}

	return result, nil
}

// Delete removes a user-specific setting or a global setting if userID is nil.
func (s *Service) Delete(ctx context.Context, key string, userID *uuid.UUID) error {
	if s.repo == nil {
		return nil
	}
	if userID != nil {
		return s.repo.DeleteUserSetting(ctx, *userID, key)
	}
	return s.repo.DeleteGlobalSetting(ctx, key)
}

// Exists checks if a setting exists.
func (s *Service) Exists(ctx context.Context, key string, userID *uuid.UUID) (bool, error) {
	if s.repo == nil {
		return false, nil
	}
	if userID != nil {
		count, err := s.repo.CountUserSetting(ctx, *userID, key)
		if err != nil {
			return false, err
		}
		if count > 0 {
			return true, nil
		}
	}

	count, err := s.repo.CountGlobalSetting(ctx, key)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetRegistrationStatus queries settings for registration flags matching NestJS AuthController.
func (s *Service) GetRegistrationStatus(ctx context.Context) (*RegistrationStatusResponse, error) {
	regVal, err := s.Get(ctx, "registration_enabled", nil, nil)
	if err != nil {
		return nil, err
	}

	allowOrgVal, err := s.Get(ctx, "allow_organization_creation", nil, nil)
	if err != nil {
		return nil, err
	}

	defaultOrgVal, err := s.Get(ctx, "default_organization_id", nil, nil)
	if err != nil {
		return nil, err
	}

	enabled := true
	if regVal != nil && *regVal == "false" {
		enabled = false
	}

	allowOrgs := true
	if allowOrgVal != nil && *allowOrgVal == "false" {
		allowOrgs = false
	}

	return &RegistrationStatusResponse{
		Enabled:            enabled,
		AllowOrganizations: allowOrgs,
		DefaultOrgID:       defaultOrgVal,
	}, nil
}

func (s *Service) extractValue(text pgtype.Text, isEncrypted bool) *string {
	if !text.Valid {
		return nil
	}
	val := text.String
	if isEncrypted && s.crypto != nil {
		val = s.crypto.Decrypt(val)
	}
	return &val
}
