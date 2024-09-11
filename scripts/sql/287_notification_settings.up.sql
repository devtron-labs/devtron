ALTER TABLE notification_settings drop constraint IF EXISTS notification_settings_env_id_fkey;
ALTER TABLE notification_settings IF NOT EXISTS ADD COLUMN cluster_id INT;
