-- name: CreateActivityLogRecord :one
INSERT INTO "activity_logs" (
    id, type, description, entity_type, entity_id,
    old_value, new_value, user_id, organization_id,
    created_by_id, updated_by_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9,
    $10, $10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: ResolveActivityTaskByIDOrSlug :one
SELECT id, slug
FROM "tasks"
WHERE (sqlc.arg(is_uuid)::boolean = true AND id = sqlc.narg(task_uuid)::uuid)
   OR (sqlc.arg(is_uuid)::boolean = false AND slug = sqlc.arg(task_slug)::text)
LIMIT 1;

-- name: ResolveActivityProjectByIDOrSlug :one
SELECT p.id, p.slug, w.slug AS workspace_slug
FROM "projects" p
LEFT JOIN "workspaces" w ON p.workspace_id = w.id
WHERE (sqlc.arg(is_uuid)::boolean = true AND p.id = sqlc.narg(proj_uuid)::uuid)
   OR (sqlc.arg(is_uuid)::boolean = false AND p.slug = sqlc.arg(proj_slug)::text)
LIMIT 1;

-- name: CheckWorkspaceExistsForActivity :one
SELECT id, organization_id
FROM "workspaces"
WHERE id = $1
LIMIT 1;

-- name: GetOrganizationIDForTask :one
SELECT w.organization_id
FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
WHERE (sqlc.arg(is_uuid)::boolean = true AND t.id = sqlc.narg(task_uuid)::uuid)
   OR (sqlc.arg(is_uuid)::boolean = false AND t.slug = sqlc.arg(task_slug)::text)
LIMIT 1;

-- name: GetOrganizationIDForProject :one
SELECT w.organization_id
FROM "projects" p
JOIN "workspaces" w ON p.workspace_id = w.id
WHERE p.id = $1
LIMIT 1;

-- name: GetOrganizationIDForWorkspace :one
SELECT organization_id
FROM "workspaces"
WHERE id = $1
LIMIT 1;

-- name: GetOrganizationIDForSprint :one
SELECT w.organization_id
FROM "sprints" s
JOIN "projects" p ON s.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
WHERE s.id = $1
LIMIT 1;

-- name: CountTaskActivities :one
SELECT COUNT(*)::bigint
FROM "activity_logs"
WHERE entity_id = $1;

-- name: ListTaskActivities :many
SELECT
    al.id, al.type, al.description, al.entity_type, al.entity_id,
    al.old_value, al.new_value, al.user_id, al.organization_id,
    al.created_by_id, al.updated_by_id, al.created_at, al.updated_at,
    u.id AS actor_id, u.first_name AS actor_first_name, u.last_name AS actor_last_name,
    u.email AS actor_email, u.avatar AS actor_avatar
FROM "activity_logs" al
JOIN "users" u ON al.user_id = u.id
WHERE al.entity_id = $1
ORDER BY al.created_at DESC
LIMIT $2
OFFSET $3;

-- name: CountWorkspaceActivityOptimized :one
SELECT COUNT(*)::bigint
FROM "activity_logs" al
WHERE
    (al.entity_type = 'Workspace' AND al.entity_id = sqlc.arg(workspace_id)::uuid)
    OR (al.entity_type = 'Project' AND al.entity_id IN (
        SELECT p.id FROM "projects" p WHERE p.workspace_id = sqlc.arg(workspace_id)::uuid
    ))
    OR (al.entity_type IN ('Task', 'Task Label', 'TaskAttachment', 'Task Attchment') AND al.entity_id IN (
        SELECT t.id FROM "tasks" t
        JOIN "projects" p ON t.project_id = p.id
        WHERE p.workspace_id = sqlc.arg(workspace_id)::uuid
    ))
    OR (al.entity_type IN ('Task Comment', 'TaskComment') AND al.entity_id IN (
        SELECT tc.id FROM "task_comments" tc
        JOIN "tasks" t ON tc.task_id = t.id
        JOIN "projects" p ON t.project_id = p.id
        WHERE p.workspace_id = sqlc.arg(workspace_id)::uuid
    ))
    OR (al.entity_type = 'Sprint' AND (
        al.entity_id IN (
            SELECT s.id FROM "sprints" s
            JOIN "projects" p ON s.project_id = p.id
            WHERE p.workspace_id = sqlc.arg(workspace_id)::uuid
        )
        OR (al.new_value->>'projectId' IN (
            SELECT p.id::text FROM "projects" p WHERE p.workspace_id = sqlc.arg(workspace_id)::uuid
        ))
        OR (al.old_value->>'projectId' IN (
            SELECT p.id::text FROM "projects" p WHERE p.workspace_id = sqlc.arg(workspace_id)::uuid
        ))
    ));

-- name: ListWorkspaceActivityOptimized :many
SELECT
    al.id,
    al.type,
    al.description,
    al.entity_type,
    al.entity_id,
    al.created_at,
    u.id AS user_id,
    u.first_name,
    u.last_name,
    u.email,
    u.avatar,
    COALESCE(t.slug, ct.slug)::text AS task_slug,
    COALESCE(tp.slug, ctp.slug)::text AS task_project_slug,
    COALESCE(tw.slug, ctw.slug)::text AS task_workspace_slug,
    pp.slug::text AS project_slug,
    pw.slug::text AS project_workspace_slug,
    ws.slug::text AS workspace_slug,
    sp.slug::text AS sprint_slug,
    spp.slug::text AS sprint_project_slug,
    spw.slug::text AS sprint_workspace_slug
FROM "activity_logs" al
JOIN "users" u ON al.user_id = u.id
LEFT JOIN "tasks" t ON (al.entity_type IN ('Task', 'Task Label', 'TaskAttachment', 'Task Attchment')) AND al.entity_id = t.id
LEFT JOIN "projects" tp ON t.project_id = tp.id
LEFT JOIN "workspaces" tw ON tp.workspace_id = tw.id
LEFT JOIN "task_comments" tc ON (al.entity_type IN ('Task Comment', 'TaskComment')) AND al.entity_id = tc.id
LEFT JOIN "tasks" ct ON tc.task_id = ct.id
LEFT JOIN "projects" ctp ON ct.project_id = ctp.id
LEFT JOIN "workspaces" ctw ON ctp.workspace_id = ctw.id
LEFT JOIN "projects" pp ON al.entity_type = 'Project' AND al.entity_id = pp.id
LEFT JOIN "workspaces" pw ON pp.workspace_id = pw.id
LEFT JOIN "workspaces" ws ON al.entity_type = 'Workspace' AND al.entity_id = ws.id
LEFT JOIN "sprints" sp ON al.entity_type = 'Sprint' AND al.entity_id = sp.id
LEFT JOIN "projects" spp ON sp.project_id = spp.id
LEFT JOIN "workspaces" spw ON spp.workspace_id = spw.id
WHERE
    (al.entity_type = 'Workspace' AND al.entity_id = sqlc.arg(workspace_id)::uuid)
    OR (al.entity_type = 'Project' AND al.entity_id IN (
        SELECT p.id FROM "projects" p WHERE p.workspace_id = sqlc.arg(workspace_id)::uuid
    ))
    OR (al.entity_type IN ('Task', 'Task Label', 'TaskAttachment', 'Task Attchment') AND al.entity_id IN (
        SELECT t2.id FROM "tasks" t2
        JOIN "projects" p2 ON t2.project_id = p2.id
        WHERE p2.workspace_id = sqlc.arg(workspace_id)::uuid
    ))
    OR (al.entity_type IN ('Task Comment', 'TaskComment') AND al.entity_id IN (
        SELECT tc2.id FROM "task_comments" tc2
        JOIN "tasks" t3 ON tc2.task_id = t3.id
        JOIN "projects" p3 ON t3.project_id = p3.id
        WHERE p3.workspace_id = sqlc.arg(workspace_id)::uuid
    ))
    OR (al.entity_type = 'Sprint' AND (
        al.entity_id IN (
            SELECT s2.id FROM "sprints" s2
            JOIN "projects" p4 ON s2.project_id = p4.id
            WHERE p4.workspace_id = sqlc.arg(workspace_id)::uuid
        )
        OR (al.new_value->>'projectId' IN (
            SELECT p5.id::text FROM "projects" p5 WHERE p5.workspace_id = sqlc.arg(workspace_id)::uuid
        ))
        OR (al.old_value->>'projectId' IN (
            SELECT p6.id::text FROM "projects" p6 WHERE p6.workspace_id = sqlc.arg(workspace_id)::uuid
        ))
    ))
ORDER BY al.created_at DESC
LIMIT sqlc.arg(limit_val)::integer
OFFSET sqlc.arg(offset_val)::integer;

-- name: CountProjectActivities :one
SELECT COUNT(*)::bigint
FROM "activity_logs" al
WHERE (al.entity_type = 'Project' AND al.entity_id = sqlc.arg(project_id)::uuid)
   OR (al.entity_type = 'Task' AND al.entity_id IN (
       SELECT t.id FROM "tasks" t WHERE t.project_id = sqlc.arg(project_id)::uuid
   ));

-- name: ListProjectActivities :many
SELECT
    al.id, al.type, al.description, al.entity_type, al.entity_id, al.created_at,
    u.id AS user_id, u.first_name, u.last_name, u.email, u.avatar
FROM "activity_logs" al
JOIN "users" u ON al.user_id = u.id
WHERE (al.entity_type = 'Project' AND al.entity_id = sqlc.arg(project_id)::uuid)
   OR (al.entity_type = 'Task' AND al.entity_id IN (
       SELECT t.id FROM "tasks" t WHERE t.project_id = sqlc.arg(project_id)::uuid
   ))
ORDER BY al.created_at DESC
LIMIT sqlc.arg(limit_val)::integer
OFFSET sqlc.arg(offset_val)::integer;

-- name: CountOrganizationActivityComprehensive :one
SELECT COUNT(*)::bigint
FROM "activity_logs" al
WHERE (
    al.organization_id = sqlc.arg(organization_id)::uuid
    OR (al.entity_type = 'Workspace' AND al.entity_id IN (
        SELECT w.id FROM "workspaces" w WHERE w.organization_id = sqlc.arg(organization_id)::uuid
    ))
    OR (al.entity_type = 'Project' AND al.entity_id IN (
        SELECT p.id FROM "projects" p
        JOIN "workspaces" w ON p.workspace_id = w.id
        WHERE w.organization_id = sqlc.arg(organization_id)::uuid
    ))
    OR (al.entity_type IN ('Task', 'Task Label', 'TaskAttachment', 'Task Attchment') AND al.entity_id IN (
        SELECT t.id FROM "tasks" t
        JOIN "projects" p ON t.project_id = p.id
        JOIN "workspaces" w ON p.workspace_id = w.id
        WHERE w.organization_id = sqlc.arg(organization_id)::uuid
    ))
    OR (al.entity_type IN ('Task Comment', 'TaskComment') AND al.entity_id IN (
        SELECT tc.id FROM "task_comments" tc
        JOIN "tasks" t ON tc.task_id = t.id
        JOIN "projects" p ON t.project_id = p.id
        JOIN "workspaces" w ON p.workspace_id = w.id
        WHERE w.organization_id = sqlc.arg(organization_id)::uuid
    ))
    OR (al.entity_type = 'Sprint' AND al.entity_id IN (
        SELECT s.id FROM "sprints" s
        JOIN "projects" p ON s.project_id = p.id
        JOIN "workspaces" w ON p.workspace_id = w.id
        WHERE w.organization_id = sqlc.arg(organization_id)::uuid
    ))
)
AND (sqlc.narg(entity_type)::text IS NULL OR al.entity_type = sqlc.narg(entity_type)::text)
AND (sqlc.narg(user_id)::uuid IS NULL OR al.user_id = sqlc.narg(user_id)::uuid)
AND (sqlc.narg(start_date)::timestamp IS NULL OR al.created_at >= sqlc.narg(start_date)::timestamp)
AND (sqlc.narg(end_date)::timestamp IS NULL OR al.created_at <= sqlc.narg(end_date)::timestamp);

-- name: ListOrganizationActivityComprehensive :many
SELECT
    al.id,
    al.type,
    al.description,
    al.entity_type,
    al.entity_id,
    al.new_value,
    al.created_at,
    u.id AS user_id,
    u.first_name,
    u.last_name,
    u.email,
    u.avatar,
    COALESCE(t.slug, ct.slug)::text AS task_slug,
    COALESCE(tp.slug, ctp.slug)::text AS task_project_slug,
    COALESCE(tw.slug, ctw.slug)::text AS task_workspace_slug,
    pp.slug::text AS project_slug,
    pw.slug::text AS project_workspace_slug,
    ws.slug::text AS workspace_slug,
    sp.slug::text AS sprint_slug,
    spp.slug::text AS sprint_project_slug,
    spw.slug::text AS sprint_workspace_slug
FROM "activity_logs" al
JOIN "users" u ON al.user_id = u.id
LEFT JOIN "tasks" t ON (al.entity_type IN ('Task', 'Task Label', 'TaskAttachment', 'Task Attchment')) AND al.entity_id = t.id
LEFT JOIN "projects" tp ON t.project_id = tp.id
LEFT JOIN "workspaces" tw ON tp.workspace_id = tw.id
LEFT JOIN "task_comments" tc ON (al.entity_type IN ('Task Comment', 'TaskComment')) AND al.entity_id = tc.id
LEFT JOIN "tasks" ct ON tc.task_id = ct.id
LEFT JOIN "projects" ctp ON ct.project_id = ctp.id
LEFT JOIN "workspaces" ctw ON ctp.workspace_id = ctw.id
LEFT JOIN "projects" pp ON al.entity_type = 'Project' AND al.entity_id = pp.id
LEFT JOIN "workspaces" pw ON pp.workspace_id = pw.id
LEFT JOIN "workspaces" ws ON al.entity_type = 'Workspace' AND al.entity_id = ws.id
LEFT JOIN "sprints" sp ON al.entity_type = 'Sprint' AND al.entity_id = sp.id
LEFT JOIN "projects" spp ON sp.project_id = spp.id
LEFT JOIN "workspaces" spw ON spp.workspace_id = spw.id
WHERE (
    al.organization_id = sqlc.arg(organization_id)::uuid
    OR (al.entity_type = 'Workspace' AND al.entity_id IN (
        SELECT w.id FROM "workspaces" w WHERE w.organization_id = sqlc.arg(organization_id)::uuid
    ))
    OR (al.entity_type = 'Project' AND al.entity_id IN (
        SELECT p.id FROM "projects" p
        JOIN "workspaces" w ON p.workspace_id = w.id
        WHERE w.organization_id = sqlc.arg(organization_id)::uuid
    ))
    OR (al.entity_type IN ('Task', 'Task Label', 'TaskAttachment', 'Task Attchment') AND al.entity_id IN (
        SELECT t2.id FROM "tasks" t2
        JOIN "projects" p2 ON t2.project_id = p2.id
        JOIN "workspaces" w2 ON p2.workspace_id = w2.id
        WHERE w2.organization_id = sqlc.arg(organization_id)::uuid
    ))
    OR (al.entity_type IN ('Task Comment', 'TaskComment') AND al.entity_id IN (
        SELECT tc2.id FROM "task_comments" tc2
        JOIN "tasks" t3 ON tc2.task_id = t3.id
        JOIN "projects" p3 ON t3.project_id = p3.id
        JOIN "workspaces" w3 ON p3.workspace_id = w3.id
        WHERE w3.organization_id = sqlc.arg(organization_id)::uuid
    ))
    OR (al.entity_type = 'Sprint' AND al.entity_id IN (
        SELECT s2.id FROM "sprints" s2
        JOIN "projects" p4 ON s2.project_id = p4.id
        JOIN "workspaces" w4 ON p4.workspace_id = w4.id
        WHERE w4.organization_id = sqlc.arg(organization_id)::uuid
    ))
)
AND (sqlc.narg(entity_type)::text IS NULL OR al.entity_type = sqlc.narg(entity_type)::text)
AND (sqlc.narg(user_id)::uuid IS NULL OR al.user_id = sqlc.narg(user_id)::uuid)
AND (sqlc.narg(start_date)::timestamp IS NULL OR al.created_at >= sqlc.narg(start_date)::timestamp)
AND (sqlc.narg(end_date)::timestamp IS NULL OR al.created_at <= sqlc.narg(end_date)::timestamp)
ORDER BY al.created_at DESC
LIMIT sqlc.arg(limit_val)::integer
OFFSET sqlc.arg(offset_val)::integer;

-- name: ListOrganizationActivitiesForStats :many
SELECT
    al.type, al.user_id, al.created_at,
    u.first_name, u.last_name
FROM "activity_logs" al
JOIN "users" u ON al.user_id = u.id
WHERE al.organization_id = $1
  AND al.created_at >= $2;

-- name: CountUserRecentActivities :one
SELECT COUNT(*)::bigint
FROM "activity_logs"
WHERE user_id = sqlc.arg(user_id)::uuid
  AND (sqlc.narg(organization_id)::uuid IS NULL OR organization_id = sqlc.narg(organization_id)::uuid);

-- name: ListUserRecentActivities :many
SELECT
    al.id, al.type, al.description, al.entity_type, al.entity_id,
    al.old_value, al.new_value, al.user_id, al.organization_id,
    al.created_by_id, al.updated_by_id, al.created_at, al.updated_at,
    u.id AS actor_id, u.first_name AS actor_first_name, u.last_name AS actor_last_name,
    u.email AS actor_email, u.avatar AS actor_avatar
FROM "activity_logs" al
JOIN "users" u ON al.user_id = u.id
WHERE al.user_id = sqlc.arg(user_id)::uuid
  AND (sqlc.narg(organization_id)::uuid IS NULL OR al.organization_id = sqlc.narg(organization_id)::uuid)
ORDER BY al.created_at DESC
LIMIT sqlc.arg(limit_val)::integer
OFFSET sqlc.arg(offset_val)::integer;

