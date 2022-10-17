CREATE SEQUENCE IF NOT EXISTS id_seq_terminal_access_templates;

-- Table Definition
CREATE TABLE "public"."terminal_access_templates"
(
    "id"            integer NOT NULL DEFAULT nextval('id_seq_terminal_access_templates'::regclass),
    "template_name" VARCHAR(1000),
    "template_kind" VARCHAR(1000),
    "template_data" text,
    "created_on"    timestamptz,
    "created_by"    int4,
    "updated_on"    timestamptz,
    "updated_by"    int4,
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_user_terminal_access_data;

-- Table Definition
CREATE TABLE "public"."user_terminal_access_data"
(
    "id"         integer NOT NULL DEFAULT nextval('id_seq_user_terminal_access_data'::regclass),
    "user_id"    int4,
    "cluster_id" integer,
    "pod_name"   VARCHAR(1000),
    "node_name"  VARCHAR(1000),
    "status"     VARCHAR(1000),
    "metadata"   text,
    "created_on" timestamptz,
    "created_by" int4,
    "updated_on" timestamptz,
    "updated_by" int4,
    PRIMARY KEY ("id")
);