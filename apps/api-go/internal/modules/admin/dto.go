package admin

// DashboardStatsDTO represents aggregated system metrics for the admin console.
type DashboardStatsDTO struct {
	TotalUsers          int64 `json:"totalUsers"`
	TotalOrganizations  int64 `json:"totalOrganizations"`
	TotalWorkspaces     int64 `json:"totalWorkspaces"`
	TotalProjects       int64 `json:"totalProjects"`
	TotalTasks          int64 `json:"totalTasks"`
	NewUsersThisWeek    int64 `json:"newUsersThisWeek"`
	NewOrgsThisWeek     int64 `json:"newOrgsThisWeek"`
	NewProjectsThisWeek int64 `json:"newProjectsThisWeek"`
	NewTasksThisWeek    int64 `json:"newTasksThisWeek"`
	ActiveUsers         int64 `json:"activeUsers"`
}

// UpdateUserRoleDTO payload to change a user's role.
type UpdateUserRoleDTO struct {
	Role string `json:"role" validate:"required,oneof=SUPER_ADMIN OWNER MANAGER MEMBER VIEWER"`
}

// UpdateUserStatusDTO payload to change a user's account status.
type UpdateUserStatusDTO struct {
	Status string `json:"status" validate:"required,oneof=ACTIVE INACTIVE SUSPENDED PENDING"`
}

// TransferOwnershipDTO payload to assign a new organization owner.
type TransferOwnershipDTO struct {
	NewOwnerID string `json:"newOwnerId" validate:"required,uuid"`
}

// SettingItemDTO represents an individual configuration entry.
type SettingItemDTO struct {
	Key         string `json:"key" validate:"required"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	IsEncrypted bool   `json:"isEncrypted,omitempty"`
}

// SaveSystemConfigDTO payload to update global settings in batch.
type SaveSystemConfigDTO struct {
	Settings []SettingItemDTO `json:"settings" validate:"required"`
}

// TestSMTPDTO payload to verify outbound mail parameters.
type TestSMTPDTO struct {
	Email string `json:"email,omitempty" validate:"omitempty,email"`
}

// UserListItemDTO represents summary user data for admin directory.
type UserListItemDTO struct {
	ID                 string  `json:"id"`
	Email              string  `json:"email"`
	FirstName          string  `json:"firstName"`
	LastName           string  `json:"lastName"`
	Avatar             *string `json:"avatar,omitempty"`
	Role               string  `json:"role"`
	Status             string  `json:"status"`
	LastLoginAt        *string `json:"lastLoginAt,omitempty"`
	CreatedAt          string  `json:"createdAt"`
	OrganizationsCount int64   `json:"_count,omitempty"`
}

// PaginatedUsersDTO wraps user list pagination matching NestJS admin service.
type PaginatedUsersDTO struct {
	Data       []UserListItemDTO `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"totalPages"`
}

// UserDetailDTO represents complete user record with memberships for admin view.
type UserDetailDTO struct {
	ID                  string                    `json:"id"`
	Email               string                    `json:"email"`
	FirstName           string                    `json:"firstName"`
	LastName            string                    `json:"lastName"`
	Avatar              *string                   `json:"avatar,omitempty"`
	Bio                 *string                   `json:"bio,omitempty"`
	Timezone            string                    `json:"timezone"`
	Language            string                    `json:"language"`
	Role                string                    `json:"role"`
	Status              string                    `json:"status"`
	LastLoginAt         *string                   `json:"lastLoginAt,omitempty"`
	EmailVerified       bool                      `json:"emailVerified"`
	CreatedAt           string                    `json:"createdAt"`
	UpdatedAt           string                    `json:"updatedAt"`
	Source              string                    `json:"source"`
	OrganizationMembers []UserOrgMembershipDetail `json:"organizationMembers"`
}

// UserOrgMembershipDetail represents organization membership attached to user detail.
type UserOrgMembershipDetail struct {
	Role         string            `json:"role"`
	Organization OrganizationBrief `json:"organization"`
}

// OrganizationBrief represents brief organization details.
type OrganizationBrief struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Slug   string            `json:"slug"`
	Counts OrganizationCount `json:"_count"`
}

// OrganizationCount counts associated sub-entities.
type OrganizationCount struct {
	Members    int64 `json:"members"`
	Workspaces int64 `json:"workspaces"`
}

// OrgListItemDTO represents organization row in admin directory.
type OrgListItemDTO struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Slug        string            `json:"slug"`
	Description *string           `json:"description,omitempty"`
	Archive     bool              `json:"archive"`
	CreatedAt   string            `json:"createdAt"`
	Owner       *UserOwnerSummary `json:"owner,omitempty"`
	Counts      OrganizationCount `json:"_count"`
}

// UserOwnerSummary brief owner user info.
type UserOwnerSummary struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

// PaginatedOrgsDTO wraps organization list pagination matching NestJS admin service.
type PaginatedOrgsDTO struct {
	Data       []OrgListItemDTO `json:"data"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalPages int              `json:"totalPages"`
}

// OrgDetailDTO represents full organization details for admin drill-down.
type OrgDetailDTO struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Slug        string             `json:"slug"`
	Description *string            `json:"description,omitempty"`
	Archive     bool               `json:"archive"`
	CreatedAt   string             `json:"createdAt"`
	UpdatedAt   string             `json:"updatedAt"`
	Owner       *UserOwnerSummary  `json:"owner,omitempty"`
	Members     []OrgMemberItemDTO `json:"members"`
	Workspaces  []OrgWorkspaceDTO  `json:"workspaces"`
	Counts      OrganizationCount  `json:"_count"`
}

// OrgMemberItemDTO represents member row in organization detail.
type OrgMemberItemDTO struct {
	Role     string           `json:"role"`
	JoinedAt string           `json:"joinedAt"`
	User     UserMemberDetail `json:"user"`
}

// UserMemberDetail represents user summary for member list.
type UserMemberDetail struct {
	ID        string  `json:"id"`
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Email     string  `json:"email"`
	Avatar    *string `json:"avatar,omitempty"`
	Status    string  `json:"status"`
}

// OrgWorkspaceDTO represents workspace attached to organization.
type OrgWorkspaceDTO struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Slug   string         `json:"slug"`
	Counts WorkspaceCount `json:"_count"`
}

// WorkspaceCount counts projects and members for a workspace.
type WorkspaceCount struct {
	Projects int64 `json:"projects"`
	Members  int64 `json:"members"`
}

// SmtpSourceDTO represents SMTP setting status and origin.
type SmtpSourceDTO struct {
	Key      string `json:"key"`
	Value    string `json:"value"`
	Source   string `json:"source"` // "env" | "db" | "none"
	Readonly bool   `json:"readonly"`
}

// ConfigResponseDTO wraps system settings and SMTP source status.
type ConfigResponseDTO struct {
	Settings    any             `json:"settings"`
	SmtpSources []SmtpSourceDTO `json:"smtpSources"`
}
