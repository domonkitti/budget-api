ALTER TABLE scenario_sub_jobs
    ADD COLUMN IF NOT EXISTS cut_transfer NUMERIC(15,3) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS under_budget NUMERIC(15,3) NOT NULL DEFAULT 0;

UPDATE scenario_sub_jobs ssj
SET cut_transfer = sj.cut_transfer,
    under_budget = sj.under_budget
FROM scenario_sub_jobs ssj2
JOIN sub_jobs sj ON sj.project_id = ssj2.project_id
                AND sj.name       = ssj2.name
                AND sj.fund_type  = ssj2.fund_type
                AND sj.data_year  = ssj2.data_year
WHERE ssj.id = ssj2.id;
