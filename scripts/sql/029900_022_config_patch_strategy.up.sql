ALTER TABLE chart_env_config_override ADD COLUMN merge_strategy VARCHAR(100);
ALTER TABLE deployment_template_history ADD COLUMN merge_strategy VARCHAR(100);
ALTER TABLE deployment_template_history ADD COLUMN  template_patch_data TEXT ;