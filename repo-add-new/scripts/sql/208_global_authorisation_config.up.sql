CREATE SEQUENCE IF NOT EXISTS id_seq_global_authorisation_config;

CREATE TABLE IF NOT EXISTS "public"."global_authorisation_config"
(
    "id"                int          NOT NULL DEFAULT nextval('id_seq_global_authorisation_config'::regclass),
    "config_type"       varchar(100) NOT NULL,
    "active"            boolean      NOT NULL,
    "created_on"        timestamptz,
    "created_by"        integer,
    "updated_on"        timestamptz,
    "updated_by"        integer,
    PRIMARY KEY ("id")
 );


INSERT into "public"."global_authorisation_config" (config_type,active,created_on,created_by,updated_on,updated_by)
VALUES ('devtron-system-managed',true,'now()',1,'now()',1);

CREATE SEQUENCE IF NOT EXISTS id_seq_user_groups;

CREATE TABLE "public"."user_groups"
(
    "id"                        integer          NOT NULL DEFAULT nextval('id_seq_user_groups'::regclass),
    "user_id"                   integer,
    "group_name"                text,
    "is_group_claims_data"      boolean NOT NULL,
    "active"                    boolean NOT NULL,
    "created_on"                timestamptz,
    "created_by"                integer,
    "updated_on"                timestamptz,
    "updated_by"                integer,
    PRIMARY KEY ("id")
);