-- name: CreateInvitation :one
INSERT INTO invitations (
    id, inviter_id, invitee_email, organization_id, workspace_id, project_id,
    role, status, token, expires_at, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: GetInvitationByID :one
SELECT 
    i.id, i.inviter_id, i.invitee_email, i.organization_id, i.workspace_id, i.project_id,
    i.role, i.status, i.token, i.expires_at, i.created_at, i.updated_at,
    u.id AS inviter_user_id, u.email AS inviter_email, u.first_name AS inviter_first_name, u.last_name AS inviter_last_name, u.username AS inviter_username,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug,
    w.id AS ws_id, w.name AS ws_name, w.slug AS ws_slug,
    p.id AS proj_id, p.name AS proj_name, p.slug AS proj_slug
FROM invitations i
JOIN users u ON i.inviter_id = u.id
LEFT JOIN organizations o ON i.organization_id = o.id
LEFT JOIN workspaces w ON i.workspace_id = w.id
LEFT JOIN projects p ON i.project_id = p.id
WHERE i.id = $1;

-- name: GetInvitationByToken :one
SELECT 
    i.id, i.inviter_id, i.invitee_email, i.organization_id, i.workspace_id, i.project_id,
    i.role, i.status, i.token, i.expires_at, i.created_at, i.updated_at,
    u.id AS inviter_user_id, u.email AS inviter_email, u.first_name AS inviter_first_name, u.last_name AS inviter_last_name, u.username AS inviter_username,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug,
    w.id AS ws_id, w.name AS ws_name, w.slug AS ws_slug,
    p.id AS proj_id, p.name AS proj_name, p.slug AS proj_slug,
    pw.organization_id AS proj_org_id, pw.id AS proj_ws_id
FROM invitations i
JOIN users u ON i.inviter_id = u.id
LEFT JOIN organizations o ON i.organization_id = o.id
LEFT JOIN workspaces w ON i.workspace_id = w.id
LEFT JOIN projects p ON i.project_id = p.id
LEFT JOIN workspaces pw ON p.workspace_id = pw.id
WHERE i.token = $1;

-- name: FindPendingInvitation :one
SELECT * FROM invitations
WHERE invitee_email = $1
  AND status = 'PENDING'
  AND (organization_id = $2 OR ($2 IS NULL AND organization_id IS NULL))
  AND (workspace_id = $3 OR ($3 IS NULL AND workspace_id IS NULL))
  AND (project_id = $4 OR ($4 IS NULL AND project_id IS NULL))
LIMIT 1;

-- name: GetUserInvitations :many
SELECT 
    i.id, i.inviter_id, i.invitee_email, i.organization_id, i.workspace_id, i.project_id,
    i.role, i.status, i.token, i.expires_at, i.created_at, i.updated_at,
    u.id AS inviter_user_id, u.email AS inviter_email, u.first_name AS inviter_first_name, u.last_name AS inviter_last_name, u.username AS inviter_username,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug,
    w.id AS ws_id, w.name AS ws_name, w.slug AS ws_slug,
    p.id AS proj_id, p.name AS proj_name, p.slug AS proj_slug
FROM invitations i
JOIN users u ON i.inviter_id = u.id
LEFT JOIN organizations o ON i.organization_id = o.id
LEFT JOIN workspaces w ON i.workspace_id = w.id
LEFT JOIN projects p ON i.project_id = p.id
WHERE i.invitee_email = $1
  AND i.status = 'PENDING'
  AND i.expires_at > CURRENT_TIMESTAMP
ORDER BY i.created_at DESC;

-- name: ExpirePendingEntityInvitations :exec
UPDATE invitations
SET status = 'EXPIRED', updated_at = CURRENT_TIMESTAMP
WHERE status = 'PENDING'
  AND expires_at < CURRENT_TIMESTAMP
  AND (
      (sqlc.arg(entity_type)::text = 'organization' AND organization_id = sqlc.arg(entity_id)) OR
      (sqlc.arg(entity_type)::text = 'workspace' AND workspace_id = sqlc.arg(entity_id)) OR
      (sqlc.arg(entity_type)::text = 'project' AND project_id = sqlc.arg(entity_id))
  );

-- name: GetEntityInvitations :many
SELECT 
    i.id, i.inviter_id, i.invitee_email, i.organization_id, i.workspace_id, i.project_id,
    i.role, i.status, i.token, i.expires_at, i.created_at, i.updated_at,
    u.id AS inviter_user_id, u.email AS inviter_email, u.first_name AS inviter_first_name, u.last_name AS inviter_last_name, u.username AS inviter_username,
    o.id AS org_id, o.name AS org_name, o.slug AS org_slug,
    w.id AS ws_id, w.name AS ws_name, w.slug AS ws_slug,
    p.id AS proj_id, p.name AS proj_name, p.slug AS proj_slug
FROM invitations i
JOIN users u ON i.inviter_id = u.id
LEFT JOIN organizations o ON i.organization_id = o.id
LEFT JOIN workspaces w ON i.workspace_id = w.id
LEFT JOIN projects p ON i.project_id = p.id
WHERE (
      (sqlc.arg(entity_type)::text = 'organization' AND i.organization_id = sqlc.arg(entity_id)) OR
      (sqlc.arg(entity_type)::text = 'workspace' AND i.workspace_id = sqlc.arg(entity_id)) OR
      (sqlc.arg(entity_type)::text = 'project' AND i.project_id = sqlc.arg(entity_id))
  )
  AND i.status IN ('PENDING', 'DECLINED', 'EXPIRED')
ORDER BY i.created_at DESC;

-- name: UpdateInvitationStatus :exec
UPDATE invitations
SET status = $2, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: ResendInvitation :one
UPDATE invitations
SET token = $2, expires_at = $3, status = 'PENDING', inviter_id = $4, updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteInvitation :exec
DELETE FROM invitations WHERE id = $1;

-- name: ListRelatedProjectInvitations :many
SELECT i.*
FROM invitations i
JOIN projects p ON i.project_id = p.id
WHERE i.invitee_email = $1
  AND i.status = 'PENDING'
  AND p.workspace_id = $2;

-- name: GetProjectWithWorkspaceAndOrg :one
SELECT 
    p.id, p.name, p.slug, p.workspace_id,
    w.organization_id, w.name AS workspace_name, w.slug AS workspace_slug,
    o.name AS organization_name
FROM projects p
JOIN workspaces w ON p.workspace_id = w.id
JOIN organizations o ON w.organization_id = o.id
WHERE p.id = $1;

-- name: CheckProjectBelongsToWorkspace :one
SELECT count(*) FROM projects
WHERE id = $1 AND workspace_id = $2;

-- name: CheckOrgMemberExists :one
SELECT count(*) FROM organization_members
WHERE user_id = $1 AND organization_id = $2;

-- name: CheckWorkspaceMemberExists :one
SELECT count(*) FROM workspace_members
WHERE user_id = $1 AND workspace_id = $2;

-- name: CheckProjectMemberExists :one
SELECT count(*) FROM project_members
WHERE user_id = $1 AND project_id = $2;

-- name: CreateProjectMemberDirect :one
INSERT INTO project_members (
    id, user_id, project_id, role, joined_at, created_by_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, CURRENT_TIMESTAMP, $5, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;
