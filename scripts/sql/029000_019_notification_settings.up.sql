ALTER TABLE notification_settings drop constraint IF EXISTS notification_settings_env_id_fkey;
ALTER TABLE notification_settings ADD COLUMN IF NOT EXISTS cluster_id INT;
