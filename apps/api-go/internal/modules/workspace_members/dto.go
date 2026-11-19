package workspace_members

import (
	"time"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// AddWorkspaceMemberRequest contains payload to add a user to a workspace.
type AddWorkspaceMemberRequest struct {
	UserID      uuid.UUID `json:"userId" validate:"required"`
	WorkspaceID uuid.UUID `json:"workspaceId" validate:"required"`
	Role        sqlc.Role `json:"role" validate:"required,oneof=OWNER MANAGER MEMBER VIEWER"`
}

// UpdateWorkspaceMemberRoleRequest contains payload to change member role.
type UpdateWorkspaceMemberRoleRequest struct {
	Role sqlc.Role `json:"role" validate:"required,oneof=OWNER MANAGER MEMBER VIEWER"`
}

// MemberUserDTO represents user profile information.
type MemberUserDTO struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Avatar    *string   `json:"avatar,omitempty"`
	Status    string    `json:"status"`
}

// MemberWorkspaceDTO represents workspace summary information.
type MemberWorkspaceDTO struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Slug           string    `json:"slug"`
	OrganizationID uuid.UUID `json:"organizationId"`
}

// WorkspaceMemberResponse represents a workspace membership payload.
type WorkspaceMemberResponse struct {
	ID          uuid.UUID           `json:"id"`
	Role        string              `json:"role"`
	JoinedAt    time.Time           `json:"joinedAt"`
	UserID      uuid.UUID           `json:"userId"`
	WorkspaceID uuid.UUID           `json:"workspaceId"`
	User        *MemberUserDTO      `json:"user,omitempty"`
	Workspace   *MemberWorkspaceDTO `json:"workspace,omitempty"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
}

// PaginatedWorkspaceMembersResponse contains paginated members list.
type PaginatedWorkspaceMembersResponse struct {
	Items      []WorkspaceMemberResponse `json:"items"`
	TotalItems int64                     `json:"totalItems"`
	Page       int                       `json:"page"`
	Limit      int                       `json:"limit"`
	TotalPages int                       `json:"totalPages"`
}
