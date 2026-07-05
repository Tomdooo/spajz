-- -- name: GetCache :one
-- SELECT data, file_size FROM cache_index
-- WHERE hash = ? AND variant = ?;

-- -- name: UpdateAccessTime :exec
-- UPDATE cache_index
-- SET last_accessed_at = CURRENT_TIMESTAMP
-- WHERE hash = ?;

-- -- name: InsertCache :exec
-- INSERT INTO cache_index (hash, file_size, data)
-- VALUES (?, ?, ?)
-- ON CONFLICT(hash) DO UPDATE SET
--     data = excluded.data,
--     file_size = excluded.file_size,
--     last_accessed_at = CURRENT_TIMESTAMP;

-- -- name: GetCacheSize :one
-- SELECT COALESCE(SUM(file_size), 0) FROM cache_index;

-- -- name: DeleteOldestCache :exec
-- DELETE FROM cache_index
-- WHERE hash IN (
--     SELECT hash FROM cache_index
--     ORDER BY last_accessed_at ASC
--     LIMIT 1
-- );
