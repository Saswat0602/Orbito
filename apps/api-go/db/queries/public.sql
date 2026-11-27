-- name: GetPublicSharedTaskByToken :one
SELECT
    pts.id AS share_id, pts.token, pts.expires_at, pts.revoked_at,
    t.id, t.title, t.description, t.type, t.priority, t.slug, t.due_date, t.created_at, t.updated_at,
    t.project_id, t.status_id,
    ts.id AS status_row_id, ts.name AS status_name, ts.color AS status_color, ts.category AS status_category,
    cu.id AS creator_id, cu.first_name AS creator_first_name, cu.last_name AS creator_last_name, cu.avatar AS creator_avatar,
    p.name AS project_name, p.slug AS project_slug, p.visibility AS project_visibility,
    w.slug AS workspace_slug, w.name AS workspace_name,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug
FROM "public_task_shares" pts
JOIN "tasks" t ON pts.task_id = t.id
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
JOIN "organizations" o ON w.organization_id = o.id
LEFT JOIN "task_statuses" ts ON t.status_id = ts.id
LEFT JOIN "users" cu ON t.created_by_id = cu.id
WHERE pts.token = $1
LIMIT 1;

-- name: GetPublicShareAttachmentByID :one
SELECT ta.id, ta.file_name, ta.file_path, ta.file_size, ta.mime_type, ta.url, ta.storage_key, ta.task_id
FROM "task_attachments" ta
JOIN "public_task_shares" pts ON pts.task_id = ta.task_id
WHERE pts.token = $1
  AND ta.id = $2
  AND pts.revoked_at IS NULL
  AND pts.expires_at > CURRENT_TIMESTAMP
LIMIT 1;

-- name: ListWorkspacePublicProjects :many
SELECT
    p.id, p.name, p.slug, p.description, p.color, p.avatar, p.status, p.priority,
    p.visibility, p.start_date, p.end_date, p.created_at, p.updated_at, p.workflow_id,
    w.slug AS workspace_slug, w.name AS workspace_name,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug
FROM "projects" p
JOIN "workspaces" w ON p.workspace_id = w.id
JOIN "organizations" o ON w.organization_id = o.id
WHERE w.slug = $1 AND p.visibility = 'PUBLIC'
ORDER BY p.updated_at DESC;

-- name: GetPublicProjectBySlugs :one
SELECT
    p.id, p.name, p.slug, p.description, p.color, p.avatar, p.status, p.priority,
    p.visibility, p.start_date, p.end_date, p.created_at, p.updated_at, p.workflow_id, p.settings,
    w.slug AS workspace_slug, w.name AS workspace_name,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug,
    (SELECT COUNT(*)::bigint FROM "tasks" t WHERE t.project_id = p.id) AS tasks_count,
    (SELECT COUNT(*)::bigint FROM "tasks" t2 JOIN "task_statuses" ts2 ON t2.status_id = ts2.id WHERE t2.project_id = p.id AND ts2.category = 'DONE') AS done_count,
    (SELECT COUNT(*)::bigint FROM "sprints" sp WHERE sp.project_id = p.id) AS sprints_count
FROM "projects" p
JOIN "workspaces" w ON p.workspace_id = w.id
JOIN "organizations" o ON w.organization_id = o.id
WHERE w.slug = $1 AND p.slug = $2 AND p.visibility = 'PUBLIC'
LIMIT 1;

-- name: GetPublicProjectBySlugOnly :one
SELECT
    p.id, p.name, p.slug, p.visibility, p.workflow_id
FROM "projects" p
WHERE p.slug = $1 AND p.visibility = 'PUBLIC'
LIMIT 1;

-- name: ListPublicProjectStatusesByWorkflow :many
SELECT
    ts.id, ts.name, ts.color, ts.category, ts.position, ts.workflow_id, ts.created_at, ts.updated_at,
    wf.name AS workflow_name, wf.description AS workflow_description,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug,
    (SELECT COUNT(*)::bigint FROM "tasks" t WHERE t.status_id = ts.id) AS task_count
FROM "task_statuses" ts
JOIN "workflows" wf ON ts.workflow_id = wf.id
JOIN "organizations" o ON wf.organization_id = o.id
WHERE ts.workflow_id = $1
ORDER BY ts.position ASC;

-- name: ListPublicProjectTaskCountsByType :many
SELECT type, COUNT(*)::bigint AS count FROM "tasks" WHERE project_id = $1 GROUP BY type;

-- name: ListPublicProjectTaskCountsByPriority :many
SELECT priority, COUNT(*)::bigint AS count FROM "tasks" WHERE project_id = $1 GROUP BY priority;

-- name: ListPublicProjectTaskCountsByStatus :many
SELECT ts.id AS status_id, ts.name AS status_name, ts.color AS status_color, ts.category AS status_category, COUNT(t.id)::bigint AS count
FROM "task_statuses" ts
LEFT JOIN "tasks" t ON t.status_id = ts.id AND t.project_id = $1
WHERE ts.workflow_id = $2
GROUP BY ts.id, ts.name, ts.color, ts.category, ts.position
ORDER BY ts.position ASC;

-- name: ListPublicProjectSprintVelocity :many
SELECT sp.id, sp.name, sp.status, sp.start_date, sp.end_date,
    COUNT(t.id)::bigint AS total_tasks,
    COALESCE(SUM(CASE WHEN ts.category = 'DONE' THEN 1 ELSE 0 END), 0)::bigint AS completed_tasks,
    COALESCE(SUM(COALESCE(t.story_points, 0)), 0)::bigint AS total_points,
    COALESCE(SUM(CASE WHEN ts.category = 'DONE' THEN COALESCE(t.story_points, 0) ELSE 0 END), 0)::bigint AS completed_points
FROM "sprints" sp
LEFT JOIN "tasks" t ON t.sprint_id = sp.id
LEFT JOIN "task_statuses" ts ON t.status_id = ts.id
WHERE sp.project_id = $1
GROUP BY sp.id, sp.name, sp.status, sp.start_date, sp.end_date
ORDER BY sp.created_at DESC;

-- name: ListPublicCalendarTasks :many
SELECT
    t.id, t.title, t.due_date,
    ts.id AS status_id, ts.name AS status_name, ts.color AS status_color, ts.category AS status_category,
    sp.id AS sprint_id, sp.name AS sprint_name, sp.start_date AS sprint_start_date, sp.end_date AS sprint_end_date
FROM "tasks" t
LEFT JOIN "task_statuses" ts ON t.status_id = ts.id
LEFT JOIN "sprints" sp ON t.sprint_id = sp.id
WHERE t.project_id = sqlc.arg(project_id)::uuid
  AND (
      (sqlc.narg(start_date)::timestamp IS NULL AND sqlc.narg(end_date)::timestamp IS NULL AND (
          t.due_date IS NOT NULL OR sp.start_date IS NOT NULL OR sp.end_date IS NOT NULL
      ))
      OR (sqlc.narg(start_date)::timestamp IS NOT NULL AND sqlc.narg(end_date)::timestamp IS NOT NULL AND (
          (t.due_date >= sqlc.narg(start_date)::timestamp AND t.due_date <= sqlc.narg(end_date)::timestamp)
          OR (sp.start_date >= sqlc.narg(start_date)::timestamp AND sp.start_date <= sqlc.narg(end_date)::timestamp)
          OR (sp.end_date >= sqlc.narg(start_date)::timestamp AND sp.end_date <= sqlc.narg(end_date)::timestamp)
      ))
  )
ORDER BY t.due_date ASC;

-- name: ListPublicProjectSprints :many
SELECT
    sp.id, sp.name, sp.goal, sp.status, sp.start_date, sp.end_date, sp.created_at,
    COUNT(t.id)::bigint AS total_tasks,
    COALESCE(SUM(CASE WHEN ts.category = 'DONE' THEN 1 ELSE 0 END), 0)::bigint AS completed_tasks
FROM "sprints" sp
LEFT JOIN "tasks" t ON t.sprint_id = sp.id
LEFT JOIN "task_statuses" ts ON t.status_id = ts.id
WHERE sp.project_id = $1
GROUP BY sp.id, sp.name, sp.goal, sp.status, sp.start_date, sp.end_date, sp.created_at
ORDER BY sp.created_at DESC;

-- name: GetPublicSprintByID :one
SELECT
    sp.id, sp.name, sp.goal, sp.status, sp.start_date, sp.end_date, sp.created_at,
    COUNT(t.id)::bigint AS total_tasks,
    COALESCE(SUM(CASE WHEN ts.category = 'DONE' THEN 1 ELSE 0 END), 0)::bigint AS completed_tasks
FROM "sprints" sp
JOIN "projects" p ON sp.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
LEFT JOIN "tasks" t ON t.sprint_id = sp.id
LEFT JOIN "task_statuses" ts ON t.status_id = ts.id
WHERE sp.id = $1 AND w.slug = $2 AND p.slug = $3 AND p.visibility = 'PUBLIC'
GROUP BY sp.id, sp.name, sp.goal, sp.status, sp.start_date, sp.end_date, sp.created_at
LIMIT 1;

-- name: ListPublicSprintTasks :many
SELECT
    t.id, t.title, t.description, t.type, t.priority, t.slug, t.due_date, t.created_at,
    ts.id AS status_id, ts.name AS status_name, ts.color AS status_color, ts.category AS status_category
FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
LEFT JOIN "task_statuses" ts ON t.status_id = ts.id
WHERE t.sprint_id = $1 AND w.slug = $2 AND p.slug = $3 AND p.visibility = 'PUBLIC'
ORDER BY t.created_at DESC;

-- name: CountPublicProjectTasks :one
SELECT COUNT(*)::bigint
FROM "tasks" t
LEFT JOIN "task_statuses" ts ON t.status_id = ts.id
WHERE t.project_id = sqlc.arg(project_id)::uuid
  AND (
      (sqlc.narg(parent_mode)::text = 'all')
      OR (sqlc.narg(parent_task_id)::uuid IS NOT NULL AND t.parent_task_id = sqlc.narg(parent_task_id)::uuid)
      OR (sqlc.narg(parent_mode)::text IS NULL AND sqlc.narg(parent_task_id)::uuid IS NULL AND t.parent_task_id IS NULL)
  )
  AND (sqlc.narg(status_name)::text IS NULL OR ts.name = sqlc.narg(status_name)::text)
  AND (sqlc.narg(priority)::"TaskPriority" IS NULL OR t.priority = sqlc.narg(priority)::"TaskPriority")
  AND (sqlc.narg(task_type)::"TaskType" IS NULL OR t.type = sqlc.narg(task_type)::"TaskType");

-- name: ListPublicProjectTasks :many
SELECT
    t.id, t.title, t.description, t.type, t.priority, t.slug, t.due_date, t.created_at, t.parent_task_id,
    ts.id AS status_id, ts.name AS status_name, ts.color AS status_color, ts.category AS status_category
FROM "tasks" t
LEFT JOIN "task_statuses" ts ON t.status_id = ts.id
WHERE t.project_id = sqlc.arg(project_id)::uuid
  AND (
      (sqlc.narg(parent_mode)::text = 'all')
      OR (sqlc.narg(parent_task_id)::uuid IS NOT NULL AND t.parent_task_id = sqlc.narg(parent_task_id)::uuid)
      OR (sqlc.narg(parent_mode)::text IS NULL AND sqlc.narg(parent_task_id)::uuid IS NULL AND t.parent_task_id IS NULL)
  )
  AND (sqlc.narg(status_name)::text IS NULL OR ts.name = sqlc.narg(status_name)::text)
  AND (sqlc.narg(priority)::"TaskPriority" IS NULL OR t.priority = sqlc.narg(priority)::"TaskPriority")
  AND (sqlc.narg(task_type)::"TaskType" IS NULL OR t.type = sqlc.narg(task_type)::"TaskType")
ORDER BY t.created_at DESC
LIMIT sqlc.arg(limit_val)::integer OFFSET sqlc.arg(offset_val)::integer;

-- name: ListPublicTaskLabelsBatch :many
SELECT tl.task_id, l.id, l.name, l.color
FROM "task_labels" tl
JOIN "labels" l ON tl.label_id = l.id
WHERE tl.task_id = ANY($1::uuid[]);

-- name: GetPublicTaskByIDOrSlug :one
SELECT
    t.id, t.title, t.description, t.type, t.priority, t.slug, t.due_date, t.created_at, t.updated_at,
    t.project_id, t.status_id, t.sprint_id, t.parent_task_id,
    ts.id AS status_row_id, ts.name AS status_name, ts.color AS status_color, ts.category AS status_category,
    sp.id AS sprint_row_id, sp.name AS sprint_name, sp.status AS sprint_status, sp.start_date AS sprint_start_date, sp.end_date AS sprint_end_date,
    pt.title AS parent_title, pt.slug AS parent_slug, pt.type AS parent_type,
    p.name AS project_name, p.slug AS project_slug, p.visibility AS project_visibility,
    w.id AS workspace_id, w.name AS workspace_name, w.slug AS workspace_slug,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug
FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
JOIN "organizations" o ON w.organization_id = o.id
LEFT JOIN "task_statuses" ts ON t.status_id = ts.id
LEFT JOIN "sprints" sp ON t.sprint_id = sp.id
LEFT JOIN "tasks" pt ON t.parent_task_id = pt.id
WHERE ((sqlc.arg(is_uuid)::boolean AND t.id = sqlc.narg(task_uuid)::uuid) OR (NOT sqlc.arg(is_uuid)::boolean AND t.slug = sqlc.arg(task_slug)::text))
  AND p.visibility = 'PUBLIC'
LIMIT 1;

-- name: ListPublicTaskActivities :many
SELECT al.id, al.type, al.description, al.created_at, u.id AS user_id, u.first_name, u.last_name, u.avatar
FROM "activity_logs" al
JOIN "tasks" t ON al.entity_id = t.id::text
JOIN "projects" p ON t.project_id = p.id
LEFT JOIN "users" u ON al.user_id = u.id
WHERE t.id = $1 AND p.visibility = 'PUBLIC'
ORDER BY al.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListPublicTaskComments :many
SELECT tc.id, tc.content, tc.parent_comment_id, tc.created_at, tc.updated_at, u.id AS author_id, u.first_name, u.last_name, u.avatar
FROM "task_comments" tc
JOIN "tasks" t ON tc.task_id = t.id
JOIN "projects" p ON t.project_id = p.id
JOIN "users" u ON tc.author_id = u.id
WHERE tc.task_id = $1 AND p.visibility = 'PUBLIC'
ORDER BY tc.created_at ASC;

-- name: ListPublicTaskAttachments :many
SELECT ta.id, ta.file_name, ta.file_size, ta.mime_type, ta.url, ta.created_at
FROM "task_attachments" ta
JOIN "tasks" t ON ta.task_id = t.id
JOIN "projects" p ON t.project_id = p.id
WHERE ta.task_id = $1 AND p.visibility = 'PUBLIC'
ORDER BY ta.created_at DESC;

-- name: ListPublicTaskAssigneesBatch :many
SELECT ta.task_id, u.id, u.first_name, u.last_name, u.avatar
FROM "task_assignees" ta
JOIN "users" u ON ta.user_id = u.id
WHERE ta.task_id = ANY($1::uuid[]);

-- name: ListPublicSubtasksBatch :many
SELECT
    t.id, t.title, t.description, t.type, t.priority, t.slug, t.due_date, t.created_at, t.parent_task_id,
    ts.id AS status_id, ts.name AS status_name, ts.color AS status_color, ts.category AS status_category
FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
LEFT JOIN "task_statuses" ts ON t.status_id = ts.id
WHERE t.parent_task_id = ANY($1::uuid[]) AND p.visibility = 'PUBLIC'
ORDER BY t.created_at ASC;

-- name: CountPublicTaskActivities :one
SELECT COUNT(*)::bigint
FROM "activity_logs" al
JOIN "tasks" t ON al.entity_id = t.id::text
JOIN "projects" p ON t.project_id = p.id
WHERE t.id = $1 AND p.visibility = 'PUBLIC';

-- name: GetPublicProjectKPIs :one
SELECT
    (SELECT COUNT(*)::bigint FROM "tasks" t WHERE t.project_id = $1) AS total_tasks,
    (SELECT COUNT(*)::bigint FROM "tasks" t WHERE t.project_id = $1 AND t.completed_at IS NOT NULL) AS completed_tasks,
    (SELECT COUNT(*)::bigint FROM "sprints" s WHERE s.project_id = $1 AND s.archive = false AND s.status = 'ACTIVE') AS active_sprints,
    (SELECT COUNT(*)::bigint FROM "tasks" t WHERE t.project_id = $1 AND t.type = 'BUG') AS total_bugs,
    (SELECT COUNT(*)::bigint FROM "tasks" t WHERE t.project_id = $1 AND t.type = 'BUG' AND t.completed_at IS NOT NULL) AS resolved_bugs;
