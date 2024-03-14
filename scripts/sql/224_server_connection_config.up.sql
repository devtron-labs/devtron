ALTER TABLE cluster ADD COLUMN server_connection_config_id INT;
ALTER TABLE docker_artifact_store ADD COLUMN server_connection_config_id INT;

CREATE SEQUENCE IF NOT EXISTS id_seq_server_connection_config;
CREATE TABLE IF NOT EXISTS public.server_connection_config
(
    "id"                           int          NOT NULL DEFAULT nextval('id_seq_server_connection_config'::regclass),
    "connection_method"            VARCHAR(50)  NOT NULL,
    "proxy_url"                    VARCHAR(300)  ,
    "ssh_server_address"           VARCHAR(300),
    "ssh_username"                 VARCHAR(300),
    "ssh_password"                 VARCHAR(300),
    "ssh_auth_key"                 VARCHAR(300),
    "deleted"                      bool         NOT NULL,
    "created_on"                   timestamptz  NOT NULL,
    "created_by"                   int4         NOT NULL,
    "updated_on"                   timestamptz  NOT NULL,
    "updated_by"                   int4         NOT NULL,
    PRIMARY KEY ("id")
);

ALTER TABLE cluster
    ADD CONSTRAINT fk_cluster_server_connection_config
        FOREIGN KEY (server_connection_config_id)
            REFERENCES server_connection_config (id);

ALTER TABLE docker_artifact_store
    ADD CONSTRAINT fk_docker_artifact_store_server_connection_config
        FOREIGN KEY (server_connection_config_id)
            REFERENCES server_connection_config (id);

