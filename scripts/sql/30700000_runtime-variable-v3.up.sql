BEGIN;

-- Create Sequence for value_constraint
CREATE SEQUENCE IF NOT EXISTS id_seq_value_constraint;

-- Table Definition: value_constraint
CREATE TABLE IF NOT EXISTS "public"."value_constraint" (
    "id"                    int         NOT NULL DEFAULT nextval('id_seq_value_constraint'::regclass),
    "choices"               text[],
    "value_of"              VARCHAR(50) NOT NULL,
    "block_custom_value"    boolean     NOT NULL DEFAULT FALSE,
    "deleted"               boolean     NOT NULL DEFAULT FALSE,
    "constraint"            jsonb       NOT NULL,
    "created_on"            timestamptz NOT NULL,
    "created_by"            int4        NOT NULL,
    "updated_on"            timestamptz NOT NULL,
    "updated_by"            int4        NOT NULL,
    PRIMARY KEY ("id")
);

-- Create Sequence for value_constraint
CREATE SEQUENCE IF NOT EXISTS id_seq_file_reference;

-- Table Definition: file_reference
CREATE TABLE IF NOT EXISTS "public"."file_reference" (
    "id"                    int             NOT NULL DEFAULT nextval('id_seq_file_reference'::regclass),
    "data"                  bytea,
    "name"                  VARCHAR(255)    NOT NULL,
    "size"                  bigint          NOT NULL,
    "mime_type"             VARCHAR(255)    NOT NULL,
    "extension"             VARCHAR(50)     NOT NULL,
    "file_type"             VARCHAR(50)     NOT NULL,
    "created_on"            timestamptz     NOT NULL,
    "created_by"            int4            NOT NULL,
    "updated_on"            timestamptz     NOT NULL,
    "updated_by"            int4            NOT NULL,
    PRIMARY KEY ("id")
);

-- Alter Table: pipeline_stage_step_variable; Add column value_constraint_id and is_runtime_arg
ALTER TABLE "public"."pipeline_stage_step_variable"
    ADD COLUMN IF NOT EXISTS "value_constraint_id"      int,
    ADD COLUMN IF NOT EXISTS "file_reference_id"        int,
    ADD COLUMN IF NOT EXISTS "file_mount_dir"           VARCHAR(255),
    ADD COLUMN IF NOT EXISTS "is_runtime_arg" boolean   NOT NULL DEFAULT FALSE;

-- Drop Foreign Key Constraint: pipeline_stage_step_value_constraint_id_fkey if exists; to obtain idempotency
ALTER TABLE "public"."pipeline_stage_step_variable"
    DROP CONSTRAINT IF EXISTS pipeline_stage_step_value_constraint_id_fkey;

-- Drop Foreign Key Constraint: pipeline_stage_step_file_reference_id_fkey if exists; to obtain idempotency
ALTER TABLE "public"."pipeline_stage_step_variable"
    DROP CONSTRAINT IF EXISTS pipeline_stage_step_file_reference_id_fkey;

-- Add Foreign Key Constraint: pipeline_stage_step_value_constraint_id_fkey
ALTER TABLE "public"."pipeline_stage_step_variable"
    ADD CONSTRAINT "pipeline_stage_step_value_constraint_id_fkey" FOREIGN KEY ("value_constraint_id") REFERENCES "public"."value_constraint" ("id");

-- Add Foreign Key Constraint: pipeline_stage_step_file_reference_id_fkey
ALTER TABLE "public"."pipeline_stage_step_variable"
    ADD CONSTRAINT "pipeline_stage_step_file_reference_id_fkey" FOREIGN KEY ("file_reference_id") REFERENCES "public"."file_reference" ("id");

-- Alter Table: plugin_step_variable; Add column value_constraint_id and is_runtime_arg
ALTER TABLE "public"."plugin_step_variable"
    ADD COLUMN IF NOT EXISTS "file_reference_id"        int,
    ADD COLUMN IF NOT EXISTS "file_mount_dir"           VARCHAR(255);

-- Drop Foreign Key Constraint: plugin_step_file_reference_id_fkey if exists; to obtain idempotency
ALTER TABLE "public"."plugin_step_variable"
    DROP CONSTRAINT IF EXISTS plugin_step_file_reference_id_fkey;

-- Add Foreign Key Constraint: plugin_step_file_reference_id_fkey
ALTER TABLE "public"."plugin_step_variable"
    ADD CONSTRAINT "plugin_step_file_reference_id_fkey" FOREIGN KEY ("file_reference_id") REFERENCES "public"."file_reference" ("id");

COMMIT;