CREATE SEQUENCE IF NOT EXISTS id_seq_deployment_config;

CREATE TABLE IF NOT EXISTS "public"."deployment_config"
(
    "id"                                int NOT NULL DEFAULT nextval('id_seq_deployment_config'::regclass),
    "app_id"                            int,
    "environment_id"                    int,
    "deployment_app_type"               VARCHAR(50),
    "config_type"                       VARCHAR(50),
    "repo_url"                          VARCHAR(50),
    "repo_name"                         VARCHAR(50),
    "chart_location"                    VARCHAR(50),
    "credential_type"                   VARCHAR(50),
    "credential_id_int"                 int,
    "credential_id_string"              VARCHAR(50),
    "active"                            bool,
    "created_on"                        timestamptz,
    "created_by"                        integer,
    "updated_on"                        timestamptz,
    "updated_by"                        integer,
    PRIMARY KEY ("id")
);