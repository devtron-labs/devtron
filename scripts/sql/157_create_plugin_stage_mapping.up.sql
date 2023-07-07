-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_plugin_stage_mapping;

-- Table Definition
CREATE TABLE  "public"."plugin_stage_mapping"
(
    "id"             int4    NOT NULL DEFAULT nextval('id_seq_plugin_stage_mapping'::regclass),
    "plugin_id"      int4    NOT NULL,
    "stage_type"     int4 ,
    "created_on"     timestamptz,
    "created_by"     int4,
    "updated_on"     timestamptz,
    "updated_by"     int4,
    PRIMARY KEY ("id")
);

INSERT INTO "public"."plugin_stage_mapping" ("plugin_id", "created_on","created_by","updated_on","updated_by")
SELECT "id", "created_on","created_by","updated_on","updated_by"
FROM "public"."plugin_metadata";

UPDATE "public"."plugin_stage_mapping" set "stage_type" = 0;