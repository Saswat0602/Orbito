package auth

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

type memoryAuthRepo struct {
	users         map[uuid.UUID]*sqlc.User
	invitations   map[string]*sqlc.FindInvitationByTokenRow
	organizations map[uuid.UUID]*sqlc.Organization
	orgMembers    map[string]bool
}

func newMemoryAuthRepo() *memoryAuthRepo {
	return &memoryAuthRepo{
		users:         make(map[uuid.UUID]*sqlc.User),
		invitations:   make(map[string]*sqlc.FindInvitationByTokenRow),
		organizations: make(map[uuid.UUID]*sqlc.Organization),
		orgMembers:    make(map[string]bool),
	}
}

func (m *memoryAuthRepo) GetUserByID(_ context.Context, id uuid.UUID) (*sqlc.User, error) {
	u, ok := m.users[id]
	if !ok || u.DeletedAt.Valid {
		return nil, nil
	}
	return u, nil
}

func (m *memoryAuthRepo) GetUserByEmail(_ context.Context, email string) (*sqlc.User, error) {
	for _, u := range m.users {
		if strings.EqualFold(u.Email, email) && !u.DeletedAt.Valid {
			return u, nil
		}
	}
	return nil, nil
}

func (m *memoryAuthRepo) GetUserByUsername(_ context.Context, username string) (*sqlc.User, error) {
	for _, u := range m.users {
		if u.Username.Valid && strings.EqualFold(u.Username.String, username) && !u.DeletedAt.Valid {
			return u, nil
		}
	}
	return nil, nil
}

func (m *memoryAuthRepo) CountSuperAdmins(_ context.Context) (int64, error) {
	var count int64
	for _, u := range m.users {
		if u.Role == sqlc.RoleSUPERADMIN && !u.DeletedAt.Valid {
			count++
		}
	}
	return count, nil
}

func (m *memoryAuthRepo) CountTotalUsers(_ context.Context) (int64, error) {
	var count int64
	for _, u := range m.users {
		if !u.DeletedAt.Valid {
			count++
		}
	}
	return count, nil
}

func (m *memoryAuthRepo) CreateUser(_ context.Context, params sqlc.CreateUserParams) (*sqlc.User, error) {
	u := &sqlc.User{
		ID:        params.ID,
		Email:     params.Email,
		Username:  params.Username,
		FirstName: params.FirstName,
		LastName:  params.LastName,
		Password:  params.Password,
		Role:      params.Role,
		Status:    params.Status,
		Timezone:  params.Timezone,
		Language:  params.Language,
		Source:    params.Source,
	}
	m.users[params.ID] = u
	return u, nil
}

func (m *memoryAuthRepo) UpdateLastLogin(_ context.Context, id uuid.UUID) error {
	if u, ok := m.users[id]; ok {
		u.LastLoginAt = pgtype.Timestamp{Valid: true}
	}
	return nil
}

func (m *memoryAuthRepo) UpdateRole(_ context.Context, id uuid.UUID, role sqlc.Role) error {
	if u, ok := m.users[id]; ok {
		u.Role = role
	}
	return nil
}

func (m *memoryAuthRepo) UpdateRefreshToken(_ context.Context, id uuid.UUID, token pgtype.Text) error {
	if u, ok := m.users[id]; ok {
		u.RefreshToken = token
	}
	return nil
}

func (m *memoryAuthRepo) SetResetToken(_ context.Context, id uuid.UUID, token pgtype.Text, expiry pgtype.Timestamp) error {
	if u, ok := m.users[id]; ok {
		u.ResetToken = token
		u.ResetTokenExpiry = expiry
	}
	return nil
}

func (m *memoryAuthRepo) GetUserByResetToken(_ context.Context, token string) (*sqlc.User, error) {
	for _, u := range m.users {
		if u.ResetToken.Valid && u.ResetToken.String == token && !u.DeletedAt.Valid {
			return u, nil
		}
	}
	return nil, nil
}

func (m *memoryAuthRepo) UpdatePassword(_ context.Context, id uuid.UUID, hash pgtype.Text) error {
	if u, ok := m.users[id]; ok {
		u.Password = hash
		u.ResetToken = pgtype.Text{Valid: false}
		u.ResetTokenExpiry = pgtype.Timestamp{Valid: false}
	}
	return nil
}

func (m *memoryAuthRepo) FindInvitationByToken(_ context.Context, token string) (*sqlc.FindInvitationByTokenRow, error) {
	inv, ok := m.invitations[token]
	if !ok {
		return nil, nil
	}
	return inv, nil
}

func (m *memoryAuthRepo) AddOrganizationMember(_ context.Context, _, userID, orgID uuid.UUID, _ sqlc.Role) error {
	key := userID.String() + ":" + orgID.String()
	m.orgMembers[key] = true
	return nil
}

func (m *memoryAuthRepo) GetOrganizationByID(_ context.Context, id uuid.UUID) (*sqlc.Organization, error) {
	org, ok := m.organizations[id]
	if !ok {
		return nil, nil
	}
	return org, nil
}
