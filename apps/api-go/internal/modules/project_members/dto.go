package project_members

import (
	"time"

	"github.com/google/uuid"
)

// CreateProjectMemberRequest specifies user, project and assigned role.
type CreateProjectMemberRequest struct {
	UserID    uuid.UUID `json:"userId"`
	ProjectID uuid.UUID `json:"projectId"`
	Role      *string   `json:"role"`
}

// UpdateProjectMemberRoleRequest contains the new role to assign.
type UpdateProjectMemberRoleRequest struct {
	Role string `json:"role"`
}

// ProjectMemberUserProfile contains user display information.
type ProjectMemberUserProfile struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Avatar    *string   `json:"avatar,omitempty"`
	Username  *string   `json:"username,omitempty"`
}

// ProjectMemberResponse formats the project membership with nested user details.
type ProjectMemberResponse struct {
	ID          uuid.UUID                `json:"id"`
	Role        string                   `json:"role"`
	JoinedAt    time.Time                `json:"joinedAt"`
	UserID      uuid.UUID                `json:"userId"`
	ProjectID   uuid.UUID                `json:"projectId"`
	ProjectName string                   `json:"projectName,omitempty"`
	ProjectSlug string                   `json:"projectSlug,omitempty"`
	CreatedAt   time.Time                `json:"createdAt"`
	UpdatedAt   time.Time                `json:"updatedAt"`
	User        ProjectMemberUserProfile `json:"user"`
}

// ProjectMemberStatsResponse contains membership counts by role.
type ProjectMemberStatsResponse struct {
	TotalMembers int64            `json:"totalMembers"`
	ByRole       map[string]int64 `json:"byRole"`
}
