-- name: ListProjectMembers :many
SELECT 
    pm.id, pm.role, pm.joined_at, pm.user_id, pm.project_id, pm.created_at, pm.updated_at,
    u.id AS member_user_id, u.email AS user_email, u.first_name AS user_first_name,
    u.last_name AS user_last_name, u.avatar AS user_avatar, u.username AS user_username,
    u.role AS user_global_role,
    p.name AS project_name, p.slug AS project_slug
FROM project_members pm
JOIN users u ON pm.user_id = u.id
JOIN projects p ON pm.project_id = p.id
WHERE pm.project_id = sqlc.arg(project_id)
  AND (sqlc.narg(search)::text IS NULL OR u.first_name ILIKE '%' || sqlc.narg(search) || '%' OR u.last_name ILIKE '%' || sqlc.narg(search) || '%' OR u.email ILIKE '%' || sqlc.narg(search) || '%')
ORDER BY pm.joined_at ASC;

-- name: CountProjectMembersFiltered :one
SELECT count(*)
FROM project_members pm
JOIN users u ON pm.user_id = u.id
WHERE pm.project_id = sqlc.arg(project_id)
  AND (sqlc.narg(search)::text IS NULL OR u.first_name ILIKE '%' || sqlc.narg(search) || '%' OR u.last_name ILIKE '%' || sqlc.narg(search) || '%' OR u.email ILIKE '%' || sqlc.narg(search) || '%');

-- name: GetProjectMemberByID :one
SELECT 
    pm.id, pm.role, pm.joined_at, pm.user_id, pm.project_id, pm.created_at, pm.updated_at,
    u.id AS member_user_id, u.email AS user_email, u.first_name AS user_first_name,
    u.last_name AS user_last_name, u.avatar AS user_avatar, u.username AS user_username,
    p.name AS project_name, p.slug AS project_slug, p.workspace_id
FROM project_members pm
JOIN users u ON pm.user_id = u.id
JOIN projects p ON pm.project_id = p.id
WHERE pm.id = $1;

-- name: GetProjectMemberByUserAndProject :one
SELECT 
    pm.id, pm.role, pm.joined_at, pm.user_id, pm.project_id, pm.created_at, pm.updated_at,
    u.id AS member_user_id, u.email AS user_email, u.first_name AS user_first_name,
    u.last_name AS user_last_name, u.avatar AS user_avatar, u.username AS user_username
FROM project_members pm
JOIN users u ON pm.user_id = u.id
WHERE pm.user_id = $1 AND pm.project_id = $2;

-- name: CreateProjectMemberRecord :one
INSERT INTO project_members (
    id, user_id, project_id, role, joined_at, created_by_id, updated_by_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, CURRENT_TIMESTAMP, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: UpdateProjectMemberRoleByID :one
UPDATE project_members
SET role = $2, updated_at = CURRENT_TIMESTAMP, updated_by_id = $3
WHERE id = $1
RETURNING *;

-- name: DeleteProjectMemberByID :exec
DELETE FROM project_members
WHERE id = $1;

-- name: CountProjectOwnersInProject :one
SELECT count(*)
FROM project_members
WHERE project_id = $1 AND role = 'OWNER';

-- name: ListProjectMembersByWorkspace :many
SELECT 
    pm.id, pm.role, pm.joined_at, pm.user_id, pm.project_id,
    u.email AS user_email, u.first_name AS user_first_name, u.last_name AS user_last_name,
    p.name AS project_name, p.slug AS project_slug
FROM project_members pm
JOIN users u ON pm.user_id = u.id
JOIN projects p ON pm.project_id = p.id
WHERE p.workspace_id = $1
ORDER BY pm.joined_at ASC;

-- name: ListUserProjects :many
SELECT 
    p.id, p.name, p.slug, p.task_prefix, p.avatar, p.color, p.status, p.priority,
    pm.role AS member_role, pm.joined_at
FROM project_members pm
JOIN projects p ON pm.project_id = p.id
WHERE pm.user_id = $1 AND p.archive = false
ORDER BY pm.joined_at DESC;

-- name: ProjectMembersRoleStats :many
SELECT role, count(*) AS count
FROM project_members
WHERE project_id = $1
GROUP BY role;
