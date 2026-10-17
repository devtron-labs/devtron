-- Supports CiArtifactRepositoryImpl.GetArtifactsByCDPipelineV3 for external-source
-- (webhook) CD pipelines, i.e. the "Select Image" listing built by
-- BuildQueryForParentTypeCIOrWebhook:
--
--   SELECT cia.* ... FROM ci_artifact cia
--   WHERE cia.external_ci_pipeline_id = ? ... ORDER BY cia.id DESC LIMIT ? OFFSET ?
--
-- With only the single-column idx_ci_artifact_external_ci_pipeline_id, the planner can
-- instead walk ci_artifact_pkey backwards (newest id first) and filter every row. It does
-- so whenever it over-estimates the pipeline's row count, which is common on large tables:
-- ANALYZE samples ~30k rows and most rows have external_ci_pipeline_id NULL, so per-pipeline
-- estimates are noisy. If the pipeline's artifacts are old, or the query also computes a
-- COUNT(*) OVER() window, that walk covers the whole table (69M rows in production,
-- 2+ minutes, 60s PG_READ_TIMEOUT -> 500 on Select Image).
--
-- This index returns one pipeline's artifacts already ordered by id, so it reads only the
-- matching rows AND gives the ORDER BY for free. That makes it cheaper than the backward
-- primary-key walk under any row estimate, so the plan no longer depends on fresh stats.
--
-- The old single-column index is intentionally kept: dropping it in this migration would
-- take an ACCESS EXCLUSIVE lock on ci_artifact during pod startup. It is small (partial on
-- NOT NULL) and can be dropped later with:
--
--   DROP INDEX CONCURRENTLY IF EXISTS idx_ci_artifact_external_ci_pipeline_id;
--
-- Note: CREATE INDEX CONCURRENTLY cannot run inside a transaction (golang-migrate
-- wraps each migration in a transaction). We use plain CREATE INDEX IF NOT EXISTS
-- here so the migration is idempotent. For large existing ci_artifact tables,
-- operators SHOULD create this index manually with CONCURRENTLY BEFORE deploying
-- this release, to avoid blocking writes during pod startup:
--
--   CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ci_artifact_external_ci_pipeline_id_id
--     ON ci_artifact (external_ci_pipeline_id, id) WHERE external_ci_pipeline_id IS NOT NULL;
--
-- Once the index exists, this migration is a no-op (IF NOT EXISTS).

CREATE INDEX IF NOT EXISTS idx_ci_artifact_external_ci_pipeline_id_id
  ON ci_artifact (external_ci_pipeline_id, id)
  WHERE external_ci_pipeline_id IS NOT NULL;
