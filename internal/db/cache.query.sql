-- name: GetCached :one
SELECT preset_config_hash, mime_type, file_size, etag, is_stored_on_disk, data FROM cache_index
WHERE file_hash = ? AND preset = ?;

-- name: UpdateAccessTime :exec
UPDATE cache_index
SET last_accessed_at = CURRENT_TIMESTAMP
WHERE file_hash = ? AND preset = ?;

-- name: InsertCache :exec
INSERT INTO cache_index (file_hash, preset, preset_config_hash, data, is_stored_on_disk, mime_type, file_size, etag)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(file_hash, preset) DO UPDATE SET
    preset_config_hash = excluded.preset_config_hash,
    data = excluded.data,
    is_stored_on_disk = excluded.is_stored_on_disk,
    mime_type = excluded.mime_type,
    file_size = excluded.file_size,
    last_accessed_at = CURRENT_TIMESTAMP;

-- name: GetCacheSize :one
SELECT COALESCE(SUM(file_size), 0) FROM cache_index;

-- name: GetOldestCacheItem :many
SELECT file_hash, preset, file_size, is_stored_on_disk
FROM cache_index
ORDER BY last_accessed_at ASC
LIMIT ?;

-- name: DeleteCacheItem :exec
DELETE FROM cache_index
WHERE file_hash = ? AND preset = ?;
