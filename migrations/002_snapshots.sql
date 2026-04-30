CREATE TABLE IF NOT EXISTS snapshots (
    id          SERIAL PRIMARY KEY,
    label       VARCHAR(200) NOT NULL,
    note        TEXT,
    created_at  TIMESTAMP DEFAULT NOW(),
    data        JSONB NOT NULL DEFAULT '[]'
);
