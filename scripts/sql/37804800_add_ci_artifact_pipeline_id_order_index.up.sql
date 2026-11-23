-- Composite indexes on ci_artifact for lookups by pipeline_id, or by
-- (component_id, data_source), that also need rows ordered by id and filtered/
-- returned by image.
--
-- id is a key column (so these also serve as a sorted, index-only source of rows
-- for a given pipeline_id / component_id+data_source). image is carried via INCLUDE
-- (not a key column, so it doesn't affect sort order) so callers that also filter or
-- project on image can be served without a separate heap lookup.
--
-- The existing single-column idx_ci_artifact_pipeline_id and composite
-- idx_ci_artifact_component_id_data_source indexes are kept as-is and continue to
-- serve any existing usage; these are additive.
--
-- 1. ci_artifact.(pipeline_id, id) INCLUDE (image)
CREATE INDEX IF NOT EXISTS idx_ci_artifact_pipeline_id_id
  ON ci_artifact (pipeline_id, id) INCLUDE (image)
  WHERE pipeline_id IS NOT NULL;

-- 2. ci_artifact.(component_id, data_source, id) INCLUDE (image)
CREATE INDEX IF NOT EXISTS idx_ci_artifact_component_id_data_source_id
  ON ci_artifact (component_id, data_source, id) INCLUDE (image)
  WHERE component_id IS NOT NULL;

-- Note: CREATE INDEX CONCURRENTLY cannot run inside a transaction (golang-migrate
-- wraps each migration in a transaction). We use plain CREATE INDEX IF NOT EXISTS
-- here so the migration is idempotent. For large existing ci_artifact tables,
-- operators SHOULD create these indexes manually with CONCURRENTLY BEFORE
-- deploying this release, to avoid blocking writes during pod startup:
--
--   CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ci_artifact_pipeline_id_id
--     ON ci_artifact (pipeline_id, id) INCLUDE (image) WHERE pipeline_id IS NOT NULL;
--   CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_ci_artifact_component_id_data_source_id
--     ON ci_artifact (component_id, data_source, id) INCLUDE (image) WHERE component_id IS NOT NULL;
--
-- The old single-column indexes are intentionally kept: dropping them here would
-- take an ACCESS EXCLUSIVE lock during pod startup. They can be dropped later with:
--
--   DROP INDEX CONCURRENTLY IF EXISTS idx_ci_artifact_pipeline_id;
--   DROP INDEX CONCURRENTLY IF EXISTS idx_ci_artifact_component_id_data_source;
--
-- Once the indexes exist, this migration is a no-op (IF NOT EXISTS).
