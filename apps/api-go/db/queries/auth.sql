-- name: GetUserByID :one
SELECT id, email, username, first_name, last_name, avatar, bio, timezone, language, role, status,
       password, refresh_token, reset_token, reset_token_expiry, default_organization_id,
       last_login_at, created_at, updated_at, deleted_at, source, external_id, external_provider
FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT id, email, username, first_name, last_name, avatar, bio, timezone, language, role, status,
       password, refresh_token, reset_token, reset_token_expiry, default_organization_id,
       last_login_at, created_at, updated_at, deleted_at, source, external_id, external_provider
FROM users
WHERE LOWER(email) = LOWER($1) AND deleted_at IS NULL;

-- name: GetUserByUsername :one
SELECT id, email, username, first_name, last_name, avatar, bio, timezone, language, role, status,
       password, refresh_token, reset_token, reset_token_expiry, default_organization_id,
       last_login_at, created_at, updated_at, deleted_at, source, external_id, external_provider
FROM users
WHERE LOWER(username) = LOWER($1) AND deleted_at IS NULL;

-- name: CountSuperAdmins :one
SELECT COUNT(*) FROM users
WHERE role = 'SUPER_ADMIN' AND deleted_at IS NULL;

-- name: CountTotalUsers :one
SELECT COUNT(*) FROM users
WHERE deleted_at IS NULL;

-- name: CreateUser :one
INSERT INTO users (
    id, email, username, first_name, last_name, password, role, status,
    timezone, language, email_verified, preferences, "onboardInfo",
    default_organization_id, source, external_id, external_provider, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8,
    $9, $10, $11, $12, $13,
    $14, $15, $16, $17, NOW(), NOW()
)
RETURNING id, email, username, first_name, last_name, avatar, bio, timezone, language, role, status,
          password, refresh_token, reset_token, reset_token_expiry, default_organization_id,
          last_login_at, created_at, updated_at, deleted_at, source, external_id, external_provider;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateUserRole :exec
UPDATE users
SET role = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateUserRefreshToken :exec
UPDATE users
SET refresh_token = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: SetUserResetToken :exec
UPDATE users
SET reset_token = $2,
    reset_token_expiry = $3,
    updated_at = NOW()
WHERE id = $1;

-- name: GetUserByResetToken :one
SELECT id, email, username, first_name, last_name, avatar, bio, timezone, language, role, status,
       password, refresh_token, reset_token, reset_token_expiry, default_organization_id,
       last_login_at, created_at, updated_at, deleted_at, source, external_id, external_provider
FROM users
WHERE reset_token = $1 AND reset_token_expiry > NOW() AND deleted_at IS NULL;

-- name: UpdateUserPassword :exec
UPDATE users
SET password = $2,
    reset_token = NULL,
    reset_token_expiry = NULL,
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateUserStatus :one
UPDATE users
SET status = $2,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, email, username, first_name, last_name, avatar, bio, timezone, language, role, status,
          last_login_at, created_at, updated_at, deleted_at;

-- name: UpdateUserProfile :one
UPDATE users
SET first_name = COALESCE(NULLIF($2, ''), first_name),
    last_name = COALESCE(NULLIF($3, ''), last_name),
    username = COALESCE(NULLIF($4, ''), username),
    avatar = COALESCE(NULLIF($5, ''), avatar),
    bio = COALESCE(NULLIF($6, ''), bio),
    timezone = COALESCE(NULLIF($7, ''), timezone),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, email, username, first_name, last_name, avatar, bio, timezone, language, role, status,
          last_login_at, created_at, updated_at, deleted_at;

-- name: ListUsers :many
SELECT id, email, username, first_name, last_name, avatar, bio, timezone, language, role, status,
       last_login_at, created_at, updated_at, deleted_at
FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: DeleteUserSoft :exec
UPDATE users
SET deleted_at = NOW(),
    deleted_by = $2
WHERE id = $1;

-- name: FindInvitationByToken :one
SELECT id, token, invitee_email, status, expires_at, workspace_id, role
FROM invitations
WHERE token = $1;

-- name: AddOrganizationMember :exec
INSERT INTO organization_members (
    id, user_id, organization_id, role, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, NOW(), NOW()
) ON CONFLICT (user_id, organization_id) DO NOTHING;

-- name: GetOrganizationByID :one
SELECT id, name, slug FROM organizations WHERE id = $1 AND deleted_at IS NULL;
