INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'Devtron CD Trigger v1.0.0','Helps in Deployment of CD pipeline in PRE CD and POST CD steps','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/sonarqube-plugin-icon.png',false,'now()',1,'now()',1);


INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)VALUES (nextval('id_seq_plugin_stage_mapping'),
(SELECT id from plugin_metadata where name='Devtron CD Trigger v1.0.0'), 0,'now()',1,'now()',1);


INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES ( nextval('id_seq_plugin_pipeline_script'),
E'#!/bin/sh

cdWorkflowType="DEPLOY"
devtron_endpoint=$(echo "${input_devtron_endpoint}" | sed \'s:/*$::\')

fetch_app_id() {
    local api_response=$(curl -s -H "Cookie: argocd.token=$api_token" "${devtron_endpoint}/orchestrator/app/autocomplete")
    local app_id=$(echo "$api_response" | jq -r --arg app_name "$user_input_app_name" \'.result[] | select(.name == $app_name) | .id\')
    
    if [[ -z "$app_id" || "$app_id" == "null" ]]; then
        echo "Error: Application "$user_input_app_name" not found."
        exit 1
    fi
    echo "$app_id"
}

fetch_pipeline_id() {
    local app_id=$1
    local api_response=$(curl -s -H "Cookie: argocd.token=$api_token" "${devtron_endpoint}/orchestrator/app/app-wf/view/$app_id")
    local pipeline_id=$(echo "$api_response" | jq -r --arg env_name "$user_input_env_name" \'.result.cdConfig.pipelines[] | select(.environmentName == $env_name) | .id\')
    
    if [[ -z "$pipeline_id" || "$pipeline_id" == "null" ]]; then
        echo "Error: Pipeline for Environment "$user_input_env_name" not found in Application with ID "$app_id"."
        exit 1
    fi
    echo "$pipeline_id"
}

fetch_ci_artifact_id() {
    local pipelineId=$1
    local apiUrl="${devtron_endpoint}/orchestrator/app/cd-pipeline/${pipelineId}/material?offset=0&size=20&stage=$cdWorkflowType"
    local apiResponse=$(curl -s -H "Cookie: argocd.token=$api_token" "$apiUrl")
    local ciArtifactId=""

    if [[ -n "$source_commit_hash" ]]; then
        ciArtifactId=$(echo "$apiResponse" | jq -r --arg hash "$source_commit_hash" \'.result.ci_artifacts[] | select(.material_info[].revision == $hash) | .id\')
    fi

    if [[ -z "$ciArtifactId" || "$ciArtifactId" == "null" ]]; then
        ciArtifactId=$(echo "$apiResponse" | jq -r \'.result.ci_artifacts[0].id\')
    fi

    if [[ -z "$ciArtifactId" ]]; then
        echo "Error: CI Artifact ID not found."
        exit 1
    fi

    echo "$ciArtifactId"
}

trigger_cd_pipeline() {
    local pipelineId=$1
    local appId=$2
    local ciArtifactId=$3
    local jsonPayload=$(jq -n \\
                            --arg pipelineId "$pipelineId" \\
                            --arg appId "$appId" \\
                            --arg ciArtifactId "$ciArtifactId" \\
                            --arg cdWorkflowType "$cdWorkflowType" \\
                            --arg deploymentWithConfig "LAST_SAVED_CONFIG" \\
                            \'{
                                pipelineId: ($pipelineId | tonumber),
                                appId: ($appId | tonumber),
                                ciArtifactId: ($ciArtifactId | tonumber),
                                cdWorkflowType: $cdWorkflowType,
                                deploymentWithConfig: $deploymentWithConfig
                            }\')

    curl -X POST "${devtron_endpoint}/orchestrator/app/cd-pipeline/trigger" \\
        -H "Content-Type: application/json" \\
        -H "Cookie: argocd.token=$api_token" \\
        --data "$jsonPayload" \\
        --compressed
}

check_deploy_status() {
    local appId=$1
    local pipelineId=$2
    local max_wait=$3
    local start_time=$(date +%s)

    while :; do
        local current_time=$(date +%s)
        local elapsed_time=$((current_time - start_time))

        if [ "$elapsed_time" -ge "$max_wait" ]; then
            echo "Timeout reached without success. Exiting..."
            exit 1
        fi

        local statusUrl="${devtron_endpoint}/orchestrator/app/workflow/status/$appId/v2"
        local response=$(curl -s -H "Cookie: argocd.token=$api_token" "$statusUrl")
        local code=$(echo "$response" | jq -r \'.code\')
        
        if [ "$code" != "200" ]; then
            echo "Error: Received response - $response. Exiting..."
            exit 1
        fi

        local status=$(echo "$response" | jq -r --arg pipelineId "$pipelineId" \'.result.cdWorkflowStatus[] | select(.pipeline_id == ($pipelineId | tonumber)) | .deploy_status\')
        echo "Current deploy status: $status"
        
        if [[ "$status" == "Succeeded" ]]; then
            echo "Deployment succeeded."
            break
        elif [[ "$status" == "Failed" ]]; then
            echo "Deployment failed."
            exit 1
        fi

        sleep 15
    done
}

app_id=$(fetch_app_id)
pipeline_id=$(fetch_pipeline_id "$app_id")
ciArtifactId=$(fetch_ci_artifact_id "$pipeline_id")

echo "Triggering CD Pipeline for App ID: $app_id, Pipeline ID: $pipeline_id, CI Artifact ID: $ciArtifactId, and CD Workflow Type: $cdWorkflowType"
trigger_cd_pipeline "$pipeline_id" "$app_id" "$ciArtifactId" "$cdWorkflowType"

check_deploy_status "$app_id" "$pipeline_id" "$max_wait"'

,
    'SHELL',
    'f',
    'now()',
    1,
    'now()',
    1
);
INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Devtron CD Trigger v1.0.0'),'Step 1','Step 1 - Devtron CD Trigger v1.0.0','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES 
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CD Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'api_token','STRING','API Token','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CD Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'input_devtron_endpoint','STRING','URL of Devtron','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CD Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'user_input_app_name','STRING','Name of the Application that user wants to deploy','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CD Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'user_input_env_name','STRING','Name of the Environment in which user wants to deploy the Application','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CD Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'max_wait','STRING','Maximum time for user can wait for the application to deploy.','t','f',false,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CD Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'source_commit_hash','STRING','Commit hash.','t','t',false,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);