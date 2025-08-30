-- Add configuration for auto-abort previous builds feature
ALTER TABLE ci_pipeline 
ADD COLUMN IF NOT EXISTS auto_abort_previous_builds BOOLEAN DEFAULT FALSE;

-- Add index for performance when querying by pipeline id and auto abort setting
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ci_pipeline_auto_abort 
ON ci_pipeline (id, auto_abort_previous_builds) 
WHERE auto_abort_previous_builds = TRUE;