INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_metadata'),'BitBucket Runner Trigger v1.0.0' , 'The plugin enables users to trigger the pipeline in the BitBucket.','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/devtron-logo-plugin.png',false,'now()',1,'now()',1);


INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)VALUES (nextval('id_seq_plugin_stage_mapping'),
(SELECT id from plugin_metadata where name='BitBucket Runner Trigger v1.0.0'), 0,'now()',1,'now()',1);


INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES ( nextval('id_seq_plugin_pipeline_script'),
E'#!/bin/bash

# Extract git username, password, and git repository URL
if [[ -z "$BitBucketUsername" || -z "$BitBucketToken" ]]; then
    BitBucketUsername=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.ciProjectDetails[0].gitOptions.userName\')
    BitBucketToken=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.ciProjectDetails[0].gitOptions.password\')
fi

git_repository_url=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.ciProjectDetails[0].gitRepository\')

# Extract the workspace name and repository name from the gitRepository URL
if [[ -z "$WorkspaceName" ]]; then
    WorkspaceName=$(echo "$git_repository_url" | awk -F\'/\' \'{print $(NF-1)}\' | cut -d\'@\' -f2)
    if [[ "$WorkspaceName" == *"bitbucket.org:"* ]]; then
        # Extract everything after \'bitbucket.org:\'
        WorkspaceName="${WorkspaceName#*bitbucket.org:}"
    fi
fi

if [[ -z "$RepoName" ]]; then
    RepoName=$(echo "$git_repository_url" | awk -F\'/\' \'{print $NF}\' | sed \'s/.git//\')
fi

if [[ -z "$BranchName" ]]; then

    # Set a default value for sourceValue
    default_source_value="main"

    # Extract sourceType and sourceValue.
    source_type=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.ciProjectDetails[0].sourceType\')
    source_value=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.ciProjectDetails[0].sourceValue\')

    # Conditionally assign sourceValue based on sourceType
    if [ "$source_type" == "SOURCE_TYPE_BRANCH_FIXED" ]; then
        BranchName="$source_value"
    else
        BranchName="$default_source_value"
    fi
fi

if [[ -z "$BitBucketUsername" ||  -z "$BitBucketToken"  ]]; then
    echo "Enter the BitBucket username or api token. Exiting..."
    exit 1
fi

# Set default value for StatusTimeOutSeconds to 0 if not provided or not an integer
if ! [[ "$StatusTimeOutSeconds" =~ ^[0-9]+$ ]]; then
    StatusTimeOutSeconds=0
fi

# Determine sleep interval based on StatusTimeOutSeconds
if [ "$StatusTimeOutSeconds" -lt "60" ]; then
    sleepInterval=$(($StatusTimeOutSeconds / 2))
else
    sleepInterval=2  
fi

# Function for verify the workspaceName, RepoName and bitbucket Access API token
verify(){
    curl -s -u $BitBucketUsername:$BitBucketToken --request GET "https://api.bitbucket.org/2.0/repositories/$WorkspaceName/$RepoName/pipelines/?page=1&pagelen=1&sort=-created_on" --compressed
}

# call the verify function to get the response and act accordingly
verify_response=$(verify)


if [[  -z "$verify_response" ]]; then
    echo "Error: Unauthorized! Please check the API token or Username provided. Exiting..."
    exit 1
elif true ; then

    verify_response=$(verify | jq -r \'.error.message\') 
    if [[ "$verify_response" == "Token is invalid or not supported for this endpoint." ]]; then
        echo "Error: Unauthorized! Please check the API token or Username provided. Exiting..."
        exit 1
    elif [[ "$verify_response" == "Your credentials lack one or more required privilege scopes." ]]; then
        echo "Error, Your credentials lack one or more required privilege scopes. Exiting..."
        exit 1
    elif [[ "$verify_response" == "Resource not found" ]]; then
        echo "Error: Workspace Name $WorkspaceName or Repository Name $RepoName not found. Please check the details entered! Exiting..."
        exit 1
    fi
fi

# For v1.0, we fixed the type name as branch
type="branch"   

# function for trigger a runner in bitbucket
trigger_pipeline() {
    curl -s -X POST \\
    -u "$BitBucketUsername:$BitBucketToken" \\
    -H \'Content-Type: application/json\' \\
    "https://api.bitbucket.org/2.0/repositories/$WorkspaceName/$RepoName/pipelines/" \\
    -d \'{"target": {"ref_type": "\'$type\'", "type": "pipeline_ref_target", "ref_name": "\'$BranchName\'"}}\'

}

# call the trigger_pipeline function to get the error if we get error otherwise it will triggered successfully.
error=$(trigger_pipeline | jq -r \'.error.message\')
if [[ "$error" == "Not found" ]]; then
    echo "Error, Enter the correct branch name $BranchName. Exiting..." 
    exit 1
elif [[ "$error" == "null" ]]; then
    echo "Pipeline triggered successfully..."
fi


# check the status of the pipeline in the bitbucket 
check_healthy_status() {

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
            echo "StatusTimeOutSeconds reached without success. Exiting..."
            exit 1
        fi

        resources=$(verify)

        stateName=$(echo $resources | jq -r \'.values[].state.name\' )

        if [[ $stateName == "COMPLETED" ]]; then
            status=$(echo $resources | jq -r \'.values[].state.result.type\')
        elif [[ $stateName == "IN_PROGRESS" ]]; then
            status=$(echo $resources | jq -r \'.values[].state.stage.type\')
        elif [[ $stateName == "PENDING" ]]; then
            status="PENDING"
        fi


        if [[ "$status" == "pipeline_state_in_progress_running" ]]; then
                echo "Triggered Pipeline status is progressing..."
        elif [[ "$status" == "PENDING" ]]; then
                echo "Triggered Pipeline status is pending..."
        elif [[ "$status" == "pipeline_state_completed_successful" ]]; then
                echo "Pipeline succeeded."
                break
        elif [[ "$status" == "pipeline_state_completed_failed" ]]; then
                echo "Pipeline Failed."
                exit 1
        elif [[ "$status" == "pipeline_state_in_progress_halted" ]]; then
                echo "Pipeline Paused."
                exit 1
        fi
        sleep $sleepInterval
    done
}

# # Optionally check the healthy status of the pipeline in bitbucket
sleep 2
check_healthy_status',

    'SHELL',
    'f',
    'now()',
    1,
    'now()',
    1
);




INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='BitBucket Runner Trigger v1.0.0'),'Step 1','Step 1 - BitBucket Runner Trigger v1.0.0','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES 
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='BitBucket Runner Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'BitBucketUsername','STRING','Enter BitBucket Username','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='BitBucket Runner Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'BitBucketToken','STRING','Enter BitBucket Api Token','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='BitBucket Runner Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'WorkspaceName','STRING','Enter Workspace Name','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='BitBucket Runner Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'RepoName','STRING','Enter the repository name in the bitbucket','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='BitBucket Runner Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'BranchName','STRING','Enter the branch name','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='BitBucket Runner Trigger v1.0.0' and ps."index"=1 and ps.deleted=false),'StatusTimeOutSeconds','STRING','Enter the maximum time (in seconds) a user can wait for the application to deploy.Enter a postive integer value','t','t',0,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);





