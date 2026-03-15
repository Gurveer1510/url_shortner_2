CREATE TABLE IF NOT EXISTS url_clicks (
    id         BIGSERIAL PRIMARY KEY,
    code       TEXT NOT NULL REFERENCES links(code) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address TEXT
);

CREATE INDEX IF NOT EXISTS idx_url_clicks_code ON url_clicks(code);