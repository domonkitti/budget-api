CREATE TABLE IF NOT EXISTS scenarios (
    id          SERIAL PRIMARY KEY,
    label       VARCHAR(200) NOT NULL,
    note        TEXT,
    created_at  TIMESTAMP DEFAULT NOW(),
    updated_at  TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS scenario_sub_jobs (
    id          SERIAL PRIMARY KEY,
    scenario_id INT NOT NULL REFERENCES scenarios(id) ON DELETE CASCADE,
    project_id  INT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name        VARCHAR NOT NULL,
    sort_order  INT,
    fund_type   VARCHAR(10) NOT NULL,
    data_year   INT NOT NULL,
    budget      NUMERIC(15,3) DEFAULT 0,
    target      NUMERIC(15,3) DEFAULT 0
);

CREATE TABLE IF NOT EXISTS scenario_budget_sources (
    id          SERIAL PRIMARY KEY,
    scenario_id INT NOT NULL REFERENCES scenarios(id) ON DELETE CASCADE,
    project_id  INT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    source      VARCHAR(50) NOT NULL,
    fund_type   VARCHAR(10) NOT NULL,
    data_year   INT NOT NULL,
    budget      NUMERIC(15,3) DEFAULT 0,
    target      NUMERIC(15,3) DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_scen_sj_scenario  ON scenario_sub_jobs(scenario_id);
CREATE INDEX IF NOT EXISTS idx_scen_sj_project   ON scenario_sub_jobs(project_id);
CREATE INDEX IF NOT EXISTS idx_scen_bs_scenario  ON scenario_budget_sources(scenario_id);
CREATE INDEX IF NOT EXISTS idx_scen_bs_project   ON scenario_budget_sources(project_id);
