ALTER TABLE public.image_scan_execution_history
    DROP COLUMN "scan_event_json" ,
    DROP COLUMN "execution_history_directory_path";

ALTER TABLE public.image_scan_execution_result
DROP COLUMN "scan_tool_id";

DROP TABLE IF EXISTS scan_step_condition_mapping;
DROP TABLE IF EXISTS scan_tool_step;
DROP TABLE IF EXISTS scan_step_condition;
DROP TABLE IF EXISTS scan_tool_execution_history_mapping;
DROP TABLE IF EXISTS scan_tool_metadata;
DROP TABLE IF EXISTS registry_index_mapping;
