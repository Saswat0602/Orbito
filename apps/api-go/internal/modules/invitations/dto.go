package invitations

import (
	"time"

	"github.com/google/uuid"
)

// CreateInvitationRequest models the payload for sending invitations.
type CreateInvitationRequest struct {
	InviteeEmail   string      `json:"inviteeEmail"`
	Email          string      `json:"email"`
	OrganizationID *uuid.UUID  `json:"organizationId"`
	WorkspaceID    *uuid.UUID  `json:"workspaceId"`
	ProjectID      *uuid.UUID  `json:"projectId"`
	ProjectIDs     []uuid.UUID `json:"projectIds"`
	Role           string      `json:"role"`
}

// EntityTarget represents an organization, workspace, or project stub.
type EntityTarget struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

// InviterUser represents the user who sent the invitation.
type InviterUser struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Username  *string   `json:"username,omitempty"`
}

// InvitationResponse represents an invitation record returned by the API.
type InvitationResponse struct {
	ID             uuid.UUID     `json:"id"`
	InviterID      uuid.UUID     `json:"inviterId"`
	InviteeEmail   string        `json:"inviteeEmail"`
	OrganizationID *uuid.UUID    `json:"organizationId"`
	WorkspaceID    *uuid.UUID    `json:"workspaceId"`
	ProjectID      *uuid.UUID    `json:"projectId"`
	Role           string        `json:"role"`
	Status         string        `json:"status"`
	Token          string        `json:"token"`
	ExpiresAt      time.Time     `json:"expiresAt"`
	CreatedAt      time.Time     `json:"createdAt"`
	UpdatedAt      time.Time     `json:"updatedAt"`
	Inviter        *InviterUser  `json:"inviter,omitempty"`
	Organization   *EntityTarget `json:"organization,omitempty"`
	Workspace      *EntityTarget `json:"workspace,omitempty"`
	Project        *EntityTarget `json:"project,omitempty"`
	EmailSent      bool          `json:"emailSent"`
}

// DirectAddResponse is returned when an existing org member is directly added to a workspace/project.
type DirectAddResponse struct {
	Type             string         `json:"type"`
	Message          string         `json:"message"`
	Member           any            `json:"member"`
	Entity           map[string]any `json:"entity"`
	User             map[string]any `json:"user"`
	AddedToProjects  []uuid.UUID    `json:"addedToProjects,omitempty"`
	AddedToWorkspace bool           `json:"addedToWorkspace,omitempty"`
}

// VerifyInvitationResponse represents the verification status of an invitation token.
type VerifyInvitationResponse struct {
	Success       bool           `json:"success"`
	Message       string         `json:"message"`
	InviteeExists bool           `json:"inviteeExists"`
	Invitation    map[string]any `json:"invitation"`
}

// ResendInvitationResponse is returned after updating and resending an invitation.
type ResendInvitationResponse struct {
	Message    string              `json:"message"`
	Invitation *InvitationResponse `json:"invitation"`
	EmailSent  bool                `json:"emailSent"`
	EmailError *string             `json:"emailError"`
}

// AcceptInvitationResponse is returned when an invitation is accepted.
type AcceptInvitationResponse struct {
	Message    string            `json:"message"`
	Invitation map[string]string `json:"invitation"`
}

// MessageResponse is a simple success message.
type MessageResponse struct {
	Message string `json:"message"`
}
