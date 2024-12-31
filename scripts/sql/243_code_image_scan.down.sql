
DROP TABLE IF EXISTS "public"."resource_scan_execution_result";

DROP SEQUENCE IF EXISTS resource_scan_execution_result_id_seq;

ALTER TABLE public.image_scan_execution_history DROP column IF EXISTS source_type;
ALTER TABLE public.image_scan_execution_history DROP column IF EXISTS source_sub_type;
ALTER TABLE public.image_scan_execution_history RENAME COLUMN source_metadata_json TO scan_event_json ;

UPDATE scan_tool_step
SET cli_command = 'trivy image -f json -o {{.OUTPUT_FILE_PATH}} --timeout {{.timeout}} {{.IMAGE_NAME}} --username {{.USERNAME}} --password {{.PASSWORD}}'
WHERE scan_tool_id=3 and  index=1 and step_execution_type='CLI';
UPDATE scan_tool_step
SET cli_command = '(export AWS_ACCESS_KEY_ID={{.AWS_ACCESS_KEY_ID}} AWS_SECRET_ACCESS_KEY={{.AWS_SECRET_ACCESS_KEY}} AWS_DEFAULT_REGION={{.AWS_DEFAULT_REGION}}; trivy image -f json -o {{.OUTPUT_FILE_PATH}} --timeout {{.timeout}} {{.IMAGE_NAME}})'
WHERE scan_tool_id=3 and index=2 and step_execution_type='CLI';
UPDATE scan_tool_step
SET cli_command = 'GOOGLE_APPLICATION_CREDENTIALS="{{.FILE_PATH}}/credentials.json" trivy image -f json -o {{.OUTPUT_FILE_PATH}} --timeout {{.timeout}} {{.IMAGE_NAME}}'
WHERE scan_tool_id=3 and index=3 and step_execution_type='CLI';
UPDATE scan_tool_step
SET cli_command = 'trivy image -f json -o {{.OUTPUT_FILE_PATH}} --timeout {{.timeout}} {{.IMAGE_NAME}}'
WHERE scan_tool_id=3 and index=5 and step_execution_type='CLI';