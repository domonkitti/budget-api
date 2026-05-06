ALTER TABLE sub_jobs ADD COLUMN IF NOT EXISTS cut_transfer NUMERIC(15,3) NOT NULL DEFAULT 0;
ALTER TABLE sub_jobs ADD COLUMN IF NOT EXISTS under_budget NUMERIC(15,3) NOT NULL DEFAULT 0;
ALTER TABLE budget_sources ADD COLUMN IF NOT EXISTS cut_transfer NUMERIC(15,3) NOT NULL DEFAULT 0;
ALTER TABLE budget_sources ADD COLUMN IF NOT EXISTS under_budget NUMERIC(15,3) NOT NULL DEFAULT 0;
ALTER TABLE scenario_sub_jobs ADD COLUMN IF NOT EXISTS cut_transfer NUMERIC(15,3) NOT NULL DEFAULT 0;
ALTER TABLE scenario_sub_jobs ADD COLUMN IF NOT EXISTS under_budget NUMERIC(15,3) NOT NULL DEFAULT 0;
ALTER TABLE scenario_budget_sources ADD COLUMN IF NOT EXISTS cut_transfer NUMERIC(15,3) NOT NULL DEFAULT 0;
ALTER TABLE scenario_budget_sources ADD COLUMN IF NOT EXISTS under_budget NUMERIC(15,3) NOT NULL DEFAULT 0;

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
    IF OLD.cut_transfer IS DISTINCT FROM NEW.cut_transfer THEN
        INSERT INTO change_log(table_name, row_id, project_id, row_name, fund_type, data_year, field, old_value, new_value, batch_id)
        VALUES ('sub_jobs', NEW.id, NEW.project_id, NEW.name, NEW.fund_type, NEW.data_year, 'cut_transfer', OLD.cut_transfer, NEW.cut_transfer, v_batch_id);
    END IF;
    IF OLD.under_budget IS DISTINCT FROM NEW.under_budget THEN
        INSERT INTO change_log(table_name, row_id, project_id, row_name, fund_type, data_year, field, old_value, new_value, batch_id)
        VALUES ('sub_jobs', NEW.id, NEW.project_id, NEW.name, NEW.fund_type, NEW.data_year, 'under_budget', OLD.under_budget, NEW.under_budget, v_batch_id);
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
    IF OLD.cut_transfer IS DISTINCT FROM NEW.cut_transfer THEN
        INSERT INTO change_log(table_name, row_id, project_id, row_name, fund_type, data_year, field, old_value, new_value, batch_id)
        VALUES ('budget_sources', NEW.id, NEW.project_id, NEW.source, NEW.fund_type, NEW.data_year, 'cut_transfer', OLD.cut_transfer, NEW.cut_transfer, v_batch_id);
    END IF;
    IF OLD.under_budget IS DISTINCT FROM NEW.under_budget THEN
        INSERT INTO change_log(table_name, row_id, project_id, row_name, fund_type, data_year, field, old_value, new_value, batch_id)
        VALUES ('budget_sources', NEW.id, NEW.project_id, NEW.source, NEW.fund_type, NEW.data_year, 'under_budget', OLD.under_budget, NEW.under_budget, v_batch_id);
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
