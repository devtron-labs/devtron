CREATE SEQUENCE IF NOT EXISTS id_seq_app_group;

-- Table Definition
CREATE TABLE "public"."app_group"
(
    "id"             int4         NOT NULL DEFAULT nextval('id_seq_app_group'::regclass),
    "name"           varchar(250) NOT NULL,
    "description"    varchar(50),
    "environment_id" int4         NOT NULL,
    "active"         bool         NOT NULL,
    "created_on"     timestamptz,
    "created_by"     integer,
    "updated_on"     timestamptz,
    "updated_by"     integer,
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_app_group_mapping;

-- Table Definition
CREATE TABLE "public"."app_group_mapping"
(
    "id"           int4 NOT NULL DEFAULT nextval('id_seq_app_group_mapping'::regclass),
    "app_group_id" int4 NOT NULL,
    "app_id"       int4 NOT NULL,
    "created_on"   timestamptz,
    "created_by"   integer,
    "updated_on"   timestamptz,
    "updated_by"   integer,
    PRIMARY KEY ("id")
);

ALTER TABLE "public"."app_group_mapping"
    ADD FOREIGN KEY ("app_group_id") REFERENCES "public"."app_group" ("id");
ALTER TABLE "public"."app_group_mapping"
    ADD FOREIGN KEY ("app_id") REFERENCES "public"."app" ("id");