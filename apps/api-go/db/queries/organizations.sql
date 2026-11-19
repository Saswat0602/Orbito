-- name: CreateOrganization :one
INSERT INTO "organizations" (
    id, name, slug, description, avatar, website, settings, owner_id, created_by_id, updated_by_id, created_at, updated_at, archive
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, false
) RETURNING *;

-- name: GetOrganizationDetailsByID :one
SELECT o.*, 
       u.id AS owner_user_id, u.first_name AS owner_first_name, u.last_name AS owner_last_name, u.email AS owner_email, u.avatar AS owner_avatar
FROM "organizations" o
JOIN "users" u ON o.owner_id = u.id
WHERE o.id = $1 AND o.archive = false;

-- name: GetOrganizationBySlug :one
SELECT o.*,
       u.id AS owner_user_id, u.first_name AS owner_first_name, u.last_name AS owner_last_name, u.email AS owner_email, u.avatar AS owner_avatar
FROM "organizations" o
JOIN "users" u ON o.owner_id = u.id
WHERE o.slug = $1 AND o.archive = false;

-- name: ListOrganizationsForUser :many
SELECT o.*, om.role AS user_role,
       u.id AS owner_user_id, u.first_name AS owner_first_name, u.last_name AS owner_last_name, u.email AS owner_email
FROM "organizations" o
JOIN "organization_members" om ON o.id = om.organization_id
JOIN "users" u ON o.owner_id = u.id
WHERE om.user_id = $1 AND o.archive = false
ORDER BY o.created_at DESC;

-- name: UpdateOrganization :one
UPDATE "organizations"
SET name = COALESCE(NULLIF($2, ''), name),
    slug = COALESCE(NULLIF($3, ''), slug),
    description = COALESCE($4, description),
    avatar = COALESCE($5, avatar),
    website = COALESCE($6, website),
    settings = COALESCE($7, settings),
    updated_by_id = $8,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND archive = false
RETURNING *;

-- name: ArchiveOrganization :exec
UPDATE "organizations"
SET archive = true,
    updated_by_id = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteOrganization :exec
DELETE FROM "organizations"
WHERE id = $1;

-- name: TransferOrganizationOwner :exec
UPDATE "organizations"
SET owner_id = $2,
    updated_by_id = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: CreateOrganizationMember :one
INSERT INTO "organization_members" (
    id, role, joined_at, user_id, organization_id, created_by_id, updated_by_id, created_at, updated_at, is_default
) VALUES (
    $1, $2, CURRENT_TIMESTAMP, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $7
) RETURNING *;

-- name: GetOrganizationMember :one
SELECT * FROM "organization_members"
WHERE user_id = $1 AND organization_id = $2;

-- name: UpdateOrganizationMemberRole :exec
UPDATE "organization_members"
SET role = $3,
    updated_by_id = $4,
    updated_at = CURRENT_TIMESTAMP
WHERE organization_id = $1 AND user_id = $2;

-- name: CountOrganizationMembers :one
SELECT COUNT(*) FROM "organization_members"
WHERE organization_id = $1;

-- name: CountOrganizationWorkspaces :one
SELECT COUNT(*) FROM "workspaces"
WHERE organization_id = $1 AND archive = false;

-- name: CountOrganizationProjects :one
SELECT COUNT(*) FROM "projects" p
JOIN "workspaces" w ON p.workspace_id = w.id
WHERE w.organization_id = $1 AND p.archive = false;

-- name: CountOrganizationTasks :one
SELECT COUNT(*) FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
WHERE w.organization_id = $1 AND t.deleted_at IS NULL;

-- name: CountOrganizationTasksCompleted :one
SELECT COUNT(*) FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
JOIN "task_statuses" ts ON t.status_id = ts.id
WHERE w.organization_id = $1 AND ts.category = 'DONE' AND t.deleted_at IS NULL;

-- name: CheckOrganizationSlugExists :one
SELECT EXISTS(
    SELECT 1 FROM "organizations"
    WHERE slug = $1 AND ($2::uuid IS NULL OR id != $2)
);

-- name: SetUserDefaultOrganization :exec
UPDATE "users"
SET default_organization_id = $2
WHERE id = $1 AND default_organization_id IS NULL;

-- name: ClearUserDefaultOrganization :exec
UPDATE "users"
SET default_organization_id = NULL
WHERE default_organization_id = $1;

-- name: CreateWorkflow :one
INSERT INTO "workflows" (
    id, name, description, is_default, organization_id, created_by_id, updated_by_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: CreateTaskStatus :one
INSERT INTO "task_statuses" (
    id, name, color, category, position, is_default, workflow_id, created_by_id, updated_by_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: CreateStatusTransition :one
INSERT INTO "status_transitions" (
    id, name, description, workflow_id, from_status_id, to_status_id, created_by_id, updated_by_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;
