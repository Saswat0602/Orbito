-- name: FindTaskIDForDependency :one
SELECT id
FROM "tasks"
WHERE (sqlc.arg(is_uuid)::boolean = true AND id = sqlc.narg(task_uuid)::uuid)
   OR (sqlc.arg(is_uuid)::boolean = false AND slug = sqlc.arg(task_slug)::text)
LIMIT 1;

-- name: CheckDependencyExists :one
SELECT EXISTS(
    SELECT 1 FROM "task_dependencies"
    WHERE dependent_task_id = $1 AND blocking_task_id = $2
);

-- name: ListBlockingIDsForTask :many
SELECT blocking_task_id
FROM "task_dependencies"
WHERE dependent_task_id = $1;

-- name: CreateTaskDependencyRecord :one
INSERT INTO "task_dependencies" (
    id, type, dependent_task_id, blocking_task_id, created_by_id, updated_by_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING id;

-- name: GetTaskDependencyDetailsByID :one
SELECT
    td.id, td.type, td.dependent_task_id, td.blocking_task_id, td.created_by_id, td.updated_by_id, td.created_at, td.updated_at,
    dt.title AS dep_title, dt.slug AS dep_slug, dt.project_id AS dep_project_id, dts.name AS dep_status_name,
    bt.title AS blk_title, bt.slug AS blk_slug, bt.project_id AS blk_project_id, bts.name AS blk_status_name,
    cu.id AS creator_id, cu.first_name AS creator_first_name, cu.last_name AS creator_last_name, cu.email AS creator_email
FROM "task_dependencies" td
JOIN "tasks" dt ON td.dependent_task_id = dt.id
JOIN "task_statuses" dts ON dt.status_id = dts.id
JOIN "tasks" bt ON td.blocking_task_id = bt.id
JOIN "task_statuses" bts ON bt.status_id = bts.id
LEFT JOIN "users" cu ON td.created_by_id = cu.id
WHERE td.id = $1;

-- name: ListTaskDependencies :many
SELECT
    td.id, td.type, td.dependent_task_id, td.blocking_task_id, td.created_by_id, td.updated_by_id, td.created_at, td.updated_at,
    dt.title AS dep_title, dt.slug AS dep_slug, dt.project_id AS dep_project_id, dts.name AS dep_status_name,
    bt.title AS blk_title, bt.slug AS blk_slug, bt.project_id AS blk_project_id, bts.name AS blk_status_name,
    cu.id AS creator_id, cu.first_name AS creator_first_name, cu.last_name AS creator_last_name, cu.email AS creator_email
FROM "task_dependencies" td
JOIN "tasks" dt ON td.dependent_task_id = dt.id
JOIN "task_statuses" dts ON dt.status_id = dts.id
JOIN "tasks" bt ON td.blocking_task_id = bt.id
JOIN "task_statuses" bts ON bt.status_id = bts.id
LEFT JOIN "users" cu ON td.created_by_id = cu.id
WHERE (sqlc.narg(project_id)::uuid IS NULL OR dt.project_id = sqlc.narg(project_id))
ORDER BY td.created_at DESC;

-- name: ListDependenciesTaskDependsOn :many
SELECT
    td.id, td.type, td.dependent_task_id, td.blocking_task_id, td.created_by_id, td.updated_by_id, td.created_at, td.updated_at,
    bt.title AS blk_title, bt.slug AS blk_slug, bts.name AS blk_status_name
FROM "task_dependencies" td
JOIN "tasks" bt ON td.blocking_task_id = bt.id
JOIN "task_statuses" bts ON bt.status_id = bts.id
WHERE td.dependent_task_id = $1;

-- name: ListDependenciesTaskBlocks :many
SELECT
    td.id, td.type, td.dependent_task_id, td.blocking_task_id, td.created_by_id, td.updated_by_id, td.created_at, td.updated_at,
    dt.title AS dep_title, dt.slug AS dep_slug, dts.name AS dep_status_name
FROM "task_dependencies" td
JOIN "tasks" dt ON td.dependent_task_id = dt.id
JOIN "task_statuses" dts ON dt.status_id = dts.id
WHERE td.blocking_task_id = $1;

-- name: UpdateTaskDependencyByID :one
UPDATE "task_dependencies"
SET
    type = COALESCE(sqlc.narg(type)::"DependencyType", type),
    dependent_task_id = COALESCE(sqlc.narg(dependent_task_id), dependent_task_id),
    blocking_task_id = COALESCE(sqlc.narg(blocking_task_id), blocking_task_id),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id;

-- name: DeleteTaskDependencyByID :execrows
DELETE FROM "task_dependencies" WHERE id = $1;

-- name: DeleteTaskDependencyByPair :execrows
DELETE FROM "task_dependencies"
WHERE dependent_task_id = $1 AND blocking_task_id = $2;

