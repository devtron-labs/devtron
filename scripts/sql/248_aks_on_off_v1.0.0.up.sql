INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'AKS Cluster ON/OFF v1.0.0', 'The plugin helps user to ON/OFF AKS cluster.','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/devtron-logo-plugin.png',false,'now()',1,'now()',1);


INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)VALUES (nextval('id_seq_plugin_stage_mapping'),
(SELECT id from plugin_metadata where name='AKS Cluster ON/OFF v1.0.0'), 0,'now()',1,'now()',1);


INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES ( nextval('id_seq_plugin_pipeline_script'),
E'#!/bin/bash


# Check if required directories exist, create them if not
if [ ! -d "/devtroncd/process" ]; then
    mkdir -p /devtroncd/process
fi

# Check if required files exist, create them if not
if [ ! -f "/devtroncd/process/stage-0_out.env" ]; then
    touch /devtroncd/process/stage-0_out.env
fi

# Convert Action to lowercase
Action=$(echo "$Action" | tr \'[:upper:]\' \'[:lower:]\')

if [[ $Action != "start" && $Action != "stop" ]]; then
    echo "Action must be \'start\' or \'stop\'."
    exit 1
fi

# Check if StatusTimeOutSeconds is a positive integer, else set it to 0
if ! [[ "$StatusTimeOutSeconds" =~ ^[1-9][0-9]*$ ]]; then
    StatusTimeOutSeconds=0
fi

# Convert ServicePrinciple to lowercase
ServicePrinciple=$(echo "$ServicePrinciple" | tr \'[:upper:]\' \'[:lower:]\')

# Validate ServicePrinciple input
if [[ "$ServicePrinciple" != "true" && "$ServicePrinciple" != "false" ]]; then
    echo "ServicePrinciple must be \'true\' or \'false\'."
    exit 1
fi

# Function to obtain Azure access token using client credentials
get_access_token() {
    local TenantID="$1"
    local ClientID="$2"
    local ClientSecret="$3"

    # Obtain access token from Azure AD
    response=$(curl -s -X POST \\
        -d "grant_type=client_credentials&client_id=$ClientID&client_secret=$ClientSecret&resource=https://management.azure.com/" \\
        "https://login.microsoftonline.com/$TenantID/oauth2/token")

    # Extract the access token from the response
    token=$(echo $response | jq -r .access_token)
    if [[ $token == null || $token == "" ]]; then
        echo "Error: Failed to retrieve access token."
        exit 1
    fi
    echo $token
}

# Function to get the current status of the AKS cluster using the REST API
get_cluster_status() {
    local token=$1
    local api_version="2023-10-01"
    local url="https://management.azure.com/subscriptions/$SubscriptionID/resourceGroups/$ResourceGroup/providers/Microsoft.ContainerService/managedClusters/$ClusterName?api-version=$api_version"

    # Curl command to get the current cluster status
    response=$(curl -s -X GET -H "Authorization: Bearer $token" -H \'Content-Type: application/json\' "$url")
    echo $response | jq -r \'.properties | .provisioningState + "," + .powerState.code\'
}

# Function to manage AKS cluster start or stop using the Azure REST API with adjusted sleep interval
manage_cluster() {
    local cluster_action="$1"
    local token="$2"
    local StatusTimeOutSeconds="$3"
    local sleep_interval="$4"
    local api_version="2023-10-01"
    local url="https://management.azure.com/subscriptions/$SubscriptionID/resourceGroups/$ResourceGroup/providers/Microsoft.ContainerService/managedClusters/$ClusterName/$cluster_action?api-version=$api_version"

    # Curl command to start/stop the AKS cluster using token
    curl -s -X POST -H "Authorization: Bearer $token" -H \'Content-Type: application/json\' --data-raw "{}" "$url" | jq

    if [[ "$StatusTimeOutSeconds" -le 0 ]]; then
        echo "StatusTimeOutSeconds is less than or equal to zero or not a positive integer. Skipping cluster status check."
        return
    fi

    echo "Checking AKS cluster status..."
    local start_time=$(date +%s)
    local end_time=$((start_time + StatusTimeOutSeconds))
    local current_status=""
    while true; do
        current_status=$(get_cluster_status "$token")
        local provisioning_state=$(echo $current_status | cut -d \',\' -f 1)
        local power_state=$(echo $current_status | cut -d \',\' -f 2)
        echo "Provisioning State: $provisioning_state, Power State: $power_state"
        if [[ "$provisioning_state" == *"Succeeded"* && "$power_state" == *"Running"* && "$cluster_action" == "start" ]] || [[ "$provisioning_state" == *"Succeeded"* && "$power_state" == *"Stopped"* && "$cluster_action" == "stop" ]]; then
            echo "Cluster $ClusterName has reached the desired state."
            break
        elif [[ "$provisioning_state" == *"Starting"* || "$provisioning_state" == *"Stopping"* ]]; then
            local current_time=$(date +%s)
            if [[ "$current_time" -ge "$end_time" ]]; then
                echo "Failed to $Action the cluster within $StatusTimeOutSeconds seconds."
                exit 1
            fi
            sleep $sleep_interval
        else
            echo "Error: Invalid input"
            exit 1
        fi
    done
}

# Check if StatusTimeOutSeconds is greater than 60 seconds
if [[ "$StatusTimeOutSeconds" -gt 60 ]]; then
    sleep_interval=30
else
    # Calculate sleep interval as modulus (StatusTimeOutSeconds divided by 2)
    sleep_interval=$((StatusTimeOutSeconds / 2))
fi

# Check if ServicePrinciple is enabled and use Azure CLI for operations
if [[ "$ServicePrinciple" == "true" ]]; then
    echo "ServicePrinciple enabled. Running operations via Azure CLI in Docker..."
    
    # Additional check using Azure CLI to determine the AKS cluster power state
    power_state=$(docker run --rm mcr.microsoft.com/azure-cli /bin/bash -c "
        az login --identity > /dev/null;
        az aks show --resource-group $ResourceGroup --name $ClusterName --query \'powerState.code\' -o tsv
    ")
    echo "Current power state of AKS cluster: $power_state"
    if [[ "$Action" == "start" && "$power_state" == "Running" ]] || [[ "$Action" == "stop" && "$power_state" == "Stopped" ]]; then
        echo "Cluster is already in the $Action state."
        exit 1  # Exit with non-zero status code to indicate failure
    fi

    echo "Proceeding to $Action the cluster..."
    if [[ "$StatusTimeOutSeconds" -le 0 ]]; then
        docker run --rm mcr.microsoft.com/azure-cli /bin/bash -c "
            az login --identity > /dev/null;
            az aks $Action --name $ClusterName --resource-group $ResourceGroup --no-wait;
            echo \'Operation started. Skipping cluster status check due to StatusTimeOutSeconds being 0.\';
        "
        echo "Operation $Action on cluster $ClusterName completed."
    else
        docker run --rm mcr.microsoft.com/azure-cli /bin/bash -c "
            az login --identity > /dev/null;
            az aks $Action --name $ClusterName --resource-group $ResourceGroup --no-wait;
            echo \'Operation started. Checking cluster status...\';
            start_time=\$(date +%s);
            end_time=\$((start_time + $StatusTimeOutSeconds));
            while true; do
                current_status=\$(az aks show --resource-group $ResourceGroup --name $ClusterName --query \'provisioningState\' -o tsv);
                echo \\"Current Provisioning State: \\$current_status\\";
                if [[ \\"\\$current_status\\" == \'Succeeded\' ]]; then
                    echo \'Cluster $ClusterName has reached the desired state.\';
                    break;
                fi
                current_time=\\$(date +%s);
                if [[ \\$current_time -ge \\$end_time ]]; then
                    echo \'Failed to $Action the cluster within $StatusTimeOutSeconds seconds.\';
                    exit 1
                fi
                sleep $sleep_interval;
            done
        "
        echo "Operation $Action on cluster $ClusterName completed."
    fi
    
    exit 0
fi

# If ServicePrinciple is not enabled, continue with managing cluster using curl
# Get access token
token=$(get_access_token "$TenantID" "$ClientID" "$ClientSecret")
cluster_status=$(get_cluster_status "$token")
desired_status="Running"
if [[ "$Action" == "stop" ]]; then
    desired_status="Stopped"
fi

# Decide on action based on the current status
if [[ "$cluster_status" == *"$desired_status"* ]]; then
    echo "Cluster $ClusterName is already in the desired state: $desired_status."
    exit 1  # Exit with non-zero status code to indicate failure
elif [[ "$cluster_status" == *"Starting"* || "$cluster_status" == *"Stopping"* ]]; then
    echo "Cluster $ClusterName is currently changing states. Please wait until the current operation is complete."
    exit 1
else
    echo "Cluster $ClusterName is not in $Action state. Proceeding with $Action operation..."
    manage_cluster "$Action" "$token" "$StatusTimeOutSeconds" "$sleep_interval"
fi

echo "Operation $Action on cluster $ClusterName completed."


'


,
    'SHELL',
    'f',
    'now()',
    1,
    'now()',
    1
);
INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='AKS Cluster ON/OFF v1.0.0'),'Step 1','Step 1 - AKS Cluster ON/OFF v1.0.0','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES 
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='AKS Cluster ON/OFF v1.0.0' and ps."index"=1 and ps.deleted=false),'ResourceGroup','STRING','Enter Resource group name','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='AKS Cluster ON/OFF v1.0.0' and ps."index"=1 and ps.deleted=false),'ClusterName','STRING','Enter the Cluster name ','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='AKS Cluster ON/OFF v1.0.0' and ps."index"=1 and ps.deleted=false),'Action','STRING','Enter the action want to perform. (Stop/start)','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='AKS Cluster ON/OFF v1.0.0' and ps."index"=1 and ps.deleted=false),'ServicePrinciple','STRING','True if permission given to VM, False if not','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='AKS Cluster ON/OFF v1.0.0' and ps."index"=1 and ps.deleted=false),'TenantID','STRING','Enter the Tenant ID','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='AKS Cluster ON/OFF v1.0.0' and ps."index"=1 and ps.deleted=false),'ClientSecret','STRING','Enter the client secret','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='AKS Cluster ON/OFF v1.0.0' and ps."index"=1 and ps.deleted=false),'SubscriptionID','STRING','Enter the subscription ID','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='AKS Cluster ON/OFF v1.0.0' and ps."index"=1 and ps.deleted=false),'ClientID','STRING','Enter the client ID','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='AKS Cluster ON/OFF v1.0.0' and ps."index"=1 and ps.deleted=false),'StatusTimeOutSeconds','STRING','Enter the maximum time (in seconds) a user can wait for the application to deploy.Enter a postive integer value','t','t',0,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);
