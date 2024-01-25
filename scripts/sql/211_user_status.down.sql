ALTER TABLE users DROP COLUMN timeout_window_configuration_id;

ALTER TABLE DROP FOREIGN KEY timeout_window_configuration_id;

ALTER TABLE timeout_window_configuration;