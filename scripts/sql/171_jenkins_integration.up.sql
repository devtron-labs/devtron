CREATE SEQUENCE IF NOT EXISTS id_seq_external_tool_metadata;

CREATE TABLE "public"."external_tool_metadata"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_external_tool_metadata'::regclass),
    "name"                        text,
    "description"                 text,
    "icon"                        text,
    "deleted"                     bool,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    PRIMARY KEY ("id")
);

INSERT INTO "public"."external_tool_metadata" ("id", "name", "description","icon","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
('1', 'Jenkins','K6 is an open-source tool and cloud service that makes load testing easy for developers and QA engineers.','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/k6-plugin-icon.png','f', 'now()', '1', 'now()', '1');

ALTER TABLE "plugin_metadata" ADD COLUMN "external_tool_id" int4;
ALTER TABLE "plugin_metadata" ADD COLUMN  "is_external_tool_configuration" bool;

ALTER TABLE "ci_pipeline" ADD COLUMN "is_task" bool;


