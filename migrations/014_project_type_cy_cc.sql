ALTER TABLE projects ALTER COLUMN project_type TYPE VARCHAR(2);
ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_project_type_check;
ALTER TABLE projects ADD CONSTRAINT projects_project_type_check CHECK (project_type IN ('Y', 'C', 'L', 'CY', 'CC'));
