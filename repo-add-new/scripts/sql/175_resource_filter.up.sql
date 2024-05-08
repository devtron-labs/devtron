CREATE SEQUENCE IF NOT EXISTS resource_filter_seq;
CREATE TABLE IF NOT EXISTS "public"."resource_filter"
(
     "id"  integer not null default nextval('resource_filter_seq' :: regclass),
     "name"  varchar(300)     NOT NULL,
     "target_object"  integer     NOT NULL,
     "condition_expression"  text     NOT NULL,
     "deleted" bool     NOT NULL,
     "description" text,
     "created_on"        timestamptz,
     "created_by"        integer,
     "updated_on"        timestamptz,
     "updated_by"        integer,
     PRIMARY KEY ("id")
);

ALTER TABLE variable_scope DROP CONSTRAINT variable_scope_variable_id_fkey;

ALTER TABLE variable_data DROP CONSTRAINT variable_data_variable_id_fkey;

ALTER TABLE variable_scope RENAME TO resource_qualifier_mapping;

ALTER TABLE resource_qualifier_mapping RENAME COLUMN variable_definition_id TO resource_id;

ALTER TABLE resource_qualifier_mapping ADD COLUMN resource_type integer DEFAULT 0;

INSERT INTO devtron_resource_searchable_key(name, is_removed, created_on, created_by, updated_on, updated_by)
VALUES ('PROJECT_ID', false, now(), 1, now(), 1);