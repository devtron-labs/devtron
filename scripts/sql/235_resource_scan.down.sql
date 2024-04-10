ALTER TABLE public.scan_tool_execution_history_mapping DROP COLUMN error_message;

DELETE FROM public.scan_tool_step
WHERE scan_tool_id = 3
  AND index = 5;
