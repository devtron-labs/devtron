BEGIN;

-- Drop the column release_config from deployment_config
ALTER TABLE deployment_config
    DROP COLUMN IF EXISTS release_config;

END;