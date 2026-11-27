-- name: GetProjectMetaForTaskCreate :one
SELECT
    p.id, p.slug, p.task_prefix, p.workflow_id, p.workspace_id,
    w.organization_id, o.owner_id AS org_owner_id, o.archive AS org_archived
FROM "projects" p
JOIN "workspaces" w ON p.workspace_id = w.id
JOIN "organizations" o ON w.organization_id = o.id
WHERE p.id = $1;

-- name: LockProjectRowForTaskNumber :one
SELECT slug, task_prefix
FROM "projects"
WHERE id = $1
FOR UPDATE;

-- name: GetMaxTaskNumberForProject :one
SELECT COALESCE(MAX(task_number), 0)::integer AS max_num
FROM "tasks"
WHERE project_id = $1;

-- name: GetDefaultSprintForProject :one
SELECT id
FROM "sprints"
WHERE project_id = $1 AND is_default = true
LIMIT 1;

-- name: GetFirstStatusInWorkflow :one
SELECT id
FROM "task_statuses"
WHERE workflow_id = $1
ORDER BY position ASC
LIMIT 1;

-- name: GetAnyFirstStatus :one
SELECT id
FROM "task_statuses"
ORDER BY position ASC
LIMIT 1;

-- name: ResolveTaskIDBySlug :one
SELECT id
FROM "tasks"
WHERE slug = $1
LIMIT 1;

-- name: CreateTaskRecord :one
INSERT INTO "tasks" (
    id, title, description, type, priority, task_number, slug,
    start_date, due_date, completed_at, story_points, original_estimate, remaining_estimate,
    custom_fields, project_id, status_id, sprint_id, parent_task_id,
    created_by_id, updated_by_id, created_at, updated_at, allow_email_replies, is_recurring
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11, $12, $13,
    $14, $15, $16, $17, $18,
    $19, $20, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, $21, $22
) RETURNING *;

-- name: InsertTaskAssignee :exec
INSERT INTO "task_assignees" (task_id, user_id, created_at)
VALUES ($1, $2, CURRENT_TIMESTAMP)
ON CONFLICT (task_id, user_id) DO NOTHING;

-- name: InsertTaskReporter :exec
INSERT INTO "task_reporters" (task_id, user_id, created_at)
VALUES ($1, $2, CURRENT_TIMESTAMP)
ON CONFLICT (task_id, user_id) DO NOTHING;

-- name: DeleteTaskAssigneesByTaskID :exec
DELETE FROM "task_assignees" WHERE task_id = $1;

-- name: DeleteTaskReportersByTaskID :exec
DELETE FROM "task_reporters" WHERE task_id = $1;

-- name: CheckUsersExistByIDs :many
SELECT id FROM "users" WHERE id = ANY($1::uuid[]);

-- name: CreateTaskAttachmentForTask :one
INSERT INTO "task_attachments" (
    id, file_name, file_path, file_size, mime_type, url, storage_key,
    task_id, created_by_id, updated_by_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING id, file_name, file_size, mime_type, url, storage_key, created_at;

-- name: GetTaskAccessDetails :one
SELECT
    t.id, t.project_id, t.created_by_id, t.start_date, t.due_date, t.status_id,
    p.visibility, p.workspace_id,
    w.organization_id, o.owner_id AS org_owner_id, o.archive AS org_archived,
    om.role AS org_role, wm.role AS ws_role, pm.role AS project_role,
    EXISTS(SELECT 1 FROM "task_assignees" ta WHERE ta.task_id = t.id AND ta.user_id = sqlc.arg(user_id)::uuid) AS is_assignee,
    EXISTS(SELECT 1 FROM "task_reporters" tr WHERE tr.task_id = t.id AND tr.user_id = sqlc.arg(user_id)::uuid) AS is_reporter
FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
JOIN "organizations" o ON w.organization_id = o.id
LEFT JOIN "organization_members" om ON om.organization_id = o.id AND om.user_id = sqlc.arg(user_id)::uuid
LEFT JOIN "workspace_members" wm ON wm.workspace_id = w.id AND wm.user_id = sqlc.arg(user_id)::uuid
LEFT JOIN "project_members" pm ON pm.project_id = p.id AND pm.user_id = sqlc.arg(user_id)::uuid
WHERE (sqlc.arg(is_uuid)::boolean = true AND t.id = sqlc.narg(task_uuid)::uuid)
   OR (sqlc.arg(is_uuid)::boolean = false AND t.slug = sqlc.arg(task_slug)::text)
LIMIT 1;

-- name: GetTaskFullByID :one
SELECT
    t.*,
    p.name AS project_name, p.slug AS project_slug,
    w.id AS workspace_id, w.name AS workspace_name, w.slug AS workspace_slug, w.organization_id,
    o.name AS org_name, o.slug AS org_slug,
    ts.name AS status_name, ts.color AS status_color, ts.category AS status_category,
    sp.name AS sprint_name, sp.slug AS sprint_slug, sp.status AS sprint_status, sp.start_date AS sprint_start_date, sp.end_date AS sprint_end_date,
    pt.title AS parent_title, pt.slug AS parent_slug, pt.type AS parent_type,
    cu.first_name AS creator_first_name, cu.last_name AS creator_last_name,
    (SELECT COUNT(*) FROM "tasks" ct WHERE ct.parent_task_id = t.id) AS child_task_count,
    (SELECT COUNT(*) FROM "task_comments" tc WHERE tc.task_id = t.id) AS comment_count,
    (SELECT COUNT(*) FROM "task_attachments" tat WHERE tat.task_id = t.id) AS attachment_count,
    (SELECT COUNT(*) FROM "task_watchers" tw WHERE tw.task_id = t.id) AS watcher_count,
    (SELECT COUNT(*) FROM "time_entries" te WHERE te.task_id = t.id) AS time_entry_count
FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
JOIN "organizations" o ON w.organization_id = o.id
JOIN "task_statuses" ts ON t.status_id = ts.id
LEFT JOIN "sprints" sp ON t.sprint_id = sp.id
LEFT JOIN "tasks" pt ON t.parent_task_id = pt.id
LEFT JOIN "users" cu ON t.created_by_id = cu.id
WHERE t.id = $1;

-- name: ListTaskAssigneesBatch :many
SELECT ta.task_id, u.id, u.email, u.first_name, u.last_name, u.avatar
FROM "task_assignees" ta
JOIN "users" u ON ta.user_id = u.id
WHERE ta.task_id = ANY($1::uuid[]);

-- name: ListTaskReportersBatch :many
SELECT tr.task_id, u.id, u.email, u.first_name, u.last_name, u.avatar
FROM "task_reporters" tr
JOIN "users" u ON tr.user_id = u.id
WHERE tr.task_id = ANY($1::uuid[]);

-- name: ListTaskLabelsBatch :many
SELECT tl.task_id, tl.label_id, l.name, l.color, l.description
FROM "task_labels" tl
JOIN "labels" l ON tl.label_id = l.id
WHERE tl.task_id = ANY($1::uuid[]);

-- name: ListTaskAttachmentsByTaskID :many
SELECT id, file_name, file_size, mime_type, url, storage_key, created_at
FROM "task_attachments"
WHERE task_id = $1
ORDER BY created_at DESC;

-- name: ListTaskWatchersByTaskID :many
SELECT tw.id, tw.task_id, tw.user_id, u.first_name, u.last_name, u.avatar
FROM "task_watchers" tw
JOIN "users" u ON tw.user_id = u.id
WHERE tw.task_id = $1;

-- name: ListTaskCommentsByTaskID :many
SELECT tc.id, tc.content, tc.task_id, tc.author_id, tc.parent_comment_id, tc.created_at, tc.updated_at,
       u.first_name, u.last_name, u.avatar
FROM "task_comments" tc
JOIN "users" u ON tc.author_id = u.id
WHERE tc.task_id = $1
ORDER BY tc.created_at DESC;

-- name: ListTaskTimeEntriesByTaskID :many
SELECT te.id, te.description, te.time_spent, te.start_time, te.end_time, te.date, te.task_id, te.user_id, te.created_at, te.updated_at,
       u.first_name, u.last_name, u.avatar
FROM "time_entries" te
JOIN "users" u ON te.user_id = u.id
WHERE te.task_id = $1
ORDER BY te.date DESC;

-- name: ListTaskChildrenForDetail :many
SELECT t.id, t.title, t.slug, t.type, t.priority, t.created_by_id,
       ts.name AS status_name, ts.color AS status_color, ts.category AS status_category
FROM "tasks" t
JOIN "task_statuses" ts ON t.status_id = ts.id
WHERE t.parent_task_id = $1;

-- name: GetProjectInboxForTaskDetail :one
SELECT id, project_id, enabled, email_address
FROM "project_inboxes"
WHERE project_id = $1
LIMIT 1;

-- name: GetTaskStatusBasicByID :one
SELECT id, name, color, category, workflow_id
FROM "task_statuses"
WHERE id = $1;

-- name: UpdateTaskCoreByID :one
UPDATE "tasks"
SET
    title = COALESCE(sqlc.narg(title), title),
    description = COALESCE(sqlc.narg(description), description),
    type = COALESCE(sqlc.narg(type)::"TaskType", type),
    priority = COALESCE(sqlc.narg(priority)::"TaskPriority", priority),
    start_date = COALESCE(sqlc.narg(start_date), start_date),
    due_date = COALESCE(sqlc.narg(due_date), due_date),
    completed_at = CASE WHEN sqlc.arg(clear_completed_at)::boolean THEN NULL ELSE COALESCE(sqlc.narg(completed_at), completed_at) END,
    story_points = COALESCE(sqlc.narg(story_points), story_points),
    original_estimate = COALESCE(sqlc.narg(original_estimate), original_estimate),
    remaining_estimate = COALESCE(sqlc.narg(remaining_estimate), remaining_estimate),
    custom_fields = COALESCE(sqlc.narg(custom_fields), custom_fields),
    status_id = COALESCE(sqlc.narg(status_id), status_id),
    sprint_id = CASE WHEN sqlc.arg(clear_sprint_id)::boolean THEN NULL ELSE COALESCE(sqlc.narg(sprint_id), sprint_id) END,
    parent_task_id = CASE WHEN sqlc.arg(clear_parent_task_id)::boolean THEN NULL ELSE COALESCE(sqlc.narg(parent_task_id), parent_task_id) END,
    allow_email_replies = COALESCE(sqlc.narg(allow_email_replies), allow_email_replies),
    updated_by_id = sqlc.narg(updated_by_id),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
RETURNING id;

-- name: DeleteTaskByID :execrows
DELETE FROM "tasks" WHERE id = $1;

-- name: CreateQuickTaskComment :one
INSERT INTO "task_comments" (
    id, content, task_id, author_id, created_by_id, updated_by_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $4, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING id, content, task_id, author_id, parent_comment_id, created_at, updated_at;

-- name: GetCommentAuthorBrief :one
SELECT id, first_name, last_name, avatar FROM "users" WHERE id = $1;

-- name: CreateRecurringTaskConfig :one
INSERT INTO "recurring_tasks" (
    id, task_id, recurrence_type, interval, days_of_week, day_of_month, month_of_year,
    end_type, end_date, occurrence_count, current_occurrence, is_active, next_occurrence,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11, true, $12,
    CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: GetRecurringTaskConfigByTaskID :one
SELECT * FROM "recurring_tasks" WHERE task_id = $1 LIMIT 1;

-- name: UpdateRecurringTaskConfigByID :one
UPDATE "recurring_tasks"
SET
    recurrence_type = $2,
    interval = $3,
    days_of_week = $4,
    day_of_month = $5,
    month_of_year = $6,
    end_type = $7,
    end_date = $8,
    occurrence_count = $9,
    next_occurrence = $10,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: IncrementRecurringOccurrence :exec
UPDATE "recurring_tasks"
SET current_occurrence = $2, next_occurrence = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteRecurringConfigByID :exec
DELETE FROM "recurring_tasks" WHERE id = $1;

-- name: SetTaskIsRecurring :one
UPDATE "tasks" SET is_recurring = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $1 RETURNING *;

-- name: ListRecurringTasksByProjectID :many
SELECT
    t.*,
    rt.id AS rec_id, rt.recurrence_type, rt.interval, rt.days_of_week, rt.day_of_month, rt.month_of_year,
    rt.end_type, rt.end_date, rt.occurrence_count, rt.current_occurrence, rt.is_active AS rec_is_active, rt.next_occurrence,
    ts.name AS status_name, ts.color AS status_color, ts.category AS status_category
FROM "tasks" t
LEFT JOIN "recurring_tasks" rt ON rt.task_id = t.id
JOIN "task_statuses" ts ON t.status_id = ts.id
WHERE t.project_id = $1 AND t.is_recurring = true;

-- name: CreatePublicTaskShare :one
INSERT INTO "public_task_shares" (id, token, expires_at, task_id, created_by_id, created_at)
VALUES ($1, $2, $3, $4, $5, CURRENT_TIMESTAMP)
RETURNING *;

-- name: ListActiveSharesForTask :many
SELECT * FROM "public_task_shares"
WHERE task_id = $1 AND revoked_at IS NULL AND expires_at > CURRENT_TIMESTAMP
ORDER BY created_at DESC;

-- name: GetTaskShareWithAccess :one
SELECT
    s.*,
    t.project_id,
    p.workspace_id,
    EXISTS(SELECT 1 FROM "project_members" pm WHERE pm.project_id = p.id AND pm.user_id = $2) AS is_project_member,
    EXISTS(SELECT 1 FROM "workspace_members" wm WHERE wm.workspace_id = p.workspace_id AND wm.user_id = $2) AS is_ws_member
FROM "public_task_shares" s
JOIN "tasks" t ON s.task_id = t.id
JOIN "projects" p ON t.project_id = p.id
WHERE s.id = $1;

-- name: RevokePublicTaskShare :exec
UPDATE "public_task_shares" SET revoked_at = CURRENT_TIMESTAMP WHERE id = $1;

-- name: BulkDeleteTasksByIDs :execrows
DELETE FROM "tasks" WHERE id = ANY($1::uuid[]);

