-- Rollback: drop indexes added by 37504600_add_artifact_tagging_indexes.up.sql

DROP INDEX IF EXISTS idx_release_tags_artifact_id;
DROP INDEX IF EXISTS idx_image_comments_artifact_id;
