CREATE SEQUENCE IF NOT EXISTS id_seq_push_config;

CREATE TABLE IF NOT EXISTS manifest_push_config
(
    "id"         integer     NOT NULL DEFAULT nextval('id_seq_push_config'::regclass),
    "app_id"     integer,
    "env_id"     integer,
    "credentials_config" text,
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
    "id"         integer     NOT NULL DEFAULT nextval('id_seq_oci_config'::regclass),
    "docker_artifact_store_id" varchar(250) NOT NULL,
    "repository_type" varchar(100),
    "repository_action" varchar(100),
    "created_on" timestamptz NOT NULL,
    "created_by" int4        NOT NULL,
    "updated_on" timestamptz NOT NULL,
    "updated_by" int4        NOT NULL,
    "deleted"    bool,
    CONSTRAINT oci_registry_config_docker_artifact_store_id_fkey
        FOREIGN KEY(docker_artifact_store_id)
        REFERENCES public.docker_artifact_store(id),
    PRIMARY KEY ("id")
);

-- Adding a CHECK constraint to ensure UNIQUE(container_registry_id, repository_type) if delete=false
CREATE UNIQUE INDEX idx_unique_oci_registry_config
    ON oci_registry_config (docker_artifact_store_id, repository_type)
    WHERE oci_registry_config.deleted = false;

ALTER TABLE docker_artifact_store ADD is_oci_compliant_registry boolean;