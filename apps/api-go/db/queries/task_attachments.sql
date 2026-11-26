-- name: CheckAttachmentAccessTaskAndUser :one
SELECT
    t.id AS task_id,
    t.title AS task_title,
    t.slug AS task_slug,
    t.project_id,
    p.name AS project_name,
    p.slug AS project_slug,
    p.workspace_id,
    w.organization_id,
    EXISTS(SELECT 1 FROM "users" u WHERE u.id = sqlc.arg(user_id)::uuid) AS user_exists,
    pm.role AS project_role,
    wm.role AS workspace_role,
    om.role AS organization_role
FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
LEFT JOIN "project_members" pm ON pm.project_id = t.project_id AND pm.user_id = sqlc.arg(user_id)::uuid
LEFT JOIN "workspace_members" wm ON wm.workspace_id = p.workspace_id AND wm.user_id = sqlc.arg(user_id)::uuid
LEFT JOIN "organization_members" om ON om.organization_id = w.organization_id AND om.user_id = sqlc.arg(user_id)::uuid
WHERE (sqlc.arg(is_uuid)::boolean = true AND t.id = sqlc.narg(task_uuid)::uuid)
   OR (sqlc.arg(is_uuid)::boolean = false AND t.slug = sqlc.arg(task_slug)::text)
LIMIT 1;

-- name: CreateTaskAttachmentRecord :one
INSERT INTO "task_attachments" (
    id, file_name, file_path, file_size, mime_type, url, storage_key, task_id, created_by_id, updated_by_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $9, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING id;

-- name: GetTaskAttachmentDetailByID :one
SELECT
    ta.id, ta.file_name, ta.file_path, ta.file_size, ta.mime_type, ta.url, ta.storage_key,
    ta.task_id, ta.created_by_id, ta.updated_by_id, ta.created_at, ta.updated_at,
    t.title AS task_title, t.slug AS task_slug, t.description AS task_description, t.task_number, t.priority AS task_priority, t.type AS task_type,
    ts.id AS status_id, ts.name AS status_name, ts.color AS status_color, ts.category AS status_category,
    p.id AS project_id, p.name AS project_name, p.slug AS project_slug,
    w.id AS workspace_id, w.name AS workspace_name, w.slug AS workspace_slug,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug,
    cu.id AS creator_id, cu.first_name AS creator_first_name, cu.last_name AS creator_last_name,
    cu.email AS creator_email, cu.username AS creator_username, cu.avatar AS creator_avatar
FROM "task_attachments" ta
JOIN "tasks" t ON ta.task_id = t.id
JOIN "task_statuses" ts ON t.status_id = ts.id
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
JOIN "organizations" o ON w.organization_id = o.id
LEFT JOIN "users" cu ON ta.created_by_id = cu.id
WHERE ta.id = $1;

-- name: ListTaskAttachmentDetailsByTaskID :many
SELECT
    ta.id, ta.file_name, ta.file_path, ta.file_size, ta.mime_type, ta.url, ta.storage_key,
    ta.task_id, ta.created_by_id, ta.updated_by_id, ta.created_at, ta.updated_at,
    t.title AS task_title, t.slug AS task_slug,
    p.id AS project_id, p.name AS project_name, p.slug AS project_slug,
    cu.id AS creator_id, cu.first_name AS creator_first_name, cu.last_name AS creator_last_name,
    cu.email AS creator_email, cu.avatar AS creator_avatar
FROM "task_attachments" ta
JOIN "tasks" t ON ta.task_id = t.id
JOIN "projects" p ON t.project_id = p.id
LEFT JOIN "users" cu ON ta.created_by_id = cu.id
WHERE ta.task_id = $1
ORDER BY ta.created_at DESC;

-- name: GroupAttachmentStatsByMime :many
SELECT
    mime_type,
    COUNT(*)::bigint AS count_files,
    COALESCE(SUM(file_size), 0)::bigint AS total_size
FROM "task_attachments"
WHERE task_id = $1
GROUP BY mime_type;

-- name: CountRecentTaskAttachments :one
SELECT COUNT(*)::bigint
FROM "task_attachments"
WHERE task_id = $1 AND created_at >= $2::timestamp;

-- name: DeleteTaskAttachmentByID :execrows
DELETE FROM "task_attachments" WHERE id = $1;
