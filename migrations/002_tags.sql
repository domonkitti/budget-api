CREATE TABLE IF NOT EXISTS tag_categories (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tag_values (
    id          SERIAL PRIMARY KEY,
    category_id INT NOT NULL REFERENCES tag_categories(id) ON DELETE CASCADE,
    code        VARCHAR NOT NULL,
    created_at  TIMESTAMP DEFAULT NOW(),
    UNIQUE(category_id, code)
);

-- Tags applied at logical sub-job level (project_id + name covers both fund types)
CREATE TABLE IF NOT EXISTS sub_job_tags (
    id           SERIAL PRIMARY KEY,
    project_id   INT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    sub_job_name VARCHAR NOT NULL,
    tag_value_id INT NOT NULL REFERENCES tag_values(id) ON DELETE CASCADE,
    percentage   NUMERIC(5,2) NOT NULL DEFAULT 100,
    created_at   TIMESTAMP DEFAULT NOW(),
    UNIQUE(project_id, sub_job_name, tag_value_id)
);

CREATE INDEX IF NOT EXISTS idx_tag_values_category ON tag_values(category_id);
CREATE INDEX IF NOT EXISTS idx_sub_job_tags_project ON sub_job_tags(project_id);
CREATE INDEX IF NOT EXISTS idx_sub_job_tags_value   ON sub_job_tags(tag_value_id);
