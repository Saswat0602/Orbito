-- name: CreateProject :one
INSERT INTO projects (
    id, name, slug, task_prefix, description, avatar, color,
    status, priority, visibility, start_date, end_date, settings,
    workflow_id, workspace_id, created_by_id, updated_by_id,
    created_at, updated_at, archive
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11, $12, $13,
    $14, $15, $16, $17,
    CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, false
) RETURNING *;

-- name: GetProjectByIDRow :one
SELECT 
    p.id, p.name, p.slug, p.task_prefix, p.description, p.avatar, p.color,
    p.status, p.priority, p.visibility, p.start_date, p.end_date, p.settings,
    p.workflow_id, p.workspace_id, p.created_by_id, p.updated_by_id,
    p.created_at, p.updated_at, p.archive,
    w.name AS workspace_name, w.slug AS workspace_slug, w.organization_id,
    o.name AS organization_name, o.slug AS organization_slug, o.owner_id AS organization_owner_id,
    wf.name AS workflow_name, wf.is_default AS workflow_is_default,
    u.email AS creator_email, u.first_name AS creator_first_name, u.last_name AS creator_last_name
FROM projects p
JOIN workspaces w ON p.workspace_id = w.id
JOIN organizations o ON w.organization_id = o.id
LEFT JOIN workflows wf ON p.workflow_id = wf.id
LEFT JOIN users u ON p.created_by_id = u.id
WHERE p.id = $1;

-- name: GetProjectBySlugRow :one
SELECT 
    p.id, p.name, p.slug, p.task_prefix, p.description, p.avatar, p.color,
    p.status, p.priority, p.visibility, p.start_date, p.end_date, p.settings,
    p.workflow_id, p.workspace_id, p.created_by_id, p.updated_by_id,
    p.created_at, p.updated_at, p.archive,
    w.name AS workspace_name, w.slug AS workspace_slug, w.organization_id,
    o.name AS organization_name, o.slug AS organization_slug, o.owner_id AS organization_owner_id,
    wf.name AS workflow_name, wf.is_default AS workflow_is_default,
    u.email AS creator_email, u.first_name AS creator_first_name, u.last_name AS creator_last_name
FROM projects p
JOIN workspaces w ON p.workspace_id = w.id
JOIN organizations o ON w.organization_id = o.id
LEFT JOIN workflows wf ON p.workflow_id = wf.id
LEFT JOIN users u ON p.created_by_id = u.id
WHERE p.slug = $1;

-- name: GetProjectByKeyRow :one
SELECT 
    p.id, p.name, p.slug, p.task_prefix, p.description, p.avatar, p.color,
    p.status, p.priority, p.visibility, p.start_date, p.end_date, p.settings,
    p.workflow_id, p.workspace_id, p.created_by_id, p.updated_by_id,
    p.created_at, p.updated_at, p.archive,
    w.name AS workspace_name, w.slug AS workspace_slug, w.organization_id,
    o.name AS organization_name, o.slug AS organization_slug, o.owner_id AS organization_owner_id
FROM projects p
JOIN workspaces w ON p.workspace_id = w.id
JOIN organizations o ON w.organization_id = o.id
WHERE p.workspace_id = $1 AND (p.task_prefix = $2 OR p.slug = $2);

-- name: ListProjectsForUser :many
SELECT 
    p.id, p.name, p.slug, p.task_prefix, p.description, p.avatar, p.color,
    p.status, p.priority, p.visibility, p.start_date, p.end_date,
    p.workflow_id, p.workspace_id, p.created_at, p.updated_at, p.archive,
    w.name AS workspace_name, w.slug AS workspace_slug, w.organization_id,
    (SELECT count(*) FROM tasks t WHERE t.project_id = p.id AND t.is_archived = false) AS task_count,
    (SELECT count(*) FROM project_members pm WHERE pm.project_id = p.id) AS member_count
FROM projects p
JOIN workspaces w ON p.workspace_id = w.id
WHERE p.archive = false
  AND (sqlc.narg(workspace_id)::uuid IS NULL OR p.workspace_id = sqlc.narg(workspace_id))
  AND (sqlc.narg(status)::text IS NULL OR p.status::text = sqlc.narg(status))
  AND (sqlc.narg(priority)::text IS NULL OR p.priority::text = sqlc.narg(priority))
  AND (sqlc.narg(search)::text IS NULL OR p.name ILIKE '%' || sqlc.narg(search) || '%' OR p.slug ILIKE '%' || sqlc.narg(search) || '%')
  AND (
      sqlc.arg(is_admin)::boolean = true
      OR p.visibility = 'PUBLIC'
      OR (p.visibility = 'INTERNAL' AND EXISTS (SELECT 1 FROM workspace_members wm WHERE wm.workspace_id = p.workspace_id AND wm.user_id = sqlc.arg(user_id)))
      OR EXISTS (SELECT 1 FROM project_members pm WHERE pm.project_id = p.id AND pm.user_id = sqlc.arg(user_id))
  )
ORDER BY p.created_at DESC
LIMIT sqlc.arg(lim)::int OFFSET sqlc.arg(offs)::int;

-- name: CountProjectsForUser :one
SELECT count(*)
FROM projects p
JOIN workspaces w ON p.workspace_id = w.id
WHERE p.archive = false
  AND (sqlc.narg(workspace_id)::uuid IS NULL OR p.workspace_id = sqlc.narg(workspace_id))
  AND (sqlc.narg(status)::text IS NULL OR p.status::text = sqlc.narg(status))
  AND (sqlc.narg(priority)::text IS NULL OR p.priority::text = sqlc.narg(priority))
  AND (sqlc.narg(search)::text IS NULL OR p.name ILIKE '%' || sqlc.narg(search) || '%' OR p.slug ILIKE '%' || sqlc.narg(search) || '%')
  AND (
      sqlc.arg(is_admin)::boolean = true
      OR p.visibility = 'PUBLIC'
      OR (p.visibility = 'INTERNAL' AND EXISTS (SELECT 1 FROM workspace_members wm WHERE wm.workspace_id = p.workspace_id AND wm.user_id = sqlc.arg(user_id)))
      OR EXISTS (SELECT 1 FROM project_members pm WHERE pm.project_id = p.id AND pm.user_id = sqlc.arg(user_id))
  );

-- name: ListProjectsByOrganization :many
SELECT 
    p.id, p.name, p.slug, p.task_prefix, p.description, p.avatar, p.color,
    p.status, p.priority, p.visibility, p.start_date, p.end_date,
    p.workflow_id, p.workspace_id, p.created_at, p.updated_at, p.archive,
    w.name AS workspace_name, w.slug AS workspace_slug, w.organization_id,
    (SELECT count(*) FROM tasks t WHERE t.project_id = p.id AND t.is_archived = false) AS task_count,
    (SELECT count(*) FROM project_members pm WHERE pm.project_id = p.id) AS member_count
FROM projects p
JOIN workspaces w ON p.workspace_id = w.id
WHERE p.archive = false
  AND w.organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(workspace_id)::uuid IS NULL OR p.workspace_id = sqlc.narg(workspace_id))
  AND (sqlc.narg(status)::text IS NULL OR p.status::text = sqlc.narg(status))
  AND (sqlc.narg(priority)::text IS NULL OR p.priority::text = sqlc.narg(priority))
  AND (sqlc.narg(search)::text IS NULL OR p.name ILIKE '%' || sqlc.narg(search) || '%' OR p.slug ILIKE '%' || sqlc.narg(search) || '%')
  AND (
      sqlc.arg(is_admin)::boolean = true
      OR p.visibility = 'PUBLIC'
      OR (p.visibility = 'INTERNAL' AND EXISTS (SELECT 1 FROM workspace_members wm WHERE wm.workspace_id = p.workspace_id AND wm.user_id = sqlc.arg(user_id)))
      OR EXISTS (SELECT 1 FROM project_members pm WHERE pm.project_id = p.id AND pm.user_id = sqlc.arg(user_id))
  )
ORDER BY p.created_at DESC
LIMIT sqlc.arg(lim)::int OFFSET sqlc.arg(offs)::int;

-- name: CountProjectsByOrganization :one
SELECT count(*)
FROM projects p
JOIN workspaces w ON p.workspace_id = w.id
WHERE p.archive = false
  AND w.organization_id = sqlc.arg(organization_id)
  AND (sqlc.narg(workspace_id)::uuid IS NULL OR p.workspace_id = sqlc.narg(workspace_id))
  AND (sqlc.narg(status)::text IS NULL OR p.status::text = sqlc.narg(status))
  AND (sqlc.narg(priority)::text IS NULL OR p.priority::text = sqlc.narg(priority))
  AND (sqlc.narg(search)::text IS NULL OR p.name ILIKE '%' || sqlc.narg(search) || '%' OR p.slug ILIKE '%' || sqlc.narg(search) || '%')
  AND (
      sqlc.arg(is_admin)::boolean = true
      OR p.visibility = 'PUBLIC'
      OR (p.visibility = 'INTERNAL' AND EXISTS (SELECT 1 FROM workspace_members wm WHERE wm.workspace_id = p.workspace_id AND wm.user_id = sqlc.arg(user_id)))
      OR EXISTS (SELECT 1 FROM project_members pm WHERE pm.project_id = p.id AND pm.user_id = sqlc.arg(user_id))
  );

-- name: ListArchivedProjectsByWorkspace :many
SELECT p.*, w.name AS workspace_name
FROM projects p
JOIN workspaces w ON p.workspace_id = w.id
WHERE p.archive = true AND p.workspace_id = $1
ORDER BY p.updated_at DESC;

-- name: ListArchivedProjectsByOrganization :many
SELECT p.*, w.name AS workspace_name
FROM projects p
JOIN workspaces w ON p.workspace_id = w.id
WHERE p.archive = true AND w.organization_id = $1
ORDER BY p.updated_at DESC;

-- name: UpdateProject :one
UPDATE projects
SET 
    name = COALESCE(sqlc.narg(name), name),
    slug = COALESCE(sqlc.narg(slug), slug),
    task_prefix = COALESCE(sqlc.narg(task_prefix), task_prefix),
    description = COALESCE(sqlc.narg(description), description),
    avatar = COALESCE(sqlc.narg(avatar), avatar),
    color = COALESCE(sqlc.narg(color), color),
    status = COALESCE(sqlc.narg(status), status),
    priority = COALESCE(sqlc.narg(priority), priority),
    visibility = COALESCE(sqlc.narg(visibility), visibility),
    start_date = COALESCE(sqlc.narg(start_date), start_date),
    end_date = COALESCE(sqlc.narg(end_date), end_date),
    settings = COALESCE(sqlc.narg(settings), settings),
    workflow_id = COALESCE(sqlc.narg(workflow_id), workflow_id),
    updated_by_id = sqlc.narg(updated_by_id),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: ArchiveProject :exec
UPDATE projects
SET archive = true, updated_at = CURRENT_TIMESTAMP, updated_by_id = $2
WHERE id = $1;

-- name: UnarchiveProject :exec
UPDATE projects
SET archive = false, updated_at = CURRENT_TIMESTAMP, updated_by_id = $2
WHERE id = $1;

-- name: DeleteProject :exec
DELETE FROM projects WHERE id = $1;

-- name: CheckProjectSlugExists :one
SELECT count(*) FROM projects
WHERE slug = $1 AND (sqlc.narg(exclude_id)::uuid IS NULL OR id != sqlc.narg(exclude_id));

-- name: ListProjectSlugsStartingWith :many
SELECT slug FROM projects
WHERE slug LIKE $1 || '%';

-- name: GetDefaultWorkflowForOrg :one
SELECT * FROM workflows
WHERE organization_id = $1 AND is_default = true
LIMIT 1;

-- name: GetWorkflowByID :one
SELECT * FROM workflows
WHERE id = $1;

-- name: CreateInitialSprint :one
INSERT INTO sprints (
    id, name, goal, status, is_default, archive, start_date, end_date,
    project_id, created_by_id, updated_by_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, false, $6, $7,
    $8, $9, $10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: CountProjectTasks :one
SELECT count(*) FROM tasks
WHERE project_id = $1 AND is_archived = false;

-- name: CountProjectCompletedTasks :one
SELECT count(*) FROM tasks t
JOIN task_statuses ts ON t.status_id = ts.id
WHERE t.project_id = $1 AND t.is_archived = false AND ts.category = 'DONE';

-- name: CountProjectMembers :one
SELECT count(*) FROM project_members
WHERE project_id = $1;

-- name: ProjectTaskStatusStats :many
SELECT ts.id AS status_id, ts.name AS status_name, ts.category, count(t.id) AS task_count
FROM task_statuses ts
LEFT JOIN tasks t ON t.status_id = ts.id AND t.project_id = $1 AND t.is_archived = false
WHERE ts.workflow_id = (SELECT workflow_id FROM projects WHERE id = $1)
GROUP BY ts.id, ts.name, ts.category, ts.position
ORDER BY ts.position ASC;

-- name: ProjectTaskPriorityStats :many
SELECT t.priority, count(t.id) AS task_count
FROM tasks t
WHERE t.project_id = $1 AND t.is_archived = false
GROUP BY t.priority;
