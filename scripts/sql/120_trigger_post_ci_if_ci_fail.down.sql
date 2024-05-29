/*
 * Copyright (c) 2024. Devtron Inc.
 */

---- drop trigger_if_parent_stage_fail column
ALTER TABLE pipeline_stage_step DROP COLUMN IF EXISTS trigger_if_parent_stage_fail;