ALTER TABLE sub_jobs ADD COLUMN IF NOT EXISTS data_year INT;
UPDATE sub_jobs SET data_year = (SELECT year FROM projects WHERE id = sub_jobs.project_id) WHERE data_year IS NULL;
ALTER TABLE sub_jobs ALTER COLUMN data_year SET NOT NULL;

ALTER TABLE budget_sources ADD COLUMN IF NOT EXISTS data_year INT;
UPDATE budget_sources SET data_year = (SELECT year FROM projects WHERE id = budget_sources.project_id) WHERE data_year IS NULL;
ALTER TABLE budget_sources ALTER COLUMN data_year SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_sub_jobs_data_year ON sub_jobs(data_year);
CREATE INDEX IF NOT EXISTS idx_budget_sources_data_year ON budget_sources(data_year);
