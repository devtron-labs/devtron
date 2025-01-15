ALTER TABLE public.plugin_parent_metadata ADD COLUMN IF NOT EXISTS is_exposed BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE public.plugin_metadata ADD COLUMN IF NOT EXISTS is_exposed BOOLEAN NOT NULL DEFAULT TRUE;
-- Plugin Id is added to scan_tool_metadata as foreign key
ALTER TABLE public.scan_tool_metadata ADD COLUMN IF NOT EXISTS plugin_id int;
ALTER TABLE public.scan_tool_metadata ADD COLUMN IF NOT EXISTS is_preset int;
ALTER TABLE public.scan_tool_metadata ADD FOREIGN KEY ("plugin_id") REFERENCES "public"."plugin_metadata" ("id");
ALTER TABLE public.scan_tool_metadata ADD CONSTRAINT scan_tool_metadata_name_version_unique UNIQUE ("name", "version");