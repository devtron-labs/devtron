ALTER TABLE public.image_scan_execution_history
    DROP COLUMN "scan_event_json" ,
    DROP COLUMN "execution_history_directory_path";

ALTER TABLE public.image_scan_execution_result
DROP COLUMN "scan_tool_id";

DROP TABLE scan_step_condition_mapping,
    scan_tool_step,
    scan_step_condition,
    scan_tool_execution_history_mapping,
    scan_tool_metadata,
    registry_index_mapping;
