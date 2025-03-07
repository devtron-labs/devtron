BEGIN;

-- Add release_config column to deployment_config
ALTER TABLE deployment_config
    ADD COLUMN IF NOT EXISTS release_config jsonb;

-- Set active to false for all deployment_configs that have an app_id that is inactive
UPDATE deployment_config
    SET active = false
    WHERE active = true
    AND app_id IN (
        SELECT id FROM app WHERE active = false
    );

END;