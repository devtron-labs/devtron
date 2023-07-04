-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_plugin_stage_mapping;

-- Table Definition
CREATE TABLE  "public"."plugin_stage_mapping"
(
    "id"             int4    NOT NULL DEFAULT nextval('id_seq_plugin_stage_mapping'::regclass),
    "plugin_id"      int4    NOT NULL,
    "stage_type"    int4 NOT NULL,
    "created_on"     timestamptz,
    "created_by"     int4,
    "updated_on"     timestamptz,
    "updated_by"     int4,
    PRIMARY KEY ("id")
);

INSERT INTO "public"."plugin_stage_mapping" ("plugin_id","stage_type", "created_on", "created_by", "updated_on", "updated_by") VALUES
    (1,0,'now()', 1, 'now()', 1);

INSERT INTO "public"."plugin_stage_mapping" ("plugin_id","stage_type", "created_on", "created_by", "updated_on", "updated_by") VALUES
    (2,0,'now()', 1, 'now()', 1);

INSERT INTO "public"."plugin_stage_mapping" ("plugin_id","stage_type", "created_on", "created_by", "updated_on", "updated_by") VALUES
    (3,0,'now()', 1, 'now()', 1);

INSERT INTO "public"."plugin_stage_mapping" ("plugin_id","stage_type", "created_on", "created_by", "updated_on", "updated_by") VALUES
    (4,0,'now()', 1, 'now()', 1);

INSERT INTO "public"."plugin_stage_mapping" ("plugin_id","stage_type", "created_on", "created_by", "updated_on", "updated_by") VALUES
    (5,0,'now()', 1, 'now()', 1);

INSERT INTO "public"."plugin_stage_mapping" ("plugin_id","stage_type", "created_on", "created_by", "updated_on", "updated_by") VALUES
    (6,0,'now()', 1, 'now()', 1);

INSERT INTO "public"."plugin_stage_mapping" ("plugin_id","stage_type", "created_on", "created_by", "updated_on", "updated_by") VALUES
    (7,0,'now()', 1, 'now()', 1);

INSERT INTO "public"."plugin_stage_mapping" ("plugin_id","stage_type", "created_on", "created_by", "updated_on", "updated_by") VALUES
    (8,0,'now()', 1, 'now()', 1);

INSERT INTO "public"."plugin_stage_mapping" ("plugin_id","stage_type", "created_on", "created_by", "updated_on", "updated_by") VALUES
    (9,0,'now()', 1, 'now()', 1);

INSERT INTO "public"."plugin_stage_mapping" ("plugin_id","stage_type", "created_on", "created_by", "updated_on", "updated_by") VALUES
    (10,0,'now()', 1, 'now()', 1);