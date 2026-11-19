package organizations

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// CreateOrganizationRequest contains input fields for creating an organization.
type CreateOrganizationRequest struct {
	Name        string          `json:"name" validate:"required,min=1,max=100"`
	Description *string         `json:"description,omitempty" validate:"omitempty,max=500"`
	Avatar      *string         `json:"avatar,omitempty"`
	Website     *string         `json:"website,omitempty"`
	Settings    json.RawMessage `json:"settings,omitempty"`
	OwnerID     *uuid.UUID      `json:"ownerId,omitempty"`
}

// UpdateOrganizationRequest contains fields for updating an organization.
type UpdateOrganizationRequest struct {
	Name        *string         `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Slug        *string         `json:"slug,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string         `json:"description,omitempty"`
	Avatar      *string         `json:"avatar,omitempty"`
	Website     *string         `json:"website,omitempty"`
	Settings    json.RawMessage `json:"settings,omitempty"`
}

// TransferOwnershipRequest contains the target new owner's UUID.
type TransferOwnershipRequest struct {
	NewOwnerID uuid.UUID `json:"newOwnerId" validate:"required"`
}

// OrganizationOwnerDTO represents the nested owner user in responses.
type OrganizationOwnerDTO struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Avatar    *string   `json:"avatar,omitempty"`
}

// OrganizationResponse represents an organization payload.
type OrganizationResponse struct {
	ID          uuid.UUID             `json:"id"`
	Name        string                `json:"name"`
	Slug        string                `json:"slug"`
	Description *string               `json:"description,omitempty"`
	Avatar      *string               `json:"avatar,omitempty"`
	Website     *string               `json:"website,omitempty"`
	Settings    json.RawMessage       `json:"settings,omitempty"`
	OwnerID     uuid.UUID             `json:"ownerId"`
	Owner       *OrganizationOwnerDTO `json:"owner,omitempty"`
	Role        *string               `json:"role,omitempty"`
	Archive     bool                  `json:"archive"`
	CreatedAt   time.Time             `json:"createdAt"`
	UpdatedAt   time.Time             `json:"updatedAt"`
}

// OrganizationStatsResponse contains aggregate metrics for an organization.
type OrganizationStatsResponse struct {
	TotalMembers    int64 `json:"totalMembers"`
	TotalWorkspaces int64 `json:"totalWorkspaces"`
	TotalProjects   int64 `json:"totalProjects"`
	TotalTasks      int64 `json:"totalTasks"`
	CompletedTasks  int64 `json:"completedTasks"`
}
