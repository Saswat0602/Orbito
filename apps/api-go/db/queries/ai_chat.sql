-- name: ListConversationsByUserID :many
SELECT id, title, session_id, user_id, created_at, updated_at
FROM "conversations"
WHERE user_id = $1
ORDER BY updated_at DESC
LIMIT 100;

-- name: ListChatMessagesBatch :many
SELECT id, role, content, conversation_id, created_at
FROM "chat_messages"
WHERE conversation_id = ANY($1::uuid[])
ORDER BY created_at ASC;

-- name: ListChatMessagesByConversationID :many
SELECT id, role, content, conversation_id, created_at
FROM "chat_messages"
WHERE conversation_id = $1
ORDER BY created_at ASC
LIMIT $2;

-- name: GetConversationByIDAndUser :one
SELECT id, title, session_id, user_id, created_at, updated_at
FROM "conversations"
WHERE id = $1 AND user_id = $2
LIMIT 1;

-- name: GetConversationBySessionID :one
SELECT id, title, session_id, user_id, created_at, updated_at
FROM "conversations"
WHERE session_id = $1
LIMIT 1;

-- name: CreateConversationRecord :one
INSERT INTO "conversations" (
    id, title, session_id, user_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: RenameConversationByIDAndUser :one
UPDATE "conversations"
SET title = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: TouchConversationUpdatedAt :exec
UPDATE "conversations"
SET updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteConversationByIDAndUser :execrows
DELETE FROM "conversations" WHERE id = $1 AND user_id = $2;

-- name: DeleteConversationBySessionIDAndUser :execrows
DELETE FROM "conversations" WHERE session_id = $1 AND user_id = $2;

-- name: DeleteMessagesByConversationID :exec
DELETE FROM "chat_messages" WHERE conversation_id = $1;

-- name: InsertChatMessageRecord :one
INSERT INTO "chat_messages" (
    id, role, content, conversation_id, created_at
) VALUES (
    $1, $2, $3, $4, CURRENT_TIMESTAMP
) RETURNING *;

