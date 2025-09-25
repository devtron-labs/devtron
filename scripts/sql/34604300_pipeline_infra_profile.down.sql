-- Rollback migration for pipeline-level infra profile support

BEGIN;

-- Drop the foreign key constraint
ALTER TABLE ci_pipeline DROP CONSTRAINT IF EXISTS fk_ci_pipeline_infra_profile;

-- Drop the infra_profile_id column
ALTER TABLE ci_pipeline DROP COLUMN IF EXISTS infra_profile_id;

COMMIT;