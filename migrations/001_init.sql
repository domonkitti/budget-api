CREATE TABLE IF NOT EXISTS projects (
    id            SERIAL PRIMARY KEY,
    project_code  VARCHAR(20) UNIQUE NOT NULL,
    year          INT NOT NULL,
    project_type  VARCHAR(1) NOT NULL CHECK (project_type IN ('Y', 'C', 'L')),
    item_no       VARCHAR,
    name          VARCHAR NOT NULL,
    division      VARCHAR,
    department    VARCHAR,
    created_at    TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sub_jobs (
    id          SERIAL PRIMARY KEY,
    project_id  INT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        VARCHAR NOT NULL,
    sort_order  INT,
    fund_type   VARCHAR(10) NOT NULL CHECK (fund_type IN ('ผูกพัน', 'ลงทุน')),
    budget      NUMERIC(15,3) DEFAULT 0,
    target      NUMERIC(15,3) DEFAULT 0,
    created_at  TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS budget_sources (
    id          SERIAL PRIMARY KEY,
    project_id  INT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    source      VARCHAR(50) NOT NULL,
    fund_type   VARCHAR(10) NOT NULL CHECK (fund_type IN ('ผูกพัน', 'ลงทุน')),
    budget      NUMERIC(15,3) DEFAULT 0,
    target      NUMERIC(15,3) DEFAULT 0,
    created_at  TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_projects_year ON projects(year);
CREATE INDEX IF NOT EXISTS idx_projects_type ON projects(project_type);
CREATE INDEX IF NOT EXISTS idx_projects_division ON projects(division);
CREATE INDEX IF NOT EXISTS idx_sub_jobs_project ON sub_jobs(project_id);
CREATE INDEX IF NOT EXISTS idx_budget_sources_project ON budget_sources(project_id);
