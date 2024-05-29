/*
 * Copyright (c) 2024. Devtron Inc.
 */

ALTER TABLE ci_workflow
    ADD COLUMN IF NOT EXISTS  blob_storage_enabled boolean NOT NULL DEFAULT true;