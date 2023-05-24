ALTER TABLE charts ADD COLUMN is_basic_view_locked bool NOT NULL DEFAULT FALSE;

ALTER TABLE charts ADD COLUMN current_view_editor text DEFAULT 'UNDEFINED';

ALTER TABLE chart_env_config_override ADD COLUMN is_basic_view_locked bool NOT NULL DEFAULT FALSE;

ALTER TABLE chart_env_config_override ADD COLUMN current_view_editor text DEFAULT 'UNDEFINED';
