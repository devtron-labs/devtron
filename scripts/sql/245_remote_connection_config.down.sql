/*
 * Copyright (c) 2024. Devtron Inc.
 */

DROP TABLE remote_connection_config;
ALTER TABLE cluster DROP COLUMN remote_connection_config_id;
ALTER TABLE docker_artifact_store DROP COLUMN remote_connection_config_id;