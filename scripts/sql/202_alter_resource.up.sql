alter table devtron_resource
    add column description text;

alter table devtron_resource
    add column is_exposed bool not null default true;

update devtron_resource
set is_exposed= false
where kind = 'application';

update devtron_resource_schema
set is_exposed= false
where kind = 'cd-pipeline';


alter table devtron_resource_schema
    add column sample_schema json;

CREATE SEQUENCE IF NOT EXISTS id_seq_devtron_resource_schema_audit;

CREATE TABLE "public"."devtron_resource_schema_audit"
(
    "id"                         int         NOT NULL DEFAULT nextval('id_seq_devtron_resource_schema_audit'::regclass),
    "devtron_resource_schema_id" int,
    "schema"                     json,
    "audit_operation"            VARCHAR(10) NOT NULL,
    "created_on"                 timestamptz,
    "created_by"                 integer,
    "updated_on"                 timestamptz,
    "updated_by"                 integer,
    PRIMARY KEY ("id")
);