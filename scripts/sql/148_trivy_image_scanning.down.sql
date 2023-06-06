ALTER TABLE public.module
    DROP COLUMN "enabled" ,
    DROP COLUMN "module_type";

DROP TABLE public.scan_tool_execution_history_mapping CASCADE;
DROP TABLE public.scan_tool_step CASCADE;
DROP TABLE public.registry_index_mapping CASCADE;
DROP TABLE public.scan_tool_metadata CASCADE;