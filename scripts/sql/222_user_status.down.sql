ALTER TABLE users DROP CONSTRAINT users_timeout_window_configuration_id_fkey;

ALTER TABLE users DROP COLUMN timeout_window_configuration_id;

DROP TABLE timeout_window_configuration;