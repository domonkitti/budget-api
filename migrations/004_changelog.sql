CREATE TABLE IF NOT EXISTS change_log (
    id          SERIAL PRIMARY KEY,
    table_name  VARCHAR(30)     NOT NULL,
    row_id      INT             NOT NULL,
    project_id  INT             NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    row_name    VARCHAR(200),
    fund_type   VARCHAR(10),
    data_year   INT,
    field       VARCHAR(20)     NOT NULL,
    old_value   NUMERIC(15,3),
    new_value   NUMERIC(15,3),
    changed_at  TIMESTAMP       DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_change_log_project ON change_log(project_id, changed_at DESC);

CREATE OR REPLACE FUNCTION log_sub_job_change() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.budget IS DISTINCT FROM NEW.budget THEN
        INSERT INTO change_log(table_name, row_id, project_id, row_name, fund_type, data_year, field, old_value, new_value)
        VALUES ('sub_jobs', NEW.id, NEW.project_id, NEW.name, NEW.fund_type, NEW.data_year, 'budget', OLD.budget, NEW.budget);
    END IF;
    IF OLD.target IS DISTINCT FROM NEW.target THEN
        INSERT INTO change_log(table_name, row_id, project_id, row_name, fund_type, data_year, field, old_value, new_value)
        VALUES ('sub_jobs', NEW.id, NEW.project_id, NEW.name, NEW.fund_type, NEW.data_year, 'target', OLD.target, NEW.target);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION log_budget_source_change() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.budget IS DISTINCT FROM NEW.budget THEN
        INSERT INTO change_log(table_name, row_id, project_id, row_name, fund_type, data_year, field, old_value, new_value)
        VALUES ('budget_sources', NEW.id, NEW.project_id, NEW.source, NEW.fund_type, NEW.data_year, 'budget', OLD.budget, NEW.budget);
    END IF;
    IF OLD.target IS DISTINCT FROM NEW.target THEN
        INSERT INTO change_log(table_name, row_id, project_id, row_name, fund_type, data_year, field, old_value, new_value)
        VALUES ('budget_sources', NEW.id, NEW.project_id, NEW.source, NEW.fund_type, NEW.data_year, 'target', OLD.target, NEW.target);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_sub_jobs_log ON sub_jobs;
CREATE TRIGGER trg_sub_jobs_log
    AFTER UPDATE ON sub_jobs
    FOR EACH ROW EXECUTE FUNCTION log_sub_job_change();

DROP TRIGGER IF EXISTS trg_budget_sources_log ON budget_sources;
CREATE TRIGGER trg_budget_sources_log
    AFTER UPDATE ON budget_sources
    FOR EACH ROW EXECUTE FUNCTION log_budget_source_change();
