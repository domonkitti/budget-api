ALTER TABLE change_log
    ADD COLUMN IF NOT EXISTS batch_id      TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS batch_comment TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_change_log_batch ON change_log(batch_id) WHERE batch_id <> '';

CREATE OR REPLACE FUNCTION log_sub_job_change() RETURNS TRIGGER AS $$
DECLARE
    v_batch_id TEXT;
BEGIN
    v_batch_id := COALESCE(current_setting('app.batch_id', true), '');
    IF OLD.budget IS DISTINCT FROM NEW.budget THEN
        INSERT INTO change_log(table_name, row_id, project_id, row_name, fund_type, data_year, field, old_value, new_value, batch_id)
        VALUES ('sub_jobs', NEW.id, NEW.project_id, NEW.name, NEW.fund_type, NEW.data_year, 'budget', OLD.budget, NEW.budget, v_batch_id);
    END IF;
    IF OLD.target IS DISTINCT FROM NEW.target THEN
        INSERT INTO change_log(table_name, row_id, project_id, row_name, fund_type, data_year, field, old_value, new_value, batch_id)
        VALUES ('sub_jobs', NEW.id, NEW.project_id, NEW.name, NEW.fund_type, NEW.data_year, 'target', OLD.target, NEW.target, v_batch_id);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION log_budget_source_change() RETURNS TRIGGER AS $$
DECLARE
    v_batch_id TEXT;
BEGIN
    v_batch_id := COALESCE(current_setting('app.batch_id', true), '');
    IF OLD.budget IS DISTINCT FROM NEW.budget THEN
        INSERT INTO change_log(table_name, row_id, project_id, row_name, fund_type, data_year, field, old_value, new_value, batch_id)
        VALUES ('budget_sources', NEW.id, NEW.project_id, NEW.source, NEW.fund_type, NEW.data_year, 'budget', OLD.budget, NEW.budget, v_batch_id);
    END IF;
    IF OLD.target IS DISTINCT FROM NEW.target THEN
        INSERT INTO change_log(table_name, row_id, project_id, row_name, fund_type, data_year, field, old_value, new_value, batch_id)
        VALUES ('budget_sources', NEW.id, NEW.project_id, NEW.source, NEW.fund_type, NEW.data_year, 'target', OLD.target, NEW.target, v_batch_id);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
