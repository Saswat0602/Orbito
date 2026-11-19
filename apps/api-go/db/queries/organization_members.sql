-- name: ListOrganizationMembers :many
SELECT om.id, om.role, om.joined_at, om.user_id, om.organization_id, om.is_default, om.created_at, om.updated_at,
       u.id AS user_user_id, u.email AS user_email, u.first_name AS user_first_name, u.last_name AS user_last_name, u.avatar AS user_avatar, u.status AS user_status,
       o.id AS org_id, o.name AS org_name, o.slug AS org_slug, o.owner_id AS org_owner_id
FROM "organization_members" om
JOIN "users" u ON om.user_id = u.id
JOIN "organizations" o ON om.organization_id = o.id
WHERE om.organization_id = $1
  AND ($2::text = '' OR u.first_name ILIKE '%' || $2 || '%' OR u.last_name ILIKE '%' || $2 || '%' OR u.email ILIKE '%' || $2 || '%')
ORDER BY om.joined_at ASC;

-- name: ListOrganizationMembersBySlug :many
SELECT om.id, om.role, om.joined_at, om.user_id, om.organization_id, om.is_default, om.created_at, om.updated_at,
       u.id AS user_user_id, u.email AS user_email, u.first_name AS user_first_name, u.last_name AS user_last_name, u.avatar AS user_avatar, u.status AS user_status,
       o.id AS org_id, o.name AS org_name, o.slug AS org_slug, o.owner_id AS org_owner_id
FROM "organization_members" om
JOIN "users" u ON om.user_id = u.id
JOIN "organizations" o ON om.organization_id = o.id
WHERE o.slug = $1 AND o.archive = false
  AND ($2::text = '' OR u.first_name ILIKE '%' || $2 || '%' OR u.last_name ILIKE '%' || $2 || '%' OR u.email ILIKE '%' || $2 || '%')
ORDER BY om.joined_at ASC
LIMIT $3 OFFSET $4;

-- name: CountOrganizationMembersBySlug :one
SELECT COUNT(*)
FROM "organization_members" om
JOIN "users" u ON om.user_id = u.id
JOIN "organizations" o ON om.organization_id = o.id
WHERE o.slug = $1 AND o.archive = false
  AND ($2::text = '' OR u.first_name ILIKE '%' || $2 || '%' OR u.last_name ILIKE '%' || $2 || '%' OR u.email ILIKE '%' || $2 || '%');

-- name: GetOrganizationMemberByID :one
SELECT om.*,
       u.id AS user_user_id, u.email AS user_email, u.first_name AS user_first_name, u.last_name AS user_last_name, u.avatar AS user_avatar, u.status AS user_status,
       o.id AS org_id, o.name AS org_name, o.slug AS org_slug, o.owner_id AS org_owner_id
FROM "organization_members" om
JOIN "users" u ON om.user_id = u.id
JOIN "organizations" o ON om.organization_id = o.id
WHERE om.id = $1;

-- name: GetOrganizationMemberByUserAndOrg :one
SELECT om.*,
       u.id AS user_user_id, u.email AS user_email, u.first_name AS user_first_name, u.last_name AS user_last_name, u.avatar AS user_avatar, u.status AS user_status,
       o.id AS org_id, o.name AS org_name, o.slug AS org_slug, o.owner_id AS org_owner_id
FROM "organization_members" om
JOIN "users" u ON om.user_id = u.id
JOIN "organizations" o ON om.organization_id = o.id
WHERE om.user_id = $1 AND om.organization_id = $2;

-- name: UpdateOrganizationMemberRoleByID :one
UPDATE "organization_members"
SET role = $2,
    updated_by_id = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteOrganizationMemberByID :exec
DELETE FROM "organization_members"
WHERE id = $1;

-- name: SetUserDefaultOrgForUser :exec
UPDATE "users"
SET default_organization_id = $2
WHERE id = $1;

-- name: ClearUserDefaultOrgIfMatches :exec
UPDATE "users"
SET default_organization_id = NULL
WHERE id = $1 AND default_organization_id = $2;
