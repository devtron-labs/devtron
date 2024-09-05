ALTER TABLE notification_settings drop constraint notification_settings_env_id_fkey;
ALTER TABLE notification_settings ADD COLUMN cluster_id INT;
