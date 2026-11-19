package admin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/config"
	"github.com/orbito/api-go/internal/modules/email"
	"github.com/orbito/api-go/internal/modules/settings"
	"github.com/orbito/api-go/internal/platform/crypto"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
	"github.com/orbito/api-go/internal/platform/mailer"
	"github.com/orbito/api-go/internal/shared/apperror"
	"github.com/orbito/api-go/internal/shared/httpx"
)

// Service coordinates administrator capabilities across users, orgs, and configuration.
type Service struct {
	repo        Repository
	settingsSvc *settings.Service
	emailSvc    *email.Service
	cfg         *config.Config
}

// NewService constructs a new admin service.
func NewService(repo Repository, settingsSvc *settings.Service, emailSvc *email.Service, cfg *config.Config) *Service {
	return &Service{
		repo:        repo,
		settingsSvc: settingsSvc,
		emailSvc:    emailSvc,
		cfg:         cfg,
	}
}

// GetDashboardStats aggregates system-wide statistics for the admin dashboard.
func (s *Service) GetDashboardStats(ctx context.Context) (*DashboardStatsDTO, error) {
	if s.repo == nil {
		return &DashboardStatsDTO{}, nil
	}
	since := time.Now().AddDate(0, 0, -7)
	stats, err := s.repo.GetDashboardStats(ctx, since)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate dashboard stats: %w", err)
	}

	return &DashboardStatsDTO{
		TotalUsers:          stats.TotalUsers,
		TotalOrganizations:  stats.TotalOrganizations,
		TotalWorkspaces:     stats.TotalWorkspaces,
		TotalProjects:       stats.TotalProjects,
		TotalTasks:          stats.TotalTasks,
		NewUsersThisWeek:    stats.NewUsersThisWeek,
		NewOrgsThisWeek:     stats.NewOrgsThisWeek,
		NewProjectsThisWeek: stats.NewProjectsThisWeek,
		NewTasksThisWeek:    stats.NewTasksThisWeek,
		ActiveUsers:         stats.ActiveUsers,
	}, nil
}

// GetUsers retrieves a paginated and filterable list of users.
func (s *Service) GetUsers(ctx context.Context, page, limit int, search, status, role string) (*PaginatedUsersDTO, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	//nolint:gosec // limit and offset are bounded above
	rows, err := s.repo.ListUsers(ctx, search, sqlc.UserStatus(status), sqlc.Role(role), int32(limit), int32(offset))
	if err != nil {
		return nil, fmt.Errorf("failed to list admin users: %w", err)
	}

	total, err := s.repo.CountUsers(ctx, search, sqlc.UserStatus(status), sqlc.Role(role))
	if err != nil {
		return nil, fmt.Errorf("failed to count admin users: %w", err)
	}

	items := make([]UserListItemDTO, len(rows))
	for i, r := range rows {
		var av, ll *string
		if r.Avatar.Valid {
			av = &r.Avatar.String
		}
		if r.LastLoginAt.Valid {
			ts := r.LastLoginAt.Time.Format(httpx.ISO8601Format)
			ll = &ts
		}

		items[i] = UserListItemDTO{
			ID:          r.ID.String(),
			Email:       r.Email,
			FirstName:   r.FirstName,
			LastName:    r.LastName,
			Avatar:      av,
			Role:        string(r.Role),
			Status:      string(r.Status),
			LastLoginAt: ll,
			CreatedAt:   r.CreatedAt.Time.Format(httpx.ISO8601Format),
			Count: UserOrgCountDTO{
				OrganizationMembers: r.OrganizationCount,
			},
		}
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return &PaginatedUsersDTO{
		Data:       items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// GetUserDetail returns comprehensive details of a single user.
func (s *Service) GetUserDetail(ctx context.Context, id uuid.UUID) (*UserDetailDTO, error) {
	u, err := s.repo.GetUserDetail(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup user: %w", err)
	}
	if u == nil {
		return nil, apperror.NotFound("User not found")
	}

	orgRows, err := s.repo.GetUserOrganizations(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup user organizations: %w", err)
	}

	var av, bio, ll *string
	if u.Avatar.Valid {
		av = &u.Avatar.String
	}
	if u.Bio.Valid {
		bio = &u.Bio.String
	}
	if u.LastLoginAt.Valid {
		ts := u.LastLoginAt.Time.Format(httpx.ISO8601Format)
		ll = &ts
	}

	memberships := make([]UserOrgMembershipDetail, len(orgRows))
	for i, o := range orgRows {
		memberships[i] = UserOrgMembershipDetail{
			Role: string(o.Role),
			Organization: OrganizationBrief{
				ID:   o.ID.String(),
				Name: o.Name,
				Slug: o.Slug,
				Counts: OrganizationCount{
					Members:    o.MembersCount,
					Workspaces: o.WorkspacesCount,
				},
			},
		}
	}

	return &UserDetailDTO{
		ID:                  u.ID.String(),
		Email:               u.Email,
		FirstName:           u.FirstName,
		LastName:            u.LastName,
		Avatar:              av,
		Bio:                 bio,
		Timezone:            u.Timezone,
		Language:            u.Language,
		Role:                string(u.Role),
		Status:              string(u.Status),
		LastLoginAt:         ll,
		EmailVerified:       u.EmailVerified,
		CreatedAt:           u.CreatedAt.Time.Format(httpx.ISO8601Format),
		UpdatedAt:           u.UpdatedAt.Time.Format(httpx.ISO8601Format),
		Source:              string(u.Source),
		OrganizationMembers: memberships,
	}, nil
}

// UpdateUserRole changes a user's role, prohibiting self-demotion or removing the last super admin.
func (s *Service) UpdateUserRole(ctx context.Context, id uuid.UUID, role string, currentUserID uuid.UUID) (map[string]any, error) {
	if id == currentUserID {
		return nil, apperror.Forbidden("Cannot change your own role")
	}

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil || user == nil {
		return nil, apperror.NotFound("User not found")
	}

	if user.Role == sqlc.RoleSUPERADMIN && role != string(sqlc.RoleSUPERADMIN) {
		superAdminCount, countErr := s.repo.CountSuperAdmins(ctx)
		if countErr != nil {
			return nil, fmt.Errorf("failed to count super admins: %w", countErr)
		}
		if superAdminCount <= 1 {
			return nil, apperror.Forbidden("Cannot demote the last super admin")
		}
	}

	updated, err := s.repo.UpdateUserRole(ctx, id, sqlc.Role(role))
	if err != nil {
		return nil, fmt.Errorf("failed to update user role: %w", err)
	}

	return map[string]any{
		"id":        updated.ID.String(),
		"email":     updated.Email,
		"firstName": updated.FirstName,
		"lastName":  updated.LastName,
		"role":      string(updated.Role),
	}, nil
}

// UpdateUserStatus updates an account's active state, invalidating sessions on deactivation.
func (s *Service) UpdateUserStatus(ctx context.Context, id uuid.UUID, status string, currentUserID uuid.UUID) (map[string]any, error) {
	if id == currentUserID {
		return nil, apperror.Forbidden("Cannot change your own status")
	}

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil || user == nil {
		return nil, apperror.NotFound("User not found")
	}

	updated, err := s.repo.UpdateUserStatus(ctx, id, sqlc.UserStatus(status))
	if err != nil {
		return nil, fmt.Errorf("failed to update user status: %w", err)
	}

	return map[string]any{
		"id":        updated.ID.String(),
		"email":     updated.Email,
		"firstName": updated.FirstName,
		"lastName":  updated.LastName,
		"status":    string(updated.Status),
	}, nil
}

// DeleteUser deletes an account after asserting safety rules.
func (s *Service) DeleteUser(ctx context.Context, id, currentUserID uuid.UUID) (map[string]any, error) {
	if id == currentUserID {
		return nil, apperror.Forbidden("Cannot delete your own account")
	}

	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil || user == nil {
		return nil, apperror.NotFound("User not found")
	}

	if user.Role == sqlc.RoleSUPERADMIN {
		superAdminCount, countErr := s.repo.CountSuperAdmins(ctx)
		if countErr != nil {
			return nil, fmt.Errorf("failed to check super admin count: %w", countErr)
		}
		if superAdminCount <= 1 {
			return nil, apperror.Forbidden("Cannot delete the last super admin")
		}
	}

	ownedOrgs, err := s.repo.GetUserOwnedOrganizations(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to check owned organizations: %w", err)
	}
	if len(ownedOrgs) > 0 {
		names := make([]string, len(ownedOrgs))
		for i, o := range ownedOrgs {
			names[i] = o.Name
		}
		return nil, apperror.Forbidden(fmt.Sprintf("Cannot delete user who owns organizations: %s. Transfer ownership first.", strings.Join(names, ", ")))
	}

	if err := s.repo.SoftDeleteUser(ctx, id, currentUserID); err != nil {
		return nil, fmt.Errorf("failed to soft delete user: %w", err)
	}

	_ = s.repo.RemoveUserMemberships(ctx, id)

	return map[string]any{
		"success": true,
		"message": fmt.Sprintf("User \"%s %s\" (%s) has been deleted.", user.FirstName, user.LastName, user.Email),
	}, nil
}

// ResetUserPassword creates a password reset token and dispatches reset notification.
func (s *Service) ResetUserPassword(ctx context.Context, id uuid.UUID) (map[string]any, error) {
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil || user == nil {
		return nil, apperror.NotFound("User not found")
	}

	tokenBytes := make([]byte, 32)
	if _, randErr := rand.Read(tokenBytes); randErr != nil {
		return nil, fmt.Errorf("failed to generate random token: %w", randErr)
	}
	plainToken := hex.EncodeToString(tokenBytes)

	hashedToken, err := crypto.HashPasswordWithCost(plainToken, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to hash reset token: %w", err)
	}

	expiry := time.Now().Add(24 * time.Hour)
	err = s.repo.SetUserResetToken(ctx, id,
		pgtype.Text{String: hashedToken, Valid: true},
		pgtype.Timestamp{Time: expiry, Valid: true},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save reset token: %w", err)
	}

	frontendURL := "http://localhost:3000"
	if s.cfg != nil && len(s.cfg.CORSOrigins) > 0 && s.cfg.CORSOrigins[0] != "" {
		frontendURL = s.cfg.CORSOrigins[0]
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, plainToken)
	if s.emailSvc != nil {
		_ = s.emailSvc.SendEmail(ctx, email.SendEmailDTO{
			To:       user.Email,
			Subject:  "Password Reset Request",
			Template: "password-reset",
			Data: map[string]any{
				"userName":   user.FirstName,
				"resetToken": plainToken,
				"resetUrl":   resetURL,
			},
		})
	}

	return map[string]any{
		"success":   true,
		"resetLink": fmt.Sprintf("/reset-password?token=%s", plainToken),
		"message":   fmt.Sprintf("Password reset email sent to %s. Valid for 24 hours.", user.Email),
	}, nil
}

// GetOrganizations returns a paginated list of organizations.
func (s *Service) GetOrganizations(ctx context.Context, page, limit int, search string) (*PaginatedOrgsDTO, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	//nolint:gosec // limit and offset are bounded above
	rows, err := s.repo.ListOrganizations(ctx, search, int32(limit), int32(offset))
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	total, err := s.repo.CountOrganizations(ctx, search)
	if err != nil {
		return nil, fmt.Errorf("failed to count organizations: %w", err)
	}

	items := make([]OrgListItemDTO, len(rows))
	for i, r := range rows {
		var desc *string
		if r.Description.Valid {
			desc = &r.Description.String
		}

		var owner *UserOwnerSummary
		if r.OwnerID.Valid {
			ownerID := uuid.UUID(r.OwnerID.Bytes)
			owner = &UserOwnerSummary{
				ID:        ownerID.String(),
				FirstName: r.OwnerFirstName.String,
				LastName:  r.OwnerLastName.String,
				Email:     r.OwnerEmail.String,
			}
		}

		items[i] = OrgListItemDTO{
			ID:          r.ID.String(),
			Name:        r.Name,
			Slug:        r.Slug,
			Description: desc,
			Archive:     r.Archive,
			CreatedAt:   r.CreatedAt.Time.Format(httpx.ISO8601Format),
			Owner:       owner,
			Counts: OrganizationCount{
				Members:    r.MembersCount,
				Workspaces: r.WorkspacesCount,
			},
		}
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(limit)))
	}

	return &PaginatedOrgsDTO{
		Data:       items,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// GetOrganizationDetail returns full details of a single organization.
func (s *Service) GetOrganizationDetail(ctx context.Context, id uuid.UUID) (*OrgDetailDTO, error) {
	o, err := s.repo.GetOrganizationDetail(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup organization: %w", err)
	}
	if o == nil {
		return nil, apperror.NotFound("Organization not found")
	}

	members, err := s.repo.GetOrganizationMembers(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve members: %w", err)
	}

	workspaces, err := s.repo.GetOrganizationWorkspaces(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve workspaces: %w", err)
	}

	var desc *string
	if o.Description.Valid {
		desc = &o.Description.String
	}

	var owner *UserOwnerSummary
	if o.OwnerID.Valid {
		ownerID := uuid.UUID(o.OwnerID.Bytes)
		owner = &UserOwnerSummary{
			ID:        ownerID.String(),
			FirstName: o.OwnerFirstName.String,
			LastName:  o.OwnerLastName.String,
			Email:     o.OwnerEmail.String,
		}
	}

	memberItems := make([]OrgMemberItemDTO, len(members))
	for i, m := range members {
		var av *string
		if m.Avatar.Valid {
			av = &m.Avatar.String
		}
		memberItems[i] = OrgMemberItemDTO{
			Role:      string(m.Role),
			CreatedAt: m.CreatedAt.Time.Format(httpx.ISO8601Format),
			JoinedAt:  m.JoinedAt.Time.Format(httpx.ISO8601Format),
			User: UserMemberDetail{
				ID:        m.UserID.String(),
				FirstName: m.FirstName,
				LastName:  m.LastName,
				Email:     m.Email,
				Avatar:    av,
				Status:    string(m.Status),
			},
		}
	}

	wsItems := make([]OrgWorkspaceDTO, len(workspaces))
	for i, w := range workspaces {
		wsItems[i] = OrgWorkspaceDTO{
			ID:   w.ID.String(),
			Name: w.Name,
			Slug: w.Slug,
			Counts: WorkspaceCount{
				Projects: w.ProjectsCount,
				Members:  w.MembersCount,
			},
		}
	}

	return &OrgDetailDTO{
		ID:          o.ID.String(),
		Name:        o.Name,
		Slug:        o.Slug,
		Description: desc,
		Archive:     o.Archive,
		CreatedAt:   o.CreatedAt.Time.Format(httpx.ISO8601Format),
		UpdatedAt:   o.UpdatedAt.Time.Format(httpx.ISO8601Format),
		Owner:       owner,
		Members:     memberItems,
		Workspaces:  wsItems,
		Counts: OrganizationCount{
			Members:    o.MembersCount,
			Workspaces: o.WorkspacesCount,
		},
	}, nil
}

// ToggleOrganizationArchive suspends or activates an organization.
func (s *Service) ToggleOrganizationArchive(ctx context.Context, id uuid.UUID) (map[string]any, error) {
	org, err := s.repo.GetOrganizationDetail(ctx, id)
	if err != nil || org == nil {
		return nil, apperror.NotFound("Organization not found")
	}

	updated, err := s.repo.ToggleOrganizationArchive(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to toggle archive: %w", err)
	}

	return map[string]any{
		"id":      updated.ID.String(),
		"name":    updated.Name,
		"archive": updated.Archive,
	}, nil
}

// DeleteOrganization permanently deletes an organization and clears associations.
func (s *Service) DeleteOrganization(ctx context.Context, id uuid.UUID) (map[string]any, error) {
	org, err := s.repo.GetOrganizationDetail(ctx, id)
	if err != nil || org == nil {
		return nil, apperror.NotFound("Organization not found")
	}

	_ = s.repo.ClearUsersDefaultOrg(ctx, id)

	if s.settingsSvc != nil {
		defaultOrg, _ := s.settingsSvc.Get(ctx, "default_organization_id", nil, nil)
		if defaultOrg != nil && *defaultOrg == id.String() {
			empty := ""
			cat := "registration"
			_ = s.settingsSvc.Set(ctx, settings.SetSettingDTO{
				Key:      "default_organization_id",
				Value:    empty,
				Category: &cat,
			}, nil)
		}
	}

	if err := s.repo.DeleteOrganization(ctx, id); err != nil {
		return nil, fmt.Errorf("failed to delete organization: %w", err)
	}

	return map[string]any{
		"success": true,
		"message": fmt.Sprintf("Organization \"%s\" deleted", org.Name),
	}, nil
}

// TransferOrganizationOwnership transfers organization ownership to a new active user.
func (s *Service) TransferOrganizationOwnership(ctx context.Context, orgID, newOwnerID uuid.UUID) (map[string]any, error) {
	org, err := s.repo.GetOrganizationDetail(ctx, orgID)
	if err != nil || org == nil {
		return nil, apperror.NotFound("Organization not found")
	}

	if org.OwnerID.Valid && uuid.UUID(org.OwnerID.Bytes) == newOwnerID {
		return nil, apperror.Forbidden("User is already the owner of this organization")
	}

	newOwner, err := s.repo.GetUserByID(ctx, newOwnerID)
	if err != nil || newOwner == nil {
		return nil, apperror.NotFound("New owner user not found")
	}
	if newOwner.Status != sqlc.UserStatusACTIVE {
		return nil, apperror.Forbidden("Cannot transfer ownership to an inactive or suspended user")
	}

	// Membership check
	membership, _ := s.repo.GetOrganizationMember(ctx, newOwnerID, orgID)
	if membership == nil {
		_ = s.repo.CreateOrganizationMember(ctx, uuid.New(), sqlc.RoleOWNER, newOwnerID, orgID, newOwnerID)
	} else if membership.Role != sqlc.RoleOWNER {
		_ = s.repo.UpdateOrganizationMemberRole(ctx, membership.ID, sqlc.RoleOWNER)
	}

	// Demote prior owner to MANAGER
	if org.OwnerID.Valid && uuid.UUID(org.OwnerID.Bytes) != newOwnerID {
		prevOwnerID := uuid.UUID(org.OwnerID.Bytes)
		prevMember, _ := s.repo.GetOrganizationMember(ctx, prevOwnerID, orgID)
		if prevMember != nil && prevMember.Role == sqlc.RoleOWNER {
			_ = s.repo.UpdateOrganizationMemberRole(ctx, prevMember.ID, sqlc.RoleMANAGER)
		}
	}

	updated, err := s.repo.TransferOrgOwnership(ctx, orgID, newOwnerID)
	if err != nil {
		return nil, fmt.Errorf("failed to transfer organization ownership: %w", err)
	}

	return map[string]any{
		"id":   updated.ID.String(),
		"name": updated.Name,
		"owner": map[string]any{
			"id":        newOwner.ID.String(),
			"firstName": newOwner.FirstName,
			"lastName":  newOwner.LastName,
			"email":     newOwner.Email,
		},
	}, nil
}

// GetSystemConfig returns all settings with environment source breakdown.
func (s *Service) GetSystemConfig(ctx context.Context, category string) (*ConfigResponseDTO, error) {
	var catPtr *string
	if category != "" {
		catPtr = &category
	}

	var globalSettings any
	if s.settingsSvc != nil {
		res, err := s.settingsSvc.GetGlobalSettings(ctx, catPtr)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve global settings: %w", err)
		}
		globalSettings = res
	}

	smtpSources := s.GetSmtpConfigWithSource(ctx)

	return &ConfigResponseDTO{
		Settings:    globalSettings,
		SmtpSources: smtpSources,
	}, nil
}

// SaveSystemConfig persists global system settings.
func (s *Service) SaveSystemConfig(ctx context.Context, items []SettingItemDTO) error {
	if s.settingsSvc == nil {
		return nil
	}

	for _, setting := range items {
		if setting.IsEncrypted && setting.Value == "••••••••" {
			continue
		}

		var desc, cat *string
		if setting.Description != "" {
			desc = &setting.Description
		}
		if setting.Category != "" {
			cat = &setting.Category
		}
		isEnc := setting.IsEncrypted

		err := s.settingsSvc.Set(ctx, settings.SetSettingDTO{
			Key:         setting.Key,
			Value:       setting.Value,
			Description: desc,
			Category:    cat,
			IsEncrypted: &isEnc,
		}, nil)
		if err != nil {
			return fmt.Errorf("failed to save setting '%s': %w", setting.Key, err)
		}
	}

	return nil
}

// GetSmtpConfigWithSource reports the origin of each SMTP setting (env vs database).
func (s *Service) GetSmtpConfigWithSource(ctx context.Context) []SmtpSourceDTO {
	fields := []struct {
		key    string
		envKey string
	}{
		{key: "smtp_host", envKey: "SMTP_HOST"},
		{key: "smtp_port", envKey: "SMTP_PORT"},
		{key: "smtp_user", envKey: "SMTP_USER"},
		{key: "smtp_pass", envKey: "SMTP_PASS"},
		{key: "smtp_from", envKey: "SMTP_FROM"},
	}

	result := make([]SmtpSourceDTO, 0, len(fields))

	for _, field := range fields {
		var envVal string
		if s.cfg != nil {
			switch field.key {
			case "smtp_host":
				envVal = s.cfg.SMTPHost
			case "smtp_port":
				if s.cfg.SMTPPort > 0 {
					envVal = strconv.Itoa(s.cfg.SMTPPort)
				}
			case "smtp_user":
				envVal = s.cfg.SMTPUser
			case "smtp_pass":
				envVal = s.cfg.SMTPPass
			case "smtp_from":
				envVal = s.cfg.SMTPFrom
			}
		}

		var dbVal *string
		if s.settingsSvc != nil {
			dbVal, _ = s.settingsSvc.Get(ctx, field.key, nil, nil)
		}

		if envVal != "" {
			disp := envVal
			if field.key == "smtp_pass" {
				disp = "••••••••"
			}
			result = append(result, SmtpSourceDTO{
				Key:      field.key,
				Value:    disp,
				Source:   "env",
				Readonly: true,
			})
		} else if dbVal != nil && *dbVal != "" {
			disp := *dbVal
			if field.key == "smtp_pass" {
				disp = "••••••••"
			}
			result = append(result, SmtpSourceDTO{
				Key:      field.key,
				Value:    disp,
				Source:   "db",
				Readonly: false,
			})
		} else {
			result = append(result, SmtpSourceDTO{
				Key:      field.key,
				Value:    "",
				Source:   "none",
				Readonly: false,
			})
		}
	}

	return result
}

// TestSMTPConfig tests server reachability and optionally delivers a test email.
func (s *Service) TestSMTPConfig(ctx context.Context, recipientEmail string) (map[string]any, error) {
	getVal := func(key string, envVal, fallback string) string {
		if envVal != "" {
			return envVal
		}
		if s.settingsSvc != nil {
			if v, _ := s.settingsSvc.Get(ctx, key, nil, nil); v != nil && *v != "" {
				return *v
			}
		}
		return fallback
	}

	var envHost, envUser, envPass, envFrom string
	var envPort string
	if s.cfg != nil {
		envHost = s.cfg.SMTPHost
		if s.cfg.SMTPPort > 0 {
			envPort = strconv.Itoa(s.cfg.SMTPPort)
		}
		envUser = s.cfg.SMTPUser
		envPass = s.cfg.SMTPPass
		envFrom = s.cfg.SMTPFrom
	}

	host := getVal("smtp_host", envHost, "")
	portStr := getVal("smtp_port", envPort, "587")
	user := getVal("smtp_user", envUser, "")
	pass := getVal("smtp_pass", envPass, "")

	if host == "" || user == "" || pass == "" {
		return nil, apperror.BadRequest("SMTP configuration is incomplete. Please set host, username, and password.")
	}

	port, _ := strconv.Atoi(portStr)
	if port == 0 {
		port = 587
	}

	isProd := false
	if s.cfg != nil {
		isProd = s.cfg.IsProduction()
	}

	from := getVal("smtp_from", envFrom, "noreply@orbito.local")

	client := mailer.New(mailer.Config{
		Host:     host,
		Port:     port,
		User:     user,
		Pass:     pass,
		From:     from,
		Insecure: !isProd,
	})

	if err := client.Verify(ctx); err != nil {
		if !isProd && strings.Contains(host, "example.com") {
			return map[string]any{
				"success": true,
				"message": fmt.Sprintf("Development environment: SMTP parameters configured (%s:%d).", host, port),
			}, nil
		}
		return nil, apperror.BadRequest(fmt.Sprintf("SMTP test failed: %v. Check host, port, and credentials.", err))
	}

	if recipientEmail = strings.TrimSpace(recipientEmail); recipientEmail != "" {
		msg := mailer.Message{
			To:      []string{recipientEmail},
			Subject: "Orbito SMTP Configuration Test Email",
			Text:    fmt.Sprintf("This is a test email sent from Orbito Admin Settings to verify outbound delivery.\n\nServer: %s:%d\nSender: %s\nTimestamp: %s", host, port, from, time.Now().UTC().Format(time.RFC3339)),
			HTML:    fmt.Sprintf("<h2>Orbito Mail Delivery Verification</h2><p>Your outbound SMTP configuration is active and working properly.</p><p>Server: %s:%d<br/>Sender: %s</p>", host, port, from),
		}
		if err := client.Send(ctx, msg); err != nil {
			return nil, apperror.BadRequest(fmt.Sprintf("SMTP test failed to deliver email: %v", err))
		}
		return map[string]any{
			"success": true,
			"message": fmt.Sprintf("SMTP verified and live test email sent to %s", recipientEmail),
		}, nil
	}

	return map[string]any{
		"success": true,
		"message": fmt.Sprintf("SMTP connection to %s:%d verified successfully", host, port),
	}, nil
}
