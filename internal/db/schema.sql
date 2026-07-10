CREATE TABLE IF NOT EXISTS cache_index (
    file_hash         TEXT NOT NULL,     -- Unikátní hash originálního souboru (např. SHA-256)
    preset       TEXT NOT NULL,     -- Název presetu z TOMLu (např. 'thumb_small')
    preset_config_hash       TEXT NOT NULL,     -- Otisk (ETag) konfigurace presetu pro detekci změn v TOMLu

    etag              TEXT NOT NULL,
    mime_type         TEXT NOT NULL,     -- Např. image/webp (důležité pro Echo HTTP hlavičky)
    file_size         INTEGER NOT NULL,  -- Velikost v bajtech pro hlídání celkového limitu cache
    is_stored_on_disk INTEGER NOT NULL DEFAULT 0, -- Fáze 1: 0 (BLOB v DB), Fáze 2: 1 (Soubor na disku)
    created_at        TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,     -- ISO 8601 čas vytvoření
    last_accessed_at  TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,     -- Klíčové pro LRU algoritmus (čištění starých dat)
    data              BLOB,              -- Samotná binární data (v budoucnu u velkých souborů NULL)

    PRIMARY KEY (file_hash, preset)
);

CREATE INDEX IF NOT EXISTS idx_cache_last_accessed ON cache_index(last_accessed_at);
