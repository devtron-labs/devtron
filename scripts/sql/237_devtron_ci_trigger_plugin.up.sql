-- Plugin metadata
INSERT INTO "plugin_metadata" ("id", "name", "description","type","icon","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_metadata'), 'Devtron CI Trigger v1.0.0','Triggers the CI pipeline of Devtron Application','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/devtron-logo-plugin.png','f', 'now()', 1, 'now()', 1);

-- Plugin Stage Mapping

INSERT INTO "plugin_stage_mapping" ("plugin_id","stage_type","created_on", "created_by", "updated_on", "updated_by")
VALUES ((SELECT id FROM plugin_metadata WHERE name='Devtron CI Trigger v1.0.0'),0,'now()', 1, 'now()', 1);

-- Plugin Script

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (
    nextval('id_seq_plugin_pipeline_script'),
    E'#!/bin/bash
triggeredFromAppName=$(echo $CI_CD_EVENT | jq \'.commonWorkflowRequest.appName\')
triggeredFromPipelineName=$(echo $CI_CD_EVENT | jq \'.commonWorkflowRequest.pipelineName\')

sleep_time=5
is_number() {
    [[ "$1" =~ ^[0-9]+$ ]]
}
if [ $Timeout -gt 100 ]; then
    sleep_time=15
fi
echo "------------------------------------------"
echo "Starting plugin Devtron CI Trigger v1.0.0"
echo "------------------------------------------"
app_list=$(curl -s -H "token:$DevtronApiToken" "$DevtronEndpoint/orchestrator/app/autocomplete")
code=$(echo $app_list | jq -r ".code")
if [ $code -ne 200 ];then
    result=$(echo $app_list | jq -r \'.result\')
    echo "Error: $result! Please check the API token provided"
    exit 1
fi
if is_number $DevtronApp;then
    app_id=$DevtronApp
else
    app_id=$(echo "$app_list" | jq -r --arg DevtronApp "$DevtronApp" \'.result[] | select(.name == $DevtronApp) | .id\')
fi
if ! [ $app_id ];then
    echo "App $DevtronApp not found! Please check the details entered. for eg.(DevtronApp,DevtronEnv,DevtronEndpoint)"
    exit 1
fi
app_workflow=$(curl -s -H "token:$DevtronApiToken" "$DevtronEndpoint/orchestrator/app/app-wf/view/$app_id")
if is_number $CiPipeline;then
    ci_pipeline_id=$CiPipeline
else
    if [ $DevtronEnv ];then
        if is_number $DevtronEnv;then
            ci_pipeline_id=$(echo "$app_workflow" | jq -r --argjson DevtronEnv $DevtronEnv \'.result.cdConfig.pipelines[] | select(.environmentId == $DevtronEnv) | .ciPipelineId\')
        else
            ci_pipeline_id=$(echo "$app_workflow" | jq -r --arg DevtronEnv "$DevtronEnv" \'.result.cdConfig.pipelines[] | select(.environmentName == $DevtronEnv) | .ciPipelineId\')
        fi
    elif [ $CiPipeline ];then
        ci_pipeline_id=$(echo "$app_workflow" | jq -r --arg CiPipeline "$CiPipeline" \'.result.ciConfig.ciPipelines[] | select(.name == $CiPipeline) | .id\')
    else
        echo "You must provide one of the fields: DevtronEnv or CiPipeline"
        echo "Error: DevtronEnv or ciPipelineId not provided" 
        exit 1   
    fi
fi
if [ ! $ci_pipeline_id ];then
    echo "Please check the CI Pipeline Name or DevtronEnv"
    exit 1
fi
git_material_id=$(echo "$app_workflow" | jq -r --argjson ci_pipeline_id "$ci_pipeline_id" \'.result.ciConfig.ciPipelines[] | select(.id == $ci_pipeline_id) | .ciMaterial[0].gitMaterialId\')
if [ $GitCommitHash ];then
    curl_req=$(curl -s "$DevtronEndpoint/orchestrator/app/ci-pipeline/trigger" \\
        -H "token: $DevtronApiToken" \\
        --data-raw \'{"triggered_from_app":\'"${triggeredFromAppName}"\',"triggered_from_pipeline":\'"${triggeredFromPipelineName}"\',"pipelineId":\'"$ci_pipeline_id"\',"ciPipelineMaterials":[{"Id":\'"$git_material_id"\',"GitCommit":{"Commit":"\'"$GitCommitHash"\'"}}],"pipelineType":"CI_BUILD"}\')
    code=$(echo "$curl_req" | jq -r \'.code\')
    if [ $code -ne 200 ];then
        error=$(echo "$curl_req" | jq -r \'.errors[]\')
        echo "$error"
        echo "CI Pipeline details could not be found. Please check!"
        exit 1
    fi
    echo "The build with CI pipeline ID $ci_pipeline_id of application $DevtronApp is triggered using the commit: $GitCommitHash"
    else
        git_material=$(curl -s "$DevtronEndpoint/orchestrator/app/ci-pipeline/$ci_pipeline_id/material" \\
            -H "token: $DevtronApiToken")
        result=$(echo $git_material | jq -r \'.result[]\')
        history=$(echo $result | jq -r \'.history[0]\')
        GitCommitHash=$(echo $history | jq -r \'.Commit\')
        curl_req=$(curl -s "$DevtronEndpoint/orchestrator/app/ci-pipeline/trigger" \\
            -H "token: $DevtronApiToken" \\
            --data-raw \'{"triggered_from_app":\'"${triggeredFromAppName}"\',"triggered_from_pipeline":\'"${triggeredFromPipelineName}"\',"pipelineId":\'"$ci_pipeline_id"\',"ciPipelineMaterials":[{"Id":\'"$git_material_id"\',"GitCommit":{"Commit":"\'"$GitCommitHash"\'"}}],"pipelineType":"CI_BUILD"}\')
        code=$(echo "$curl_req" | jq -r \'.code\')
        if [ $code -ne 200 ];then
            error=$(echo "$curl_req" | jq -r \'.errors[]\')
            echo "$error"
            echo "CI Pipeline details could not be found. Please check!"
            exit 1
        fi
        echo "The build with CI pipeline ID $ci_pipeline_id of application $DevtronApp is triggered using the latest commit: $GitCommitHash"
fi
if [ $Timeout -eq -1 ] || [ $Timeout -eq 0 ];then
    echo "Pipeline has been Triggered"
else
    sleep 1
    fetch_status() {
        curl --silent "$DevtronEndpoint/orchestrator/app/workflow/status/$app_id/v2" \\
            -H "token: $DevtronApiToken" 
    }
    num=$(fetch_status)
    ci_status=$(echo $num | jq -r --argjson ci_pipeline_id $ci_pipeline_id \'.result.ciWorkflowStatus[] | select(.ciPipelineId == $ci_pipeline_id) | .ciStatus\');
    echo "The current status of the build is: $ci_status";
    echo "Maximum waiting time is : $Timeout seconds"
    echo "Waiting for the process to complete......"
    start_time=$(date +%s)
    job_completed=false
    while [ "$ci_status" != "Succeeded" ]; do
        if [ "$ci_status" == "Failed" ]; then
            echo "The build has been Failed"
            exit 1
        elif [ "$ci_status" == "CANCELLED" ];then
            echo "Build has been Cancelled"
            exit 1
        fi
        current_time=$(date +%s)
        elapsed_time=$((current_time - start_time))
        if [ "$elapsed_time" -ge "$Timeout" ]; then
            echo "Timeout reached. Terminating the current process...."
            exit 1
            break
        fi
        num=$(fetch_status)
        ci_status=$(echo $num | jq -r --argjson ci_pipeline_id $ci_pipeline_id \'.result.ciWorkflowStatus[] | select(.ciPipelineId == $ci_pipeline_id) | .ciStatus\')
        sleep $sleep_time
    done
    if [ "$ci_status" = "Succeeded" ]; then
        echo "The final status of the build is: $ci_status"
        job_completed=true
    elif [ "$ci_status" = "Failed" ]; then
        echo "The final status of the Build is: $ci_status"
    else
        echo "The final status of the Build is: $ci_status (Timeout)"
    fi
    
    if [ "$job_completed" = true ]; then
        echo "The triggered Build is Scuccessfully completed"
    else
        exit 1 
    fi
fi',
    'SHELL',
    'f',
    'now()',
    1,
    'now()',
    1
);


--Plugin Step

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Devtron CI Trigger v1.0.0'),'Step 1','Runnig the plugin','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);


-- Input Variables

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format, description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by)VALUES
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CI Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'DevtronApiToken','STRING','Enter Devtron API Token with required permissions.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CI Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'DevtronEndpoint','STRING','Enter the URL of Devtron Dashboard for.eg (https://abc.xyz).','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CI Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'DevtronApp','STRING','Enter the name or ID of the Application whose build is to be triggered.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CI Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'DevtronEnv','STRING','Enter the name or ID of the Environment to which the CI is attached. Required if CiPipeline is not given.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CI Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'CiPipeline','STRING','Enter the name or ID of the CI pipeline to be triggered. Required if DevtronEnv is not given.','t','t',null,null,'INPUT','NEW',null,1,null,null, 'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CI Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'GitCommitHash','STRING','Enter the commit hash from which the build is to be triggered. If not given then will pick the latest.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron CI Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'Timeout','NUMBER','Enter the maximum time to wait for the build status.', 't','t',-1,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);
