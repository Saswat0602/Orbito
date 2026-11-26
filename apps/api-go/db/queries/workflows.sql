-- name: GetWorkflowDetailsByID :one
SELECT 
    w.id, w.name, w.description, w.is_default, w.organization_id, 
    w.created_by_id, w.updated_by_id, w.created_at, w.updated_at,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug,
    cu.id AS created_by_user_id, cu.email AS created_by_email, cu.first_name AS created_by_first_name, cu.last_name AS created_by_last_name,
    uu.id AS updated_by_user_id, uu.email AS updated_by_email, uu.first_name AS updated_by_first_name, uu.last_name AS updated_by_last_name
FROM "workflows" w
JOIN "organizations" o ON w.organization_id = o.id
LEFT JOIN "users" cu ON w.created_by_id = cu.id
LEFT JOIN "users" uu ON w.updated_by_id = uu.id
WHERE w.id = $1;

-- name: GetWorkflowStatusesByWorkflowID :many
SELECT 
    id, name, color, category, position, is_default, workflow_id, created_at, updated_at, created_by_id, updated_by_id, deleted_at, deleted_by
FROM "task_statuses"
WHERE workflow_id = $1 AND deleted_at IS NULL
ORDER BY position ASC;

-- name: GetWorkflowTransitionsByWorkflowID :many
SELECT 
    t.id, t.name, t.description, t.workflow_id, t.from_status_id, t.to_status_id,
    t.created_by_id, t.updated_by_id, t.created_at, t.updated_at,
    fs.id AS from_status_id, fs.name AS from_status_name, fs.color AS from_status_color, fs.category AS from_status_category,
    ts.id AS to_status_id, ts.name AS to_status_name, ts.color AS to_status_color, ts.category AS to_status_category
FROM "status_transitions" t
JOIN "task_statuses" fs ON t.from_status_id = fs.id
JOIN "task_statuses" ts ON t.to_status_id = ts.id
WHERE t.workflow_id = $1
ORDER BY t.created_at ASC;

-- name: CountWorkflowActiveStatuses :one
SELECT COUNT(*) FROM "task_statuses" WHERE workflow_id = $1 AND deleted_at IS NULL;

-- name: CountWorkflowTransitions :one
SELECT COUNT(*) FROM "status_transitions" WHERE workflow_id = $1;

-- name: ListWorkflowsByOrganizationID :many
SELECT 
    w.id, w.name, w.description, w.is_default, w.organization_id, 
    w.created_by_id, w.updated_by_id, w.created_at, w.updated_at,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug,
    cu.id AS created_by_user_id, cu.email AS created_by_email, cu.first_name AS created_by_first_name, cu.last_name AS created_by_last_name,
    uu.id AS updated_by_user_id, uu.email AS updated_by_email, uu.first_name AS updated_by_first_name, uu.last_name AS updated_by_last_name
FROM "workflows" w
JOIN "organizations" o ON w.organization_id = o.id
LEFT JOIN "users" cu ON w.created_by_id = cu.id
LEFT JOIN "users" uu ON w.updated_by_id = uu.id
WHERE w.organization_id = $1
ORDER BY w.is_default DESC, w.created_at DESC;

-- name: ListWorkflowsByOrganizationIDs :many
SELECT 
    w.id, w.name, w.description, w.is_default, w.organization_id, 
    w.created_by_id, w.updated_by_id, w.created_at, w.updated_at,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug,
    cu.id AS created_by_user_id, cu.email AS created_by_email, cu.first_name AS created_by_first_name, cu.last_name AS created_by_last_name,
    uu.id AS updated_by_user_id, uu.email AS updated_by_email, uu.first_name AS updated_by_first_name, uu.last_name AS updated_by_last_name
FROM "workflows" w
JOIN "organizations" o ON w.organization_id = o.id
LEFT JOIN "users" cu ON w.created_by_id = cu.id
LEFT JOIN "users" uu ON w.updated_by_id = uu.id
WHERE w.organization_id = ANY($1::uuid[])
ORDER BY w.is_default DESC, w.created_at DESC;

-- name: ListAllWorkflows :many
SELECT 
    w.id, w.name, w.description, w.is_default, w.organization_id, 
    w.created_by_id, w.updated_by_id, w.created_at, w.updated_at,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug,
    cu.id AS created_by_user_id, cu.email AS created_by_email, cu.first_name AS created_by_first_name, cu.last_name AS created_by_last_name,
    uu.id AS updated_by_user_id, uu.email AS updated_by_email, uu.first_name AS updated_by_first_name, uu.last_name AS updated_by_last_name
FROM "workflows" w
JOIN "organizations" o ON w.organization_id = o.id
LEFT JOIN "users" cu ON w.created_by_id = cu.id
LEFT JOIN "users" uu ON w.updated_by_id = uu.id
ORDER BY w.is_default DESC, w.created_at DESC;

-- name: GetDefaultWorkflowByOrgID :one
SELECT 
    w.id, w.name, w.description, w.is_default, w.organization_id, 
    w.created_by_id, w.updated_by_id, w.created_at, w.updated_at,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug
FROM "workflows" w
JOIN "organizations" o ON w.organization_id = o.id
WHERE w.organization_id = $1 AND w.is_default = true
LIMIT 1;

-- name: UnsetDefaultWorkflowsInOrg :exec
UPDATE "workflows"
SET is_default = false, updated_by_id = $2, updated_at = CURRENT_TIMESTAMP
WHERE organization_id = $1 AND is_default = true;

-- name: UnsetDefaultWorkflowsExcept :exec
UPDATE "workflows"
SET is_default = false, updated_by_id = $2, updated_at = CURRENT_TIMESTAMP
WHERE organization_id = $1 AND is_default = true AND id != $3;

-- name: UpdateWorkflowDetails :one
UPDATE "workflows"
SET 
    name = COALESCE(sqlc.narg(name), name),
    description = COALESCE(sqlc.narg(description), description),
    is_default = COALESCE(sqlc.narg(is_default), is_default),
    updated_by_id = sqlc.narg(updated_by_id),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: SetWorkflowDefault :one
UPDATE "workflows"
SET 
    is_default = true,
    updated_by_id = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteWorkflow :exec
DELETE FROM "workflows" WHERE id = $1;

-- name: CountProjectsUsingWorkflow :one
SELECT COUNT(*) FROM "projects" WHERE workflow_id = $1;

-- name: CountTasksUsingWorkflowStatuses :one
SELECT COUNT(*) 
FROM "tasks" t
JOIN "task_statuses" s ON t.status_id = s.id
WHERE s.workflow_id = $1;

-- name: GetUserAccessibleOrgIDs :many
SELECT id FROM "organizations" WHERE owner_id = $1
UNION
SELECT organization_id AS id FROM "organization_members" WHERE user_id = $1;

-- name: GetOrganizationOwnerAndUserRole :one
SELECT 
    o.id,
    o.owner_id,
    om.role AS member_role
FROM "organizations" o
LEFT JOIN "organization_members" om ON om.organization_id = o.id AND om.user_id = $2
WHERE o.id = $1;

-- name: GetWorkflowOrgOwnerAndUserRole :one
SELECT 
    w.id AS workflow_id,
    w.organization_id,
    o.owner_id,
    om.role AS member_role
FROM "workflows" w
JOIN "organizations" o ON w.organization_id = o.id
LEFT JOIN "organization_members" om ON om.organization_id = o.id AND om.user_id = $2
WHERE w.id = $1;

-- name: GetOrganizationBySlugForWorkflow :one
SELECT 
    o.id,
    o.owner_id,
    om.role AS member_role
FROM "organizations" o
LEFT JOIN "organization_members" om ON om.organization_id = o.id AND om.user_id = $2
WHERE o.slug = $1;
