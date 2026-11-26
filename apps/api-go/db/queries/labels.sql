-- name: GetProjectWorkspaceForLabel :one
SELECT
    p.id,
    p.name,
    p.slug,
    p.workspace_id,
    EXISTS(
        SELECT 1 FROM "workspace_members" wm
        WHERE wm.workspace_id = p.workspace_id AND wm.user_id = $2
    ) AS is_ws_member
FROM "projects" p
WHERE p.id = $1;

-- name: CreateLabelRecord :one
INSERT INTO "labels" (
    id, name, color, description, project_id,
    created_by_id, updated_by_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING id;

-- name: UpsertWorkspaceLabel :exec
INSERT INTO "workspace_labels" (workspace_id, label_id, created_at)
VALUES ($1, $2, CURRENT_TIMESTAMP)
ON CONFLICT (workspace_id, label_id) DO NOTHING;

-- name: GetLabelDetailsByID :one
SELECT
    l.id, l.name, l.color, l.description, l.project_id,
    l.created_by_id, l.updated_by_id, l.created_at, l.updated_at,
    p.name AS project_name, p.slug AS project_slug, p.workspace_id,
    w.name AS workspace_name, w.slug AS workspace_slug, w.organization_id,
    o.name AS org_name, o.slug AS org_slug,
    cu.id AS created_by_user_id, cu.email AS created_by_email, cu.first_name AS created_by_first_name, cu.last_name AS created_by_last_name,
    uu.id AS updated_by_user_id, uu.email AS updated_by_email, uu.first_name AS updated_by_first_name, uu.last_name AS updated_by_last_name,
    (SELECT COUNT(*) FROM "task_labels" tl WHERE tl.label_id = l.id) AS task_label_count,
    EXISTS(
        SELECT 1 FROM "workspace_members" wm
        WHERE wm.workspace_id = p.workspace_id AND wm.user_id = $2
    ) AS is_ws_member
FROM "labels" l
JOIN "projects" p ON l.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
JOIN "organizations" o ON w.organization_id = o.id
LEFT JOIN "users" cu ON l.created_by_id = cu.id
LEFT JOIN "users" uu ON l.updated_by_id = uu.id
WHERE l.id = $1;

-- name: ListLabelsByWorkspaceID :many
SELECT
    l.id, l.name, l.color, l.description, l.project_id,
    l.created_by_id, l.updated_by_id, l.created_at, l.updated_at,
    p.name AS project_name, p.slug AS project_slug, p.workspace_id,
    w.name AS workspace_name, w.slug AS workspace_slug,
    (SELECT COUNT(*) FROM "task_labels" tl WHERE tl.label_id = l.id) AS task_label_count
FROM "labels" l
JOIN "projects" p ON l.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
WHERE p.workspace_id = $1
   OR EXISTS(
       SELECT 1 FROM "workspace_labels" wl
       WHERE wl.label_id = l.id AND wl.workspace_id = $1
   )
ORDER BY l.name ASC;

-- name: ListAccessibleLabelsForUser :many
SELECT
    l.id, l.name, l.color, l.description, l.project_id,
    l.created_by_id, l.updated_by_id, l.created_at, l.updated_at,
    p.name AS project_name, p.slug AS project_slug, p.workspace_id,
    w.name AS workspace_name, w.slug AS workspace_slug,
    (SELECT COUNT(*) FROM "task_labels" tl WHERE tl.label_id = l.id) AS task_label_count
FROM "labels" l
JOIN "projects" p ON l.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
WHERE EXISTS(
    SELECT 1 FROM "workspace_members" wm
    WHERE wm.workspace_id = p.workspace_id AND wm.user_id = $1
)
ORDER BY l.name ASC;

-- name: ListTaskLabelEntriesByLabelID :many
SELECT
    tl.task_id, tl.label_id, tl.created_by_id, tl.updated_by_id, tl.created_at, tl.updated_at,
    t.title AS task_title, t.slug AS task_slug, t.type AS task_type, t.priority AS task_priority,
    ts.id AS status_id, ts.name AS status_name, ts.color AS status_color, ts.category AS status_category
FROM "task_labels" tl
JOIN "tasks" t ON tl.task_id = t.id
JOIN "task_statuses" ts ON t.status_id = ts.id
WHERE tl.label_id = $1;

-- name: ListTaskAssigneesForTaskIDs :many
SELECT
    ta.task_id,
    u.id, u.first_name, u.last_name, u.avatar
FROM "task_assignees" ta
JOIN "users" u ON ta.user_id = u.id
WHERE ta.task_id = ANY($1::uuid[]);

-- name: UpdateLabelByID :one
UPDATE "labels"
SET
    name = COALESCE(sqlc.narg(name), name),
    color = COALESCE(sqlc.narg(color), color),
    description = COALESCE(sqlc.narg(description), description),
    updated_by_id = sqlc.narg(updated_by_id),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id;

-- name: DeleteLabelByID :execrows
DELETE FROM "labels"
WHERE id = $1;

-- name: FindTaskWorkspaceAccessByIDOrSlug :one
SELECT
    t.id AS task_id,
    t.project_id,
    p.workspace_id,
    EXISTS(
        SELECT 1 FROM "workspace_members" wm
        WHERE wm.workspace_id = p.workspace_id AND wm.user_id = sqlc.arg(user_id)::uuid
    ) AS is_ws_member
FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
WHERE (sqlc.arg(is_uuid)::boolean = true AND t.id = sqlc.narg(task_uuid)::uuid)
   OR (sqlc.arg(is_uuid)::boolean = false AND t.slug = sqlc.arg(task_slug)::text)
LIMIT 1;

-- name: GetLabelWorkspaceCheck :one
SELECT
    l.id,
    p.workspace_id AS project_workspace_id,
    EXISTS(
        SELECT 1 FROM "workspace_labels" wl
        WHERE wl.label_id = l.id AND wl.workspace_id = $2
    ) AS in_target_workspace
FROM "labels" l
JOIN "projects" p ON l.project_id = p.id
WHERE l.id = $1;

-- name: AssignTaskLabel :exec
INSERT INTO "task_labels" (task_id, label_id, created_at, updated_at)
VALUES ($1, $2, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- name: RemoveTaskLabel :execrows
DELETE FROM "task_labels"
WHERE task_id = $1 AND label_id = $2;

-- name: DeleteAllTaskLabelsByTaskID :exec
DELETE FROM "task_labels"
WHERE task_id = $1;

-- name: ListWorkspaceValidLabelIDs :many
SELECT l.id
FROM "labels" l
JOIN "projects" p ON l.project_id = p.id
WHERE l.id = ANY($1::uuid[])
  AND (
      p.workspace_id = $2
      OR EXISTS(
          SELECT 1 FROM "workspace_labels" wl
          WHERE wl.label_id = l.id AND wl.workspace_id = $2
      )
  );

-- name: ListLabelsByTaskID :many
SELECT
    l.id, l.name, l.color, l.description, l.project_id,
    l.created_by_id, l.updated_by_id, l.created_at, l.updated_at,
    p.name AS project_name, p.slug AS project_slug
FROM "task_labels" tl
JOIN "labels" l ON tl.label_id = l.id
JOIN "projects" p ON l.project_id = p.id
WHERE tl.task_id = $1;
