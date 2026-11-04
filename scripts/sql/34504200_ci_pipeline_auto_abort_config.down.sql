-- Remove auto-abort configuration from ci_pipeline
DROP INDEX IF EXISTS idx_ci_pipeline_auto_abort;
ALTER TABLE ci_pipeline DROP COLUMN IF EXISTS auto_abort_previous_builds;