ALTER TABLE pipeline_config_override ADD COLUMN commit_time timestamptz,
Update pipeline_config_override SET commit_time = updated_on;