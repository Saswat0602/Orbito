-- name: CheckCommentAccessTaskAndUser :one
SELECT
    t.id AS task_id,
    t.title AS task_title,
    t.slug AS task_slug,
    t.is_archived,
    t.allow_email_replies,
    t.project_id,
    p.workspace_id,
    w.organization_id,
    o.archive AS org_archived,
    EXISTS(SELECT 1 FROM "users" u WHERE u.id = sqlc.arg(user_id)::uuid) AS user_exists,
    pm.role AS project_role,
    wm.role AS workspace_role,
    om.role AS organization_role
FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
JOIN "organizations" o ON w.organization_id = o.id
LEFT JOIN "project_members" pm ON pm.project_id = t.project_id AND pm.user_id = sqlc.arg(user_id)::uuid
LEFT JOIN "workspace_members" wm ON wm.workspace_id = p.workspace_id AND wm.user_id = sqlc.arg(user_id)::uuid
LEFT JOIN "organization_members" om ON om.organization_id = w.organization_id AND om.user_id = sqlc.arg(user_id)::uuid
WHERE (sqlc.arg(is_uuid)::boolean = true AND t.id = sqlc.narg(task_uuid)::uuid)
   OR (sqlc.arg(is_uuid)::boolean = false AND t.slug = sqlc.arg(task_slug)::text)
LIMIT 1;

-- name: GetCommentBasicByID :one
SELECT id, author_id, task_id, content, parent_comment_id
FROM "task_comments"
WHERE id = $1;

-- name: CreateTaskCommentRecord :one
INSERT INTO "task_comments" (
    id, content, task_id, author_id, parent_comment_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING id;

-- name: GetTaskCommentDetailByID :one
SELECT
    tc.id, tc.content, tc.task_id, tc.author_id, tc.parent_comment_id,
    tc.created_by_id, tc.updated_by_id, tc.created_at, tc.updated_at,
    tc.email_message_id, tc.sent_as_email, tc.email_recipients, tc.email_recipient_names, tc.email_sent_at,
    au.email AS author_email, au.first_name AS author_first_name, au.last_name AS author_last_name, au.avatar AS author_avatar,
    t.title AS task_title, t.slug AS task_slug, t.allow_email_replies AS task_allow_email_replies,
    p.id AS project_id, p.name AS project_name, p.slug AS project_slug,
    pc.id AS parent_id, pc.content AS parent_content, pc.created_at AS parent_created_at,
    pau.id AS parent_author_id, pau.first_name AS parent_author_first_name, pau.last_name AS parent_author_last_name, pau.avatar AS parent_author_avatar,
    (SELECT COUNT(*) FROM "task_comments" r WHERE r.parent_comment_id = tc.id)::bigint AS reply_count
FROM "task_comments" tc
JOIN "users" au ON tc.author_id = au.id
JOIN "tasks" t ON tc.task_id = t.id
JOIN "projects" p ON t.project_id = p.id
LEFT JOIN "task_comments" pc ON tc.parent_comment_id = pc.id
LEFT JOIN "users" pau ON pc.author_id = pau.id
WHERE tc.id = $1;

-- name: CountTopLevelTaskComments :one
SELECT COUNT(*)::bigint
FROM "task_comments"
WHERE task_id = $1 AND parent_comment_id IS NULL;

-- name: ListTopLevelTaskCommentsAsc :many
SELECT
    tc.id, tc.content, tc.task_id, tc.author_id, tc.parent_comment_id,
    tc.created_by_id, tc.updated_by_id, tc.created_at, tc.updated_at,
    tc.email_message_id, tc.sent_as_email, tc.email_recipients, tc.email_recipient_names, tc.email_sent_at,
    au.email AS author_email, au.first_name AS author_first_name, au.last_name AS author_last_name, au.avatar AS author_avatar,
    t.title AS task_title, t.slug AS task_slug,
    (SELECT COUNT(*) FROM "task_comments" r WHERE r.parent_comment_id = tc.id)::bigint AS reply_count
FROM "task_comments" tc
JOIN "users" au ON tc.author_id = au.id
JOIN "tasks" t ON tc.task_id = t.id
WHERE tc.task_id = $1 AND tc.parent_comment_id IS NULL
ORDER BY tc.created_at ASC
LIMIT $2 OFFSET $3;

-- name: ListTopLevelTaskCommentsDesc :many
SELECT
    tc.id, tc.content, tc.task_id, tc.author_id, tc.parent_comment_id,
    tc.created_by_id, tc.updated_by_id, tc.created_at, tc.updated_at,
    tc.email_message_id, tc.sent_as_email, tc.email_recipients, tc.email_recipient_names, tc.email_sent_at,
    au.email AS author_email, au.first_name AS author_first_name, au.last_name AS author_last_name, au.avatar AS author_avatar,
    t.title AS task_title, t.slug AS task_slug,
    (SELECT COUNT(*) FROM "task_comments" r WHERE r.parent_comment_id = tc.id)::bigint AS reply_count
FROM "task_comments" tc
JOIN "users" au ON tc.author_id = au.id
JOIN "tasks" t ON tc.task_id = t.id
WHERE tc.task_id = $1 AND tc.parent_comment_id IS NULL
ORDER BY tc.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListRepliesByParentIDs :many
SELECT
    tc.id, tc.content, tc.task_id, tc.author_id, tc.parent_comment_id,
    tc.created_by_id, tc.updated_by_id, tc.created_at, tc.updated_at,
    au.first_name AS author_first_name, au.last_name AS author_last_name, au.avatar AS author_avatar,
    (SELECT COUNT(*) FROM "task_comments" r WHERE r.parent_comment_id = tc.id)::bigint AS reply_count
FROM "task_comments" tc
JOIN "users" au ON tc.author_id = au.id
WHERE tc.parent_comment_id = ANY($1::uuid[])
ORDER BY tc.created_at ASC;

-- name: UpdateTaskCommentContentByID :one
UPDATE "task_comments"
SET content = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id;

-- name: DeleteTaskCommentByID :execrows
DELETE FROM "task_comments" WHERE id = $1;

