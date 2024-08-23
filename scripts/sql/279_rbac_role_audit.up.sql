CREATE SEQUENCE IF NOT EXISTS id_seq_rbac_role_audit;

CREATE TABLE IF NOT EXISTS "public"."rbac_role_audit"
(
    "id"                     integer     NOT NULL DEFAULT nextval('id_seq_rbac_role_audit'::regclass),
    "entity"                 varchar(250) NOT NULL,
    "access_type"            varchar(250) ,
    "role"                   varchar(250) NOT NULL,
    "policy_data"            jsonb,
    "role_data"              jsonb,
    "audit_operation"        varchar(20) NOT NULL,
    "created_on" timestamptz NOT NULL,
    "created_by" int4        NOT NULL,
    "updated_on" timestamptz NOT NULL,
    "updated_by" int4        NOT NULL,
    PRIMARY KEY ("id")
    );
