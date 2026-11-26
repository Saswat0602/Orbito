-- name: CheckTimeEntryTaskProjectAccess :one
SELECT
    t.id AS task_id,
    t.title AS task_title,
    t.slug AS task_slug,
    p.id AS project_id,
    p.visibility,
    p.workspace_id,
    EXISTS(SELECT 1 FROM "project_members" pm WHERE pm.project_id = p.id AND pm.user_id = $2) AS is_pm,
    EXISTS(SELECT 1 FROM "workspace_members" wm WHERE wm.workspace_id = p.workspace_id AND wm.user_id = $2) AS is_wm
FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
WHERE t.id = $1;

-- name: CreateTimeEntryRecord :one
INSERT INTO "time_entries" (
    id, task_id, user_id, time_spent, description, start_time, end_time, date, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING id;

-- name: GetTimeEntryDetailByID :one
SELECT
    te.id, te.description, te.time_spent, te.start_time, te.end_time, te.date,
    te.task_id, te.user_id, te.created_by_id, te.updated_by_id, te.created_at, te.updated_at,
    t.title AS task_title, t.slug AS task_slug,
    p.id AS project_id, p.name AS project_name, p.slug AS project_slug,
    w.id AS workspace_id, w.name AS workspace_name, w.slug AS workspace_slug,
    u.email AS user_email, u.first_name AS user_first_name, u.last_name AS user_last_name, u.avatar AS user_avatar
FROM "time_entries" te
JOIN "tasks" t ON te.task_id = t.id
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
JOIN "users" u ON te.user_id = u.id
WHERE te.id = $1;

-- name: ListTimeEntriesFiltered :many
SELECT
    te.id, te.description, te.time_spent, te.start_time, te.end_time, te.date,
    te.task_id, te.user_id, te.created_by_id, te.updated_by_id, te.created_at, te.updated_at,
    t.title AS task_title, t.slug AS task_slug,
    p.id AS project_id, p.name AS project_name, p.slug AS project_slug,
    u.first_name AS user_first_name, u.last_name AS user_last_name, u.avatar AS user_avatar
FROM "time_entries" te
JOIN "tasks" t ON te.task_id = t.id
JOIN "projects" p ON t.project_id = p.id
JOIN "users" u ON te.user_id = u.id
WHERE (sqlc.narg(task_id)::uuid IS NULL OR te.task_id = sqlc.narg(task_id))
  AND (sqlc.narg(user_id)::uuid IS NULL OR te.user_id = sqlc.narg(user_id))
  AND (sqlc.narg(start_date)::timestamp IS NULL OR te.date >= sqlc.narg(start_date))
  AND (sqlc.narg(end_date)::timestamp IS NULL OR te.date <= sqlc.narg(end_date))
ORDER BY te.date DESC;

-- name: GetActiveTimerForUser :one
SELECT
    te.id, te.description, te.time_spent, te.start_time, te.end_time, te.date,
    te.task_id, te.user_id, te.created_by_id, te.updated_by_id, te.created_at, te.updated_at,
    t.title AS task_title, t.slug AS task_slug,
    p.id AS project_id, p.name AS project_name, p.slug AS project_slug,
    u.first_name AS user_first_name, u.last_name AS user_last_name, u.avatar AS user_avatar
FROM "time_entries" te
JOIN "tasks" t ON te.task_id = t.id
JOIN "projects" p ON t.project_id = p.id
JOIN "users" u ON te.user_id = u.id
WHERE te.user_id = $1
  AND te.start_time IS NOT NULL
  AND te.end_time IS NULL
LIMIT 1;

-- name: UpdateTimeEntryByID :one
UPDATE "time_entries"
SET
    description = COALESCE(sqlc.narg(description), description),
    time_spent = COALESCE(sqlc.narg(time_spent), time_spent),
    start_time = COALESCE(sqlc.narg(start_time), start_time),
    end_time = COALESCE(sqlc.narg(end_time), end_time),
    date = COALESCE(sqlc.narg(date), date),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id;

-- name: StopActiveTimerByID :one
UPDATE "time_entries"
SET
    end_time = $2,
    time_spent = $3,
    description = COALESCE(sqlc.narg(description), description),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id;

-- name: DeleteTimeEntryByID :execrows
DELETE FROM "time_entries" WHERE id = $1;
