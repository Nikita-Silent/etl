-- Migration: 000004_add_etl_file_load_state
-- Description: Track successfully committed logical file loads atomically with tx_* writes

CREATE TABLE etl_file_load_state (
  logical_key TEXT PRIMARY KEY,
  remote_path TEXT NOT NULL,
  requested_date DATE,
  source_folder TEXT NOT NULL,
  content_hash TEXT NOT NULL,
  transaction_manifest JSONB,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX etl_file_load_state_source_folder_idx
  ON etl_file_load_state (source_folder);

CREATE INDEX etl_file_load_state_requested_date_idx
  ON etl_file_load_state (requested_date);
