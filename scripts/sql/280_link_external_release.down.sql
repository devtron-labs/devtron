DROP  INDEX IF EXISTS unique_deployment_app_name;

ALTER TABLE deployment_config DROP COLUMN IF EXISTS release_mode;
