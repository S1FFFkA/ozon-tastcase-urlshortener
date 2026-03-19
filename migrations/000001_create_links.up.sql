CREATE TABLE IF NOT EXISTS links (
    id            BIGSERIAL PRIMARY KEY,
    original_url  TEXT NOT NULL UNIQUE,
    short_url     TEXT NOT NULL UNIQUE,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at  TIMESTAMPTZ,
    use_count     BIGINT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_links_last_used_at ON links (last_used_at);
CREATE INDEX IF NOT EXISTS idx_links_created_at ON links (created_at);


