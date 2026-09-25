-- Supports CiWorkflowRepositoryImpl.FindByStatusesIn, called every 5 minutes by
-- the CI workflow status cron (UpdateCiWorkflowStatusFailure) to find builds
-- stuck in Starting/Running so they can be timed out. With no index on status,
-- this does a full sequential scan of ci_workflow (tens of millions of rows in
-- production) on every run.
--
-- Partial index scoped to only the two statuses this query filters on, so it
-- stays small (proportional to active build count, not total table size) and
-- cheap to maintain on INSERT/UPDATE, regardless of how large ci_workflow grows.
--
-- Note: CREATE INDEX CONCURRENTLY cannot run inside a transaction (golang-migrate
-- wraps each migration in a transaction). We use plain CREATE INDEX IF NOT EXISTS
-- here so the migration is idempotent. For large existing ci_workflow tables,
-- operators SHOULD create this index manually with CONCURRENTLY BEFORE deploying
-- this release, to avoid blocking writes during pod startup (~2 minutes to build
-- on a table with 40 million rows, benchmarked locally):
--
--   CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ci_workflow_active_status
--     ON ci_workflow (status) WHERE status IN ('Starting', 'Running');
--
-- Once the index exists, this migration is a no-op (IF NOT EXISTS).

CREATE INDEX IF NOT EXISTS idx_ci_workflow_active_status
  ON ci_workflow (status) WHERE status IN ('Starting', 'Running');
