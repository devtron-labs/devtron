-- Begin Transaction
BEGIN;
---------------------------------------
ALTER TABLE public.plugin_parent_metadata DROP COLUMN is_exposed;
ALTER TABLE public.plugin_metadata DROP COLUMN is_exposed ;
ALTER TABLE public.scan_tool_metadata DROP COLUMN is_preset ;
ALTER TABLE public.scan_tool_metadata DROP COLUMN plugin_id;
ALTER TABLE public.scan_tool_metadata DROP CONSTRAINT IF EXISTS scan_tool_metadata_name_version_unique;
ALTER TABLE public.scan_tool_metadata ADD COLUMN url;

-- ---------------------------------------------------
COMMIT;