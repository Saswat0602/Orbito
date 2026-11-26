-- name: GetProjectBasicForGantt :one
SELECT id, name, start_date, end_date
FROM "projects"
WHERE id = $1;

-- name: GetSprintBasicForGantt :one
SELECT id, name, start_date, end_date, project_id
FROM "sprints"
WHERE id = $1;

-- name: ListSprintsForGanttByProject :many
SELECT id, name, start_date, end_date, status
FROM "sprints"
WHERE project_id = $1;

-- name: ListTasksForProjectGantt :many
SELECT
    t.id, t.title, t.slug, t.start_date, t.due_date, t.priority, t.type, t.parent_task_id, t.story_points,
    ts.name AS status_name, ts.color AS status_color
FROM "tasks" t
JOIN "task_statuses" ts ON t.status_id = ts.id
WHERE t.project_id = $1
ORDER BY t.created_at ASC;

-- name: ListTasksForSprintGantt :many
SELECT
    t.id, t.title, t.slug, t.start_date, t.due_date, t.priority, t.type, t.parent_task_id, t.story_points,
    ts.name AS status_name, ts.color AS status_color
FROM "tasks" t
JOIN "task_statuses" ts ON t.status_id = ts.id
WHERE t.sprint_id = $1
ORDER BY t.created_at ASC;

-- name: ListTaskAssigneesForGantt :many
SELECT
    ta.task_id,
    u.id AS user_id, u.first_name, u.last_name, u.avatar
FROM "task_assignees" ta
JOIN "users" u ON ta.user_id = u.id
WHERE ta.task_id = ANY($1::uuid[]);

-- name: ListTaskDependenciesForGantt :many
SELECT
    td.dependent_task_id,
    td.blocking_task_id,
    td.type
FROM "task_dependencies" td
JOIN "tasks" dt ON td.dependent_task_id = dt.id
WHERE dt.project_id = $1;

-- name: ListResourceTasksForProjectGantt :many
SELECT
    t.id, t.title, t.start_date, t.due_date, t.story_points
FROM "tasks" t
WHERE t.project_id = $1
  AND t.start_date IS NOT NULL
  AND t.due_date IS NOT NULL
  AND EXISTS (SELECT 1 FROM "task_assignees" ta WHERE ta.task_id = t.id);
