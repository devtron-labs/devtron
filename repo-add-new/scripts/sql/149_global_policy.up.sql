CREATE SEQUENCE IF NOT EXISTS id_seq_devtron_resource_searchable_key;

CREATE TABLE IF NOT EXISTS devtron_resource_searchable_key
(
    "id"         integer     NOT NULL DEFAULT nextval('id_seq_devtron_resource_searchable_key'::regclass),
    "name"       varchar(100),
    "is_removed" bool,
    "created_on" timestamptz NOT NULL,
    "created_by" int4        NOT NULL,
    "updated_on" timestamptz NOT NULL,
    "updated_by" int4        NOT NULL,
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_global_policy;

CREATE TABLE IF NOT EXISTS global_policy
(
    "id"          integer     NOT NULL DEFAULT nextval('id_seq_global_policy'::regclass),
    "name"        varchar(200),
    "policy_of"   varchar(100),
    "version"     varchar(20),
    "policy_json" text,
    "description" text,
    "enabled"     bool,
    "deleted"     bool,
    "created_on"  timestamptz NOT NULL,
    "created_by"  int4        NOT NULL,
    "updated_on"  timestamptz NOT NULL,
    "updated_by"  int4        NOT NULL,
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_global_policy_searchable_field;

CREATE TABLE IF NOT EXISTS global_policy_searchable_field
(
    "id"                integer     NOT NULL DEFAULT nextval('id_seq_global_policy_searchable_field'::regclass),
    "global_policy_id"  integer,
    "searchable_key_id" integer,
    "value"             text,
    "is_regex"          bool        NOT NULL,
    "policy_component"  integer,
    "created_on"        timestamptz NOT NULL,
    "created_by"        int4        NOT NULL,
    "updated_on"        timestamptz NOT NULL,
    "updated_by"        int4        NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "global_policy_searchable_field_global_policy_id_fkey" FOREIGN KEY ("global_policy_id") REFERENCES "global_policy" ("id"),
    CONSTRAINT "global_policy_searchable_field_searchable_key_id_fkey" FOREIGN KEY ("searchable_key_id") REFERENCES "devtron_resource_searchable_key" ("id")
);

CREATE INDEX "gpsf_value_idx" ON global_policy_searchable_field USING BTREE ("value");


CREATE SEQUENCE IF NOT EXISTS id_seq_global_policy_history;

CREATE TABLE IF NOT EXISTS global_policy_history
(
    "id"                integer     NOT NULL DEFAULT nextval('id_seq_global_policy_history'::regclass),
    "global_policy_id"  integer,
    "history_of_action" varchar(50),
    "history_data"      jsonb,
    "policy_of"         varchar(100),
    "policy_version"    varchar(20),
    "policy_data"       text,
    "description"       text,
    "enabled"           bool,
    "created_on"        timestamptz NOT NULL,
    "created_by"        int4        NOT NULL,
    "updated_on"        timestamptz NOT NULL,
    "updated_by"        int4        NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "global_policy_history_global_policy_id_fkey" FOREIGN KEY ("global_policy_id") REFERENCES "global_policy" ("id")
);


INSERT INTO devtron_resource_searchable_key(name, is_removed, created_on, created_by, updated_on, updated_by)
VALUES ('PROJECT_APP_NAME', false, now(), 1, now(), 1),
       ('CLUSTER_ENV_NAME', false, now(), 1, now(), 1),
       ('IS_ALL_PRODUCTION_ENV', false, now(), 1, now(), 1),
       ('CI_PIPELINE_BRANCH', false, now(), 1, now(), 1),
       ('CI_PIPELINE_TRIGGER_ACTION', false, now(), 1, now(), 1);
