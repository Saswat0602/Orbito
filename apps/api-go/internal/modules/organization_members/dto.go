package organization_members

import (
	"time"

	"github.com/google/uuid"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// AddMemberRequest contains input for adding a member to an organization.
type AddMemberRequest struct {
	UserID         uuid.UUID `json:"userId" validate:"required"`
	OrganizationID uuid.UUID `json:"organizationId" validate:"required"`
	Role           sqlc.Role `json:"role" validate:"required,oneof=OWNER MANAGER MEMBER VIEWER"`
}

// UpdateRoleRequest contains the target role for a member.
type UpdateRoleRequest struct {
	Role sqlc.Role `json:"role" validate:"required,oneof=OWNER MANAGER MEMBER VIEWER"`
}

// MemberUserDTO represents embedded user info in member responses.
type MemberUserDTO struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Avatar    *string   `json:"avatar,omitempty"`
	Status    string    `json:"status"`
}

// MemberOrgDTO represents embedded organization info.
type MemberOrgDTO struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Slug    string    `json:"slug"`
	OwnerID uuid.UUID `json:"ownerId"`
}

// OrganizationMemberResponse represents a membership payload.
type OrganizationMemberResponse struct {
	ID             uuid.UUID      `json:"id"`
	Role           string         `json:"role"`
	JoinedAt       time.Time      `json:"joinedAt"`
	UserID         uuid.UUID      `json:"userId"`
	OrganizationID uuid.UUID      `json:"organizationId"`
	IsDefault      bool           `json:"isDefault"`
	User           *MemberUserDTO `json:"user,omitempty"`
	Organization   *MemberOrgDTO  `json:"organization,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
}

// PaginatedMembersResponse contains pagination metadata and items.
type PaginatedMembersResponse struct {
	Items      []OrganizationMemberResponse `json:"items"`
	TotalItems int64                        `json:"totalItems"`
	Page       int                          `json:"page"`
	Limit      int                          `json:"limit"`
	TotalPages int                          `json:"totalPages"`
}
