-- Plugin metadata
INSERT INTO "plugin_metadata" ("id", "name", "description","type","icon","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_metadata'), 'Devtron Job Trigger v1.0.0','Devtronjob can be triggered with the help of name/id','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/devtron-logo-plugin.png','f', 'now()', 1, 'now()', 1);

-- Plugin Stage Mapping

INSERT INTO "plugin_stage_mapping" ("plugin_id","stage_type","created_on", "created_by", "updated_on", "updated_by")
VALUES ((SELECT id FROM plugin_metadata WHERE name='Devtron Job Trigger v1.0.0'),0,'now()', 1, 'now()', 1);

-- Plugin Script

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (
    nextval('id_seq_plugin_pipeline_script'),
    E'#!/bin/bash
if [ ! "$JobPipeline" ] && [ ! "$DevtronEnv" ];
then
    echo "Error: You must provide either JobPipeline or DevtronEnv to proceed."
    exit 1
fi
eventType=$(echo "$CI_CD_EVENT" | jq -r \'.type\')

if [ $StatusTimeoutSeconds -ge 100 ];
then
    sleep_time=30
else
    sleep_time=5
fi

#Functions
is_number() {
    [[ "$1" =~ ^[0-9]+$ ]]
}

if [ "$eventType" == "CI" ];
then
    TriggeredFrom=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.appName\')
    TriggeredFromPipeline=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.pipelineName\')
elif [ "$eventType" == "CD" ];
then
    TriggeredFrom=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.Pipeline.App.AppName\')
    TriggeredFromPipeline=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.workflowNamePrefix\')
else
    TriggeredFrom=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.appName\')
    TriggeredFromPipeline=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.pipelineName\')
fi

jobId=null
jobList=$(curl -s "$DevtronEndpoint/orchestrator/job/list" -H "token: $DevtronApiToken" --data-raw "{}")
if [ ! "$jobList" ];
then
    echo "Error: Please check the DevtronApiToken or DevtronEndpoint."
    exit 1
fi
if [ "$(echo "$jobList" | jq \'.code\')" -ne 200 ];
then
    echo "$jobList" | jq \'.result\'
    echo "Error: Please check the DevtronApiToken or DevtronEndpoint."
    exit 1
fi 


if is_number $DevtronJob; 
then
    jobId=$DevtronJob
else    
    jobId=$(echo "$jobList" | jq -r --arg DevtronJob "$DevtronJob" \'.result.jobContainers[] | select(.jobName == $DevtronJob) | .jobId\')
    if [ ! "$jobId" ];
    then
        echo "Error: Invalid DevtronJob. Please check the DevtronJob."
        exit 1
    fi
fi

pipelineId=null
DevtronEnvId=null
jobInfo=$(curl -s "$DevtronEndpoint/orchestrator/app/ci-pipeline/$jobId" -H "token: $DevtronApiToken")
jobInfoReturnCode=$(echo "$jobInfo" | jq \'.code\');
if [ "$jobInfoReturnCode" -ne 200 ];
then    
    echo "Error: Invalid DevtronJob. Please check the DevtronJob."
    exit 1
fi


if [ "$DevtronEnv" ];
then
    if is_number "$DevtronEnv";
    then
        DevtronEnvId=$DevtronEnv
        result_count=$(echo "$jobInfo" | jq -r --argjson DevtronEnv "$DevtronEnv" \'[.result.ciPipelines[] | select(.environmentId == $DevtronEnv) | .id] | length\')
        if [ "$result_count" == 1 ];
        then
            pipelineId=$(echo "$jobInfo" | jq -r --argjson DevtronEnv "$DevtronEnv" \'.result.ciPipelines[] | select(.environmentId == $DevtronEnv) | .id\')
        else
            echo "No Environment or more than one pipelines found within the same environment. Please use the JobPipeline."
            exit 1
        fi
    else
        envList=$(curl -s "$DevtronEndpoint/orchestrator/env" -H "token: $DevtronApiToken")
        DevtronEnvId=$(echo "$envList" | jq -r --arg DevtronEnv "$DevtronEnv" \'.result[] | select(.environment_name == $DevtronEnv) | .id\')
        if [ ! "$DevtronEnvId" ];
        then
            echo "Error: Invalid DevtronEnv. Please check the DevtronEnv."
            exit 1
        fi
        
        result_count=$(echo "$jobInfo" | jq -r --argjson DevtronEnvId "$DevtronEnvId" \'[.result.ciPipelines[] | select(.environmentId == $DevtronEnvId) | .id] | length\')
        if [ "$result_count" == 1 ];
        then
            pipelineId=$(echo "$jobInfo" | jq -r --argjson DevtronEnvId "$DevtronEnvId" \'.result.ciPipelines[] | select(.environmentId == $DevtronEnvId) | .id\')
        else
            echo "No Environment or more than one pipelines found within the same environment. Please use the JobPipeline."
            exit 1
        fi
    fi
else
    if is_number "$JobPipeline";
    then
        pipelineId=$JobPipeline
    else
        pipelineId=$(echo "$jobInfo" | jq -r --arg JobPipeline "$JobPipeline" \'.result.ciPipelines[] | select(.name == $JobPipeline) | .id\')
        if [ ! "$pipelineId" ];
        then
            echo "Error: Invalid JobPipeline. Please check the JobPipeline."
            exit 1
        fi
    fi
    DevtronEnvId=$(echo "$jobInfo" | jq -r --argjson pipelineId "$pipelineId" \'.result.ciPipelines[] | select(.id == $pipelineId) | .environmentId\')
    if [ ! "$DevtronEnvId" ];
    then
        echo "Error: Invalid JobPipeline. Please check the JobPipeline."
        exit 1
    fi
fi

gitMaterial=$(curl -s "$DevtronEndpoint/orchestrator/app/ci-pipeline/$pipelineId/material" -H "token: $DevtronApiToken")
if [ ! "$gitMaterial" ];
then
    echo "No git material found. Please check the details."
    exit 1
fi
gitMaterialId=$(echo "$gitMaterial" | jq \'.result[0].id\')

if [ ! "$GitCommitHash" ];
then   
    GitCommitHash=$(echo "$gitMaterial" | jq -r \'.result[0].history[0].Commit\')
fi


if [ ! "$DevtronEnvId" == null ];
then
    triggerCurl=$(curl -s "$DevtronEndpoint/orchestrator/app/ci-pipeline/trigger" -H "token: $DevtronApiToken" --data-raw \'{"TriggeredFrom": "\'"$TriggeredFrom"\'","TriggeredFromPipeline": "\'"$TriggeredFromPipeline"\'","pipelineId":\'"$pipelineId"\',"ciPipelineMaterials":[{"Id":\'"$gitMaterialId"\',"GitCommit":{"Commit":"\'"$GitCommitHash"\'"}}],"environmentId": \'"$DevtronEnvId"\'}\')
    triggerCurlReturnCode=$(echo "$triggerCurl" | jq \'.code\');
    if [ "$triggerCurlReturnCode" -ne 200 ];
    then
        echo "$triggerCurl" | jq \'.errors\'
        echo "Please check the details entered."
        exit 1
    else
        echo "The Job $DevtronJob has been triggered using GitCommitHash $GitCommitHash"
    fi
else
    triggerCurl=$(curl -s "$DevtronEndpoint/orchestrator/app/ci-pipeline/trigger" \\
    -H "token: $DevtronApiToken" \\
    --data-raw \'{"TriggeredFrom": "\'"$TriggeredFrom"\'","TriggeredFromPipeline": "\'"$TriggeredFromPipeline"\'","pipelineId":\'"$pipelineId"\',"ciPipelineMaterials":[{"Id":\'"$gitMaterialId"\',"GitCommit":{"Commit":"\'"$GitCommitHash"\'"}}]}\')
    triggerCurlReturnCode=$(echo "$triggerCurl" | jq \'.code\');
    if [ "$triggerCurlReturnCode" -ne 200 ];
    then
        echo "$triggerCurl" | jq \'.errors\'
        echo "Please check the details entered."
        exit 1
    else
        echo "The Job $DevtronJob has been triggered using GitCommitHash $GitCommitHash"
    fi
fi

if [ "$StatusTimeoutSeconds" -eq -1 ] || [ "$StatusTimeoutSeconds" -eq 0 ];
then
    echo "No waiting time provided. Hence not waiting for the Job status."
else
    sleep 1
    fetch_status() {
        curl -s "$DevtronEndpoint/orchestrator/app/workflow/status/$jobId/v2" \\
            -H "token: $DevtronApiToken" 
    }
    num=$(fetch_status)
    ci_status=$(echo "$num" | jq -r --argjson pipelineId "$pipelineId" \'.result.ciWorkflowStatus[] | select(.ciPipelineId == $pipelineId) | .ciStatus\');
    echo "The current status of the Job is: $ci_status";
    echo "Maximum waiting time is : $StatusTimeoutSeconds seconds"
    echo "Waiting for the Job to complete......"
    start_time=$(date +%s)
    job_completed=false
    while [ "$ci_status" != "Succeeded" ]; do
        if [ "$ci_status" == "Failed" ]; then
            echo "The Job has been Failed"
            exit 1
        elif [ "$ci_status" == "CANCELLED" ];then
            echo "Job has been Cancelled"
            exit 1
        fi
        current_time=$(date +%s)
        elapsed_time=$((current_time - start_time))
        if [ "$elapsed_time" -ge "$StatusTimeoutSeconds" ]; then
            echo "Timeout reached. Terminating the current process...."
            exit 1
        fi
        num=$(fetch_status)
        ci_status=$(echo "$num" | jq -r --argjson pipelineId "$pipelineId" \'.result.ciWorkflowStatus[] | select(.ciPipelineId == $pipelineId) | .ciStatus\')
        sleep $sleep_time
    done
    if [ "$ci_status" = "Succeeded" ]; then
        echo "The final status of the Job is: $ci_status"
        job_completed=true
    elif [ "$ci_status" = "Failed" ]; then
        echo "The final status of the Job is: $ci_status"
    else
        echo "The final status of the Job is: $ci_status (Timeout)"
    fi
    
    if [ "$job_completed" = true ]; then
        echo "The Job has been Scuccessfully completed"
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

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Devtron Job Trigger v1.0.0'),'Step 1','Runnig the plugin','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);


-- Input Variables

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format, description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by)VALUES
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Job Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'DevtronApiToken','STRING','Enter Devtron API Token with required permissions.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Job Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'DevtronEndpoint','STRING','Enter the URL of Devtron Dashboard for.eg (https://abc.xyz).','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Job Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'DevtronJob','STRING','Enter the name or ID of the Devtron Job to be triggered.','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Job Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'DevtronEnv','STRING','Enter the name or ID of the Environment where the Job is to be triggered. Required if JobPipeline is not given.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Job Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'JobPipeline','STRING','Enter the name or ID of the Job pipeline to be triggered. Required if DevtronEnv is not given.','t','t',null,null,'INPUT','NEW',null,1,null,null, 'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Job Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'GitCommitHash','STRING','Enter the commit hash from which the job is to be triggered. If not given then, will pick the latest.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Job Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'StatusTimeoutSeconds','NUMBER','Enter the maximum time to wait for the job status.', 't','t',-1,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);
