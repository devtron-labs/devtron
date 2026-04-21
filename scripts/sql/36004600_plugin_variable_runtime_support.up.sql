BEGIN;

-- Add is_runtime_arg and value_constraint_id to plugin_step_variable
-- These were intended in migration 30802500 (comment says so) but were missed
ALTER TABLE "public"."plugin_step_variable"
    ADD COLUMN IF NOT EXISTS "is_runtime_arg"      boolean NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS "value_constraint_id"  int;

-- Add FK to value_constraint table
ALTER TABLE "public"."plugin_step_variable"
    ADD CONSTRAINT "plugin_step_variable_value_constraint_id_fkey"
    FOREIGN KEY ("value_constraint_id") REFERENCES "public"."value_constraint" ("id");

COMMIT;