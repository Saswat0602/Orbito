-- name: GetProjectAccessInfoForSprint :one
SELECT
    p.id, p.name, p.slug, p.visibility, p.workspace_id,
    w.organization_id,
    o.owner_id AS org_owner_id,
    o.archive AS org_archived,
    om.role AS org_role,
    wm.role AS ws_role,
    pm.role AS project_role
FROM "projects" p
JOIN "workspaces" w ON p.workspace_id = w.id
JOIN "organizations" o ON w.organization_id = o.id
LEFT JOIN "organization_members" om ON om.organization_id = o.id AND om.user_id = $2
LEFT JOIN "workspace_members" wm ON wm.workspace_id = w.id AND wm.user_id = $2
LEFT JOIN "project_members" pm ON pm.project_id = p.id AND pm.user_id = $2
WHERE p.id = $1;

-- name: GetProjectAccessInfoForSprintBySlug :one
SELECT
    p.id, p.name, p.slug, p.visibility, p.workspace_id,
    w.organization_id,
    o.owner_id AS org_owner_id,
    o.archive AS org_archived,
    om.role AS org_role,
    wm.role AS ws_role,
    pm.role AS project_role
FROM "projects" p
JOIN "workspaces" w ON p.workspace_id = w.id
JOIN "organizations" o ON w.organization_id = o.id
LEFT JOIN "organization_members" om ON om.organization_id = o.id AND om.user_id = $2
LEFT JOIN "workspace_members" wm ON wm.workspace_id = w.id AND wm.user_id = $2
LEFT JOIN "project_members" pm ON pm.project_id = p.id AND pm.user_id = $2
WHERE p.slug = $1;

-- name: CheckSprintSlugExists :one
SELECT EXISTS(
    SELECT 1 FROM "sprints"
    WHERE project_id = $1
      AND slug = $2
      AND (sqlc.narg(exclude_id)::uuid IS NULL OR id != sqlc.narg(exclude_id))
);

-- name: UpdateSprintSlugByID :exec
UPDATE "sprints"
SET slug = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: HasActiveSprintInProject :one
SELECT EXISTS(
    SELECT 1 FROM "sprints"
    WHERE project_id = $1
      AND status = 'ACTIVE'
      AND (sqlc.narg(exclude_id)::uuid IS NULL OR id != sqlc.narg(exclude_id))
);

-- name: CreateSprintRecord :one
INSERT INTO "sprints" (
    id, name, slug, goal, status, is_default, archive,
    start_date, end_date, project_id, created_by_id, updated_by_id,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, false, false,
    $6, $7, $8, $9, $10,
    CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING id;

-- name: GetSprintDetailsByID :one
SELECT
    s.id, s.name, s.slug, s.goal, s.status, s.is_default, s.archive,
    s.start_date, s.end_date, s.project_id, s.created_by_id, s.updated_by_id, s.created_at, s.updated_at,
    p.name AS project_name, p.slug AS project_slug,
    w.id AS workspace_id, w.name AS workspace_name, w.slug AS workspace_slug, w.organization_id,
    o.name AS org_name, o.slug AS org_slug,
    cu.id AS created_by_user_id, cu.email AS created_by_email, cu.first_name AS created_by_first_name, cu.last_name AS created_by_last_name,
    uu.id AS updated_by_user_id, uu.email AS updated_by_email, uu.first_name AS updated_by_first_name, uu.last_name AS updated_by_last_name,
    (SELECT COUNT(*) FROM "tasks" t WHERE t.sprint_id = s.id) AS task_count
FROM "sprints" s
JOIN "projects" p ON s.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
JOIN "organizations" o ON w.organization_id = o.id
LEFT JOIN "users" cu ON s.created_by_id = cu.id
LEFT JOIN "users" uu ON s.updated_by_id = uu.id
WHERE s.id = $1;

-- name: GetSprintByProjectAndSlug :one
SELECT
    s.id, s.name, s.slug, s.goal, s.status, s.is_default, s.archive,
    s.start_date, s.end_date, s.project_id, s.created_by_id, s.updated_by_id, s.created_at, s.updated_at,
    p.name AS project_name, p.slug AS project_slug,
    w.id AS workspace_id, w.name AS workspace_name, w.slug AS workspace_slug,
    (SELECT COUNT(*) FROM "tasks" t WHERE t.sprint_id = s.id) AS task_count
FROM "sprints" s
JOIN "projects" p ON s.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
WHERE p.slug = $1 AND s.slug = $2
LIMIT 1;

-- name: GetActiveSprintByProjectID :one
SELECT
    s.id, s.name, s.slug, s.goal, s.status, s.is_default, s.archive,
    s.start_date, s.end_date, s.project_id, s.created_by_id, s.updated_by_id, s.created_at, s.updated_at,
    p.name AS project_name, p.slug AS project_slug,
    (SELECT COUNT(*) FROM "tasks" t WHERE t.sprint_id = s.id) AS task_count
FROM "sprints" s
JOIN "projects" p ON s.project_id = p.id
WHERE s.project_id = $1 AND s.status = 'ACTIVE'
LIMIT 1;

-- name: ListSprintsFiltered :many
SELECT
    s.id, s.name, s.slug, s.goal, s.status, s.is_default, s.archive,
    s.start_date, s.end_date, s.project_id, s.created_by_id, s.updated_by_id, s.created_at, s.updated_at,
    p.name AS project_name, p.slug AS project_slug,
    w.id AS workspace_id, w.name AS workspace_name, w.slug AS workspace_slug,
    (SELECT COUNT(*) FROM "tasks" t WHERE t.sprint_id = s.id) AS task_count
FROM "sprints" s
JOIN "projects" p ON s.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
WHERE s.archive = false
  AND (sqlc.narg(project_id)::uuid IS NULL OR s.project_id = sqlc.narg(project_id))
  AND (sqlc.narg(project_slug)::text IS NULL OR p.slug = sqlc.narg(project_slug))
  AND (sqlc.narg(status)::text IS NULL OR s.status::text = sqlc.narg(status))
ORDER BY s.status ASC, s.created_at DESC;

-- name: ListSprintTasksBySprintID :many
SELECT
    t.id, t.title, t.description, t.type, t.priority, t.task_number, t.slug,
    t.start_date, t.due_date, t.completed_at, t.story_points, t.original_estimate, t.remaining_estimate,
    t.project_id, t.status_id, t.sprint_id, t.parent_task_id, t.created_at, t.updated_at, t.is_archived,
    ts.name AS status_name, ts.color AS status_color, ts.category AS status_category,
    (SELECT COUNT(*) FROM "tasks" ct WHERE ct.parent_task_id = t.id) AS child_task_count,
    (SELECT COUNT(*) FROM "task_comments" tc WHERE tc.task_id = t.id) AS comment_count
FROM "tasks" t
JOIN "task_statuses" ts ON t.status_id = ts.id
WHERE t.sprint_id = $1
ORDER BY t.priority DESC;

-- name: ListSprintTaskAssignees :many
SELECT
    ta.task_id,
    u.id, u.first_name, u.last_name, u.avatar
FROM "task_assignees" ta
JOIN "users" u ON ta.user_id = u.id
WHERE ta.task_id = ANY($1::uuid[]);

-- name: ListSprintTaskReporters :many
SELECT
    tr.task_id,
    u.id, u.first_name, u.last_name, u.avatar
FROM "task_reporters" tr
JOIN "users" u ON tr.user_id = u.id
WHERE tr.task_id = ANY($1::uuid[]);

-- name: UpdateSprintByID :one
UPDATE "sprints"
SET
    name = COALESCE(sqlc.narg(name), name),
    goal = COALESCE(sqlc.narg(goal), goal),
    status = COALESCE(sqlc.narg(status)::"SprintStatus", status),
    start_date = COALESCE(sqlc.narg(start_date), start_date),
    end_date = COALESCE(sqlc.narg(end_date), end_date),
    updated_by_id = sqlc.narg(updated_by_id),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id;

-- name: DeleteSprintByID :execrows
DELETE FROM "sprints"
WHERE id = $1;

-- name: ArchiveSprintByID :execrows
UPDATE "sprints"
SET archive = true,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: CreateSprintActivityLog :exec
INSERT INTO "activity_logs" (
    id, type, description, entity_type, entity_id,
    old_value, new_value, user_id, organization_id,
    created_by_id, updated_by_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9,
    $10, $11, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
);

