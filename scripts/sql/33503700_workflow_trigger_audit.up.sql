BEGIN;

-- Create Sequence for workflow_config_snapshot
CREATE SEQUENCE IF NOT EXISTS id_seq_workflow_config_snapshot;

CREATE TABLE IF NOT EXISTS "public"."workflow_config_snapshot" (
    "id"                    int4             NOT NULL DEFAULT nextval('id_seq_workflow_config_snapshot'::regclass),
    "workflow_id"           int4             NOT NULL, -- ci_workflow.id or cd_workflow_runner.id
    "workflow_type"         varchar(20)     NOT NULL, -- CI, CD
    "pipeline_id"           int4             NOT NULL,
    "artifact_id"           int4,            -- ci_artifact.id
    "trigger_type"          varchar(20)     NOT NULL, -- MANUAL, AUTO, WEBHOOK
    "workflow_request_json" text            NOT NULL, -- complete WorkflowRequest JSON (contains everything)
    "workflow_request_schema_version"        varchar(20)             NOT NULL DEFAULT 'V1', -- for backward compatibility
    "created_on"            timestamptz     NOT NULL,
    "created_by"            integer            NOT NULL,
    PRIMARY KEY ("id"),
    UNIQUE ("workflow_id", "workflow_type")
);

COMMIT;
