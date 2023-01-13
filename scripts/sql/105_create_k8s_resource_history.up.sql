CREATE SEQUENCE IF NOT EXISTS id_seq_k8s_resource_history_sequence;

-- Table Definition
CREATE TABLE IF NOT EXISTS "public"."kubernetes_resource_history"
(
    "id"            integer NOT NULL DEFAULT nextval('id_seq_k8s_resource_history_sequence'::regclass),
    "app_id"  integer,
    "app_name" VARCHAR(100),
    "env_id"  integer,
    "namespace"  VARCHAR(100) ,
    "resource_name" VARCHAR(100),
    "kind"    VARCHAR(100),
    "group"    VARCHAR(100),
    "force_delete"   boolean,
    "action_type"   VARCHAR(100),
    "deployment_app_type"  VARCHAR(100),
    "created_on"    timestamptz,
    "created_by"    int4,
    "updated_on"    timestamptz,
    "updated_by"    int4,
    PRIMARY KEY ("id")
    );