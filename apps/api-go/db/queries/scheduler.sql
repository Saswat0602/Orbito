-- name: ListTasksDueWithinWindowForCron :many
SELECT
    t.id, t.title, t.slug, t.due_date, w.organization_id
FROM "tasks" t
JOIN "task_statuses" ts ON t.status_id = ts.id
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
WHERE t.due_date >= $1
  AND t.due_date <= $2
  AND ts.category <> 'DONE'
  AND EXISTS (SELECT 1 FROM "task_assignees" ta WHERE ta.task_id = t.id);

-- name: ListOverdueTasksWithAssigneesForCron :many
SELECT
    t.id AS task_id, t.title AS task_title, t.slug AS task_slug, t.due_date,
    p.name AS project_name,
    o.name AS organization_name,
    u.id AS user_id, u.email AS user_email, u.first_name AS user_first_name, u.last_name AS user_last_name
FROM "tasks" t
JOIN "task_statuses" ts ON t.status_id = ts.id
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
JOIN "organizations" o ON w.organization_id = o.id
JOIN "task_assignees" ta ON ta.task_id = t.id
JOIN "users" u ON ta.user_id = u.id
WHERE t.due_date < $1
  AND ts.category <> 'DONE'
  AND u.deleted_at IS NULL
ORDER BY t.due_date ASC;

-- name: ListActiveUsersForWeeklySummaryForCron :many
SELECT DISTINCT u.id, u.email, u.first_name, u.last_name
FROM "users" u
JOIN "task_assignees" ta ON ta.user_id = u.id
JOIN "tasks" t ON ta.task_id = t.id
WHERE u.status = 'ACTIVE'
  AND u.deleted_at IS NULL
  AND t.created_at >= $1;

-- name: DeleteOldNotificationsForCron :execrows
DELETE FROM "notifications" WHERE created_at < $1;

-- name: DeleteOldActivityLogsForCron :execrows
DELETE FROM "activity_logs" WHERE created_at < $1;

-- name: DeleteOldRuleExecutionsForCron :execrows
DELETE FROM "rule_executions" WHERE created_at < $1;

-- name: ActivatePlanningSprintsForCron :execrows
UPDATE "sprints"
SET status = 'ACTIVE', updated_at = CURRENT_TIMESTAMP
WHERE status = 'PLANNING' AND start_date <= $1;

-- name: CompleteActiveSprintsForCron :execrows
UPDATE "sprints"
SET status = 'COMPLETED', updated_at = CURRENT_TIMESTAMP
WHERE status = 'ACTIVE' AND end_date <= $1;

-- name: ListAllOrganizationIDsForCron :many
SELECT id FROM "organizations" WHERE archive = false;

-- name: GetWeeklyAnalyticsForOrgCron :one
SELECT
    (SELECT COUNT(*)::bigint FROM "tasks" t
     JOIN "projects" p ON t.project_id = p.id
     JOIN "workspaces" w ON p.workspace_id = w.id
     WHERE w.organization_id = sqlc.arg(org_id)::uuid AND t.created_at >= sqlc.arg(since)::timestamp) AS tasks_created,
    (SELECT COUNT(*)::bigint FROM "tasks" t2
     JOIN "projects" p2 ON t2.project_id = p2.id
     JOIN "workspaces" w2 ON p2.workspace_id = w2.id
     JOIN "task_statuses" ts ON t2.status_id = ts.id
     WHERE w2.organization_id = sqlc.arg(org_id)::uuid AND ts.category = 'DONE' AND t2.updated_at >= sqlc.arg(since)::timestamp) AS tasks_completed,
    (SELECT COALESCE(SUM(te.time_spent), 0)::bigint FROM "time_entries" te
     JOIN "tasks" t3 ON te.task_id = t3.id
     JOIN "projects" p3 ON t3.project_id = p3.id
     JOIN "workspaces" w3 ON p3.workspace_id = w3.id
     WHERE w3.organization_id = sqlc.arg(org_id)::uuid AND te.date >= sqlc.arg(since)::timestamp) AS total_time_spent;

-- name: ListActiveRecurringTasksForCron :many
SELECT
    rt.id, rt.task_id, rt.recurrence_type, rt.interval, rt.days_of_week, rt.day_of_month, rt.month_of_year,
    rt.end_type, rt.end_date, rt.occurrence_count, rt.current_occurrence, rt.is_active, rt.next_occurrence,
    t.title, t.description, t.type, t.priority, t.project_id, t.status_id, t.sprint_id, t.created_by_id,
    p.slug AS project_slug, p.workspace_id, w.organization_id
FROM "recurring_tasks" rt
JOIN "tasks" t ON rt.task_id = t.id
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
WHERE rt.is_active = true;

-- name: DeactivateRecurringTaskForCron :exec
UPDATE "recurring_tasks"
SET is_active = false, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: ClearTaskIsRecurringForCron :exec
UPDATE "tasks"
SET is_recurring = false, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: FindExistingOccurrenceTaskForCron :one
SELECT id
FROM "tasks"
WHERE project_id = $1 AND title = $2 AND due_date = $3
LIMIT 1;

-- name: ListTaskAssigneeUserIDsForCron :many
SELECT user_id FROM "task_assignees" WHERE task_id = $1;

-- name: ListTaskReporterUserIDsForCron :many
SELECT user_id FROM "task_reporters" WHERE task_id = $1;

-- name: GetLastTaskNumberInProjectForCron :one
SELECT COALESCE(MAX(task_number), 0)::integer AS max_num
FROM "tasks"
WHERE project_id = $1;

-- name: InsertRecurringOccurrenceTaskForCron :one
INSERT INTO "tasks" (
    id, title, description, type, priority, project_id, status_id, sprint_id,
    due_date, task_number, slug, created_by_id, updated_by_id, is_recurring,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8,
    $9, $10, $11, $12, $12, false,
    CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING id;

-- name: AdvanceRecurringTaskScheduleForCron :exec
UPDATE "recurring_tasks"
SET current_occurrence = current_occurrence + 1,
    next_occurrence = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;
