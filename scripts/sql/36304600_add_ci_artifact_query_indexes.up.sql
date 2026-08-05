-- Hotfix: add indexes for ci_artifact / ci_workflow / workflow_execution_stage /
-- deployment_approval_request to eliminate sequential scans that were causing
-- 60s PG_READ_TIMEOUT errors on the following hot queries:
--   * CiArtifactRepositoryImpl.GetArtifactsByCDPipelineV3 (external CI + UNION variants)
--   * CiArtifactRepositoryImpl.FindArtifactByListFilter
--   * CiWorkflowRepositoryImpl.FindByPipelineId
--   * CiWorkflowRepositoryImpl.FindByStatusesIn
--   * WorkflowStageRepositoryImpl.GetSuccessfulCIExecutionStages
--
-- Note: CREATE INDEX CONCURRENTLY cannot run inside a transaction (golang-migrate
-- wraps each migration in a transaction). We use plain CREATE INDEX IF NOT EXISTS
-- here so the migration is idempotent. For large existing tables, operators
-- SHOULD create these indexes manually with CONCURRENTLY BEFORE deploying this
-- release to avoid blocking writes during pod startup:
--
--   CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ci_artifact_external_ci_pipeline_id
--     ON ci_artifact (external_ci_pipeline_id) WHERE external_ci_pipeline_id IS NOT NULL;
--   CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ci_artifact_pipeline_id
--     ON ci_artifact (pipeline_id) WHERE pipeline_id IS NOT NULL;
--   CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ci_artifact_component_id_data_source
--     ON ci_artifact (component_id, data_source) WHERE component_id IS NOT NULL;
--   CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ci_workflow_ci_pipeline_id
--     ON ci_workflow (ci_pipeline_id);
--   CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_dar_pipeline_id_active
--     ON deployment_approval_request (pipeline_id) WHERE active = true;
--   CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_wes_ci_execution_lookup
--     ON workflow_execution_stage (created_on)
--     WHERE workflow_type = 'CI' AND stage_name = 'Execution'
--       AND status = 'SUCCEEDED' AND status_for = 'workflow';
--
-- Once the indexes exist, this migration is a no-op (IF NOT EXISTS).

-- 1. ci_artifact.external_ci_pipeline_id
--    Filter used by GetArtifactsByCDPipelineV3 (external webhook CI variant).
--    Partial index: most rows are internal CI (external_ci_pipeline_id IS NULL),
--    so a partial index is much smaller and covers the actual query.
CREATE INDEX IF NOT EXISTS idx_ci_artifact_external_ci_pipeline_id
  ON ci_artifact (external_ci_pipeline_id)
  WHERE external_ci_pipeline_id IS NOT NULL;

-- 2. ci_artifact.pipeline_id
--    Filter used by GetArtifactsByCDPipelineV3 (UNION variant) and several other
--    repository methods (Where pipeline_id = ?).
CREATE INDEX IF NOT EXISTS idx_ci_artifact_pipeline_id
  ON ci_artifact (pipeline_id)
  WHERE pipeline_id IS NOT NULL;

-- 3. ci_artifact.(component_id, data_source)
--    Composite filter used by GetArtifactsByCDPipelineV3 (UNION variant) and
--    FindArtifactByListFilter: (component_id = ? AND data_source = 'pre_cd'|'post_ci').
CREATE INDEX IF NOT EXISTS idx_ci_artifact_component_id_data_source
  ON ci_artifact (component_id, data_source)
  WHERE component_id IS NOT NULL;

-- 4. ci_workflow.ci_pipeline_id
--    Filter used by FindByPipelineId (WHERE ci_pipeline_id = ? ORDER BY started_on DESC).
CREATE INDEX IF NOT EXISTS idx_ci_workflow_ci_pipeline_id
  ON ci_workflow (ci_pipeline_id);

-- 5. deployment_approval_request.pipeline_id (partial on active = true)
--    Filter used by FindArtifactByListFilter NOT IN subquery
--    (WHERE pipeline_id = ? AND active = true AND artifact_deployment_triggered = false).
--    Partial index on active = true keeps it small.
CREATE INDEX IF NOT EXISTS idx_dar_pipeline_id_active
  ON deployment_approval_request (pipeline_id)
  WHERE active = true;

-- 6. workflow_execution_stage covering the exact GetSuccessfulCIExecutionStages filter
--    (workflow_type='CI' AND stage_name='Execution' AND status='SUCCEEDED'
--     AND status_for='workflow' AND created_on BETWEEN ? AND ?).
--    Highly selective partial index shaped to that query.
CREATE INDEX IF NOT EXISTS idx_wes_ci_execution_lookup
  ON workflow_execution_stage (created_on)
  WHERE workflow_type = 'CI'
    AND stage_name = 'Execution'
    AND status = 'SUCCEEDED'
    AND status_for = 'workflow';
