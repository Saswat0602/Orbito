-- name: CreateWorkspace :one
INSERT INTO "workspaces" (
    id, name, slug, description, avatar, color, settings, organization_id, parent_workspace_id, path, created_by_id, updated_by_id, created_at, updated_at, archive
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, false
) RETURNING *;

-- name: GetWorkspaceByID :one
SELECT w.*,
       o.name AS organization_name, o.slug AS organization_slug, o.owner_id AS organization_owner_id
FROM "workspaces" w
JOIN "organizations" o ON w.organization_id = o.id
WHERE w.id = $1;

-- name: GetWorkspaceByOrgAndSlug :one
SELECT w.*,
       o.name AS organization_name, o.slug AS organization_slug, o.owner_id AS organization_owner_id
FROM "workspaces" w
JOIN "organizations" o ON w.organization_id = o.id
WHERE w.organization_id = $1 AND w.slug = $2 AND w.archive = false;

-- name: GetWorkspaceBySlugOnly :one
SELECT w.*,
       o.name AS organization_name, o.slug AS organization_slug, o.owner_id AS organization_owner_id
FROM "workspaces" w
JOIN "organizations" o ON w.organization_id = o.id
WHERE w.slug = $1 AND w.archive = false;

-- name: ListWorkspacesForUser :many
SELECT DISTINCT w.*, wm.role AS user_role
FROM "workspaces" w
JOIN "workspace_members" wm ON w.id = wm.workspace_id
WHERE wm.user_id = $1
  AND ($2::uuid IS NULL OR w.organization_id = $2)
  AND ($3::text = '' OR w.name ILIKE '%' || $3 || '%' OR w.slug ILIKE '%' || $3 || '%')
  AND w.archive = false
ORDER BY w.created_at DESC;

-- name: ListArchivedWorkspaces :many
SELECT w.*
FROM "workspaces" w
WHERE w.organization_id = $1 AND w.archive = true
ORDER BY w.updated_at DESC;

-- name: ListWorkspaceTree :many
SELECT w.id, w.name, w.slug, w.parent_workspace_id, w.path, w.archive
FROM "workspaces" w
WHERE w.organization_id = $1 AND w.archive = false
ORDER BY w.name ASC;

-- name: UpdateWorkspace :one
UPDATE "workspaces"
SET name = COALESCE(NULLIF($2, ''), name),
    slug = COALESCE(NULLIF($3, ''), slug),
    description = COALESCE($4, description),
    avatar = COALESCE($5, avatar),
    color = COALESCE($6, color),
    settings = COALESCE($7, settings),
    updated_by_id = $8,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ArchiveWorkspace :exec
UPDATE "workspaces"
SET archive = true,
    updated_by_id = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: UnarchiveWorkspace :exec
UPDATE "workspaces"
SET archive = false,
    updated_by_id = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteWorkspace :exec
DELETE FROM "workspaces"
WHERE id = $1;

-- name: CheckWorkspaceSlugExists :one
SELECT EXISTS(
    SELECT 1 FROM "workspaces"
    WHERE organization_id = $1 AND slug = $2 AND ($3::uuid IS NULL OR id != $3)
);

-- name: ListWorkspaceMembersDetails :many
SELECT wm.id, wm.role, wm.joined_at, wm.user_id, wm.workspace_id, wm.created_at, wm.updated_at,
       u.id AS user_user_id, u.email AS user_email, u.first_name AS user_first_name, u.last_name AS user_last_name, u.avatar AS user_avatar, u.status AS user_status
FROM "workspace_members" wm
JOIN "users" u ON wm.user_id = u.id
WHERE wm.workspace_id = $1
ORDER BY wm.joined_at ASC;

-- name: CreateWorkspaceMemberDirect :one
INSERT INTO "workspace_members" (
    id, role, joined_at, user_id, workspace_id, created_by_id, updated_by_id, created_at, updated_at
) VALUES (
    $1, $2, CURRENT_TIMESTAMP, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: CountWorkspaceProjects :one
SELECT COUNT(*) FROM "projects"
WHERE workspace_id = $1 AND archive = false;

-- name: CountWorkspaceProjectsByStatus :one
SELECT COUNT(*) FROM "projects"
WHERE workspace_id = $1 AND status = $2 AND archive = false;

-- name: CountWorkspaceTasks :one
SELECT COUNT(*) FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
WHERE p.workspace_id = $1 AND p.archive = false AND t.deleted_at IS NULL;

-- name: CountWorkspaceCompletedTasks :one
SELECT COUNT(*) FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
WHERE p.workspace_id = $1 AND p.archive = false AND t.completed_at IS NOT NULL AND t.deleted_at IS NULL;

-- name: CountWorkspaceOverdueTasks :one
SELECT COUNT(*) FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
WHERE p.workspace_id = $1 AND p.archive = false AND t.due_date < CURRENT_TIMESTAMP AND t.completed_at IS NULL AND t.deleted_at IS NULL;

-- name: WorkspaceProjectStatusStats :many
SELECT status, COUNT(*) AS count
FROM "projects"
WHERE workspace_id = $1 AND archive = false
GROUP BY status;

-- name: WorkspaceTaskPriorityStats :many
SELECT t.priority, COUNT(*) AS count
FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
WHERE p.workspace_id = $1 AND p.archive = false AND t.deleted_at IS NULL
GROUP BY t.priority;
