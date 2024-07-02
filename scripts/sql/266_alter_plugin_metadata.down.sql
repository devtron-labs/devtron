ALTER TABLE plugin_metadata DROP COLUMN IF EXISTS plugin_parent_metadata_id;
ALTER TABLE plugin_metadata DROP COLUMN IF EXISTS plugin_version;
ALTER TABLE plugin_metadata DROP COLUMN IF EXISTS is_deprecated;
ALTER TABLE plugin_metadata DROP COLUMN IF EXISTS doc_link;
ALTER TABLE plugin_metadata DROP COLUMN IF EXISTS is_latest;

ALTER TABLE plugin_metadata DROP CONSTRAINT plugin_metadata_plugin_parent_metadata_id_fkey;