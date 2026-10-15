-- WS-G: add indexes for unindexed artifact_id lookups on release_tags/image_comments
-- to eliminate sequential scans on the following user-facing hot paths:
--   * ImageTaggingRepositoryImpl.GetTagsByArtifactId
--     (internal/sql/repository/imageTagging/ImageTaggingRepository.go)
--     — called from ImageTaggingService/ImageTaggingReadService when viewing/editing an
--     artifact's release tags.
--   * ImageTaggingRepositoryImpl.GetImageComment / GetImageCommentsByArtifactIds
--     (same file) — called from ImageTaggingService when rendering an artifact's comment
--     (build/deployment history rows).
--
-- Note: CREATE INDEX CONCURRENTLY cannot run inside a transaction (golang-migrate wraps
-- each migration in one). We use plain CREATE INDEX IF NOT EXISTS here so the migration is
-- idempotent, matching this repo's existing convention (see 36304600_add_ci_artifact_query_indexes).
-- For large existing tables, operators SHOULD create these indexes manually with
-- CONCURRENTLY BEFORE deploying this release, to avoid blocking writes during pod startup:
--
--   CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_release_tags_artifact_id
--     ON release_tags (artifact_id);
--   CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_image_comments_artifact_id
--     ON image_comments (artifact_id);
--
-- Once the indexes exist, this migration is a no-op (IF NOT EXISTS).

-- release_tags.artifact_id — the table's only existing index is UNIQUE(app_id, tag_name),
-- which doesn't help an artifact_id-only lookup.
CREATE INDEX IF NOT EXISTS idx_release_tags_artifact_id
  ON release_tags (artifact_id);

-- image_comments.artifact_id — no index beyond the primary key today.
CREATE INDEX IF NOT EXISTS idx_image_comments_artifact_id
  ON image_comments (artifact_id);
