-- name: ListWorkspaceMembersPaginated :many
SELECT wm.id, wm.role, wm.joined_at, wm.user_id, wm.workspace_id, wm.created_at, wm.updated_at,
       u.id AS user_user_id, u.email AS user_email, u.first_name AS user_first_name, u.last_name AS user_last_name, u.avatar AS user_avatar, u.status AS user_status,
       w.id AS ws_id, w.name AS ws_name, w.slug AS ws_slug, w.organization_id AS ws_org_id
FROM "workspace_members" wm
JOIN "users" u ON wm.user_id = u.id
JOIN "workspaces" w ON wm.workspace_id = w.id
WHERE ($1::uuid IS NULL OR wm.workspace_id = $1)
  AND ($2::text = '' OR u.first_name ILIKE '%' || $2 || '%' OR u.last_name ILIKE '%' || $2 || '%' OR u.email ILIKE '%' || $2 || '%')
ORDER BY wm.joined_at ASC
LIMIT $3 OFFSET $4;

-- name: CountWorkspaceMembersFiltered :one
SELECT COUNT(*)
FROM "workspace_members" wm
JOIN "users" u ON wm.user_id = u.id
JOIN "workspaces" w ON wm.workspace_id = w.id
WHERE ($1::uuid IS NULL OR wm.workspace_id = $1)
  AND ($2::text = '' OR u.first_name ILIKE '%' || $2 || '%' OR u.last_name ILIKE '%' || $2 || '%' OR u.email ILIKE '%' || $2 || '%');

-- name: GetWorkspaceMemberDetailsByID :one
SELECT wm.*,
       u.id AS user_user_id, u.email AS user_email, u.first_name AS user_first_name, u.last_name AS user_last_name, u.avatar AS user_avatar, u.status AS user_status,
       w.id AS ws_id, w.name AS ws_name, w.slug AS ws_slug, w.organization_id AS ws_org_id
FROM "workspace_members" wm
JOIN "users" u ON wm.user_id = u.id
JOIN "workspaces" w ON wm.workspace_id = w.id
WHERE wm.id = $1;

-- name: GetWorkspaceMemberByUserAndWorkspace :one
SELECT wm.*,
       u.id AS user_user_id, u.email AS user_email, u.first_name AS user_first_name, u.last_name AS user_last_name, u.avatar AS user_avatar, u.status AS user_status,
       w.id AS ws_id, w.name AS ws_name, w.slug AS ws_slug, w.organization_id AS ws_org_id
FROM "workspace_members" wm
JOIN "users" u ON wm.user_id = u.id
JOIN "workspaces" w ON wm.workspace_id = w.id
WHERE wm.user_id = $1 AND wm.workspace_id = $2;

-- name: UpdateWorkspaceMemberRoleByID :one
UPDATE "workspace_members"
SET role = $2,
    updated_by_id = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteWorkspaceMemberByID :exec
DELETE FROM "workspace_members"
WHERE id = $1;
