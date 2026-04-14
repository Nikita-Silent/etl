CREATE TABLE etl_operation_runs (
    operation_id TEXT PRIMARY KEY,
    request_id TEXT,
    operation_type TEXT NOT NULL,
    status TEXT NOT NULL,
    date TEXT,
    source_folder TEXT,
    component TEXT NOT NULL,
    instance_id TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    finished_at TIMESTAMPTZ,
    error_message TEXT,
    failed_stage TEXT,
    timeout_report_sent BOOLEAN NOT NULL DEFAULT FALSE,
    crash_suspected BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_etl_operation_runs_status_updated_at
    ON etl_operation_runs (status, updated_at);

CREATE INDEX idx_etl_operation_runs_operation_type_started_at
    ON etl_operation_runs (operation_type, started_at DESC);

CREATE INDEX idx_etl_operation_runs_date
    ON etl_operation_runs (date);

CREATE INDEX idx_etl_operation_runs_source_folder
    ON etl_operation_runs (source_folder);
