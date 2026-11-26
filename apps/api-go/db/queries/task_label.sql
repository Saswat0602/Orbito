-- name: FindTaskForTaskLabel :one
SELECT
    t.id AS task_id,
    t.title AS task_title,
    t.slug AS task_slug,
    p.id AS project_id,
    p.workspace_id,
    p.archive AS project_archived,
    EXISTS(
        SELECT 1 FROM "workspace_members" wm
        WHERE wm.workspace_id = p.workspace_id AND wm.user_id = sqlc.arg(user_id)::uuid
    ) AS is_ws_member
FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
WHERE (sqlc.arg(is_uuid)::boolean = true AND t.id = sqlc.narg(task_uuid)::uuid)
   OR (sqlc.arg(is_uuid)::boolean = false AND t.slug = sqlc.arg(task_slug)::text)
LIMIT 1;

-- name: GetLabelWorkspaceForTaskLabel :one
SELECT
    l.id, l.name, l.color, l.description,
    p.workspace_id AS project_workspace_id,
    EXISTS(
        SELECT 1 FROM "workspace_labels" wl
        WHERE wl.label_id = l.id AND wl.workspace_id = $2
    ) AS in_target_ws
FROM "labels" l
JOIN "projects" p ON l.project_id = p.id
WHERE l.id = $1;

-- name: CheckTaskLabelAssignmentExists :one
SELECT EXISTS(
    SELECT 1 FROM "task_labels"
    WHERE task_id = $1 AND label_id = $2
);

-- name: InsertTaskLabelWithAudit :exec
INSERT INTO "task_labels" (task_id, label_id, created_by_id, updated_by_id, created_at, updated_at)
VALUES ($1, $2, $3, $3, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- name: GetTaskLabelPairDetails :one
SELECT
    tl.task_id, tl.label_id, tl.created_by_id, tl.updated_by_id, tl.created_at, tl.updated_at,
    l.name AS label_name, l.color AS label_color, l.description AS label_description,
    t.title AS task_title, t.slug AS task_slug
FROM "task_labels" tl
JOIN "labels" l ON tl.label_id = l.id
JOIN "tasks" t ON tl.task_id = t.id
WHERE tl.task_id = $1 AND tl.label_id = $2;

-- name: ListAccessibleProjectIDsForTaskLabels :many
SELECT p.id
FROM "projects" p
JOIN "workspace_members" wm ON p.workspace_id = wm.workspace_id
WHERE wm.user_id = $1;

-- name: ListTaskLabelsByProjectIDs :many
SELECT
    tl.task_id, tl.label_id, tl.created_by_id, tl.updated_by_id, tl.created_at, tl.updated_at,
    l.name AS label_name, l.color AS label_color, l.description AS label_description,
    t.title AS task_title, t.slug AS task_slug
FROM "task_labels" tl
JOIN "labels" l ON tl.label_id = l.id
JOIN "tasks" t ON tl.task_id = t.id
WHERE t.project_id = ANY($1::uuid[])
ORDER BY tl.created_at DESC;

-- name: DeleteTaskLabelPair :execrows
DELETE FROM "task_labels"
WHERE task_id = $1 AND label_id = $2;


-- name: DeleteTaskLabelsByTaskID :exec
DELETE FROM "task_labels"
WHERE task_id = $1;

-- name: UpsertWorkspaceLabelForTaskLabel :exec
INSERT INTO "workspace_labels" (workspace_id, label_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: FindLabelsInWorkspaceByIDs :many
SELECT l.id
FROM "labels" l
JOIN "projects" p ON l.project_id = p.id
WHERE l.id = ANY($1::uuid[])
  AND (
      p.workspace_id = $2::uuid
      OR EXISTS(
          SELECT 1 FROM "workspace_labels" wl
          WHERE wl.label_id = l.id AND wl.workspace_id = $2::uuid
      )
  );
