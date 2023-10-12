BEGIN;
-- create resource filter audit table
CREATE SEQUENCE IF NOT EXISTS resource_filter_audit_seq;
CREATE TABLE IF NOT EXISTS "public"."resource_filter_audit"
(
    "id"  integer not null default nextval('resource_filter_audit_seq' :: regclass),
    "target_object"  integer     NOT NULL,
    "conditions"  text     NOT NULL,
    "filter_id"   int      NOT NULL,
    "action"      int      NOT NULL,
    "created_on"        timestamptz,
    "created_by"        integer,
    "updated_on"        timestamptz,
    "updated_by"        integer,
    CONSTRAINT "resource_filter_audit_filter_id_fkey" FOREIGN KEY ("filter_id") REFERENCES "public"."resource_filter" ("id"),
    PRIMARY KEY ("id")
);

-- create resource filter evaluation audit table
CREATE SEQUENCE IF NOT EXISTS resource_filter_evaluation_audit_seq;
CREATE TABLE IF NOT EXISTS "public"."resource_filter_evaluation_audit"
(
    "id"  integer not null default nextval('resource_filter_evaluation_audit_seq' :: regclass),
    "reference_type"           integer   NOT NULL,
    "reference_id"             integer   NOT NULL,
    "filter_history_objects"   text      NOT NULL,
    "subject_type"             int       NOT NULL,
    "subject_ids"              text      NOT NULL,
    "created_on"               timestamptz,
    "created_by"               integer,
    "updated_on"               timestamptz,
    "updated_by"               integer,
    PRIMARY KEY ("id")
);
COMMIT;