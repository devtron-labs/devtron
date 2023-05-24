DROP TABLE pipeline_status_timeline_resources CASCADE;

DROP SEQUENCE IF EXISTS id_seq_pipeline_status_timeline_resources;

DROP TABLE pipeline_status_timeline_sync_detail CASCADE;

DROP SEQUENCE IF EXISTS id_seq_pipeline_status_timeline_sync_detail;

ALTER TABLE pipeline DROP COLUMN deployment_app_name;

ALTER TABLE cd_workflow_runner DROP COLUMN created_on;

ALTER TABLE cd_workflow_runner DROP COLUMN created_by;

ALTER TABLE cd_workflow_runner DROP COLUMN updated_on;

ALTER TABLE cd_workflow_runner DROP COLUMN updated_by;