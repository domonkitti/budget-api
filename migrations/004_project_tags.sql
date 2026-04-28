CREATE TABLE IF NOT EXISTS project_tags (
    id           SERIAL PRIMARY KEY,
    project_id   INT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    tag_value_id INT NOT NULL REFERENCES tag_values(id) ON DELETE CASCADE,
    percentage   NUMERIC(5,2) NOT NULL DEFAULT 100,
    created_at   TIMESTAMP DEFAULT NOW(),
    UNIQUE(project_id, tag_value_id)
);

CREATE INDEX IF NOT EXISTS idx_project_tags_project ON project_tags(project_id);
CREATE INDEX IF NOT EXISTS idx_project_tags_value   ON project_tags(tag_value_id);
