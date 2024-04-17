
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


INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'Vulnerabilty_Scanner v1.0.0' , 'Checks code vulnerability types in the Post-CI stage','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/devtron-logo-plugin.png',false,'now()',1,'now()',1);


INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)VALUES (nextval('id_seq_plugin_stage_mapping'),
                                                                                                              (SELECT id from plugin_metadata where name='Vulnerabilty_Scanner v1.0.0'),1,'now()',1,'now()',1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES ( nextval('id_seq_plugin_pipeline_script'),'
         E'#!/bin/bash

json_data="$CI_CD_EVENT"
base_url="$IMAGE_SCANNER_ENDPOINT"


url="$base_url/scanner/image"

ciProjectDetails=$(echo "$json_data" | jq -r \'.commonWorkflowRequest.ciProjectDetails\')
ciWorkflowId=$(echo "$json_data" | jq -r \'.workflowId\')
sourceType=2
sourceSubType=1


new_payload=$(cat <<EOF
{
  "ciProjectDetails": $ciProjectDetails,
  "ciWorkflowId" : $ciWorkflowId,
  "sourceType" : $sourceType,
  "sourceSubType" : $sourceSubType

}
EOF
)


response=$(curl -s -X POST -H "Content-Type: application/json" -d "$new_payload" "$url")

  export LOW=-1
  export MEDIUM=-1
  export HIGH=-1
  export CRITICAL=-1
  export UNKNOWN=-1


if [[ $(echo "$response" | jq -r \'.status\') == "OK" ]]; then
    # Extract severity values from the response JSON and replace null with zero
    LOW=$(echo "$response" | jq -r \'.result.codeScanResponse.misConfigurations.list[0].summary.severities.LOW // 0\')
    MEDIUM=$(echo "$response" | jq -r \'.result.codeScanResponse.misConfigurations.list[0].summary.severities.MEDIUM // 0\')
             HIGH=$(echo "$response" | jq -r \'.result.codeScanResponse.misConfigurations.list[0].summary.severities.HIGH  // 0\')
             CRITICAL=$(echo "$response" | jq -r \'.result.codeScanResponse.misConfigurations.list[0].summary.severities.CRITICAL // 0\')
             UNKNOWN=$(echo "$response" | jq -r \'.result.codeScanResponse.misConfigurations.list[0].summary.severities.UNKNOWN  // 0\')
             else
             echo "Response not OK: $response"
             fi


             echo "LOW = $LOW"
             echo "MEDIUM = $MEDIUM"
             echo "HIGH = $HIGH"
             echo "CRITICAL = $CRITICAL"
             echo "UNKNOWN = $UNKNOWN"',

    'SHELL',
    'f',
    'now()',
    1,
    'now()',
    1
);'



INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Vulnerabilty_Scanner v1.0.0'),'Step 1','Step 1 - Vulnerabilty_Scanner v1.0.0','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);
INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by)
VALUES
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Vulnerabilty_Scanner v1.0.0' and ps."index"=1 and ps.deleted=false),'LOW','NUMBER','Number of LOW vulnerability,','t','f',null,null,'OUTPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Vulnerabilty_Scanner v1.0.0' and ps."index"=1 and ps.deleted=false),'MEDIUM','NUMBER','Number of MEDIUM vulnerability,','t','f',null,null,'OUTPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Vulnerabilty_Scanner v1.0.0' and ps."index"=1 and ps.deleted=false),'HIGH','NUMBER','Number of HIGH vulnerability,','t','f',null,null,'OUTPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Vulnerabilty_Scanner v1.0.0' and ps."index"=1 and ps.deleted=false),'CRITICAL','NUMBER','Number of CRITICAL vulnerability,','t','f',null,null,'OUTPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Vulnerabilty_Scanner v1.0.0' and ps."index"=1 and ps.deleted=false),'UNKNOWN','NUMBER','Number of UNKNOWN vulnerability,','t','f',null,null,'OUTPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

