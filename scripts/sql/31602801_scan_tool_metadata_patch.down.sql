ALTER TABLE public.plugin_parent_metadata DROP COLUMN IF EXISTS is_exposed;
ALTER TABLE public.plugin_metadata DROP COLUMN IF EXISTS is_exposed;
-- Plugin Id is added to scan_tool_metadata as foreign key
ALTER TABLE public.scan_tool_metadata DROP CONSTRAINT IF EXISTS scan_tool_metadata_plugin_id_fkey;
ALTER TABLE public.scan_tool_metadata DROP COLUMN IF EXISTS plugin_id;
ALTER TABLE public.scan_tool_metadata DROP COLUMN IF EXISTS is_preset;