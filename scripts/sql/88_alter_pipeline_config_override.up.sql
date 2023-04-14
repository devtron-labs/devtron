ALTER TABLE pipeline_config_override ADD COLUMN commit_time timestamptz;
UPDATE pipeline_config_override SET commit_time = updated_on;