/*
 * Copyright (c) 2024. Devtron Inc.
 */

ALTER TABLE cd_workflow_runner ADD COLUMN IF NOT EXISTS "reference_id" VARCHAR(50) NULL;