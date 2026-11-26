-- name: GetTaskRanksForRebalance :many
SELECT id
FROM "task_ranks"
WHERE scope_type = $1 AND scope_id = $2 AND view_type = $3
ORDER BY rank ASC, task_id ASC;

-- name: UpdateTaskRankValueByID :exec
UPDATE "task_ranks"
SET rank = $2
WHERE id = $1;

-- name: GetMaxTaskRanksForScopes :many
SELECT
    scope_type,
    scope_id,
    view_type,
    COALESCE(MAX(rank), 0)::double precision AS max_rank
FROM "task_ranks"
WHERE (scope_type = 'PROJECT' AND scope_id = sqlc.arg(project_id)::uuid)
   OR (scope_type = 'WORKSPACE' AND scope_id = sqlc.arg(workspace_id)::uuid)
   OR (scope_type = 'ORGANIZATION' AND scope_id = sqlc.arg(org_id)::uuid)
GROUP BY scope_type, scope_id, view_type;

-- name: InsertTaskRankSkipConflict :exec
INSERT INTO "task_ranks" (id, task_id, scope_type, scope_id, view_type, rank)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (task_id, scope_type, scope_id, view_type) DO NOTHING;

-- name: ListUnrankedTopTasksInProject :many
SELECT t.id
FROM "tasks" t
WHERE t.project_id = $1
  AND t.parent_task_id IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM "task_ranks" tr
      WHERE tr.task_id = t.id
        AND tr.scope_type = $2
        AND tr.scope_id = $3
        AND tr.view_type = $4
  )
ORDER BY t.task_number DESC;

-- name: ListUnrankedTopTasksInWorkspace :many
SELECT t.id
FROM "tasks" t
JOIN "projects" p ON p.id = t.project_id
WHERE p.workspace_id = $1
  AND t.parent_task_id IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM "task_ranks" tr
      WHERE tr.task_id = t.id
        AND tr.scope_type = $2
        AND tr.scope_id = $3
        AND tr.view_type = $4
  )
ORDER BY t.task_number DESC;

-- name: ListUnrankedTopTasksInOrg :many
SELECT t.id
FROM "tasks" t
JOIN "projects" p ON p.id = t.project_id
JOIN "workspaces" w ON w.id = p.workspace_id
WHERE w.organization_id = $1
  AND t.parent_task_id IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM "task_ranks" tr
      WHERE tr.task_id = t.id
        AND tr.scope_type = $2
        AND tr.scope_id = $3
        AND tr.view_type = $4
  )
ORDER BY t.task_number DESC;

-- name: GetMaxTaskRankSingleScope :one
SELECT COALESCE(MAX(rank), 0)::double precision AS max_rank
FROM "task_ranks"
WHERE scope_type = $1 AND scope_id = $2 AND view_type = $3;

-- name: GetTaskRankByUniqueKey :one
SELECT rank
FROM "task_ranks"
WHERE task_id = $1 AND scope_type = $2 AND scope_id = $3 AND view_type = $4
LIMIT 1;

-- name: FindTrueAfterRank :one
SELECT rank
FROM "task_ranks"
WHERE scope_type = $1 AND scope_id = $2 AND view_type = $3 AND rank < $4
ORDER BY rank DESC, task_id DESC
LIMIT 1;

-- name: FindTrueBeforeRank :one
SELECT rank
FROM "task_ranks"
WHERE scope_type = $1 AND scope_id = $2 AND view_type = $3 AND rank > $4
ORDER BY rank ASC, task_id ASC
LIMIT 1;

-- name: UpsertTaskRank :exec
INSERT INTO "task_ranks" (id, task_id, scope_type, scope_id, view_type, rank)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (task_id, scope_type, scope_id, view_type)
DO UPDATE SET rank = EXCLUDED.rank;
