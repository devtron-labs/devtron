BEGIN;

-- Drop Foreign Key Constraint: pipeline_stage_step_value_constraint_id_fkey
ALTER TABLE "public"."pipeline_stage_step_variable"
    DROP CONSTRAINT IF EXISTS pipeline_stage_step_value_constraint_id_fkey;

-- Drop Foreign Key Constraint: pipeline_stage_step_file_reference_id_fkey
ALTER TABLE "public"."pipeline_stage_step_variable"
    DROP CONSTRAINT IF EXISTS pipeline_stage_step_file_reference_id_fkey;

-- Drop Foreign Key Constraint: plugin_step_file_reference_id_fkey
ALTER TABLE "public"."plugin_step_variable"
    DROP CONSTRAINT IF EXISTS plugin_step_file_reference_id_fkey;

-- Drop Columns: value_constraint_id and is_runtime_arg from plugin_step_variable
ALTER TABLE "public"."pipeline_stage_step_variable"
    DROP COLUMN IF EXISTS "value_constraint_id",
    DROP COLUMN IF EXISTS "file_reference_id",
    DROP COLUMN IF EXISTS "file_mount_dir",
    DROP COLUMN IF EXISTS "is_runtime_arg";

-- Drop Columns: value_constraint_id and is_runtime_arg from plugin_step_variable
ALTER TABLE "public"."plugin_step_variable"
    DROP COLUMN IF EXISTS "file_reference_id",
    DROP COLUMN IF EXISTS "file_mount_dir";

-- Drop Table: value_constraint
DROP TABLE IF EXISTS "public"."value_constraint";

-- Drop Sequence: id_seq_value_constraint
DROP SEQUENCE IF EXISTS id_seq_value_constraint;

-- Drop Table: file_reference
DROP TABLE IF EXISTS "public"."file_reference";

-- Drop Sequence: id_seq_file_reference
DROP SEQUENCE IF EXISTS id_seq_file_reference;

COMMIT;