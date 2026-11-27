-- name: CreateNotificationRecord :one
INSERT INTO "notifications" (
    id, title, message, type, priority, "isRead",
    entity_type, entity_id, action_url, user_id, organization_id,
    created_by_id, updated_by_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, false,
    $6, $7, $8, $9, $10,
    $11, $11, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: ListUserNotifications :many
SELECT
    n.id, n.title, n.message, n.type, n.priority, n."isRead",
    n.entity_type, n.entity_id, n.action_url, n.user_id, n.organization_id,
    n.created_by_id, n.updated_by_id, n.created_at, n.updated_at, n.read_at,
    cb.id AS creator_id, cb.first_name AS creator_first_name,
    cb.last_name AS creator_last_name, cb.avatar AS creator_avatar
FROM "notifications" n
LEFT JOIN "users" cb ON n.created_by_id = cb.id
WHERE n.user_id = sqlc.arg(user_id)::uuid
  AND (sqlc.narg(is_read)::boolean IS NULL OR n."isRead" = sqlc.narg(is_read)::boolean)
  AND (sqlc.narg(notif_type)::"NotificationType" IS NULL OR n.type = sqlc.narg(notif_type)::"NotificationType")
  AND (sqlc.narg(organization_id)::uuid IS NULL OR n.organization_id = sqlc.narg(organization_id)::uuid)
  AND (sqlc.narg(since)::timestamp IS NULL OR n.created_at >= sqlc.narg(since)::timestamp)
ORDER BY n.created_at DESC
LIMIT sqlc.arg(limit_val)::integer
OFFSET sqlc.arg(offset_val)::integer;

-- name: CountUserNotifications :one
SELECT COUNT(*)::bigint
FROM "notifications" n
WHERE n.user_id = sqlc.arg(user_id)::uuid
  AND (sqlc.narg(is_read)::boolean IS NULL OR n."isRead" = sqlc.narg(is_read)::boolean)
  AND (sqlc.narg(notif_type)::"NotificationType" IS NULL OR n.type = sqlc.narg(notif_type)::"NotificationType")
  AND (sqlc.narg(organization_id)::uuid IS NULL OR n.organization_id = sqlc.narg(organization_id)::uuid)
  AND (sqlc.narg(since)::timestamp IS NULL OR n.created_at >= sqlc.narg(since)::timestamp);

-- name: ListNotificationsByUserOrgDetailed :many
SELECT
    n.id, n.title, n.message, n.type, n.priority, n."isRead",
    n.entity_type, n.entity_id, n.action_url, n.user_id, n.organization_id,
    n.created_by_id, n.updated_by_id, n.created_at, n.updated_at, n.read_at,
    u.id AS recipient_id, u.first_name AS recipient_first_name, u.last_name AS recipient_last_name,
    u.email AS recipient_email, u.username AS recipient_username, u.avatar AS recipient_avatar,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug, o.avatar AS org_avatar,
    cb.id AS creator_id, cb.first_name AS creator_first_name, cb.last_name AS creator_last_name,
    cb.email AS creator_email, cb.avatar AS creator_avatar
FROM "notifications" n
JOIN "users" u ON n.user_id = u.id
LEFT JOIN "organizations" o ON n.organization_id = o.id
LEFT JOIN "users" cb ON n.created_by_id = cb.id
WHERE n.user_id = sqlc.arg(user_id)::uuid
  AND n.organization_id = sqlc.arg(organization_id)::uuid
  AND (sqlc.narg(is_read)::boolean IS NULL OR n."isRead" = sqlc.narg(is_read)::boolean)
  AND (sqlc.narg(notif_type)::"NotificationType" IS NULL OR n.type = sqlc.narg(notif_type)::"NotificationType")
  AND (sqlc.narg(priority)::"NotificationPriority" IS NULL OR n.priority = sqlc.narg(priority)::"NotificationPriority")
  AND (sqlc.narg(start_date)::timestamp IS NULL OR n.created_at >= sqlc.narg(start_date)::timestamp)
  AND (sqlc.narg(end_date)::timestamp IS NULL OR n.created_at <= sqlc.narg(end_date)::timestamp)
ORDER BY n."isRead" ASC, n.created_at DESC
LIMIT sqlc.arg(limit_val)::integer
OFFSET sqlc.arg(offset_val)::integer;

-- name: CountNotificationsByUserOrgDetailed :one
SELECT COUNT(*)::bigint
FROM "notifications" n
WHERE n.user_id = sqlc.arg(user_id)::uuid
  AND n.organization_id = sqlc.arg(organization_id)::uuid
  AND (sqlc.narg(is_read)::boolean IS NULL OR n."isRead" = sqlc.narg(is_read)::boolean)
  AND (sqlc.narg(notif_type)::"NotificationType" IS NULL OR n.type = sqlc.narg(notif_type)::"NotificationType")
  AND (sqlc.narg(priority)::"NotificationPriority" IS NULL OR n.priority = sqlc.narg(priority)::"NotificationPriority")
  AND (sqlc.narg(start_date)::timestamp IS NULL OR n.created_at >= sqlc.narg(start_date)::timestamp)
  AND (sqlc.narg(end_date)::timestamp IS NULL OR n.created_at <= sqlc.narg(end_date)::timestamp);

-- name: GroupNotificationsByTypeForUserOrg :many
SELECT n.type, COUNT(*)::bigint AS cnt
FROM "notifications" n
WHERE n.user_id = sqlc.arg(user_id)::uuid
  AND n.organization_id = sqlc.arg(organization_id)::uuid
  AND (sqlc.narg(is_read)::boolean IS NULL OR n."isRead" = sqlc.narg(is_read)::boolean)
  AND (sqlc.narg(notif_type)::"NotificationType" IS NULL OR n.type = sqlc.narg(notif_type)::"NotificationType")
  AND (sqlc.narg(priority)::"NotificationPriority" IS NULL OR n.priority = sqlc.narg(priority)::"NotificationPriority")
  AND (sqlc.narg(start_date)::timestamp IS NULL OR n.created_at >= sqlc.narg(start_date)::timestamp)
  AND (sqlc.narg(end_date)::timestamp IS NULL OR n.created_at <= sqlc.narg(end_date)::timestamp)
GROUP BY n.type;

-- name: GroupNotificationsByPriorityForUserOrg :many
SELECT n.priority, COUNT(*)::bigint AS cnt
FROM "notifications" n
WHERE n.user_id = sqlc.arg(user_id)::uuid
  AND n.organization_id = sqlc.arg(organization_id)::uuid
  AND (sqlc.narg(is_read)::boolean IS NULL OR n."isRead" = sqlc.narg(is_read)::boolean)
  AND (sqlc.narg(notif_type)::"NotificationType" IS NULL OR n.type = sqlc.narg(notif_type)::"NotificationType")
  AND (sqlc.narg(priority)::"NotificationPriority" IS NULL OR n.priority = sqlc.narg(priority)::"NotificationPriority")
  AND (sqlc.narg(start_date)::timestamp IS NULL OR n.created_at >= sqlc.narg(start_date)::timestamp)
  AND (sqlc.narg(end_date)::timestamp IS NULL OR n.created_at <= sqlc.narg(end_date)::timestamp)
GROUP BY n.priority;

-- name: GetNotificationByIDForUser :one
SELECT
    n.id, n.title, n.message, n.type, n.priority, n."isRead",
    n.entity_type, n.entity_id, n.action_url, n.user_id, n.organization_id,
    n.created_by_id, n.updated_by_id, n.created_at, n.updated_at, n.read_at,
    cb.id AS creator_id, cb.first_name AS creator_first_name,
    cb.last_name AS creator_last_name, cb.avatar AS creator_avatar
FROM "notifications" n
LEFT JOIN "users" cb ON n.created_by_id = cb.id
WHERE n.id = $1 AND n.user_id = $2
LIMIT 1;

-- name: MarkNotificationAsRead :execrows
UPDATE "notifications"
SET "isRead" = true,
    read_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2;

-- name: MarkAllNotificationsAsRead :execrows
UPDATE "notifications"
SET "isRead" = true,
    read_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE user_id = sqlc.arg(user_id)::uuid
  AND "isRead" = false
  AND (sqlc.narg(organization_id)::uuid IS NULL OR organization_id = sqlc.narg(organization_id)::uuid);

-- name: CountUnreadNotifications :one
SELECT COUNT(*)::bigint
FROM "notifications"
WHERE user_id = sqlc.arg(user_id)::uuid
  AND "isRead" = false
  AND (sqlc.narg(organization_id)::uuid IS NULL OR organization_id = sqlc.narg(organization_id)::uuid);

-- name: CountUnreadNotificationsByOrganization :many
SELECT
    n.organization_id,
    COALESCE(o.name, 'Unknown')::text AS organization_name,
    COUNT(*)::bigint AS unread_count
FROM "notifications" n
LEFT JOIN "organizations" o ON n.organization_id = o.id
WHERE n.user_id = $1
  AND n."isRead" = false
  AND n.organization_id IS NOT NULL
GROUP BY n.organization_id, o.name;

-- name: GroupNotificationsByTypeForUser :many
SELECT n.type, COUNT(*)::bigint AS cnt
FROM "notifications" n
WHERE n.user_id = sqlc.arg(user_id)::uuid
  AND (sqlc.narg(organization_id)::uuid IS NULL OR n.organization_id = sqlc.narg(organization_id)::uuid)
GROUP BY n.type;

-- name: DeleteNotificationForUser :execrows
DELETE FROM "notifications"
WHERE id = $1 AND user_id = $2;

-- name: DeleteMultipleNotificationsForUser :execrows
DELETE FROM "notifications"
WHERE user_id = sqlc.arg(user_id)::uuid
  AND id = ANY(sqlc.arg(ids)::uuid[]);

-- name: GetNotificationEntityTask :one
SELECT
    t.id, t.title, t.slug, t.priority, t.task_number,
    ts.id AS status_id, ts.name AS status_name, ts.color AS status_color, ts.category AS status_category,
    p.id AS project_id, p.name AS project_name, p.slug AS project_slug
FROM "tasks" t
LEFT JOIN "task_statuses" ts ON t.status_id = ts.id
LEFT JOIN "projects" p ON t.project_id = p.id
WHERE t.id = $1
LIMIT 1;

-- name: GetNotificationEntityTaskComment :one
SELECT
    tc.id, tc.content, tc.created_at,
    t.id AS task_id, t.title AS task_title, t.slug AS task_slug, t.task_number AS task_number
FROM "task_comments" tc
LEFT JOIN "tasks" t ON tc.task_id = t.id
WHERE tc.id = $1
LIMIT 1;

-- name: GetNotificationEntityProject :one
SELECT
    p.id, p.name, p.slug, p.description, p.avatar,
    w.id AS workspace_id, w.name AS workspace_name, w.slug AS workspace_slug
FROM "projects" p
LEFT JOIN "workspaces" w ON p.workspace_id = w.id
WHERE p.id = $1
LIMIT 1;

-- name: GetNotificationEntityWorkspace :one
SELECT
    w.id, w.name, w.slug, w.description, w.avatar,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug
FROM "workspaces" w
LEFT JOIN "organizations" o ON w.organization_id = o.id
WHERE w.id = $1
LIMIT 1;

-- name: GetNotificationEntityOrganization :one
SELECT id, name, slug, description, avatar
FROM "organizations"
WHERE id = $1
LIMIT 1;

-- name: GetNotificationEntitySprint :one
SELECT
    s.id, s.name, s.goal, s.status, s.start_date, s.end_date,
    p.id AS project_id, p.name AS project_name, p.slug AS project_slug
FROM "sprints" s
LEFT JOIN "projects" p ON s.project_id = p.id
WHERE s.id = $1
LIMIT 1;

-- name: GetNotificationEntityInvitation :one
SELECT
    i.id, i.invitee_email, i.role, i.status, i.expires_at,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug,
    w.id AS workspace_id, w.name AS workspace_name, w.slug AS workspace_slug,
    p.id AS project_id, p.name AS project_name, p.slug AS project_slug
FROM "invitations" i
LEFT JOIN "organizations" o ON i.organization_id = o.id
LEFT JOIN "workspaces" w ON i.workspace_id = w.id
LEFT JOIN "projects" p ON i.project_id = p.id
WHERE i.id = $1
LIMIT 1;

-- name: GetNotificationEntityTaskAttachment :one
SELECT
    ta.id, ta.file_name, ta.file_size, ta.mime_type,
    t.id AS task_id, t.title AS task_title, t.slug AS task_slug
FROM "task_attachments" ta
LEFT JOIN "tasks" t ON ta.task_id = t.id
WHERE ta.id = $1
LIMIT 1;

-- name: GetTaskNotificationContext :one
SELECT
    t.id, t.title, t.slug, t.due_date, w.organization_id
FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
WHERE t.id = $1
LIMIT 1;

-- name: ListTaskNotificationParticipants :many
SELECT ta.user_id, 'assignee'::text AS role_kind FROM "task_assignees" ta WHERE ta.task_id = sqlc.arg(task_id)::uuid
UNION
SELECT tr.user_id, 'reporter'::text AS role_kind FROM "task_reporters" tr WHERE tr.task_id = sqlc.arg(task_id)::uuid
UNION
SELECT tw.user_id, 'watcher'::text AS role_kind FROM "task_watchers" tw WHERE tw.task_id = sqlc.arg(task_id)::uuid;

-- name: FindMentionedUsers :many
SELECT id, username, email, first_name, last_name
FROM "users"
WHERE deleted_at IS NULL
  AND (
    (cardinality(sqlc.arg(usernames)::text[]) > 0 AND username = ANY(sqlc.arg(usernames)::text[]))
    OR
    (cardinality(sqlc.arg(user_ids)::uuid[]) > 0 AND id = ANY(sqlc.arg(user_ids)::uuid[]))
  );

-- name: GetUserBriefForNotification :one
SELECT id, first_name, last_name, email, username
FROM "users"
WHERE id = $1
LIMIT 1;
