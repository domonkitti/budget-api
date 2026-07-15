CREATE TABLE IF NOT EXISTS report_groups (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR NOT NULL,
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS reports (
    id          SERIAL PRIMARY KEY,
    group_id    INT NOT NULL REFERENCES report_groups(id) ON DELETE CASCADE,
    preset_id   VARCHAR,
    data        JSONB NOT NULL,
    created_at  TIMESTAMP DEFAULT NOW(),
    updated_at  TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reports_group ON reports(group_id);

INSERT INTO report_groups (name, sort_order) VALUES
    ('การประชุมครั้งที่ 1', 0),
    ('การประชุมครั้งที่ 2', 1);
