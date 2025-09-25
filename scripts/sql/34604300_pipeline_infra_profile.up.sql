-- Migration to add pipeline-level infra profile support to ci_pipeline table
-- This enables selecting different build infrastructure per pipeline

BEGIN;

-- Add infra_profile_id column to ci_pipeline table
ALTER TABLE ci_pipeline 
ADD COLUMN IF NOT EXISTS infra_profile_id INTEGER;

-- Add foreign key constraint to infra_profile table
ALTER TABLE ci_pipeline 
ADD CONSTRAINT fk_ci_pipeline_infra_profile 
FOREIGN KEY (infra_profile_id) REFERENCES infra_profile(id);

-- Add comment to explain the column
COMMENT ON COLUMN ci_pipeline.infra_profile_id IS 'Optional pipeline-level infra profile override. If null, falls back to application-level profile.';

COMMIT;