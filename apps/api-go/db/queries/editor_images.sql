-- name: CreateMediaAssetRecord :one
INSERT INTO "media_assets" (
    id, "userId", "assetType", "storageKey", "filename",
    "mimeType", "size", "fileHash", "inCloud", created_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, CURRENT_TIMESTAMP
) RETURNING *;
