CREATE SEQUENCE IF NOT EXISTS id_seq_push_config;

CREATE TABLE IF NOT EXISTS manifest_push_config
(
    "id"         integer     NOT NULL DEFAULT nextval('id_seq_push_config'::regclass),
    "app_id"     integer,
    "env_id"     integer,
    "credentials_config" text,
    "chart_name"  varchar(100),
    "chart_base_version" varchar(100),
    "storage_type" varchar(100),
    "created_on" timestamptz NOT NULL,
    "created_by" int4        NOT NULL,
    "updated_on" timestamptz NOT NULL,
    "updated_by" int4        NOT NULL,
    "deleted"     bool,
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_oci_config;

CREATE TABLE IF NOT EXISTS oci_registry_config
(
    "id"         integer     NOT NULL DEFAULT nextval('id_seq_push_config'::regclass),
    "container_registry_id" integer,
    "repository_type" varchar(100),
    "repositor_action" varchar(100),
    "created_on" timestamptz NOT NULL,
    "created_by" int4        NOT NULL,
    "updated_on" timestamptz NOT NULL,
    "updated_by" int4        NOT NULL,
    "deleted"     bool,
    PRIMARY KEY ("id")
);

ALTER TABLE docker_artifact_store ADD is_oci_compliant_registry boolean