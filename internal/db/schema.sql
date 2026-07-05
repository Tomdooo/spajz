-- CREATE TABLE IF NOT EXISTS cache_index (
--     hash TEXT PRIMARY KEY,
--     variant TEXT PRIMARY KEY,
--     file_size INTEGER NOT NULL,
--     format TEXT NOT NULL,
--     created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
--     last_accessed_at DATETIME DEFAULT CURRENT_TIMESTAMP
-- );

-- CREATE INDEX IF NOT EXISTS idx_cache_last_accessed ON cache_index(last_accessed_at);
