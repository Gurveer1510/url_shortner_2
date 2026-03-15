CREATE TABLE IF NOT EXISTS links (
    code       TEXT PRIMARY KEY,
    url        TEXT NOT NULL,
    clicks     INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
  