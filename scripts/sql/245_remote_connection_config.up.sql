/*
 * Copyright (c) 2024. Devtron Inc.
 */

ALTER TABLE cluster ADD COLUMN remote_connection_config_id INT;
ALTER TABLE docker_artifact_store ADD COLUMN remote_connection_config_id INT;

CREATE SEQUENCE IF NOT EXISTS id_seq_remote_connection_config;
CREATE TABLE IF NOT EXISTS public.remote_connection_config
(
    "id"                           int          NOT NULL DEFAULT nextval('id_seq_remote_connection_config'::regclass),
    "connection_method"
    VARCHAR(50)  NOT NULL,
    "proxy_url"                    VARCHAR(300)  ,
    "ssh_server_address"           VARCHAR(300),
    "ssh_username"                 VARCHAR(300),
    "ssh_password"                 text,
    "ssh_auth_key"                 text,
    "deleted"                      bool         NOT NULL,
    "created_on"                   timestamptz  NOT NULL,
    "created_by"                   int4         NOT NULL,
    "updated_on"                   timestamptz  NOT NULL,
    "updated_by"                   int4         NOT NULL,
    PRIMARY KEY ("id")
);

ALTER TABLE cluster
    ADD CONSTRAINT fk_cluster_remote_connection_config
        FOREIGN KEY (remote_connection_config_id)
            REFERENCES remote_connection_config (id);

ALTER TABLE docker_artifact_store
    ADD CONSTRAINT fk_docker_artifact_store_remote_connection_config
        FOREIGN KEY (remote_connection_config_id)
            REFERENCES remote_connection_config (id);

