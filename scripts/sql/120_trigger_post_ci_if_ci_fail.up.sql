---- add trigger_if_parent_stage_fail column
ALTER TABLE pipeline_stage_step ADD COLUMN IF NOT EXISTS trigger_if_parent_stage_fail bool;