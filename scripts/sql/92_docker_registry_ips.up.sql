-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_docker_registry_ips_config;

-- Table Definition
CREATE TABLE "public"."docker_registry_ips_config"
(
    "id"                       int4         NOT NULL DEFAULT nextval('id_seq_docker_registry_ips_config'::regclass),
    "docker_artifact_store_id" varchar(250) NOT NULL,
    "credential_type"          varchar(50)  NOT NULL,
    "credential_value"         text,
    "applied_cluster_ids_csv"  varchar(256),
    "ignored_cluster_ids_csv"  varchar(256),
    PRIMARY KEY ("id"),
    UNIQUE("docker_artifact_store_id")
);

-- add foreign key
ALTER TABLE "public"."docker_registry_ips_config"
    ADD FOREIGN KEY ("docker_artifact_store_id") REFERENCES "public"."docker_artifact_store" ("id");

-- insert values
INSERT INTO docker_registry_ips_config (docker_artifact_store_id, credential_type, ignored_cluster_ids_csv)
SELECT id, 'SAME_AS_REGISTRY', '-1' from docker_artifact_store;