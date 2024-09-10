ALTER TABLE plugin_metadata DROP COLUMN IF EXISTS plugin_parent_metadata_id;
ALTER TABLE plugin_metadata DROP COLUMN IF EXISTS plugin_version;
ALTER TABLE plugin_metadata DROP COLUMN IF EXISTS is_deprecated;
ALTER TABLE plugin_metadata DROP COLUMN IF EXISTS doc_link;
ALTER TABLE plugin_metadata DROP COLUMN IF EXISTS is_latest;

---- DROP table
DROP TABLE IF EXISTS "plugin_parent_metadata";

---- DROP sequence
DROP SEQUENCE IF EXISTS public.id_seq_plugin_parent_metadata;