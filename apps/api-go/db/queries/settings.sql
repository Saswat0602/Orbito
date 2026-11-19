-- name: GetSettingByKeyAndUser :one
SELECT id, key, value, description, category, is_encrypted, "userId", created_at, updated_at
FROM settings
WHERE key = $1 AND "userId" = $2;

-- name: GetGlobalSettingByKey :one
SELECT id, key, value, description, category, is_encrypted, "userId", created_at, updated_at
FROM settings
WHERE key = $1 AND "userId" IS NULL
LIMIT 1;

-- name: UpsertUserSetting :one
INSERT INTO settings (id, key, value, description, category, is_encrypted, "userId", created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
ON CONFLICT ("userId", key)
DO UPDATE SET
    value = EXCLUDED.value,
    description = COALESCE(EXCLUDED.description, settings.description),
    category = EXCLUDED.category,
    is_encrypted = EXCLUDED.is_encrypted,
    updated_at = NOW()
RETURNING id, key, value, description, category, is_encrypted, "userId", created_at, updated_at;

-- name: CreateGlobalSetting :one
INSERT INTO settings (id, key, value, description, category, is_encrypted, "userId", created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, NULL, NOW(), NOW())
RETURNING id, key, value, description, category, is_encrypted, "userId", created_at, updated_at;

-- name: UpdateGlobalSettingByID :one
UPDATE settings
SET value = $2,
    description = COALESCE($3, description),
    category = $4,
    is_encrypted = $5,
    updated_at = NOW()
WHERE id = $1
RETURNING id, key, value, description, category, is_encrypted, "userId", created_at, updated_at;

-- name: ListUserSettings :many
SELECT id, key, value, description, category, is_encrypted, "userId", created_at, updated_at
FROM settings
WHERE "userId" = $1 AND ($2::text = '' OR category = $2::text)
ORDER BY key ASC;

-- name: ListGlobalSettings :many
SELECT id, key, value, description, category, is_encrypted, "userId", created_at, updated_at
FROM settings
WHERE "userId" IS NULL AND ($1::text = '' OR category = $1::text)
ORDER BY key ASC;

-- name: DeleteUserSetting :exec
DELETE FROM settings
WHERE "userId" = $1 AND key = $2;

-- name: DeleteGlobalSetting :exec
DELETE FROM settings
WHERE "userId" IS NULL AND key = $1;

-- name: CountUserSetting :one
SELECT COUNT(*) FROM settings
WHERE "userId" = $1 AND key = $2;

-- name: CountGlobalSetting :one
SELECT COUNT(*) FROM settings
WHERE "userId" IS NULL AND key = $1;
