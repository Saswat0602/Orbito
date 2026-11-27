-- name: CreateAutomationRule :one
INSERT INTO "automation_rules" (
    id, name, description, status, trigger_type, trigger_config,
    action_type, action_config, organization_id, workspace_id, project_id,
    created_by_id, updated_by_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10, $11,
    $12, $12, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING id;

-- name: GetAutomationRuleDetailByID :one
SELECT
    ar.id, ar.name, ar.description, ar.status, ar.trigger_type, ar.trigger_config,
    ar.action_type, ar.action_config, ar.organization_id, ar.workspace_id, ar.project_id,
    ar.execution_count, ar.last_executed, ar.created_by_id, ar.updated_by_id, ar.created_at, ar.updated_at,
    u.id AS creator_id, u.first_name AS creator_first_name, u.last_name AS creator_last_name, u.email AS creator_email,
    o.name AS org_name,
    w.name AS workspace_name,
    p.name AS project_name, p.slug AS project_slug
FROM "automation_rules" ar
LEFT JOIN "users" u ON ar.created_by_id = u.id
LEFT JOIN "organizations" o ON ar.organization_id = o.id
LEFT JOIN "workspaces" w ON ar.workspace_id = w.id
LEFT JOIN "projects" p ON ar.project_id = p.id
WHERE ar.id = $1
LIMIT 1;

-- name: ListAutomationRulesFiltered :many
SELECT
    ar.id, ar.name, ar.description, ar.status, ar.trigger_type, ar.trigger_config,
    ar.action_type, ar.action_config, ar.organization_id, ar.workspace_id, ar.project_id,
    ar.execution_count, ar.last_executed, ar.created_by_id, ar.updated_by_id, ar.created_at, ar.updated_at,
    u.id AS creator_id, u.first_name AS creator_first_name, u.last_name AS creator_last_name, u.email AS creator_email,
    o.name AS org_name,
    w.name AS workspace_name,
    p.name AS project_name, p.slug AS project_slug
FROM "automation_rules" ar
LEFT JOIN "users" u ON ar.created_by_id = u.id
LEFT JOIN "organizations" o ON ar.organization_id = o.id
LEFT JOIN "workspaces" w ON ar.workspace_id = w.id
LEFT JOIN "projects" p ON ar.project_id = p.id
WHERE (sqlc.narg(project_id)::uuid IS NULL OR ar.project_id = sqlc.narg(project_id)::uuid)
  AND (sqlc.narg(workspace_id)::uuid IS NULL OR ar.workspace_id = sqlc.narg(workspace_id)::uuid)
  AND (sqlc.narg(organization_id)::uuid IS NULL OR ar.organization_id = sqlc.narg(organization_id)::uuid)
ORDER BY ar.created_at DESC;

-- name: ListRuleExecutionsByRuleID :many
SELECT
    re.id, re.success, re.error_message, re.execution_time,
    re.trigger_data, re.action_result, re.rule_id, re.triggered_by_id,
    re.created_by_id, re.updated_by_id, re.created_at, re.updated_at,
    u.id AS trig_user_id, u.first_name AS trig_first_name, u.last_name AS trig_last_name, u.email AS trig_email
FROM "rule_executions" re
LEFT JOIN "users" u ON re.triggered_by_id = u.id
WHERE re.rule_id = $1
ORDER BY re.created_at DESC
LIMIT $2;

-- name: UpdateAutomationRuleByID :execrows
UPDATE "automation_rules"
SET
    name = COALESCE(sqlc.narg(name)::text, name),
    description = COALESCE(sqlc.narg(description)::text, description),
    status = COALESCE(sqlc.narg(status)::"RuleStatus", status),
    trigger_type = COALESCE(sqlc.narg(trigger_type)::"TriggerType", trigger_type),
    trigger_config = COALESCE(sqlc.narg(trigger_config)::jsonb, trigger_config),
    action_type = COALESCE(sqlc.narg(action_type)::"ActionType", action_type),
    action_config = COALESCE(sqlc.narg(action_config)::jsonb, action_config),
    updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg(id)::uuid;

-- name: ToggleAutomationRuleStatus :one
UPDATE "automation_rules"
SET status = CASE WHEN status = 'ACTIVE' THEN 'INACTIVE'::"RuleStatus" ELSE 'ACTIVE'::"RuleStatus" END,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteAutomationRuleByID :execrows
DELETE FROM "automation_rules" WHERE id = $1;

-- name: IncrementAutomationRuleExecution :exec
UPDATE "automation_rules"
SET execution_count = execution_count + 1,
    last_executed = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: CreateRuleExecutionRecord :one
INSERT INTO "rule_executions" (
    id, success, error_message, execution_time, trigger_data, action_result,
    rule_id, triggered_by_id, created_by_id, updated_by_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $9, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: GetAutomationRuleStatsRow :one
SELECT
    COUNT(*)::bigint AS total_executions,
    COALESCE(SUM(CASE WHEN success THEN 1 ELSE 0 END), 0)::bigint AS success_count,
    COALESCE(AVG(execution_time), 0)::double precision AS avg_execution_time
FROM "rule_executions"
WHERE rule_id = $1;

-- name: ListActiveAutomationRulesByTrigger :many
SELECT *
FROM "automation_rules"
WHERE status = 'ACTIVE'
  AND trigger_type = sqlc.arg(trigger_type)::"TriggerType"
  AND (
    (sqlc.narg(project_id)::uuid IS NOT NULL AND project_id = sqlc.narg(project_id)::uuid)
    OR (sqlc.narg(workspace_id)::uuid IS NOT NULL AND workspace_id = sqlc.narg(workspace_id)::uuid)
    OR (sqlc.narg(organization_id)::uuid IS NOT NULL AND organization_id = sqlc.narg(organization_id)::uuid)
    OR (project_id IS NULL AND workspace_id IS NULL AND organization_id IS NULL)
  );

-- name: GetTaskForAutomationAction :one
SELECT
    t.id, t.title, t.project_id, t.priority, t.status_id,
    p.workspace_id, w.organization_id
FROM "tasks" t
JOIN "projects" p ON t.project_id = p.id
JOIN "workspaces" w ON p.workspace_id = w.id
WHERE t.id = $1
LIMIT 1;

-- name: UpdateTaskStatusForAutomation :one
UPDATE "tasks"
SET status_id = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING id, project_id, status_id;

-- name: GetTaskStatusSummaryForAutomation :one
SELECT id, name, color
FROM "task_statuses"
WHERE id = $1
LIMIT 1;

-- name: UpdateTaskPriorityForAutomation :execrows
UPDATE "tasks"
SET priority = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: UpdateTaskDueDateForAutomation :execrows
UPDATE "tasks"
SET due_date = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteTaskAssigneesForAutomation :exec
DELETE FROM "task_assignees" WHERE task_id = $1;

-- name: InsertTaskAssigneeForAutomation :exec
INSERT INTO "task_assignees" (task_id, user_id, created_at)
VALUES ($1, $2, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;

-- name: InsertTaskLabelForAutomation :exec
INSERT INTO "task_labels" (task_id, label_id, created_at)
VALUES ($1, $2, CURRENT_TIMESTAMP)
ON CONFLICT DO NOTHING;

-- name: InsertTaskCommentForAutomation :one
INSERT INTO "task_comments" (id, task_id, author_id, content, created_at, updated_at)
VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
RETURNING id, task_id, author_id, content, created_at, updated_at;
