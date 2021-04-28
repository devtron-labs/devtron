-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS external_app_id_seq;

-- Table Definition
CREATE TABLE "public"."external_apps" (
    "id" int4 NOT NULL DEFAULT nextval('external_app_id_seq'::regclass),
    "app_name" varchar(250),
    "cluster_id" int4,
    "label" varchar(250),
    "chart_name" varchar(250),
    "namespace" varchar(250),
    "last_deployed_on" timestamptz,
    "created_on" timestamptz,
    "created_by" int4,
    "updated_by" int4,
    "updated_on" timestamptz,
    "active" bool DEFAULT false,
    "status" varchar(250),
    "deprecated" bool,
    "chart_version" varchar(250),
    CONSTRAINT "external_apps_cluster_id_fkey" FOREIGN KEY ("cluster_id") REFERENCES "public"."cluster"("id"),
    PRIMARY KEY ("id")
);