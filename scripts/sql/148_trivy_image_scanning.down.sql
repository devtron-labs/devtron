ALTER TABLE public.module
    DROP COLUMN "enabled" ,
    DROP COLUMN "module_type";

DELETE from public.scan_tool_step CASCADE;
DELETE from public.registry_index_mapping CASCADE;
DELETE from public.scan_tool_metadata CASCADE;