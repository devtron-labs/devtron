INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'The plugin enables users to trigger pre/post/deployment of application. It helps users deploy the application that contains dependencies for their current application.','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/devtron-logo.png',false,'now()',1,'now()',1);


INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)VALUES (nextval('id_seq_plugin_stage_mapping'),
(SELECT id from plugin_metadata where name='Devtron CD Trigger v1.0.0'), 0,'now()',1,'now()',1);


INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES ( nextval('id_seq_plugin_pipeline_script'),
E'#!/bin/bash

# Convert CD Workflow Type to uppercase to make it case-insensitive
TargetTriggerStage=$(echo "$TargetTriggerStage" | tr \'[:lower:]\' \'[:upper:]\')

# Set default value for TargetTriggerStage to DEPLOY if not provided or invalid
case $TargetTriggerStage in
"PRE")
    ;;
"POST")
    ;;
"DEPLOY")
    ;;
"")
    TargetTriggerStage="DEPLOY" # Set to DEPLOY if no input provided
    ;;
*)
    echo "Error: Invalid CD Workflow Type. Please enter PRE, DEPLOY, or POST. Exiting..."
    exit 1
    ;;
esac

# Set default value for StatusTimeOutSeconds to 0 if not provided or not an integer
if ! [[ "$StatusTimeOutSeconds" =~ ^[0-9]+$ ]]; then
    StatusTimeOutSeconds=0
fi

DevtronEndpoint=$(echo "$DevtronEndpoint" | sed \'s:/*$::\')

# Determine sleep interval based on StatusTimeOutSeconds
if [ "$StatusTimeOutSeconds" -lt "60" ]; then
    sleepInterval=$(($StatusTimeOutSeconds / 2))
else
    sleepInterval=30
fi

fetch_app_id() {
    # Check if DevtronApp is numeric, if yes, use it directly as App ID
    if [[ "$DevtronApp" =~ ^[0-9]+$ ]]; then
        echo "$DevtronApp"
    else
        local api_response=$(curl -s -H "Cookie: argocd.token=$DevtronApiToken" "$DevtronEndpoint/orchestrator/app/autocomplete")
        local app_id=$(echo "$api_response" | jq -r --arg app_name "$DevtronApp" \'.result[] | select(.name == $app_name) | .id\')

        if [[ -z "$app_id" || "$app_id" == "null" ]]; then
            echo "Error: Application \'$DevtronApp\' not found. Please verify the DevtronApp."
            exit 1
        fi
        echo "$app_id"
    fi
}

fetch_env_id() {
    # Check if DevtronEnv is numeric, if yes, use it directly as Env ID
    if [[ "$DevtronEnv" =~ ^[0-9]+$ ]]; then
        echo "$DevtronEnv"
    else
        local api_response=$(curl -s -H "Cookie: argocd.token=$DevtronApiToken" "$DevtronEndpoint/orchestrator/env/autocomplete")
        local env_id=$(echo "$api_response" | jq -r --arg env_name "$DevtronEnv" \'.result[] | select(.environment_name == $env_name) | .id\')

        if [[ -z "$env_id" || "$env_id" == "null" ]]; then
            echo "Error: Environment \'$DevtronEnv\' not found. Please verify the DevtronEnv."
            exit 1
        fi
        echo "$env_id"
    fi
}

fetch_pipeline_id() {
    local app_id=$1
    local env_id=$2
    local api_response=$(curl -s -H "Cookie: argocd.token=$DevtronApiToken" "$DevtronEndpoint/orchestrator/app/app-wf/view/$app_id")
    local pipeline_id=$(echo "$api_response" | jq -r --arg env_id "$env_id" \'.result.cdConfig.pipelines[] | select(.environmentId == ($env_id | tonumber)) | .id\')

    if [[ -z "$pipeline_id" || "$pipeline_id" == "null" ]]; then
        echo "Error: Pipeline not found for the provided Environment. Please verify the Environment ID."
        echo "Environment ID: $env_id"
        echo "API Response: $api_response"
        exit 1
    fi
    echo "$pipeline_id"
}

fetch_ci_artifact_id() {
    local pipeline_id=$1
    local apiUrl="$DevtronEndpoint/orchestrator/app/cd-pipeline/$pipeline_id/material?offset=0&size=20&stage=$TargetTriggerStage"
    local apiResponse=$(curl -s -H "Cookie: argocd.token=$DevtronApiToken" "$apiUrl")

    local ciArtifactId=""
    if [[ -n "$GitCommitHash" ]]; then
        ciArtifactId=$(echo "$apiResponse" | jq -r --arg hash "$GitCommitHash" \'.result.ci_artifacts[] | select(.material_info[].revision == $hash) | .id\' | head -n 1)
        if [[ -z "$ciArtifactId" || "$ciArtifactId" == "null" || "$ciArtifactId" == "" ]]; then
            echo "Error: CI Artifact ID for the provided commit hash \'$GitCommitHash\' not found. Please verify the commit hash."
            exit 1
        fi
    else
        ciArtifactId=$(echo "$apiResponse" | jq -r \'.result.ci_artifacts[0].id\')
        if [[ -z "$ciArtifactId" || "$ciArtifactId" == "null" ]]; then
            echo "Error: CI Artifact ID not found."
            exit 1
        fi
    fi

    echo "$ciArtifactId"
}

# Fetch the app ID. Exit the script if the app name is incorrect.
app_id=$(fetch_app_id)
if [ $? -ne 0 ]; then
    echo "Enter the correct App Name. Exiting..."
    exit 1
fi

# Fetch the env ID. Exit the script if the environment name or ID is incorrect.
env_id=$(fetch_env_id)
if [ $? -ne 0 ]; then
    echo "Enter the correct Environment Name. Exiting..."
    exit 1
fi

# Fetch the pipeline ID using the env ID. Exit the script if the environment is incorrect.
pipeline_id=$(fetch_pipeline_id "$app_id" "$env_id")
if [ $? -ne 0 ]; then
    echo "Verify your App Name/ID and Env Name/ID. Exiting..."
    exit 1
fi

# Fetch the CI Artifact ID based on the commit hash.
ciArtifactId=$(fetch_ci_artifact_id "$pipeline_id")
if [ $? -ne 0 ]; then
    echo "Failed to fetch a valid CI Artifact ID based on the provided GitCommitHash. Exiting..."
    exit 1
fi

trigger_cd_pipeline() {
    local pipeline_id=$1
    local app_id=$2
    local ciArtifactId=$3
    local jsonPayload=$(jq -n \\
        --arg pipelineId "$pipeline_id" \\
        --arg appId "$app_id" \\
        --arg ciArtifactId "$ciArtifactId" \\
        --arg TargetTriggerStage "$TargetTriggerStage" \\
        --arg deploymentWithConfig "LAST_SAVED_CONFIG" \\
        \'{
                                pipelineId: ($pipelineId | tonumber),
                                appId: ($appId | tonumber),
                                ciArtifactId: ($ciArtifactId | tonumber),
                                cdWorkflowType: $TargetTriggerStage,
                                deploymentWithConfig: $deploymentWithConfig
                            }\')

    curl -X POST "$DevtronEndpoint/orchestrator/app/cd-pipeline/trigger" \\
        -H "Content-Type: application/json" \\
        -H "Cookie: argocd.token=$DevtronApiToken" \\
        --data "$jsonPayload" \\
        --compressed
}

echo "Triggering CD Pipeline for App ID: $app_id, Pipeline ID: $pipeline_id, CI Artifact ID: $ciArtifactId, and CD Workflow Type: $TargetTriggerStage"
trigger_cd_pipeline "$pipeline_id" "$app_id" "$ciArtifactId"

check_deploy_status() {
    if [ "$StatusTimeOutSeconds" -le "0" ]; then
        echo "Skipping deployment status check. Taking 0 as a default input"
        return
    fi

    local appId=$1
    local pipelineId=$2
    local max_wait=$StatusTimeOutSeconds
    local statusKey="deploy_status" # Default status key

    if [[ "$TargetTriggerStage" == "PRE" ]]; then
        statusKey="pre_status"
    elif [[ "$TargetTriggerStage" == "POST" ]]; then
        statusKey="post_status"
    fi

    local start_time=$(date +%s)

    while :; do
        local current_time=$(date +%s)
        local elapsed_time=$((current_time - start_time))

        if [ "$elapsed_time" -ge "$max_wait" ]; then
            echo "Timeout reached without success. Exiting..."
            exit 1
        fi

        local statusUrl="$DevtronEndpoint/orchestrator/app/workflow/status/$appId/v2"
        local response=$(curl -s -H "Cookie: argocd.token=$DevtronApiToken" "$statusUrl")
        local code=$(echo "$response" | jq -r \'.code\')

        if [ "$code" != "200" ]; then
            echo "Error: Received response - $response. Exiting..."
            exit 1
        fi
        local status=$(echo "$response" | jq -r --arg pipelineId "$pipelineId" --arg statusKey "$statusKey" \'.result.cdWorkflowStatus[] | select(.pipeline_id == ($pipelineId | tonumber)) | .[$statusKey]\')

        echo "Current $TargetTriggerStage status: $status"

        if [[ "$status" == "Succeeded" ]]; then
            echo "Deployment succeeded."
            break
        elif [[ "$status" == "Failed" ]]; then
            echo "$TargetTriggerStage workflow failed."
            exit 1
        fi

        sleep $sleepInterval
    done
}

# Optionally check the deployment status based on the CD workflow type
check_deploy_status "$app_id" "$pipeline_id"'


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
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CD Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'DevtronApiToken','STRING','Enter Devtron API Token','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CD Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'DevtronEndpoint','STRING','Enter URL of Devtron','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CD Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'DevtronApp','STRING','Enter the Devtron Application name/Id','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CD Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'DevtronEnv','STRING','Enter the Environment name/Id','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CD Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'StatusTimeOutSeconds','STRING','Enter the maximum time (in seconds) a user can wait for the application to deploy.Enter a postive integer value','t','t',0,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CD Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'GitCommitHash','STRING','Enter the git hash from which user wants to deploy its application. By deault it takes latest Artifact ID to deploy the application','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CD Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'TargetTriggerStage','STRING','Enter Trigger Stage PRE/DEPLOY/POST, Default DEPLOY','t','t','DEPLOY',null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);
