DELETE FROM change_log cl
WHERE cl.field NOT IN ('budget', 'target', 'cut_transfer', 'under_budget')
   OR (cl.field = 'budget' AND cl.fund_type = 'ผูกพัน');

WITH ranked AS (
    SELECT cl.id,
           row_number() OVER (
             PARTITION BY cl.project_id
             ORDER BY cl.changed_at DESC, cl.id DESC
           ) AS rn
    FROM change_log cl
    WHERE cl.field IN ('budget', 'target', 'cut_transfer', 'under_budget')
      AND NOT (cl.field = 'budget' AND cl.fund_type = 'ผูกพัน')
)
DELETE FROM change_log cl
USING ranked r
WHERE cl.id = r.id
  AND r.rn > 20;
