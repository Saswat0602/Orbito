-- name: GetDashboardStatsCounts :one
SELECT
    (SELECT COUNT(*) FROM users) as total_users,
    (SELECT COUNT(*) FROM organizations) as total_organizations,
    (SELECT COUNT(*) FROM workspaces) as total_workspaces,
    (SELECT COUNT(*) FROM projects) as total_projects,
    (SELECT COUNT(*) FROM tasks) as total_tasks,
    (SELECT COUNT(*) FROM users WHERE created_at >= @since::timestamp) as new_users_this_week,
    (SELECT COUNT(*) FROM organizations WHERE created_at >= @since::timestamp) as new_orgs_this_week,
    (SELECT COUNT(*) FROM projects WHERE created_at >= @since::timestamp) as new_projects_this_week,
    (SELECT COUNT(*) FROM tasks WHERE created_at >= @since::timestamp) as new_tasks_this_week,
    (SELECT COUNT(*) FROM users WHERE status = 'ACTIVE' AND deleted_at IS NULL) as active_users;

-- name: AdminListUsers :many
SELECT u.id, u.email, u.first_name, u.last_name, u.avatar, u.role, u.status, u.last_login_at, u.created_at,
       (SELECT COUNT(*) FROM organization_members om WHERE om.user_id = u.id) as organization_count
FROM users u
WHERE u.deleted_at IS NULL
  AND (@search::text = '' OR u.email ILIKE '%' || @search || '%' OR u.first_name ILIKE '%' || @search || '%' OR u.last_name ILIKE '%' || @search || '%')
  AND (@status::text = '' OR u.status = @status::"UserStatus")
  AND (@role::text = '' OR u.role = @role::"Role")
ORDER BY u.created_at DESC
LIMIT @limit_val::int OFFSET @offset_val::int;

-- name: AdminCountUsers :one
SELECT COUNT(*)
FROM users u
WHERE u.deleted_at IS NULL
  AND (@search::text = '' OR u.email ILIKE '%' || @search || '%' OR u.first_name ILIKE '%' || @search || '%' OR u.last_name ILIKE '%' || @search || '%')
  AND (@status::text = '' OR u.status = @status::"UserStatus")
  AND (@role::text = '' OR u.role = @role::"Role");

-- name: AdminGetUserDetail :one
SELECT u.id, u.email, u.first_name, u.last_name, u.avatar, u.bio, u.timezone, u.language,
       u.role, u.status, u.last_login_at, u.email_verified, u.created_at, u.updated_at, u.source
FROM users u
WHERE u.id = @id AND u.deleted_at IS NULL;

-- name: AdminGetUserOrganizations :many
SELECT o.id, o.name, o.slug, om.role,
       (SELECT COUNT(*) FROM organization_members m WHERE m.organization_id = o.id) as members_count,
       (SELECT COUNT(*) FROM workspaces w WHERE w.organization_id = o.id) as workspaces_count
FROM organization_members om
JOIN organizations o ON om.organization_id = o.id
WHERE om.user_id = @user_id;

-- name: AdminCountSuperAdmins :one
SELECT COUNT(*) FROM users WHERE role = 'SUPER_ADMIN' AND deleted_at IS NULL;

-- name: AdminUpdateUserRole :one
UPDATE users
SET role = @role::"Role", updated_at = NOW()
WHERE id = @id AND deleted_at IS NULL
RETURNING id, email, first_name, last_name, role;

-- name: AdminUpdateUserStatus :one
UPDATE users
SET status = @status::"UserStatus",
    refresh_token = CASE WHEN @status::"UserStatus" IN ('INACTIVE'::"UserStatus", 'SUSPENDED'::"UserStatus") THEN NULL ELSE refresh_token END,
    reset_token = CASE WHEN @status::"UserStatus" IN ('INACTIVE'::"UserStatus", 'SUSPENDED'::"UserStatus") THEN NULL ELSE reset_token END,
    reset_token_expiry = CASE WHEN @status::"UserStatus" IN ('INACTIVE'::"UserStatus", 'SUSPENDED'::"UserStatus") THEN NULL ELSE reset_token_expiry END,
    updated_at = NOW()
WHERE id = @id AND deleted_at IS NULL
RETURNING id, email, first_name, last_name, status;

-- name: AdminSoftDeleteUser :exec
UPDATE users
SET status = 'INACTIVE',
    deleted_at = NOW(),
    deleted_by = @deleted_by,
    password = NULL,
    refresh_token = NULL,
    reset_token = NULL,
    reset_token_expiry = NULL,
    default_organization_id = NULL
WHERE id = @id;

-- name: AdminDeleteUserMemberships :exec
DELETE FROM organization_members WHERE user_id = @user_id;

-- name: AdminDeleteUserWorkspaceMemberships :exec
DELETE FROM workspace_members WHERE user_id = @user_id;

-- name: AdminDeleteUserProjectMemberships :exec
DELETE FROM project_members WHERE user_id = @user_id;

-- name: AdminGetUserOwnedOrganizations :many
SELECT id, name FROM organizations WHERE owner_id = @owner_id;

-- name: AdminSetUserResetToken :exec
UPDATE users
SET reset_token = @reset_token,
    reset_token_expiry = @reset_token_expiry,
    updated_at = NOW()
WHERE id = @id;

-- name: AdminListOrganizations :many
SELECT o.id, o.name, o.slug, o.description, o.archive, o.created_at,
       u.id as owner_id, u.first_name as owner_first_name, u.last_name as owner_last_name, u.email as owner_email,
       (SELECT COUNT(*) FROM organization_members om WHERE om.organization_id = o.id) as members_count,
       (SELECT COUNT(*) FROM workspaces w WHERE w.organization_id = o.id) as workspaces_count
FROM organizations o
LEFT JOIN users u ON o.owner_id = u.id
WHERE (@search::text = '' OR o.name ILIKE '%' || @search || '%' OR o.slug ILIKE '%' || @search || '%')
ORDER BY o.created_at DESC
LIMIT @limit_val::int OFFSET @offset_val::int;

-- name: AdminCountOrganizations :one
SELECT COUNT(*)
FROM organizations o
WHERE (@search::text = '' OR o.name ILIKE '%' || @search || '%' OR o.slug ILIKE '%' || @search || '%');

-- name: AdminGetOrganizationDetail :one
SELECT o.id, o.name, o.slug, o.description, o.archive, o.created_at, o.updated_at,
       u.id as owner_id, u.first_name as owner_first_name, u.last_name as owner_last_name, u.email as owner_email,
       (SELECT COUNT(*) FROM organization_members om WHERE om.organization_id = o.id) as members_count,
       (SELECT COUNT(*) FROM workspaces w WHERE w.organization_id = o.id) as workspaces_count
FROM organizations o
LEFT JOIN users u ON o.owner_id = u.id
WHERE o.id = @id;

-- name: AdminGetOrganizationMembers :many
SELECT om.role, om.joined_at, u.id as user_id, u.first_name, u.last_name, u.email, u.avatar, u.status
FROM organization_members om
JOIN users u ON om.user_id = u.id
WHERE om.organization_id = @organization_id
ORDER BY om.joined_at ASC;

-- name: AdminGetOrganizationWorkspaces :many
SELECT w.id, w.name, w.slug,
       (SELECT COUNT(*) FROM projects p WHERE p.workspace_id = w.id) as projects_count,
       (SELECT COUNT(*) FROM workspace_members wm WHERE wm.workspace_id = w.id) as members_count
FROM workspaces w
WHERE w.organization_id = @organization_id
ORDER BY w.created_at ASC;

-- name: AdminToggleOrganizationArchive :one
UPDATE organizations
SET archive = NOT archive, updated_at = NOW()
WHERE id = @id
RETURNING id, name, archive;

-- name: AdminDeleteOrganization :exec
DELETE FROM organizations WHERE id = @id;

-- name: AdminClearUsersDefaultOrg :exec
UPDATE users SET default_organization_id = NULL WHERE default_organization_id = @org_id;

-- name: AdminTransferOrgOwnership :one
UPDATE organizations
SET owner_id = @new_owner_id, updated_at = NOW()
WHERE id = @org_id
RETURNING id, name, owner_id;

-- name: AdminGetOrganizationMember :one
SELECT id, role FROM organization_members
WHERE user_id = @user_id AND organization_id = @organization_id;

-- name: AdminCreateOrganizationMember :exec
INSERT INTO organization_members (id, role, joined_at, user_id, organization_id, created_by_id)
VALUES (@id, @role::"Role", NOW(), @user_id, @organization_id, @created_by_id);

-- name: AdminUpdateOrganizationMemberRole :exec
UPDATE organization_members
SET role = @role::"Role"
WHERE id = @id;
