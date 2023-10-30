CREATE SEQUENCE IF NOT EXISTS id_seq_devtron_resource;

CREATE TABLE "public"."devtron_resource"
(
    "id"             int          NOT NULL DEFAULT nextval('id_seq_devtron_resource'::regclass),
    "kind"           varchar(250) NOT NULL,
    "parent_kind_id" int,
    "deleted"        boolean,
    "created_on"     timestamptz,
    "created_by"     integer,
    "updated_on"     timestamptz,
    "updated_by"     integer,
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_devtron_resource_schema;

CREATE TABLE "public"."devtron_resource_schema"
(
    "id"                  int        NOT NULL DEFAULT nextval('id_seq_devtron_resource_schema'::regclass),
    "devtron_resource_id" int,
    "version"             VARCHAR(5) NOT NULL,
    "schema"              jsonb,
    "latest"              boolean,
    "created_on"          timestamptz,
    "created_by"          integer,
    "updated_on"          timestamptz,
    "updated_by"          integer,
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_devtron_resource_object;

CREATE TABLE "public"."devtron_resource_object"
(
    "id"                         int          NOT NULL DEFAULT nextval('id_seq_devtron_resource_object'::regclass),
    "name"                       VARCHAR(250) NOT NULL,
    "devtron_resource_id"        int,
    "devtron_resource_schema_id" int,
    "schema"                     jsonb,
    "deleted"                    boolean,
    "created_on"                 timestamptz,
    "created_by"                 integer,
    "updated_on"                 timestamptz,
    "updated_by"                 integer,
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_devtron_resource_object_audit;

CREATE TABLE "public"."devtron_resource_object_audit"
(
    "id"                         int         NOT NULL DEFAULT nextval('id_seq_devtron_resource_object_audit'::regclass),
    "devtron_resource_object_id" int,
    "schema"                     jsonb,
    "audit_operation"            VARCHAR(10) NOT NULL,
    "created_on"                 timestamptz,
    "created_by"                 integer,
    "updated_on"                 timestamptz,
    "updated_by"                 integer,
    PRIMARY KEY ("id")
);


