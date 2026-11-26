-- name: GetTaskStatusAuthByID :one
SELECT
    ts.id, ts.name, ts.color, ts.category, ts.position, ts.is_default, ts.workflow_id, ts.deleted_at,
    w.organization_id,
    o.owner_id,
    om.role AS member_role
FROM "task_statuses" ts
JOIN "workflows" w ON ts.workflow_id = w.id
JOIN "organizations" o ON w.organization_id = o.id
LEFT JOIN "organization_members" om ON om.organization_id = o.id AND om.user_id = $2
WHERE ts.id = $1;

-- name: FindSoftDeletedTaskStatus :one
SELECT *
FROM "task_statuses"
WHERE workflow_id = $1 AND name = $2 AND deleted_at IS NOT NULL
LIMIT 1;

-- name: CountAllWorkflowStatuses :one
SELECT COUNT(*)
FROM "task_statuses"
WHERE workflow_id = $1;

-- name: CreateTaskStatusNode :one
INSERT INTO "task_statuses" (
    id, name, color, category, position, is_default, workflow_id,
    created_by_id, updated_by_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: RestoreSoftDeletedTaskStatusOnCreate :one
UPDATE "task_statuses"
SET deleted_at = NULL,
    deleted_by = NULL,
    updated_by_id = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: GetTaskStatusFullByID :one
SELECT
    ts.id, ts.name, ts.color, ts.category, ts.position, ts.is_default, ts.workflow_id,
    ts.created_at, ts.updated_at, ts.created_by_id, ts.updated_by_id, ts.deleted_at, ts.deleted_by,
    w.name AS workflow_name, w.description AS workflow_description,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug,
    cu.id AS created_by_user_id, cu.email AS created_by_email, cu.first_name AS created_by_first_name, cu.last_name AS created_by_last_name,
    uu.id AS updated_by_user_id, uu.email AS updated_by_email, uu.first_name AS updated_by_first_name, uu.last_name AS updated_by_last_name,
    (SELECT COUNT(*) FROM "tasks" t WHERE t.status_id = ts.id) AS task_count
FROM "task_statuses" ts
JOIN "workflows" w ON ts.workflow_id = w.id
JOIN "organizations" o ON w.organization_id = o.id
LEFT JOIN "users" cu ON ts.created_by_id = cu.id
LEFT JOIN "users" uu ON ts.updated_by_id = uu.id
WHERE ts.id = $1;

-- name: GetActiveTaskStatusFullByID :one
SELECT
    ts.id, ts.name, ts.color, ts.category, ts.position, ts.is_default, ts.workflow_id,
    ts.created_at, ts.updated_at, ts.created_by_id, ts.updated_by_id, ts.deleted_at, ts.deleted_by,
    w.name AS workflow_name, w.description AS workflow_description,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug,
    (SELECT COUNT(*) FROM "tasks" t WHERE t.status_id = ts.id) AS task_count
FROM "task_statuses" ts
JOIN "workflows" w ON ts.workflow_id = w.id
JOIN "organizations" o ON w.organization_id = o.id
WHERE ts.id = $1 AND ts.deleted_at IS NULL;

-- name: ListTaskStatusesFiltered :many
SELECT
    ts.id, ts.name, ts.color, ts.category, ts.position, ts.is_default, ts.workflow_id,
    ts.created_at, ts.updated_at, ts.created_by_id, ts.updated_by_id, ts.deleted_at, ts.deleted_by,
    w.name AS workflow_name, w.description AS workflow_description,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug,
    (SELECT COUNT(*) FROM "tasks" t WHERE t.status_id = ts.id) AS task_count
FROM "task_statuses" ts
JOIN "workflows" w ON ts.workflow_id = w.id
JOIN "organizations" o ON w.organization_id = o.id
WHERE ts.deleted_at IS NULL
  AND (sqlc.narg(workflow_id)::uuid IS NULL OR ts.workflow_id = sqlc.narg(workflow_id))
  AND (sqlc.narg(organization_id)::uuid IS NULL OR w.organization_id = sqlc.narg(organization_id))
ORDER BY ts.workflow_id ASC, ts.position ASC;

-- name: ListTaskStatusesByProjectID :many
SELECT
    ts.id, ts.name, ts.color, ts.category, ts.position, ts.is_default, ts.workflow_id,
    ts.created_at, ts.updated_at, ts.created_by_id, ts.updated_by_id, ts.deleted_at, ts.deleted_by,
    w.name AS workflow_name, w.description AS workflow_description,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug,
    (SELECT COUNT(*) FROM "tasks" t WHERE t.status_id = ts.id) AS task_count
FROM "projects" p
JOIN "task_statuses" ts ON ts.workflow_id = p.workflow_id
JOIN "workflows" w ON ts.workflow_id = w.id
JOIN "organizations" o ON w.organization_id = o.id
WHERE p.id = $1 AND ts.deleted_at IS NULL
ORDER BY ts.position ASC;

-- name: ListFromTransitionsForStatus :many
SELECT
    st.id, st.name, st.description, st.workflow_id, st.from_status_id, st.to_status_id,
    st.created_by_id, st.updated_by_id, st.created_at, st.updated_at,
    ts.id AS status_id, ts.name AS status_name, ts.color AS status_color, ts.category AS status_category
FROM "status_transitions" st
JOIN "task_statuses" ts ON st.to_status_id = ts.id
WHERE st.from_status_id = $1;

-- name: ListToTransitionsForStatus :many
SELECT
    st.id, st.name, st.description, st.workflow_id, st.from_status_id, st.to_status_id,
    st.created_by_id, st.updated_by_id, st.created_at, st.updated_at,
    fs.id AS status_id, fs.name AS status_name, fs.color AS status_color, fs.category AS status_category
FROM "status_transitions" st
JOIN "task_statuses" fs ON st.from_status_id = fs.id
WHERE st.to_status_id = $1;

-- name: ListDeletedTaskStatuses :many
SELECT
    ts.id, ts.name, ts.color, ts.category, ts.position, ts.is_default, ts.workflow_id,
    ts.created_at, ts.updated_at, ts.created_by_id, ts.updated_by_id, ts.deleted_at, ts.deleted_by,
    w.name AS workflow_name, w.description AS workflow_description,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug,
    du.id AS deleted_by_user_id, du.email AS deleted_by_email, du.first_name AS deleted_by_first_name, du.last_name AS deleted_by_last_name
FROM "task_statuses" ts
JOIN "workflows" w ON ts.workflow_id = w.id
JOIN "organizations" o ON w.organization_id = o.id
LEFT JOIN "users" du ON ts.deleted_by = du.id
WHERE ts.deleted_at IS NOT NULL
  AND (sqlc.narg(workflow_id)::uuid IS NULL OR ts.workflow_id = sqlc.narg(workflow_id))
ORDER BY ts.deleted_at DESC;

-- name: UpdateTaskStatusByID :one
UPDATE "task_statuses"
SET
    name = COALESCE(sqlc.narg(name), name),
    color = COALESCE(sqlc.narg(color), color),
    category = COALESCE(sqlc.narg(category)::"StatusCategory", category),
    position = COALESCE(sqlc.narg(position), position),
    updated_by_id = sqlc.narg(updated_by_id),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id;

-- name: UpdateTaskStatusPositionByID :one
UPDATE "task_statuses"
SET
    position = $2,
    updated_by_id = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id;

-- name: DeleteTransitionsByStatusID :exec
DELETE FROM "status_transitions"
WHERE from_status_id = $1 OR to_status_id = $1;

-- name: DecrementStatusPositionsAfterDelete :exec
UPDATE "task_statuses"
SET position = position - 1,
    updated_by_id = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE workflow_id = $1 AND position > $2;

-- name: SoftDeleteTaskStatusByID :exec
UPDATE "task_statuses"
SET deleted_at = CURRENT_TIMESTAMP,
    deleted_by = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: CheckActiveTaskStatusNameConflict :one
SELECT EXISTS(
    SELECT 1 FROM "task_statuses"
    WHERE workflow_id = $1 AND name = $2 AND deleted_at IS NULL AND id != $3
);

-- name: RestoreDeletedTaskStatusByID :one
UPDATE "task_statuses"
SET deleted_at = NULL,
    deleted_by = NULL,
    updated_by_id = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id;

-- name: GetProjectWorkflowForStatusCreate :one
SELECT p.id AS project_id, p.workflow_id
FROM "projects" p
WHERE p.id = $1;

