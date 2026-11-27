package users

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/crypto"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// PresenceChecker reports live WebSocket online state and last-seen timestamps.
type PresenceChecker interface {
	IsUserOnline(userID string) bool
	GetUserLastSeen(userID string) (string, bool)
}

// Service provides business logic for user management, directory, and status operations.
type Service struct {
	repo     Repository
	presence PresenceChecker
}

// NewService constructs a new users service.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// SetPresenceChecker attaches a live presence tracker (e.g., WebSocket gateway).
func (s *Service) SetPresenceChecker(pc PresenceChecker) {
	s.presence = pc
}

// CheckUsersExist checks if any users exist in the system.
func (s *Service) CheckUsersExist(ctx context.Context) (bool, error) {
	if s.repo == nil {
		return false, nil
	}
	count, err := s.repo.CountUsers(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check users exist: %w", err)
	}
	return count > 0, nil
}

// Create creates a new user account with administrative privileges.
func (s *Service) Create(ctx context.Context, dto CreateUserDTO) (*ResponseDTO, error) {
	emailClean := strings.ToLower(strings.TrimSpace(dto.Email))
	existing, err := s.repo.GetUserByEmail(ctx, emailClean)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup user: %w", err)
	}
	if existing != nil {
		return nil, apperror.Conflict("Email or username already exists.")
	}

	pwdHash, err := crypto.HashPassword(dto.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	role := sqlc.RoleMEMBER
	if dto.Role != nil && *dto.Role != "" {
		role = sqlc.Role(*dto.Role)
	}

	status := sqlc.UserStatusACTIVE
	if dto.Status != nil && *dto.Status != "" {
		status = sqlc.UserStatus(*dto.Status)
	}

	var uname string
	if dto.Username != nil && *dto.Username != "" {
		uname = strings.TrimSpace(*dto.Username)
		existingUname, _ := s.repo.GetUserByUsername(ctx, uname)
		if existingUname != nil {
			return nil, apperror.Conflict("Email or username already exists.")
		}
	} else {
		baseUsername := strings.ToLower(strings.Split(emailClean, "@")[0])
		uname = baseUsername
		counter := 1
		for {
			existingUname, _ := s.repo.GetUserByUsername(ctx, uname)
			if existingUname == nil {
				break
			}
			uname = fmt.Sprintf("%s%d", baseUsername, counter)
			counter++
		}
	}

	tz := "UTC"
	if dto.Timezone != nil && *dto.Timezone != "" {
		tz = *dto.Timezone
	}
	lang := "en"
	if dto.Language != nil && *dto.Language != "" {
		lang = *dto.Language
	}

	var av, bio, mob pgtype.Text
	if dto.Avatar != nil {
		av = pgtype.Text{String: *dto.Avatar, Valid: true}
	}
	if dto.Bio != nil {
		bio = pgtype.Text{String: *dto.Bio, Valid: true}
	}
	if dto.MobileNumber != nil {
		mob = pgtype.Text{String: *dto.MobileNumber, Valid: true}
	}

	row, err := s.repo.CreateUserWithProfile(ctx, sqlc.CreateUserWithProfileParams{
		ID:           uuid.New(),
		Email:        emailClean,
		Username:     pgtype.Text{String: uname, Valid: true},
		FirstName:    strings.TrimSpace(dto.FirstName),
		LastName:     strings.TrimSpace(dto.LastName),
		Password:     pgtype.Text{String: pwdHash, Valid: true},
		Role:         role,
		Status:       status,
		Timezone:     tz,
		Language:     lang,
		MobileNumber: mob,
		Avatar:       av,
		Bio:          bio,
		Source:       sqlc.UserSourceMANUAL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	res := mapCreateProfileRowToResponse(*row)
	return &res, nil
}

// FindAll retrieves paginated users.
func (s *Service) FindAll(ctx context.Context, limit, offset int32) ([]ResponseDTO, int64, error) {
	rows, err := s.repo.ListUsers(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	total, err := s.repo.CountUsers(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	results := make([]ResponseDTO, len(rows))
	for i, r := range rows {
		results[i] = mapListRowToResponse(r)
	}

	return results, total, nil
}

// FindOne returns user details, asserting permission.
func (s *Service) FindOne(ctx context.Context, id uuid.UUID, requesterID uuid.UUID, requesterRole string) (*ResponseDTO, error) {
	if requesterRole != string(sqlc.RoleSUPERADMIN) && requesterID != id {
		return nil, apperror.Forbidden("Profile access not allowed")
	}

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user: %w", err)
	}
	if user == nil {
		return nil, apperror.NotFound("User not found.")
	}

	res := mapUserToResponse(*user)
	return &res, nil
}

// Update modifies user details enforcing self-update privilege restrictions.
func (s *Service) Update(ctx context.Context, id uuid.UUID, dto UpdateUserDTO, requesterID uuid.UUID, requesterRole string) (*ResponseDTO, error) {
	isSuperAdmin := requesterRole == string(sqlc.RoleSUPERADMIN)
	if !isSuperAdmin && requesterID != id {
		return nil, apperror.Forbidden("You can only update your own profile")
	}

	if !isSuperAdmin {
		if dto.Role != nil || dto.Status != nil || dto.EmailVerified != nil {
			return nil, apperror.Forbidden("You cannot change your own role, status, or emailVerified. An administrator manages these fields.")
		}
	}

	var fn, ln, un, av, bio, tz, lang, mob string
	if dto.FirstName != nil {
		fn = *dto.FirstName
	}
	if dto.LastName != nil {
		ln = *dto.LastName
	}
	if dto.Username != nil {
		un = *dto.Username
	}
	if dto.Avatar != nil {
		av = *dto.Avatar
	}
	if dto.Bio != nil {
		bio = *dto.Bio
	}
	if dto.Timezone != nil {
		tz = *dto.Timezone
	}
	if dto.Language != nil {
		lang = *dto.Language
	}
	if dto.MobileNumber != nil {
		mob = *dto.MobileNumber
	}

	row, err := s.repo.UpdateUserProfile(ctx, sqlc.UpdateUserProfileParams{
		ID:           id,
		FirstName:    fn,
		LastName:     ln,
		Username:     un,
		Avatar:       av,
		Bio:          bio,
		Timezone:     tz,
		Language:     lang,
		MobileNumber: mob,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update user profile: %w", err)
	}

	if isSuperAdmin && dto.Status != nil && *dto.Status != "" {
		st := sqlc.UserStatus(*dto.Status)
		_, _ = s.repo.UpdateUserStatus(ctx, id, st)
		row.Status = st
	}

	res := mapUpdateRowToResponse(*row)
	return &res, nil
}

// UpdateStatus changes an account's active/inactive/suspended state (admin only).
func (s *Service) UpdateStatus(ctx context.Context, id uuid.UUID, status sqlc.UserStatus, requesterRole string) (*ResponseDTO, error) {
	if requesterRole != string(sqlc.RoleSUPERADMIN) {
		return nil, apperror.Forbidden("Only administrators can update account status")
	}

	row, err := s.repo.UpdateUserStatus(ctx, id, status)
	if err != nil {
		return nil, fmt.Errorf("failed to update user status: %w", err)
	}

	res := mapStatusRowToResponse(*row)
	return &res, nil
}

// ChangePassword updates user password verifying the current password first.
func (s *Service) ChangePassword(ctx context.Context, id uuid.UUID, dto ChangePasswordDTO) error {
	if dto.NewPassword != dto.ConfirmPassword {
		return apperror.BadRequest("New password and confirm password do not match")
	}

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil || user == nil {
		return apperror.NotFound("User not found.")
	}

	if !user.Password.Valid || !crypto.CheckPasswordHash(dto.CurrentPassword, user.Password.String) {
		return apperror.BadRequest("Current password is not correct")
	}

	if crypto.CheckPasswordHash(dto.NewPassword, user.Password.String) {
		return apperror.BadRequest("New password must be different from current password")
	}

	pwdHash, err := crypto.HashPassword(dto.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	return s.repo.UpdatePassword(ctx, id, pgtype.Text{String: pwdHash, Valid: true})
}

// Remove soft-deletes a user account.
func (s *Service) Remove(ctx context.Context, id, deletedBy uuid.UUID, requesterRole string) error {
	if requesterRole != string(sqlc.RoleSUPERADMIN) {
		return apperror.Forbidden("Only administrators can delete user accounts")
	}

	return s.repo.DeleteUser(ctx, id, deletedBy)
}

// GetOnlineStatus returns online state for user.
func (s *Service) GetOnlineStatus(ctx context.Context, id uuid.UUID) OnlineStatusDTO {
	isOnline := false
	var lastSeen string
	if s.presence != nil {
		isOnline = s.presence.IsUserOnline(id.String())
		if ts, ok := s.presence.GetUserLastSeen(id.String()); ok && ts != "" {
			lastSeen = ts
		}
	}
	if lastSeen == "" {
		user, err := s.repo.GetUserByID(ctx, id)
		lastSeen = time.Now().UTC().Format(httpx.ISO8601Format)
		if err == nil && user != nil && user.LastLoginAt.Valid {
			lastSeen = user.LastLoginAt.Time.UTC().Format(httpx.ISO8601Format)
		}
	}
	return OnlineStatusDTO{
		IsOnline: isOnline,
		LastSeen: &lastSeen,
	}
}

// GetPublicProfile retrieves public user profile for shared organization members.
func (s *Service) GetPublicProfile(ctx context.Context, identifier string, requesterID uuid.UUID, requesterRole string) (*PublicProfileDTO, error) {
	var targetUser *sqlc.GetPublicUserByIDRow
	var av, mob, bio *string
	var idStr, firstName, lastName, email, timezone, role string

	if targetUUID, err := uuid.Parse(identifier); err == nil {
		row, err := s.repo.GetPublicUserByID(ctx, targetUUID)
		if err == nil && row != nil {
			targetUser = row
			idStr = row.ID.String()
			firstName = row.FirstName
			lastName = row.LastName
			email = row.Email
			timezone = row.Timezone
			role = string(row.Role)
			if row.Avatar.Valid {
				av = &row.Avatar.String
			}
			if row.MobileNumber.Valid {
				mob = &row.MobileNumber.String
			}
			if row.Bio.Valid {
				bio = &row.Bio.String
			}
		}
	}

	if targetUser == nil {
		row, err := s.repo.GetPublicUserByUsername(ctx, identifier)
		if err != nil || row == nil {
			return nil, apperror.NotFound("User not found")
		}
		idStr = row.ID.String()
		firstName = row.FirstName
		lastName = row.LastName
		email = row.Email
		timezone = row.Timezone
		role = string(row.Role)
		if row.Avatar.Valid {
			av = &row.Avatar.String
		}
		if row.MobileNumber.Valid {
			mob = &row.MobileNumber.String
		}
		if row.Bio.Valid {
			bio = &row.Bio.String
		}
	}

	parsedID, _ := uuid.Parse(idStr)
	if requesterRole != string(sqlc.RoleSUPERADMIN) && requesterID != parsedID {
		isShared, err := s.repo.CheckUsersShareOrganization(ctx, parsedID, requesterID)
		if err != nil || !isShared {
			return nil, apperror.Forbidden("Profile access not allowed. You must belong to the same organization.")
		}
	}

	return &PublicProfileDTO{
		ID:           idStr,
		FirstName:    firstName,
		LastName:     lastName,
		Email:        email,
		Avatar:       av,
		MobileNumber: mob,
		Bio:          bio,
		Timezone:     timezone,
		Role:         role,
	}, nil
}

func mapUserToResponse(u sqlc.User) ResponseDTO {
	var uname, avatar, bio, mobile, lastLogin *string
	if u.Username.Valid {
		uname = &u.Username.String
	}
	if u.Avatar.Valid {
		avatar = &u.Avatar.String
	}
	if u.Bio.Valid {
		bio = &u.Bio.String
	}
	if u.MobileNumber.Valid {
		mobile = &u.MobileNumber.String
	}
	if u.LastLoginAt.Valid {
		ll := u.LastLoginAt.Time.Format(httpx.ISO8601Format)
		lastLogin = &ll
	}

	return ResponseDTO{
		ID:           u.ID.String(),
		Email:        u.Email,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Username:     uname,
		Role:         string(u.Role),
		Status:       string(u.Status),
		Avatar:       avatar,
		Bio:          bio,
		Timezone:     u.Timezone,
		Language:     u.Language,
		MobileNumber: mobile,
		LastLoginAt:  lastLogin,
		CreatedAt:    u.CreatedAt.Time.Format(httpx.ISO8601Format),
		UpdatedAt:    u.UpdatedAt.Time.Format(httpx.ISO8601Format),
	}
}

func mapListRowToResponse(u sqlc.ListUsersRow) ResponseDTO {
	var uname, avatar, bio, mobile, lastLogin *string
	if u.Username.Valid {
		uname = &u.Username.String
	}
	if u.Avatar.Valid {
		avatar = &u.Avatar.String
	}
	if u.Bio.Valid {
		bio = &u.Bio.String
	}
	if u.MobileNumber.Valid {
		mobile = &u.MobileNumber.String
	}
	if u.LastLoginAt.Valid {
		ll := u.LastLoginAt.Time.Format(httpx.ISO8601Format)
		lastLogin = &ll
	}

	return ResponseDTO{
		ID:           u.ID.String(),
		Email:        u.Email,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Username:     uname,
		Role:         string(u.Role),
		Status:       string(u.Status),
		Avatar:       avatar,
		Bio:          bio,
		Timezone:     u.Timezone,
		Language:     u.Language,
		MobileNumber: mobile,
		LastLoginAt:  lastLogin,
		CreatedAt:    u.CreatedAt.Time.Format(httpx.ISO8601Format),
		UpdatedAt:    u.UpdatedAt.Time.Format(httpx.ISO8601Format),
	}
}

func mapUpdateRowToResponse(u sqlc.UpdateUserProfileRow) ResponseDTO {
	var uname, avatar, bio, mobile, lastLogin *string
	if u.Username.Valid {
		uname = &u.Username.String
	}
	if u.Avatar.Valid {
		avatar = &u.Avatar.String
	}
	if u.Bio.Valid {
		bio = &u.Bio.String
	}
	if u.MobileNumber.Valid {
		mobile = &u.MobileNumber.String
	}
	if u.LastLoginAt.Valid {
		ll := u.LastLoginAt.Time.Format(httpx.ISO8601Format)
		lastLogin = &ll
	}

	return ResponseDTO{
		ID:           u.ID.String(),
		Email:        u.Email,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Username:     uname,
		Role:         string(u.Role),
		Status:       string(u.Status),
		Avatar:       avatar,
		Bio:          bio,
		Timezone:     u.Timezone,
		Language:     u.Language,
		MobileNumber: mobile,
		LastLoginAt:  lastLogin,
		CreatedAt:    u.CreatedAt.Time.Format(httpx.ISO8601Format),
		UpdatedAt:    u.UpdatedAt.Time.Format(httpx.ISO8601Format),
	}
}

func mapStatusRowToResponse(u sqlc.UpdateUserStatusRow) ResponseDTO {
	var uname, avatar, bio, mobile, lastLogin *string
	if u.Username.Valid {
		uname = &u.Username.String
	}
	if u.Avatar.Valid {
		avatar = &u.Avatar.String
	}
	if u.Bio.Valid {
		bio = &u.Bio.String
	}
	if u.MobileNumber.Valid {
		mobile = &u.MobileNumber.String
	}
	if u.LastLoginAt.Valid {
		ll := u.LastLoginAt.Time.Format(httpx.ISO8601Format)
		lastLogin = &ll
	}

	return ResponseDTO{
		ID:           u.ID.String(),
		Email:        u.Email,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Username:     uname,
		Role:         string(u.Role),
		Status:       string(u.Status),
		Avatar:       avatar,
		Bio:          bio,
		Timezone:     u.Timezone,
		Language:     u.Language,
		MobileNumber: mobile,
		LastLoginAt:  lastLogin,
		CreatedAt:    u.CreatedAt.Time.Format(httpx.ISO8601Format),
		UpdatedAt:    u.UpdatedAt.Time.Format(httpx.ISO8601Format),
	}
}

func mapCreateProfileRowToResponse(u sqlc.CreateUserWithProfileRow) ResponseDTO {
	var uname, avatar, bio, mobile, lastLogin *string
	if u.Username.Valid {
		uname = &u.Username.String
	}
	if u.Avatar.Valid {
		avatar = &u.Avatar.String
	}
	if u.Bio.Valid {
		bio = &u.Bio.String
	}
	if u.MobileNumber.Valid {
		mobile = &u.MobileNumber.String
	}
	if u.LastLoginAt.Valid {
		ll := u.LastLoginAt.Time.Format(httpx.ISO8601Format)
		lastLogin = &ll
	}

	return ResponseDTO{
		ID:           u.ID.String(),
		Email:        u.Email,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Username:     uname,
		Role:         string(u.Role),
		Status:       string(u.Status),
		Avatar:       avatar,
		Bio:          bio,
		Timezone:     u.Timezone,
		Language:     u.Language,
		MobileNumber: mobile,
		LastLoginAt:  lastLogin,
		CreatedAt:    u.CreatedAt.Time.Format(httpx.ISO8601Format),
		UpdatedAt:    u.UpdatedAt.Time.Format(httpx.ISO8601Format),
	}
}
