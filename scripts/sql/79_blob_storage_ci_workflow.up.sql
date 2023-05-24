ALTER TABLE ci_workflow
    ADD COLUMN IF NOT EXISTS  blob_storage_enabled boolean NOT NULL DEFAULT true;