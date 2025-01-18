-- Begin Transaction
BEGIN;
---------------------------------------
ALTER TABLE public.plugin_parent_metadata DROP COLUMN IF EXISTS is_exposed;
ALTER TABLE public.plugin_metadata DROP COLUMN IF EXISTS is_exposed ;
ALTER TABLE public.scan_tool_metadata DROP COLUMN IF EXISTS is_preset ;
ALTER TABLE public.scan_tool_metadata DROP COLUMN IF EXISTS plugin_id;
ALTER TABLE public.scan_tool_metadata DROP CONSTRAINT IF EXISTS scan_tool_metadata_name_version_unique;
ALTER TABLE public.scan_tool_metadata ADD COLUMN IF EXISTS url;

-- ---------------------------------------------------
COMMIT;