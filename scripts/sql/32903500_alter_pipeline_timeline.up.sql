CREATE INDEX IF NOT EXISTS "idx_pipeline_status_timeline_cd_workflow_runner_id" ON pipeline_status_timeline USING BTREE ("cd_workflow_runner_id");

CREATE INDEX IF NOT EXISTS "idx_pipeline_status_timeline_installed_app_version_history_id" ON pipeline_status_timeline USING BTREE ("installed_app_version_history_id");
