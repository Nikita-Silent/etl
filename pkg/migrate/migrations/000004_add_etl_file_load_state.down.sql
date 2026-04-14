-- Migration: 000004_add_etl_file_load_state
-- Description: Drop durable logical file load metadata table

DROP TABLE IF EXISTS etl_file_load_state;
