ALTER TABLE reports ADD COLUMN IF NOT EXISTS sort_order INT NOT NULL DEFAULT 0;

-- Backfill existing rows with their current implicit order (by id, within each group).
WITH ranked AS (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY group_id ORDER BY id) - 1 AS rn
    FROM reports
)
UPDATE reports SET sort_order = ranked.rn
FROM ranked
WHERE reports.id = ranked.id;
