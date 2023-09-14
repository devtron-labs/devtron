CREATE SEQUENCE IF NOT EXISTS variable_definition_seq;
CREATE TABLE IF NOT EXISTS "public"."variable_definition"
(
     "id"  integer not null default nextval('variable_definition_seq' :: regclass),
     "name"  varchar(300)     NOT NULL,
     "data_type"  varchar(50)     NOT NULL,
     "var_type"  varchar(50)     NOT NULL,
     "active" bool     NOT NULL,
     "description" text,
    "short_description" text,
     "created_on"        timestamptz,
     "created_by"        integer,
     "updated_on"        timestamptz,
     "updated_by"        integer,
     PRIMARY KEY ("id")

);

CREATE SEQUENCE IF NOT EXISTS variable_scope_seq;

CREATE TABLE IF NOT EXISTS "public"."variable_scope"
(
    "id" integer not null default nextval('variable_scope_seq' :: regclass),
    "variable_definition_id" int,
    "qualifier_id" int,
    "identifier_key" int,
    identifier_value_int int,
    identifier_value_string VARCHAR(255),
    "active" bool NOT NULL,
    "parent_identifier" int,
    "created_on"        timestamptz,
    "created_by"        integer,
    "updated_on"        timestamptz,
    "updated_by"        integer,
    CONSTRAINT "variable_scope_variable_id_fkey" FOREIGN KEY ("variable_definition_id") REFERENCES "public"."variable_definition" ("id"),
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS variable_data_seq;
CREATE TABLE  IF NOT EXISTS "public"."variable_data"
(
   "id" integer not null default nextval('variable_data_seq' :: regclass),
   "variable_scope_id" int,
   "data" text,
    "created_on"        timestamptz,
    "created_by"        integer,
    "updated_on"        timestamptz,
    "updated_by"        integer,
   CONSTRAINT "variable_data_variable_id_fkey" FOREIGN KEY ("variable_scope_id") REFERENCES "public"."variable_scope" ("id"),
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS variable_entity_mapping_seq;
CREATE TABLE IF NOT EXISTS "public"."variable_entity_mapping"
(
    "id"  integer NOT NULL default nextval('variable_entity_mapping_seq' :: regclass),
    "variable_name"  varchar(300)     NOT NULL,
    "entity_type"  int    NOT NULL,
    "entity_id"  int     NOT NULL,
    "is_deleted" bool     NOT NULL,
    "created_on"        timestamptz,
    "created_by"        integer,
    "updated_on"        timestamptz,
    "updated_by"        integer,
    PRIMARY KEY ("id")
    );

CREATE SEQUENCE IF NOT EXISTS variable_snapshot_history_seq;
CREATE TABLE IF NOT EXISTS "public"."variable_snapshot_history"
(
    "id"  integer NOT NULL default nextval('variable_snapshot_history_seq' :: regclass),
    "variable_snapshot"  jsonb     NOT NULL,
    "history_reference_type"  int     NOT NULL,
    "history_reference_id"  int     NOT NULL,
    "created_on"        timestamptz,
    "created_by"        integer,
    "updated_on"        timestamptz,
    "updated_by"        integer,
    PRIMARY KEY ("id")
    );


INSERT INTO devtron_resource_searchable_key(name, is_removed, created_on, created_by, updated_on, updated_by)
VALUES ('APP_ID', false, now(), 1, now(), 1),
       ('ENV_ID', false, now(), 1, now(), 1),
       ('CLUSTER_ID', false, now(), 1, now(), 1);
