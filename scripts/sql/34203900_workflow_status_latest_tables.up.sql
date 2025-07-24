BEGIN;

-- Create Sequence for ci_workflow_status_latest
CREATE SEQUENCE IF NOT EXISTS id_seq_ci_workflow_status_latest;

-- Create ci_workflow_status_latest table
CREATE TABLE IF NOT EXISTS "public"."ci_workflow_status_latest" (
    "id"                    int4             NOT NULL DEFAULT nextval('id_seq_ci_workflow_status_latest'::regclass),
    "pipeline_id"           int4             NOT NULL,
    "app_id"                int4             NOT NULL,
    "ci_workflow_id"        int4             NOT NULL,
    "created_on"            timestamptz      NOT NULL,
    "created_by"            int4             NOT NULL,
    "updated_on"            timestamptz      NOT NULL,
    "updated_by"            int4             NOT NULL,
    PRIMARY KEY ("id"),
    UNIQUE ("ci_workflow_id")
);

-- Create Sequence for cd_workflow_status_latest
CREATE SEQUENCE IF NOT EXISTS id_seq_cd_workflow_status_latest;

-- Create cd_workflow_status_latest table
CREATE TABLE IF NOT EXISTS "public"."cd_workflow_status_latest" (
    "id"                    int4             NOT NULL DEFAULT nextval('id_seq_cd_workflow_status_latest'::regclass),
    "pipeline_id"           int4             NOT NULL,
    "app_id"                int4             NOT NULL,
    "environment_id"        int4             NOT NULL,
    "workflow_type"         varchar(20)      NOT NULL, -- PRE, DEPLOY, POST
    "workflow_runner_id"    int4             NOT NULL,
    "created_on"            timestamptz      NOT NULL,
    "created_by"            int4             NOT NULL,
    "updated_on"            timestamptz      NOT NULL,
    "updated_by"            int4             NOT NULL,
    PRIMARY KEY ("id"),
    UNIQUE ("workflow_runner_id", "workflow_type")
);

COMMIT;
