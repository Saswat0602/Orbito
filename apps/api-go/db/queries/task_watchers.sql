-- name: GetTaskProjectScopeForWatcher :one
SELECT
    t.id AS task_id,
    t.title AS task_title,
    t.slug AS task_slug,
    p.id AS project_id,
    p.name AS project_name,
    p.slug AS project_slug,
    p.visibility,
    p.workspace_id,
    w.organization_id
FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
WHERE (sqlc.arg(is_uuid)::boolean = true AND t.id = sqlc.narg(task_uuid)::uuid)
   OR (sqlc.arg(is_uuid)::boolean = false AND t.slug = sqlc.arg(task_slug)::text)
LIMIT 1;

-- name: CheckTargetUserTenantAccess :one
SELECT
    u.id,
    u.first_name,
    u.last_name,
    EXISTS(SELECT 1 FROM "project_members" pm WHERE pm.project_id = sqlc.arg(project_id)::uuid AND pm.user_id = u.id) AS is_pm,
    EXISTS(SELECT 1 FROM "workspace_members" wm WHERE wm.workspace_id = sqlc.arg(workspace_id)::uuid AND wm.user_id = u.id) AS is_wm,
    EXISTS(SELECT 1 FROM "organization_members" om WHERE om.organization_id = sqlc.arg(org_id)::uuid AND om.user_id = u.id) AS is_om
FROM "users" u
WHERE u.id = sqlc.arg(user_id)::uuid;

-- name: CheckCallerTaskMembership :one
SELECT
    EXISTS(SELECT 1 FROM "project_members" pm WHERE pm.project_id = $1 AND pm.user_id = $3) AS is_pm,
    EXISTS(SELECT 1 FROM "workspace_members" wm WHERE wm.workspace_id = $2 AND wm.user_id = $3) AS is_wm;

-- name: CreateTaskWatcherRecord :one
INSERT INTO "task_watchers" (id, task_id, user_id, created_at, updated_at)
VALUES ($1, $2, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id;

-- name: GetTaskWatcherDetailsByID :one
SELECT
    tw.id, tw.task_id, tw.user_id, tw.created_by_id, tw.updated_by_id, tw.created_at, tw.updated_at,
    u.email AS user_email, u.first_name AS user_first_name, u.last_name AS user_last_name, u.avatar AS user_avatar, u.bio AS user_bio, u.status AS user_status,
    t.title AS task_title, t.slug AS task_slug, t.description AS task_description, t.type AS task_type, t.priority AS task_priority,
    ts.id AS status_id, ts.name AS status_name, ts.color AS status_color, ts.category AS status_category,
    p.id AS project_id, p.name AS project_name, p.slug AS project_slug,
    w.id AS workspace_id, w.name AS workspace_name, w.slug AS workspace_slug
FROM "task_watchers" tw
JOIN "users" u ON tw.user_id = u.id
JOIN "tasks" t ON tw.task_id = t.id
JOIN "task_statuses" ts ON t.status_id = ts.id
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
WHERE tw.id = $1;

-- name: FindTaskWatcherByTaskAndUser :one
SELECT id, task_id, user_id
FROM "task_watchers"
WHERE task_id = $1 AND user_id = $2
LIMIT 1;

-- name: DeleteTaskWatcherByPair :execrows
DELETE FROM "task_watchers"
WHERE task_id = $1 AND user_id = $2;

-- name: DeleteTaskWatcherByID :execrows
DELETE FROM "task_watchers" WHERE id = $1;

-- name: ListTaskWatchersFiltered :many
SELECT
    tw.id, tw.task_id, tw.user_id, tw.created_by_id, tw.updated_by_id, tw.created_at, tw.updated_at,
    u.email AS user_email, u.first_name AS user_first_name, u.last_name AS user_last_name, u.avatar AS user_avatar, u.status AS user_status,
    t.title AS task_title, t.slug AS task_slug, t.type AS task_type, t.priority AS task_priority,
    ts.id AS status_id, ts.name AS status_name, ts.color AS status_color, ts.category AS status_category,
    p.id AS project_id, p.name AS project_name, p.slug AS project_slug
FROM "task_watchers" tw
JOIN "users" u ON tw.user_id = u.id
JOIN "tasks" t ON tw.task_id = t.id
JOIN "task_statuses" ts ON t.status_id = ts.id
JOIN "projects" p ON t.project_id = p.id
WHERE (sqlc.narg(task_id)::uuid IS NULL OR tw.task_id = sqlc.narg(task_id))
  AND (sqlc.narg(user_id)::uuid IS NULL OR tw.user_id = sqlc.narg(user_id))
ORDER BY tw.created_at DESC;

-- name: ListWatchersForTask :many
SELECT
    tw.id, tw.task_id, tw.user_id, tw.created_by_id, tw.updated_by_id, tw.created_at, tw.updated_at,
    u.email AS user_email, u.first_name AS user_first_name, u.last_name AS user_last_name, u.avatar AS user_avatar, u.status AS user_status
FROM "task_watchers" tw
JOIN "users" u ON tw.user_id = u.id
WHERE tw.task_id = $1
ORDER BY tw.created_at ASC;

-- name: ListWatchedTasksByUser :many
SELECT
    tw.id, tw.task_id, tw.user_id, tw.created_by_id, tw.updated_by_id, tw.created_at, tw.updated_at,
    t.title AS task_title, t.slug AS task_slug, t.description AS task_description, t.type AS task_type, t.priority AS task_priority, t.due_date AS task_due_date,
    ts.id AS status_id, ts.name AS status_name, ts.color AS status_color, ts.category AS status_category,
    p.id AS project_id, p.name AS project_name, p.slug AS project_slug,
    w.id AS workspace_id, w.name AS workspace_name, w.slug AS workspace_slug,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug,
    (SELECT COUNT(*) FROM "task_comments" tc WHERE tc.task_id = t.id) AS comment_count,
    (SELECT COUNT(*) FROM "task_watchers" tw2 WHERE tw2.task_id = t.id) AS watcher_count
FROM "task_watchers" tw
JOIN "tasks" t ON tw.task_id = t.id
JOIN "task_statuses" ts ON t.status_id = ts.id
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
JOIN "organizations" o ON w.organization_id = o.id
WHERE tw.user_id = $1
ORDER BY tw.created_at DESC;

-- name: CheckUserExistsByIDSimple :one
SELECT EXISTS(SELECT 1 FROM "users" WHERE id = $1);

-- name: CountTaskWatchersTotalAndRecent :one
SELECT
    COUNT(*)::bigint AS total_watchers,
    COUNT(*) FILTER (WHERE created_at >= sqlc.arg(since)::timestamp)::bigint AS recent_watches
FROM "task_watchers"
WHERE (sqlc.narg(task_id)::uuid IS NULL OR task_id = sqlc.narg(task_id))
  AND (sqlc.narg(user_id)::uuid IS NULL OR user_id = sqlc.narg(user_id));

-- name: GroupMostWatchedTasks :many
SELECT
    task_id,
    COUNT(*)::bigint AS watch_count
FROM "task_watchers"
WHERE (sqlc.narg(user_id)::uuid IS NULL OR user_id = sqlc.narg(user_id))
GROUP BY task_id
ORDER BY watch_count DESC
LIMIT 10;
