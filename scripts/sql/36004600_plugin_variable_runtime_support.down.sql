BEGIN;

ALTER TABLE "public"."plugin_step_variable"
    DROP CONSTRAINT IF EXISTS plugin_step_variable_value_constraint_id_fkey;

ALTER TABLE "public"."plugin_step_variable"
    DROP COLUMN IF EXISTS "is_runtime_arg",
    DROP COLUMN IF EXISTS "value_constraint_id";

COMMIT;