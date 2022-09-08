ALTER TABLE cd_workflow_runner
    ADD COLUMN  blob_storage_enabled boolean NOT NULL DEFAULT true;


ALTER TABLE ci_workflow
    ADD COLUMN  blob_storage_enabled boolean NOT NULL DEFAULT true;