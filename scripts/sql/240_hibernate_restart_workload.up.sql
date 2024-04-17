INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_metadata'),'Devtron Hibernate/Restart Workload v1.0.0' , 'The plugin enables users to hibernate/unhibernate and restart workload. It helps users restart the applications that contains dependencies for their current workload.','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/devtron-logo-plugin.png',false,'now()',1,'now()',1);


INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)VALUES (nextval('id_seq_plugin_stage_mapping'),
(SELECT id from plugin_metadata where name='Devtron Hibernate/Restart Workload v1.0.0'), 0,'now()',1,'now()',1);


INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES ( nextval('id_seq_plugin_pipeline_script'),
E'#!/bin/bash


pipeline_type=$(echo $CI_CD_EVENT | jq -r \'.type\')
if [[ "$pipeline_type" == "CI" ]]; then
    triggeredFromAppName=$(echo $CI_CD_EVENT | jq \'.commonWorkflowRequest.appName\')
    triggeredFromPipelineName=$(echo $CI_CD_EVENT | jq \'.commonWorkflowRequest.pipelineName\')
elif [[ "$pipeline_type" == "CD" ]]; then
    triggeredFromAppName=$(echo $CI_CD_EVENT | jq \'.commonWorkflowRequest.extraEnvironmentVariables.APP_NAME\')
    triggeredFromPipelineName=$(echo $CI_CD_EVENT | jq \'.commonWorkflowRequest.Pipeline.Name\')
fi


# Set default value for StatusTimeOutSeconds to 0 if not provided or not an integer
if ! [[ "$StatusTimeOutSeconds" =~ ^[0-9]+$ ]]; then
    StatusTimeOutSeconds=0
fi

DevtronEndpoint=$(echo "$DevtronEndpoint" | sed \'s:/*$::\')

# Determine sleep interval based on StatusTimeOutSeconds
if [[ "$StatusTimeOutSeconds" -lt "60" ]]; then
    sleepInterval=$(($StatusTimeOutSeconds / 2))
else
    sleepInterval=30
fi

verify(){
    local response=$(curl -s -H "token: $DevtronApiToken" "$DevtronEndpoint/orchestrator/devtron/auth/verify")
    echo $response
    exit 1
}
verify_response=$(verify)
verify_status=$( echo "$verify_response" | jq \'.code\')
if [[ "$verify_status" == "401" ]]; then
    echo "Enter the valid DevtronApiToken. Exiting..."
    exit 1
elif [[ -z "$verify_status" ]]; then
    echo "Enter the valid DevtronEndpoint. Exiting..."
    exit 1 
fi

fetch_app_id() {
    # Check if DevtronApp is numeric, if yes, use it directly as App ID
    if [[ "$DevtronApp" =~ ^[0-9]+$ ]]; then
        echo "$DevtronApp"
    else
        local api_response=$(curl -s -H "token: $DevtronApiToken" "$DevtronEndpoint/orchestrator/app/autocomplete")
        local app_id=$(echo "$api_response" | jq -r --arg app_name "$DevtronApp" \'.result[] | select(.name == $app_name) | .id\')
        if [[ -z "$app_id" || "$app_id" == "null" ]]; then
            exit 1
        fi
        echo "$app_id"

    fi
}

# Fetch the app ID. Exit the script if the app name is incorrect.
app_id=$(fetch_app_id)
if [ $? -ne 0 ]; then
    echo "Error: Application \'$DevtronApp\' not found. Enter the correct App Name. Exiting...."
    exit 1
fi




fetch_env_id() {
    # Check if DevtronEnv is numeric, if yes, use it directly as Env ID
    if [[ "$DevtronEnv" =~ ^[0-9]+$ ]]; then
        echo "$DevtronEnv"
    else
        local api_response=$(curl -s -H "token: $DevtronApiToken" "$DevtronEndpoint/orchestrator/env/autocomplete")
        local env_id=$(echo "$api_response" | jq -r --arg env_name "$DevtronEnv" \'.result[] | select(.environment_name == $env_name) | .id\')

        if [[ -z "$env_id" || "$env_id" == "null" ]]; then
            exit 1
        fi
        echo "$env_id"
    fi
}


# Fetch the env ID. Exit the script if the environment name or ID is incorrect.
env_id=$(fetch_env_id)
if [ $? -ne 0 ]; then
    echo "Error: Environment \'$DevtronEnv\' not found. Enter the correct Environment Name. Exiting..."
    exit 1
fi

resource_tree() {
    # fetch the details from the resource-tree api and save into the variables
    api_response=$(curl -s -H "token: $DevtronApiToken" "$DevtronEndpoint/orchestrator/app/detail/resource-tree?app-id=$app_id&env-id=$env_id")       
    echo "$api_response"
}


# checking deployment status 
deployment_status(){
    # fetch the details from the resource-tree api and save into the variables
    deployment_status_response=$(curl -s -H "token: $DevtronApiToken" "$DevtronEndpoint/orchestrator/app/deployment-status/timeline/$app_id/$env_id?showTimeline=false")       
    echo "$deployment_status_response"
}

timeline_response=$(deployment_status)
if [ $? -ne 0 ]; then
    echo "Enter the correct App details. Exiting..."
    echo "Not able to fetch the deployment status..."
    exit 1
fi


resources=$(resource_tree)
appname_envname=$(echo $resources | jq \'.result.nodes[] | select(.canBeHibernated == true ) | .name\')
namespace=$(echo $resources | jq \'.result.nodes[] | select(.canBeHibernated == true )| .namespace\') 
group=$(echo $resources | jq \'.result.nodes[] | select(.canBeHibernated == true ) | .group\')
version=$(echo $resources | jq \'.result.nodes[] | select(.canBeHibernated == true ) | .version\')
kind=$(echo $resources | jq \'.result.nodes[] | select(.canBeHibernated == true ) | .kind\')
application_status=$(echo $resources | jq \'.result.status\' )
deployment_status=$(echo $timeline_response | jq \'.result.wfrStatus\' )



# for restart the application
restart_workload() {

    resources=$(resource_tree)

    local status=$(echo $resources | jq -r \'.result.status\')

    if [ "$status" = "HIBERNATING" ] || [ "$status" = "hibernating" ] || [ "$status" = "hibernated" ]; then
        exit 1
    fi

    curl -s "$DevtronEndpoint/orchestrator/app/rotate-pods" \\
        -X POST \\
        -H "Content-Type: application/json" \\
        -H "token: $DevtronApiToken" \\
        --data-raw \'{"triggered_from_app":\'"${triggeredFromAppName}"\',"triggered_from_pipeline":\'"${triggeredFromPipelineName}"\',"appId":\'$app_id\',"environmentId":\'$env_id\',"resources":[{"name":\'$appname_envname\',"namespace":\'$namespace\',"groupVersionKind":{"Group":\'$group\',"Version":\'$version\',"Kind":\'$kind\'}}]}\' \\
        --compressed
}



# for hibernate the application
hibernate_app() {

    resources=$(resource_tree)

    local status=$(echo $resources | jq -r \'.result.status\')

    if [ "$status" = "HIBERNATING" ] || [ "$status" = "hibernating" ] || [ "$status" = "hibernated" ]; then
        exit 1
    fi

    curl -s "$DevtronEndpoint/orchestrator/app/stop-start-app" \\
        -X POST \\
        -H "Content-Type: application/json" \\
        -H "token: $DevtronApiToken" \\
        --data-raw \'{"triggered_from_app":\'"${triggeredFromAppName}"\',"triggered_from_pipeline":\'"${triggeredFromPipelineName}"\',"AppId":\'$app_id\',"EnvironmentId":\'$env_id\',"RequestType":"STOP"}\'
}


# for unhibernate the application
un_hibernate_app() {

    resources=$(resource_tree)

    local status=$(echo $resources | jq -r \'.result.status\')

    if [ "$status" != "HIBERNATING" ]; then
        exit 1
    fi

    curl -s "$DevtronEndpoint/orchestrator/app/stop-start-app" \\
        -X POST \\
        -H "Content-Type: application/json" \\
        -H "token: $DevtronApiToken" \\
        --data-raw \'{"triggered_from_app":\'"${triggeredFromAppName}"\',"triggered_from_pipeline":\'"${triggeredFromPipelineName}"\',"AppId":\'$app_id\',"EnvironmentId":\'$env_id\',"RequestType":"START"}\'
}

# un_hibernate_app

check_application_status() {
    if [ "$StatusTimeOutSeconds" -le "0" ]; then
        echo "Skipping application status check. Taking 0 as a default input"
        return
    fi

    local max_wait=$StatusTimeOutSeconds


    local start_time=$(date +%s)

    while :; do
        local current_time=$(date +%s)
        local elapsed_time=$((current_time - start_time))

        if [ "$elapsed_time" -ge "$max_wait" ]; then
            echo "Timeout reached without success. Exiting..."
            exit 1
        fi

        resources=$(resource_tree)
        # echo $resources
        status=$(echo $resources | jq -r \'.result.status\' )

        echo "Current Application status:" $status
        # echo "1"
        status=$(echo "$status" | tr \'[:upper:]\' \'[:lower:]\')

        if [[ "$PluginAction" == "restart" ]]; then
            if [[ "$status" == "healthy" ]]; then
                echo $PluginAction "workload succeeded."
                break
            elif [[ "$status" == "unknown" || "$status" == "suspended"  || "$status" ==  "missing" || "$status" == "hibernated" || "$status" == "hibernating" ]]; then
                echo $PluginAction "workload failed due to" $status
                exit 1
            fi
        elif [[ "$PluginAction" == "hibernate" ]]; then
            if [[  "$status" == "hibernated" ||  "$status" == "hibernating" ]]; then
                echo $PluginAction "workload succeeded."
                break
            elif [[ "$status" == "unknown" || "$status" == "suspended"  || "$status" ==  "missing" ]]; then
                echo $PluginAction "workload failed due to" $status
                exit 1
            fi
        elif [[ "$PluginAction" == "unhibernate" ]]; then
            if [[  "$status" == "healthy" ]]; then
                echo $PluginAction "workload succeeded."
                break
            elif [[ "$status" == "unknown" || "$status" == "suspended"  || "$status" ==  "missing" || "$status" == "degraded" ]]; then
                echo $PluginAction "workload failed due to" $status
                exit 1
            fi
        fi

        sleep $sleepInterval
    done
   
}



PluginAction=$(echo "$PluginAction" | tr \'[:upper:]\' \'[:lower:]\')

echo "The plugin action is" $PluginAction

if [[ "${PluginAction}" == "restart" ]]; then
    result=$(restart_workload)
    code=$(echo "$result" | jq -r \'.code\')
    if [ -z "$code" ]; then
        echo "Workload is hibernating state already. Exiting..."
        exit 1
    elif [ "$code" != "200" ]; then
        echo "Error: Received response - $result. Exiting..."
        exit 1
    fi
    
elif [[ "${PluginAction}" == "hibernate" ]]; then
    result=$(hibernate_app)
    code=$(echo "$result" | jq -r \'.code\')
    if [ -z "$code" ]; then
        echo "Workload is hibernating state already. Exiting..."
        exit 1
    elif [ "$code" != "200" ]; then
        echo "Error: Received response - $result. Exiting..."
        exit 1
    fi


elif [[ "${PluginAction}" == "unhibernate" ]]; then 
    result=$(un_hibernate_app)
    code=$(echo "$result" | jq -r \'.code\')
    if [ -z "$code" ]; then
        echo "Workload is un-hibernating state already. Exiting..."
        exit 1
    elif [ "$code" != "200" ]; then
        echo "Error: Received response - $result. Exiting..."
        exit 1
    fi

    sleep 5

else 
    echo "Enter the correct Action Name, You have entered "$PluginAction 
    exit 1

fi


# Optionally check the deployment status based on the CD workflow type
check_application_status',

    'SHELL',
    'f',
    'now()',
    1,
    'now()',
    1
);




INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Devtron Hibernate/Restart Workload v1.0.0'),'Step 1','Step 1 - Devtron Hibernate/Restart Workload v1.0.0','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES 
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Hibernate/Restart Workload v1.0.0' and ps."index"=1 and ps.deleted=false),'DevtronApiToken','STRING','Enter Devtron API Token','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Hibernate/Restart Workload v1.0.0' and ps."index"=1 and ps.deleted=false),'DevtronEndpoint','STRING','Enter URL of Devtron','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Hibernate/Restart Workload v1.0.0' and ps."index"=1 and ps.deleted=false),'DevtronApp','STRING','Enter the Devtron Application name/Id','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Hibernate/Restart Workload v1.0.0' and ps."index"=1 and ps.deleted=false),'DevtronEnv','STRING','Enter the Environment name/Id','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Hibernate/Restart Workload v1.0.0' and ps."index"=1 and ps.deleted=false),'StatusTimeOutSeconds','STRING','Enter the maximum time (in seconds) a user can wait for the application to deploy.Enter a postive integer value','t','t',0,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Hibernate/Restart Workload v1.0.0' and ps."index"=1 and ps.deleted=false),'PluginAction','STRING','Enter Hibernate/Unhibernate/Restart','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);



