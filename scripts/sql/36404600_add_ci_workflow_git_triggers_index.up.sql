-- Supports CiWorkflowRepositoryImpl.FindWorkflowByMaterialAndCommit, used by
-- validateBuildSequence to check a pipeline's full build history (not just the
-- last build) for whether a given material/commit combination has already been
-- built, before rejecting a stale/redelivered automatic CI trigger.
--
-- git_triggers is stored as `json`, not `jsonb`, so a plain GIN index can't be
-- created directly on the column. We index the jsonb-cast expression instead,
-- using jsonb_path_ops (smaller and faster than the default jsonb_ops for
-- containment-only (@>) queries, which is the only operator this lookup uses).
--
-- Note: CREATE INDEX CONCURRENTLY cannot run inside a transaction (golang-migrate
-- wraps each migration in a transaction). We use plain CREATE INDEX IF NOT EXISTS
-- here so the migration is idempotent. For large existing ci_workflow tables,
-- operators SHOULD create this index manually with CONCURRENTLY BEFORE deploying
-- this release, to avoid blocking writes during pod startup:
--
--   CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ci_workflow_git_triggers_gin
--     ON ci_workflow USING gin ((git_triggers::jsonb) jsonb_path_ops);
--
-- Once the index exists, this migration is a no-op (IF NOT EXISTS).

CREATE INDEX IF NOT EXISTS idx_ci_workflow_git_triggers_gin
  ON ci_workflow USING gin ((git_triggers::jsonb) jsonb_path_ops);
