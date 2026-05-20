-- required for ON CONFLICT upsert in import accept
CREATE UNIQUE INDEX IF NOT EXISTS sub_jobs_uq ON sub_jobs (project_id, name, fund_type, data_year);

CREATE TABLE po_import_status (
    project_code TEXT PRIMARY KEY,
    last_accepted_version INT,
    last_accepted_at TIMESTAMPTZ,
    po_version INT,
    po_updated_at TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'unknown'
);

CREATE TABLE po_import_log (
    id SERIAL PRIMARY KEY,
    project_code TEXT NOT NULL,
    po_version INT NOT NULL,
    accepted_by TEXT NOT NULL DEFAULT 'system',
    accepted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    snapshot_json JSONB NOT NULL
);

CREATE INDEX po_import_log_project_code_idx ON po_import_log (project_code);
CREATE INDEX po_import_log_accepted_at_idx ON po_import_log (accepted_at DESC);
