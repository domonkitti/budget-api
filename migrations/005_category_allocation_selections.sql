CREATE TABLE IF NOT EXISTS category_allocation_selections (
    id           SERIAL PRIMARY KEY,
    category_id  INT NOT NULL REFERENCES tag_categories(id) ON DELETE CASCADE,
    project_id   INT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    target_type  VARCHAR(10) NOT NULL CHECK (target_type IN ('project', 'job')),
    sub_job_name VARCHAR,
    created_at   TIMESTAMP DEFAULT NOW(),
    UNIQUE(category_id, project_id, target_type, sub_job_name)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_category_allocation_selections_project
    ON category_allocation_selections(category_id, project_id)
    WHERE target_type = 'project';

CREATE INDEX IF NOT EXISTS idx_category_allocation_selections_category
    ON category_allocation_selections(category_id);
