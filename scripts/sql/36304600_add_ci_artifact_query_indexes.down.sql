-- Rollback: drop indexes added by 36204600_add_ci_artifact_query_indexes.up.sql

DROP INDEX IF EXISTS idx_ci_artifact_external_ci_pipeline_id;
DROP INDEX IF EXISTS idx_ci_artifact_pipeline_id;
DROP INDEX IF EXISTS idx_ci_artifact_component_id_data_source;
DROP INDEX IF EXISTS idx_ci_workflow_ci_pipeline_id;
DROP INDEX IF EXISTS idx_dar_pipeline_id_active;
DROP INDEX IF EXISTS idx_wes_ci_execution_lookup;
