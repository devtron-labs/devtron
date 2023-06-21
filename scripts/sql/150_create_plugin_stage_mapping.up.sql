-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_plugin_stage_mapping;

-- Table Definition
CREATE TABLE  "public"."plugin_stage_mapping"
(
    "id"             int4    NOT NULL DEFAULT nextval('id_seq_plugin_stage_mapping'::regclass),
    "plugin_id"      int4    NOT NULL,
    "stage_type"    varchar(50) NOT NULL,
    "created_on"     timestamptz,
    "created_by"     integer,
    "updated_on"     timestamptz,
    "updated_by"     integer,
    PRIMARY KEY ("id"),
    CONSTRAINT fk_plugin_stage FOREIGN KEY(plugin_id) REFERENCES "public"."plugin_metadata"("id")
);
