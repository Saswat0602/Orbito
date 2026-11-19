-- name: GetOrganizationMemberRole :one
SELECT role FROM "organization_members"
WHERE "user_id" = $1 AND "organization_id" = $2;

-- name: GetOrganizationIDBySlug :one
SELECT id FROM "organizations"
WHERE slug = $1 AND archive = false;

-- name: GetWorkspaceMemberRole :one
SELECT role FROM "workspace_members"
WHERE "user_id" = $1 AND "workspace_id" = $2;

-- name: GetWorkspaceIDBySlug :one
SELECT id FROM "workspaces"
WHERE slug = $1 AND archive = false;

-- name: GetProjectMemberRole :one
SELECT role FROM "project_members"
WHERE "user_id" = $1 AND "project_id" = $2;

-- name: GetProjectScopeInfo :one
SELECT id, visibility, "workspace_id" FROM "projects"
WHERE id = $1 AND archive = false;

-- name: GetProjectScopeInfoBySlug :one
SELECT id, visibility, "workspace_id" FROM "projects"
WHERE slug = $1 AND archive = false;
