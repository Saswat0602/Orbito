-- name: GetUserAccessibleOrganizationIDsForSearch :many
SELECT organization_id FROM "organization_members" WHERE user_id = $1;

-- name: GetUserAccessibleWorkspaceIDsForSearch :many
SELECT workspace_id FROM "workspace_members" WHERE user_id = $1;

-- name: GetUserAccessibleProjectIDsForSearch :many
SELECT project_id FROM "project_members" WHERE user_id = $1;

-- name: SearchTasksGlobalQuery :many
SELECT
    t.id, t.title, t.description, t.slug, t.type, t.priority, t.due_date, t.story_points,
    t.created_at, t.updated_at,
    p.id AS project_id, p.slug AS project_slug, p.workspace_id,
    w.organization_id,
    ts.name AS status_name, ts.color AS status_color, ts.category AS status_category
FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
LEFT JOIN "task_statuses" ts ON t.status_id = ts.id
WHERE (
    t.title ILIKE '%' || sqlc.arg(query_text)::text || '%' ESCAPE '\'
    OR COALESCE(t.description, '') ILIKE '%' || sqlc.arg(query_text)::text || '%' ESCAPE '\'
    OR t.slug ILIKE '%' || sqlc.arg(query_text)::text || '%' ESCAPE '\'
)
AND (
    (cardinality(sqlc.arg(org_ids)::uuid[]) > 0 AND w.organization_id = ANY(sqlc.arg(org_ids)::uuid[]))
    OR (cardinality(sqlc.arg(ws_ids)::uuid[]) > 0 AND p.workspace_id = ANY(sqlc.arg(ws_ids)::uuid[]))
    OR (cardinality(sqlc.arg(proj_ids)::uuid[]) > 0 AND t.project_id = ANY(sqlc.arg(proj_ids)::uuid[]))
)
AND (sqlc.narg(filter_org_id)::uuid IS NULL OR w.organization_id = sqlc.narg(filter_org_id)::uuid)
AND (sqlc.narg(filter_ws_id)::uuid IS NULL OR p.workspace_id = sqlc.narg(filter_ws_id)::uuid)
AND (sqlc.narg(filter_proj_id)::uuid IS NULL OR t.project_id = sqlc.narg(filter_proj_id)::uuid)
ORDER BY t.updated_at DESC
LIMIT sqlc.arg(limit_val)::integer
OFFSET sqlc.arg(offset_val)::integer;

-- name: SearchProjectsGlobalQuery :many
SELECT
    p.id, p.name, p.description, p.slug, p.status, p.priority, p.avatar,
    p.workspace_id, w.organization_id, p.created_at, p.updated_at
FROM "projects" p
JOIN "workspaces" w ON p.workspace_id = w.id
WHERE (
    p.name ILIKE '%' || sqlc.arg(query_text)::text || '%' ESCAPE '\'
    OR COALESCE(p.description, '') ILIKE '%' || sqlc.arg(query_text)::text || '%' ESCAPE '\'
    OR p.slug ILIKE '%' || sqlc.arg(query_text)::text || '%' ESCAPE '\'
)
AND (
    (cardinality(sqlc.arg(org_ids)::uuid[]) > 0 AND w.organization_id = ANY(sqlc.arg(org_ids)::uuid[]))
    OR (cardinality(sqlc.arg(ws_ids)::uuid[]) > 0 AND p.workspace_id = ANY(sqlc.arg(ws_ids)::uuid[]))
    OR (cardinality(sqlc.arg(proj_ids)::uuid[]) > 0 AND p.id = ANY(sqlc.arg(proj_ids)::uuid[]))
)
AND (sqlc.narg(filter_org_id)::uuid IS NULL OR w.organization_id = sqlc.narg(filter_org_id)::uuid)
AND (sqlc.narg(filter_ws_id)::uuid IS NULL OR p.workspace_id = sqlc.narg(filter_ws_id)::uuid)
AND (sqlc.narg(filter_proj_id)::uuid IS NULL OR p.id = sqlc.narg(filter_proj_id)::uuid)
ORDER BY p.updated_at DESC
LIMIT sqlc.arg(limit_val)::integer
OFFSET sqlc.arg(offset_val)::integer;

-- name: SearchUsersGlobalQuery :many
SELECT DISTINCT
    u.id, u.first_name, u.last_name, u.email, u.username, u.avatar, u.role, u.created_at, u.updated_at
FROM "users" u
LEFT JOIN "organization_members" om ON om.user_id = u.id
LEFT JOIN "workspace_members" wm ON wm.user_id = u.id
LEFT JOIN "project_members" pm ON pm.user_id = u.id
WHERE u.deleted_at IS NULL
AND (
    u.first_name ILIKE '%' || sqlc.arg(query_text)::text || '%' ESCAPE '\'
    OR u.last_name ILIKE '%' || sqlc.arg(query_text)::text || '%' ESCAPE '\'
    OR u.email ILIKE '%' || sqlc.arg(query_text)::text || '%' ESCAPE '\'
    OR COALESCE(u.username, '') ILIKE '%' || sqlc.arg(query_text)::text || '%' ESCAPE '\'
)
AND (
    (cardinality(sqlc.arg(org_ids)::uuid[]) > 0 AND om.organization_id = ANY(sqlc.arg(org_ids)::uuid[]))
    OR (cardinality(sqlc.arg(ws_ids)::uuid[]) > 0 AND wm.workspace_id = ANY(sqlc.arg(ws_ids)::uuid[]))
    OR (cardinality(sqlc.arg(proj_ids)::uuid[]) > 0 AND pm.project_id = ANY(sqlc.arg(proj_ids)::uuid[]))
)
LIMIT sqlc.arg(limit_val)::integer
OFFSET sqlc.arg(offset_val)::integer;

-- name: SearchCommentsGlobalQuery :many
SELECT
    tc.id, tc.content, tc.task_id, tc.created_at, tc.updated_at,
    t.title AS task_title, t.slug AS task_slug,
    p.id AS project_id, p.slug AS project_slug, p.workspace_id,
    w.organization_id,
    u.id AS author_id, u.first_name AS author_first_name, u.last_name AS author_last_name
FROM "task_comments" tc
JOIN "tasks" t ON tc.task_id = t.id
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
JOIN "users" u ON tc.author_id = u.id
WHERE tc.content ILIKE '%' || sqlc.arg(query_text)::text || '%' ESCAPE '\'
AND (
    (cardinality(sqlc.arg(org_ids)::uuid[]) > 0 AND w.organization_id = ANY(sqlc.arg(org_ids)::uuid[]))
    OR (cardinality(sqlc.arg(ws_ids)::uuid[]) > 0 AND p.workspace_id = ANY(sqlc.arg(ws_ids)::uuid[]))
    OR (cardinality(sqlc.arg(proj_ids)::uuid[]) > 0 AND t.project_id = ANY(sqlc.arg(proj_ids)::uuid[]))
)
AND (sqlc.narg(filter_org_id)::uuid IS NULL OR w.organization_id = sqlc.narg(filter_org_id)::uuid)
AND (sqlc.narg(filter_ws_id)::uuid IS NULL OR p.workspace_id = sqlc.narg(filter_ws_id)::uuid)
AND (sqlc.narg(filter_proj_id)::uuid IS NULL OR t.project_id = sqlc.narg(filter_proj_id)::uuid)
ORDER BY tc.updated_at DESC
LIMIT sqlc.arg(limit_val)::integer
OFFSET sqlc.arg(offset_val)::integer;

-- name: SearchAttachmentsGlobalQuery :many
SELECT
    ta.id, ta.file_name, ta.file_size, ta.mime_type, ta.task_id, ta.created_at, ta.updated_at,
    t.title AS task_title, t.slug AS task_slug,
    p.id AS project_id, p.slug AS project_slug, p.workspace_id,
    w.organization_id
FROM "task_attachments" ta
JOIN "tasks" t ON ta.task_id = t.id
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
WHERE ta.file_name ILIKE '%' || sqlc.arg(query_text)::text || '%' ESCAPE '\'
AND (
    (cardinality(sqlc.arg(org_ids)::uuid[]) > 0 AND w.organization_id = ANY(sqlc.arg(org_ids)::uuid[]))
    OR (cardinality(sqlc.arg(ws_ids)::uuid[]) > 0 AND p.workspace_id = ANY(sqlc.arg(ws_ids)::uuid[]))
    OR (cardinality(sqlc.arg(proj_ids)::uuid[]) > 0 AND t.project_id = ANY(sqlc.arg(proj_ids)::uuid[]))
)
AND (sqlc.narg(filter_org_id)::uuid IS NULL OR w.organization_id = sqlc.narg(filter_org_id)::uuid)
AND (sqlc.narg(filter_ws_id)::uuid IS NULL OR p.workspace_id = sqlc.narg(filter_ws_id)::uuid)
AND (sqlc.narg(filter_proj_id)::uuid IS NULL OR t.project_id = sqlc.narg(filter_proj_id)::uuid)
ORDER BY ta.updated_at DESC
LIMIT sqlc.arg(limit_val)::integer
OFFSET sqlc.arg(offset_val)::integer;

-- name: SearchSprintsGlobalQuery :many
SELECT
    s.id, s.name, s.goal, s.status, s.start_date, s.end_date, s.created_at, s.updated_at,
    p.id AS project_id, p.slug AS project_slug, p.workspace_id,
    w.organization_id
FROM "sprints" s
JOIN "projects" p ON s.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
WHERE (
    s.name ILIKE '%' || sqlc.arg(query_text)::text || '%' ESCAPE '\'
    OR COALESCE(s.goal, '') ILIKE '%' || sqlc.arg(query_text)::text || '%' ESCAPE '\'
)
AND (
    (cardinality(sqlc.arg(org_ids)::uuid[]) > 0 AND w.organization_id = ANY(sqlc.arg(org_ids)::uuid[]))
    OR (cardinality(sqlc.arg(ws_ids)::uuid[]) > 0 AND p.workspace_id = ANY(sqlc.arg(ws_ids)::uuid[]))
    OR (cardinality(sqlc.arg(proj_ids)::uuid[]) > 0 AND s.project_id = ANY(sqlc.arg(proj_ids)::uuid[]))
)
AND (sqlc.narg(filter_org_id)::uuid IS NULL OR w.organization_id = sqlc.narg(filter_org_id)::uuid)
AND (sqlc.narg(filter_ws_id)::uuid IS NULL OR p.workspace_id = sqlc.narg(filter_ws_id)::uuid)
AND (sqlc.narg(filter_proj_id)::uuid IS NULL OR s.project_id = sqlc.narg(filter_proj_id)::uuid)
ORDER BY s.updated_at DESC
LIMIT sqlc.arg(limit_val)::integer
OFFSET sqlc.arg(offset_val)::integer;

-- name: CountTasksAdvancedQuery :one
SELECT COUNT(*)::bigint
FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
WHERE (
    (cardinality(sqlc.arg(org_ids)::uuid[]) > 0 AND w.organization_id = ANY(sqlc.arg(org_ids)::uuid[]))
    OR (cardinality(sqlc.arg(ws_ids)::uuid[]) > 0 AND p.workspace_id = ANY(sqlc.arg(ws_ids)::uuid[]))
    OR (cardinality(sqlc.arg(proj_ids)::uuid[]) > 0 AND t.project_id = ANY(sqlc.arg(proj_ids)::uuid[]))
)
AND (sqlc.narg(filter_org_id)::uuid IS NULL OR w.organization_id = sqlc.narg(filter_org_id)::uuid)
AND (sqlc.narg(filter_ws_id)::uuid IS NULL OR p.workspace_id = sqlc.narg(filter_ws_id)::uuid)
AND (sqlc.narg(filter_proj_id)::uuid IS NULL OR t.project_id = sqlc.narg(filter_proj_id)::uuid)
AND (sqlc.narg(query_text)::text IS NULL OR (
    t.title ILIKE '%' || sqlc.narg(query_text)::text || '%' ESCAPE '\'
    OR COALESCE(t.description, '') ILIKE '%' || sqlc.narg(query_text)::text || '%' ESCAPE '\'
    OR t.slug ILIKE '%' || sqlc.narg(query_text)::text || '%' ESCAPE '\'
))
AND (cardinality(sqlc.arg(task_types)::"TaskType"[]) = 0 OR t.type = ANY(sqlc.arg(task_types)::"TaskType"[]))
AND (cardinality(sqlc.arg(priorities)::"TaskPriority"[]) = 0 OR t.priority = ANY(sqlc.arg(priorities)::"TaskPriority"[]))
AND (cardinality(sqlc.arg(status_ids)::uuid[]) = 0 OR t.status_id = ANY(sqlc.arg(status_ids)::uuid[]))
AND (cardinality(sqlc.arg(sprint_ids)::uuid[]) = 0 OR t.sprint_id = ANY(sqlc.arg(sprint_ids)::uuid[]))
AND (cardinality(sqlc.arg(assignee_ids)::uuid[]) = 0 OR EXISTS (
    SELECT 1 FROM "task_assignees" ta WHERE ta.task_id = t.id AND ta.user_id = ANY(sqlc.arg(assignee_ids)::uuid[])
))
AND (cardinality(sqlc.arg(reporter_ids)::uuid[]) = 0 OR EXISTS (
    SELECT 1 FROM "task_reporters" tr WHERE tr.task_id = t.id AND tr.user_id = ANY(sqlc.arg(reporter_ids)::uuid[])
))
AND (cardinality(sqlc.arg(label_ids)::uuid[]) = 0 OR EXISTS (
    SELECT 1 FROM "task_labels" tl WHERE tl.task_id = t.id AND tl.label_id = ANY(sqlc.arg(label_ids)::uuid[])
))
AND (sqlc.narg(due_from)::timestamp IS NULL OR t.due_date >= sqlc.narg(due_from)::timestamp)
AND (sqlc.narg(due_to)::timestamp IS NULL OR t.due_date <= sqlc.narg(due_to)::timestamp)
AND (sqlc.narg(created_from)::timestamp IS NULL OR t.created_at >= sqlc.narg(created_from)::timestamp)
AND (sqlc.narg(created_to)::timestamp IS NULL OR t.created_at <= sqlc.narg(created_to)::timestamp);

-- name: SearchTasksAdvancedQuery :many
SELECT
    t.id, t.title, t.description, t.slug, t.type, t.priority, t.due_date, t.story_points,
    t.created_at, t.updated_at,
    p.id AS project_id, p.slug AS project_slug, p.workspace_id,
    w.organization_id,
    ts.id AS status_id, ts.name AS status_name, ts.color AS status_color,
    sp.id AS sprint_id, sp.name AS sprint_name, sp.status AS sprint_status
FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
LEFT JOIN "task_statuses" ts ON t.status_id = ts.id
LEFT JOIN "sprints" sp ON t.sprint_id = sp.id
WHERE (
    (cardinality(sqlc.arg(org_ids)::uuid[]) > 0 AND w.organization_id = ANY(sqlc.arg(org_ids)::uuid[]))
    OR (cardinality(sqlc.arg(ws_ids)::uuid[]) > 0 AND p.workspace_id = ANY(sqlc.arg(ws_ids)::uuid[]))
    OR (cardinality(sqlc.arg(proj_ids)::uuid[]) > 0 AND t.project_id = ANY(sqlc.arg(proj_ids)::uuid[]))
)
AND (sqlc.narg(filter_org_id)::uuid IS NULL OR w.organization_id = sqlc.narg(filter_org_id)::uuid)
AND (sqlc.narg(filter_ws_id)::uuid IS NULL OR p.workspace_id = sqlc.narg(filter_ws_id)::uuid)
AND (sqlc.narg(filter_proj_id)::uuid IS NULL OR t.project_id = sqlc.narg(filter_proj_id)::uuid)
AND (sqlc.narg(query_text)::text IS NULL OR (
    t.title ILIKE '%' || sqlc.narg(query_text)::text || '%' ESCAPE '\'
    OR COALESCE(t.description, '') ILIKE '%' || sqlc.narg(query_text)::text || '%' ESCAPE '\'
    OR t.slug ILIKE '%' || sqlc.narg(query_text)::text || '%' ESCAPE '\'
))
AND (cardinality(sqlc.arg(task_types)::"TaskType"[]) = 0 OR t.type = ANY(sqlc.arg(task_types)::"TaskType"[]))
AND (cardinality(sqlc.arg(priorities)::"TaskPriority"[]) = 0 OR t.priority = ANY(sqlc.arg(priorities)::"TaskPriority"[]))
AND (cardinality(sqlc.arg(status_ids)::uuid[]) = 0 OR t.status_id = ANY(sqlc.arg(status_ids)::uuid[]))
AND (cardinality(sqlc.arg(sprint_ids)::uuid[]) = 0 OR t.sprint_id = ANY(sqlc.arg(sprint_ids)::uuid[]))
AND (cardinality(sqlc.arg(assignee_ids)::uuid[]) = 0 OR EXISTS (
    SELECT 1 FROM "task_assignees" ta WHERE ta.task_id = t.id AND ta.user_id = ANY(sqlc.arg(assignee_ids)::uuid[])
))
AND (cardinality(sqlc.arg(reporter_ids)::uuid[]) = 0 OR EXISTS (
    SELECT 1 FROM "task_reporters" tr WHERE tr.task_id = t.id AND tr.user_id = ANY(sqlc.arg(reporter_ids)::uuid[])
))
AND (cardinality(sqlc.arg(label_ids)::uuid[]) = 0 OR EXISTS (
    SELECT 1 FROM "task_labels" tl WHERE tl.task_id = t.id AND tl.label_id = ANY(sqlc.arg(label_ids)::uuid[])
))
AND (sqlc.narg(due_from)::timestamp IS NULL OR t.due_date >= sqlc.narg(due_from)::timestamp)
AND (sqlc.narg(due_to)::timestamp IS NULL OR t.due_date <= sqlc.narg(due_to)::timestamp)
AND (sqlc.narg(created_from)::timestamp IS NULL OR t.created_at >= sqlc.narg(created_from)::timestamp)
AND (sqlc.narg(created_to)::timestamp IS NULL OR t.created_at <= sqlc.narg(created_to)::timestamp)
ORDER BY t.updated_at DESC
LIMIT sqlc.arg(limit_val)::integer
OFFSET sqlc.arg(offset_val)::integer;

-- name: SearchTaskTitleSuggestions :many
SELECT t.title
FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
WHERE t.title ILIKE '%' || sqlc.arg(query_text)::text || '%' ESCAPE '\'
AND (
    (cardinality(sqlc.arg(org_ids)::uuid[]) > 0 AND w.organization_id = ANY(sqlc.arg(org_ids)::uuid[]))
    OR (cardinality(sqlc.arg(ws_ids)::uuid[]) > 0 AND p.workspace_id = ANY(sqlc.arg(ws_ids)::uuid[]))
    OR (cardinality(sqlc.arg(proj_ids)::uuid[]) > 0 AND t.project_id = ANY(sqlc.arg(proj_ids)::uuid[]))
)
LIMIT sqlc.arg(limit_val)::integer;

-- name: SearchProjectNameSuggestions :many
SELECT p.name
FROM "projects" p
JOIN "workspaces" w ON p.workspace_id = w.id
WHERE p.name ILIKE '%' || sqlc.arg(query_text)::text || '%' ESCAPE '\'
AND (
    (cardinality(sqlc.arg(org_ids)::uuid[]) > 0 AND w.organization_id = ANY(sqlc.arg(org_ids)::uuid[]))
    OR (cardinality(sqlc.arg(ws_ids)::uuid[]) > 0 AND p.workspace_id = ANY(sqlc.arg(ws_ids)::uuid[]))
    OR (cardinality(sqlc.arg(proj_ids)::uuid[]) > 0 AND p.id = ANY(sqlc.arg(proj_ids)::uuid[]))
)
LIMIT sqlc.arg(limit_val)::integer;
