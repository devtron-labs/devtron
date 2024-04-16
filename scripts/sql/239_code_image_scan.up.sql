
CREATE SEQUENCE IF NOT EXISTS public.resource_scan_execution_result_id_seq;

CREATE table if not exists public.resource_scan_execution_result (
    id integer DEFAULT nextval('public.resource_scan_execution_result_id_seq'::regclass) NOT NULL,
    image_scan_execution_history_id integer NOT NULL,
    scan_data_json text,
    format integer,
    types integer[],
    scan_tool_id int,
    PRIMARY KEY ("id"),
    CONSTRAINT image_scan_execution_history_id_fkey
    FOREIGN KEY("image_scan_execution_history_id")
    REFERENCES"public"."image_scan_execution_history" ("id")
    );

ALTER TABLE public.image_scan_execution_history ADD column IF NOT exists source_type integer NULL;
ALTER TABLE public.image_scan_execution_history ADD column IF NOT exists source_sub_type integer NULL;
ALTER TABLE public.image_scan_execution_history RENAME COLUMN if exists scan_event_json TO source_metadata_json;


UPDATE scan_tool_step
SET cli_command = 'trivy image -f json -o {{.OUTPUT_FILE_PATH}} --timeout {{.timeout}} {{.IMAGE_NAME}} --username {{.USERNAME}} --password {{.PASSWORD}} {{.EXTRA_ARGS}}'
WHERE scan_tool_id=3,index=1,step_execution_type='CLI';
UPDATE scan_tool_step
SET cli_command = '(export AWS_ACCESS_KEY_ID={{.AWS_ACCESS_KEY_ID}} AWS_SECRET_ACCESS_KEY={{.AWS_SECRET_ACCESS_KEY}} AWS_DEFAULT_REGION={{.AWS_DEFAULT_REGION}}; trivy image -f json -o {{.OUTPUT_FILE_PATH}} --timeout {{.timeout}} {{.IMAGE_NAME}} {{.EXTRA_ARGS}})'
WHERE scan_tool_id=3,index=2,step_execution_type='CLI';
UPDATE scan_tool_step
SET cli_command = 'GOOGLE_APPLICATION_CREDENTIALS="{{.FILE_PATH}}/credentials.json" trivy image -f json -o {{.OUTPUT_FILE_PATH}} --timeout {{.timeout}} {{.IMAGE_NAME}} {{.EXTRA_ARGS}}'
WHERE scan_tool_id=3,index=3,step_execution_type='CLI';
UPDATE scan_tool_step
SET cli_command = 'trivy image -f json -o {{.OUTPUT_FILE_PATH}} --timeout {{.timeout}} {{.IMAGE_NAME}} {{.EXTRA_ARGS}}'
WHERE scan_tool_id=3,index=5,step_execution_type='CLI';


