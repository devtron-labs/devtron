-- Begin Transaction
BEGIN;
-- Adding Exposed on plugin metadata and plugin parent metadata
ALTER TABLE public.plugin_parent_metadata ADD COLUMN is_exposed bool NOT NULL DEFAULT true;
ALTER TABLE public.plugin_metadata ADD COLUMN is_exposed bool NOT NULL DEFAULT true;

-- Preset flag is added to scan_tool_metadata to define tool added by user or devtron system
ALTER TABLE public.scan_tool_metadata ADD COLUMN is_preset bool NOT NULL DEFAULT true;
-- Plugin Id is added to scan_tool_metadata as foreign key
ALTER TABLE public.scan_tool_metadata ADD COLUMN plugin_id int;
ALTER TABLE "public"."scan_tool_metadata" ADD FOREIGN KEY ("plugin_id") REFERENCES "public"."plugin_metadata" ("id");
ALTER TABLE public.scan_tool_metadata ADD CONSTRAINT scan_tool_metadata_name_version_unique UNIQUE ("name", "version");

ALTER TABLE public.scan_tool_metadata ADD COLUMN url varchar(100);

UPDATE public.scan_tool_metadata SET url='https://cdn.devtron.ai/images/ic-clair.webp' WHERE name='CLAIR';
UPDATE public.scan_tool_metadata SET url='https://cdn.devtron.ai/images/ic-trivy.webp' WHERE name='TRIVY';

-- ---------------------------------------------------
-- Commit Transaction
-- ---------------------------------------------------
COMMIT;
