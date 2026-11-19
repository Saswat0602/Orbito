package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/modules/email"
	"github.com/orbito/api-go/internal/modules/settings"
	"github.com/orbito/api-go/internal/platform/crypto"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
)

// Service provides business logic for user authentication, registration, password lifecycle, and setup.
type Service struct {
	repo        Repository
	tokenSvc    *TokenService
	emailSvc    *email.Service
	settingsSvc *settings.Service
	frontendURL string
}

// NewService constructs a new auth service.
func NewService(
	repo Repository,
	tokenSvc *TokenService,
	emailSvc *email.Service,
	settingsSvc *settings.Service,
	frontendURL string,
) *Service {
	if frontendURL == "" {
		frontendURL = "http://localhost:3001"
	}
	return &Service{
		repo:        repo,
		tokenSvc:    tokenSvc,
		emailSvc:    emailSvc,
		settingsSvc: settingsSvc,
		frontendURL: frontendURL,
	}
}

// Login authenticates a user by email and password.
func (s *Service) Login(ctx context.Context, dto LoginDTO) (*ResponseDTO, error) {
	user, err := s.repo.GetUserByEmail(ctx, strings.ToLower(strings.TrimSpace(dto.Email)))
	if err != nil {
		return nil, fmt.Errorf("failed to lookup user: %w", err)
	}
	if user == nil {
		return nil, apperror.Unauthorized("Invalid credentials")
	}

	if !user.Password.Valid || !crypto.CheckPasswordHash(dto.Password, user.Password.String) {
		return nil, apperror.Unauthorized("Invalid credentials")
	}

	if err := AssertAccountActive(user.Status, user.DeletedAt); err != nil {
		return nil, err
	}

	if err := AssertNotSystemUser(user.ID); err != nil {
		return nil, err
	}

	// Bootstrap rule: first user on a fresh instance without any super admin becomes super admin
	if user.Role != sqlc.RoleSUPERADMIN {
		adminCount, countErr := s.repo.CountSuperAdmins(ctx)
		if countErr == nil && adminCount == 0 {
			_ = s.repo.UpdateRole(ctx, user.ID, sqlc.RoleSUPERADMIN)
			user.Role = sqlc.RoleSUPERADMIN
		}
	}

	_ = s.repo.UpdateLastLogin(ctx, user.ID)

	return s.generateAuthResponse(ctx, user)
}

// Register creates a new user account with unique username and default organization join.
func (s *Service) Register(ctx context.Context, dto RegisterDTO) (*ResponseDTO, error) {
	emailClean := strings.ToLower(strings.TrimSpace(dto.Email))

	// Check if public registration is enabled in settings
	if s.settingsSvc != nil {
		regStatus, err := s.settingsSvc.GetRegistrationStatus(ctx)
		if err == nil && !regStatus.Enabled {
			hasValidInvite := false
			if dto.InvitationToken != nil && *dto.InvitationToken != "" {
				inv, invErr := s.repo.FindInvitationByToken(ctx, *dto.InvitationToken)
				if invErr == nil && inv != nil &&
					inv.Status == sqlc.InvitationStatusPENDING &&
					inv.ExpiresAt.Time.After(time.Now()) &&
					strings.EqualFold(inv.InviteeEmail, emailClean) {
					hasValidInvite = true
				}
			}
			if !hasValidInvite {
				return nil, apperror.BadRequest("User registration is currently disabled. Please contact your administrator to get an account.")
			}
		}
	}

	// Check existing email
	existing, err := s.repo.GetUserByEmail(ctx, emailClean)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existing != nil {
		return nil, apperror.Conflict("User with this email already exists")
	}

	// Auto-assign unique username
	baseUsername := strings.Split(emailClean, "@")[0]
	if dto.Username != nil && *dto.Username != "" {
		baseUsername = strings.ToLower(strings.TrimSpace(*dto.Username))
	}
	finalUsername := baseUsername
	counter := 1
	for {
		existingByUname, uErr := s.repo.GetUserByUsername(ctx, finalUsername)
		if uErr != nil {
			return nil, fmt.Errorf("failed to check username uniqueness: %w", uErr)
		}
		if existingByUname == nil {
			break
		}
		finalUsername = fmt.Sprintf("%s%d", baseUsername, counter)
		counter++
	}

	// Hash password with bcrypt cost 12
	pwdHash, err := crypto.HashPassword(dto.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	userID := uuid.New()
	newUser, err := s.repo.CreateUser(ctx, sqlc.CreateUserParams{
		ID:        userID,
		Email:     emailClean,
		Username:  pgtype.Text{String: finalUsername, Valid: true},
		FirstName: strings.TrimSpace(dto.FirstName),
		LastName:  strings.TrimSpace(dto.LastName),
		Password:  pgtype.Text{String: pwdHash, Valid: true},
		Role:      sqlc.RoleMEMBER,
		Status:    sqlc.UserStatusACTIVE,
		Timezone:  "UTC",
		Language:  "en",
		Source:    sqlc.UserSourceMANUAL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Auto-join default organization if configured
	s.autoJoinDefaultOrg(ctx, newUser.ID)

	return s.generateAuthResponse(ctx, newUser)
}

// RefreshToken validates a refresh token and mints rotated tokens.
func (s *Service) RefreshToken(ctx context.Context, refreshTokenStr string) (*ResponseDTO, error) {
	claims, err := s.tokenSvc.ValidateRefreshToken(refreshTokenStr)
	if err != nil {
		return nil, apperror.Unauthorized("Invalid or expired refresh token")
	}

	userUUID, err := uuid.Parse(claims.Sub)
	if err != nil {
		return nil, apperror.Unauthorized("Invalid user identity")
	}

	user, err := s.repo.GetUserByID(ctx, userUUID)
	if err != nil || user == nil {
		return nil, apperror.Unauthorized("User not found")
	}

	if err := AssertAccountActive(user.Status, user.DeletedAt); err != nil {
		return nil, err
	}

	if !user.RefreshToken.Valid || user.RefreshToken.String != refreshTokenStr {
		return nil, apperror.Unauthorized("Invalid refresh token")
	}

	return s.generateAuthResponse(ctx, user)
}

// Logout invalidates the active refresh token.
func (s *Service) Logout(ctx context.Context, userID uuid.UUID) error {
	return s.repo.UpdateRefreshToken(ctx, userID, pgtype.Text{Valid: false})
}

// ForgotPassword creates a reset token and queues a recovery email.
func (s *Service) ForgotPassword(ctx context.Context, emailStr string) error {
	emailClean := strings.ToLower(strings.TrimSpace(emailStr))
	user, err := s.repo.GetUserByEmail(ctx, emailClean)
	if err != nil {
		return fmt.Errorf("failed to lookup user: %w", err)
	}
	if user == nil {
		return apperror.NotFound("User not found")
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("failed to generate random token: %w", err)
	}
	resetToken := hex.EncodeToString(tokenBytes)
	expiry := pgtype.Timestamp{Time: time.Now().Add(24 * time.Hour), Valid: true}

	if err := s.repo.SetResetToken(ctx, user.ID, pgtype.Text{String: resetToken, Valid: true}, expiry); err != nil {
		return fmt.Errorf("failed to save reset token: %w", err)
	}

	if s.emailSvc != nil {
		resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.frontendURL, resetToken)
		_ = s.emailSvc.SendEmail(ctx, email.SendEmailDTO{
			To:       user.Email,
			Subject:  "Orbito Password Reset Request",
			Template: "password-reset",
			Data:     map[string]any{"resetLink": resetLink},
		})
	}

	return nil
}

// VerifyResetToken checks if a reset token is valid and unexpired.
func (s *Service) VerifyResetToken(ctx context.Context, token string) (bool, error) {
	user, err := s.repo.GetUserByResetToken(ctx, token)
	if err != nil {
		return false, err
	}
	return user != nil, nil
}

// ResetPassword verifies the token and applies the new password.
func (s *Service) ResetPassword(ctx context.Context, dto ResetPasswordDTO) error {
	if dto.Password != dto.ConfirmPassword {
		return apperror.BadRequest("Passwords do not match")
	}

	user, err := s.repo.GetUserByResetToken(ctx, dto.Token)
	if err != nil || user == nil {
		return apperror.BadRequest("Invalid or expired reset token")
	}

	pwdHash, err := crypto.HashPassword(dto.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.repo.UpdatePassword(ctx, user.ID, pgtype.Text{String: pwdHash, Valid: true}); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	if s.emailSvc != nil {
		_ = s.emailSvc.SendEmail(ctx, email.SendEmailDTO{
			To:       user.Email,
			Subject:  "Password Successfully Reset",
			Template: "password-reset-confirmation",
			Data:     map[string]any{},
		})
	}

	return nil
}

// IsSetupRequired checks if the platform requires initial super admin setup.
func (s *Service) IsSetupRequired(ctx context.Context) (required bool, canSetup bool, message string, err error) {
	count, err := s.repo.CountSuperAdmins(ctx)
	if err != nil {
		return false, false, "", fmt.Errorf("failed to count super admins: %w", err)
	}

	if count == 0 {
		return true, true, "Setup is required. No super admin exists.", nil
	}
	return false, false, "Setup has already been completed.", nil
}

// SetupSuperAdmin performs one-time initialization of the first super admin.
func (s *Service) SetupSuperAdmin(ctx context.Context, dto SetupAdminDTO) (*ResponseDTO, error) {
	required, _, _, err := s.IsSetupRequired(ctx)
	if err != nil {
		return nil, err
	}
	if !required {
		return nil, apperror.Conflict("System setup has already been completed")
	}

	emailClean := strings.ToLower(strings.TrimSpace(dto.Email))
	username := strings.Split(emailClean, "@")[0]
	if dto.Username != nil && *dto.Username != "" {
		username = strings.ToLower(strings.TrimSpace(*dto.Username))
	}

	pwdHash, err := crypto.HashPassword(dto.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	adminUser, err := s.repo.CreateUser(ctx, sqlc.CreateUserParams{
		ID:        uuid.New(),
		Email:     emailClean,
		Username:  pgtype.Text{String: username, Valid: true},
		FirstName: strings.TrimSpace(dto.FirstName),
		LastName:  strings.TrimSpace(dto.LastName),
		Password:  pgtype.Text{String: pwdHash, Valid: true},
		Role:      sqlc.RoleSUPERADMIN,
		Status:    sqlc.UserStatusACTIVE,
		Timezone:  "UTC",
		Language:  "en",
		Source:    sqlc.UserSourceMANUAL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create super admin: %w", err)
	}

	return s.generateAuthResponse(ctx, adminUser)
}

// GetOIDCConfig returns the OpenID Connect SSO configuration status.
func (s *Service) GetOIDCConfig(_ context.Context) OIDCConfigResponseDTO {
	return OIDCConfigResponseDTO{
		Configured:   false,
		ProviderName: "OpenID Connect",
	}
}

// GetUserByID loads user entity for middleware or profile endpoints.
func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*sqlc.User, error) {
	return s.repo.GetUserByID(ctx, id)
}

func (s *Service) generateAuthResponse(ctx context.Context, user *sqlc.User) (*ResponseDTO, error) {
	accessToken, err := s.tokenSvc.MintAccessToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to mint access token: %w", err)
	}

	refreshToken, err := s.tokenSvc.MintRefreshToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to mint refresh token: %w", err)
	}

	_ = s.repo.UpdateRefreshToken(ctx, user.ID, pgtype.Text{String: refreshToken, Valid: true})

	var uname, avatar, bio, mobile, lang, tz *string
	if user.Username.Valid {
		uname = &user.Username.String
	}
	if user.Avatar.Valid {
		avatar = &user.Avatar.String
	}
	if user.Bio.Valid {
		bio = &user.Bio.String
	}
	if user.MobileNumber.Valid {
		mobile = &user.MobileNumber.String
	}
	if user.Language != "" {
		lang = &user.Language
	}
	if user.Timezone != "" {
		tz = &user.Timezone
	}

	return &ResponseDTO{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: UserDTO{
			ID:           user.ID.String(),
			Email:        user.Email,
			FirstName:    user.FirstName,
			LastName:     user.LastName,
			Username:     uname,
			Role:         string(user.Role),
			Avatar:       avatar,
			Bio:          bio,
			MobileNumber: mobile,
			Language:     lang,
			Timezone:     tz,
		},
	}, nil
}

func (s *Service) autoJoinDefaultOrg(ctx context.Context, userID uuid.UUID) {
	if s.settingsSvc == nil {
		return
	}
	defOrgIDStr, err := s.settingsSvc.Get(ctx, "default_organization_id", nil, nil)
	if err != nil || defOrgIDStr == nil || *defOrgIDStr == "" {
		return
	}

	orgUUID, err := uuid.Parse(*defOrgIDStr)
	if err != nil {
		return
	}

	org, err := s.repo.GetOrganizationByID(ctx, orgUUID)
	if err != nil || org == nil {
		return
	}

	_ = s.repo.AddOrganizationMember(ctx, uuid.New(), userID, orgUUID, sqlc.RoleMEMBER)
}
