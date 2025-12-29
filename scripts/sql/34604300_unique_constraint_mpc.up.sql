
-- Add unique constraint to prevent future duplicates
-- This ensures only one active manifest_push_config per (app_id, env_id) combination
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_manifest_push_config_app_env
    ON manifest_push_config (app_id, env_id)
    WHERE deleted = false;