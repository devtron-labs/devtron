INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_metadata'),'AutoSync Environmment Configuration v1.0.0' , 'The plugin enables to sync the environment with each other.','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/devtron-logo-plugin.png',false,'now()',1,'now()',1);


INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)VALUES (nextval('id_seq_plugin_stage_mapping'),
(SELECT id from plugin_metadata where name='AutoSync Environmment Configuration v1.0.0'), 0,'now()',1,'now()',1);


INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES ( nextval('id_seq_plugin_pipeline_script'),
E'#!/bin/bash
echo "================== AUTOSYNC ENVIRONMENT CONFIGURATION PLUGIN STEP STARTS FROM HERE ========================="
DevtronEndpoint=${DevtronEndpoint%/*}
echo "Entered DevtronEndpoint is: $DevtronEndpoint"

echo "==================================== DEPLOYMENT TEMPLATE ==================================================="
# Extract appId from CI_CD_EVENT 
appId=$(echo "$CI_CD_EVENT" | jq -r \'.commonWorkflowRequest.appId\')
if [ -z "$appId" ]; then
    echo "Error: Could not extract appId from CI_CD_EVENT."
    exit 1
fi

# Fetch response from API using appId
response=$(curl -s "$DevtronEndpoint/orchestrator/app/other-env/min?app-id=$appId" -H "cookie: argocd.token=$DevtronApiToken")
if [ $(echo $response | jq \'.code\') == "403" ]; then
    echo "Error: User does not authorized for the source and target environments. Exiting..."
    exit 1
elif [ $(echo $response | jq \'.code\') == "401" ]; then
    echo "Error: Enter the correct Devtron API token. Exiting..."
    exit 1
elif [ $(echo $response | jq \'.code\') -ge 400 ] && [ $(echo $response | jq \'.code\') -lt 500 ]; then
    echo "Error: Failed to fetch data of the other-environments."
    exit 1
fi

# Extract source environment ID and chartRefId 
sourceEnvId=$(echo "$response" | jq -r --arg envName "$SourceEnvironmentName" \'.result[] | select(.environmentName == $envName) | .environmentId\')
if [ -z "$sourceEnvId" ]; then
    echo "Error: Could not extract sourceEnvironmentID. Exiting..."
    exit 1
fi

sourceChartRefId=$(echo "$response" | jq -r --arg envName "$SourceEnvironmentName" \'.result[] | select(.environmentName == $envName) | .chartRefId\')
if [ -z "$sourceChartRefId" ]; then
    echo "Error: Could not extract sourceChartRefID. Exiting... "
    exit 1
fi

echo "Source Environment ID for \'$SourceEnvironmentName\' is: $sourceEnvId"
echo "ChartRefId for environment \'$SourceEnvironmentName\' is: $sourceChartRefId"

# Extract target environment ID and chartRefId 
targetEnvId=$(echo "$response" | jq -r --arg envName "$TargetEnvironmentName" \'.result[] | select(.environmentName == $envName) | .environmentId\')
if [ -z "$targetEnvId" ]; then
    echo "Error: Could not extract targetEnvironmentID. Exiting... "
    exit 1
fi

targetChartRefId=$(echo "$response" | jq -r --arg envName "$TargetEnvironmentName" \'.result[] | select(.environmentName == $envName) | .chartRefId\')
if [ -z "$targetChartRefId" ]; then
    echo "Error: Could not extract targetChartRefID. Exiting... "
    exit 1
fi

echo "Target Environment ID for \'$TargetEnvironmentName\' is: $targetEnvId"
echo "ChartRefId for environment \'$TargetEnvironmentName\' is: $targetChartRefId"

# Fetch deployment template information for source environment
url="$DevtronEndpoint/orchestrator/env?id=$targetEnvId"
targetNamespaceJson=$(curl -s "$url" -H "cookie: argocd.token=$DevtronApiToken")
if [ $(echo $targetNamespaceJson | jq \'.code\') == "403" ]; then
    echo "Error: User does not authorized for the source and target environments. Exiting..."
    exit 1
elif [ $(echo $targetNamespaceJson | jq \'.code\') == "401" ]; then
    echo "Error: Enter the correct Devtron API token. Exiting..."
    exit 1
elif [ $(echo $targetNamespaceJson | jq \'.code\') -ge 400 ] && [ $(echo $targetNamespaceJson | jq \'.code\') -lt 500 ]; then
    echo "Error: Failed to fetch data of the target environments."
    exit 1
fi

targetNamespace=$(echo $targetNamespaceJson | jq -r \'.result.namespace\')
if [ -z "$targetNamespace" ]; then
    echo "Error: Could not extract targetNamespace. Exiting... "
    exit 1
fi

url="$DevtronEndpoint/orchestrator/env?id=$sourceEnvId"
sourceNamespaceJson=$(curl -s "$url" -H "cookie: argocd.token=$DevtronApiToken")
if [ $(echo $sourceNamespaceJson | jq \'.code\') == "403" ]; then
    echo "Error: User does not authorized for the source and target environments. Exiting..."
    exit 1
elif [ $(echo $sourceNamespaceJson | jq \'.code\') == "401" ]; then
    echo "Error: Enter the correct Devtron API token. Exiting..."
    exit 1
elif [ $(echo $sourceNamespaceJson | jq \'.code\') -ge 400 ] && [ $(echo $sourceNamespaceJson | jq \'.code\') -lt 500 ]; then
    echo "Error: Failed to fetch data of the source environments."
    exit 1
fi

sourceNamespace=$(echo $sourceNamespaceJson | jq -r \'.result.namespace\')
if [ -z "$sourceNamespace" ]; then
    echo "Error: Could not extract sourceNamespace. Exiting... "
    exit 1
fi

url="$DevtronEndpoint/orchestrator/app/env/$appId/$sourceEnvId/$sourceChartRefId"
sourceTempJson=$(curl -s "$url" -H "cookie: argocd.token=$DevtronApiToken")
if [ $(echo $sourceTempJson | jq \'.code\') == "403" ]; then
    echo "Error: User does not authorized for the source and target environments. Exiting..."
    exit 1
elif [ $(echo $sourceTempJson | jq \'.code\') == "401" ]; then
    echo "Error: Enter the correct Devtron API token. Exiting..."
    exit 1
elif [ $(echo $sourceTempJson | jq \'.code\') -ge 400 ] && [ $(echo $sourceTempJson | jq \'.code\') -lt 500 ]; then
    echo "Error: Failed to fetch data of the source deployment template."
    exit 1
fi

sourceDeploymentTemplateId=$(echo $sourceTempJson | jq \'.result.environmentConfig.id\')
if [ -z "$sourceDeploymentTemplateId" ]; then
    echo "Error: Could not extract source deployment template ID. Exiting... "
    exit 1
fi

# Get the deployment template api and id
url="$DevtronEndpoint/orchestrator/app/env/$appId/$targetEnvId/$targetChartRefId"
targetTempJson=$(curl -s "$url" -H "cookie: argocd.token=$DevtronApiToken")
if [ $(echo $targetTempJson | jq \'.code\') == "403" ]; then
    echo "Error: User does not authorized for the source and target environments. Exiting..."
    exit 1
elif [ $(echo $targetTempJson | jq \'.code\') == "401" ]; then
    echo "Error: Enter the correct Devtron API token. Exiting..."
    exit 1
elif [ $(echo $targetTempJson | jq \'.code\') -ge 400 ] && [ $(echo $targetTempJson | jq \'.code\') -lt 500 ]; then
    echo "Error: Failed to fetch data of the target deployment template. Exiting..."
    exit 1
fi

targetDeploymentTemplateId=$(echo $targetTempJson | jq \'.result.environmentConfig.id\')
if [ -z "$targetDeploymentTemplateId" ]; then
    echo "Error: Could not extract target deployment template ID. Exiting... "
    exit 1
fi

# Update the deployment template for the target environment
updated_json=$(echo "$sourceTempJson" | jq \'.result.environmentConfig\')
if [ -z "$updated_json" ]; then
    echo "Error: Could not update deployment template config. Exiting... "
    exit 1
fi
updated_json=$(echo "$updated_json" | jq --argjson new_deployment_template_id $targetDeploymentTemplateId \'.id = $new_deployment_template_id\')
if [ -z "$updated_json" ]; then
    echo "Error: Could not update deployment template config. Exiting... "
    exit 1
fi
updated_json=$(echo "$updated_json" | jq --argjson new_id $targetEnvId \'.environmentId = $new_id\')
if [ -z "$updated_json" ]; then
    echo "Error: Could not update deployment template config. Exiting... "
    exit 1
fi
updated_json=$(echo "$updated_json" | jq --arg new_name "$TargetEnvironmentName" \'.environmentName = $new_name\')
if [ -z "$updated_json" ]; then
    echo "Error: Could not update deployment template config. Exiting... "
    exit 1
fi
updated_json=$(echo "$updated_json" | jq --arg new_ns "$targetNamespace" \'.namespace = $new_ns\')
if [ -z "$updated_json" ]; then
    echo "Error: Could not update deployment template config. Exiting... "
    exit 1
fi

url="$DevtronEndpoint/orchestrator/app/env"
echo "Syncing source environment to the target environment $SourceEnvironmentName and $TargetEnvironmentName"
response=$(curl -s "$url" -X PUT -H "cookie: argocd.token=$DevtronApiToken" --data-raw "$updated_json")

code=$(echo $response | jq \'.code\')

if [ -n "$code" ] && [ "$code" -ge 400 ] && [ "$code" -lt 500 ]; then
    echo "Error: Failed to sync target deployment template with the source deployment template details from API. Exiting..."
    exit 1
fi

echo "========================================= CONFIG MAP ======================================================="

# Function to delete a config map
delete_config_map() {
    local id=$1
    local appId=$2
    local targetEnvId=$3
    local configName=$4

    # Construct the DELETE request
    temp=$(curl -s -X DELETE "${DevtronEndpoint}/orchestrator/config/environment/cm/${appId}/${targetEnvId}/${id}?name=${configName}" -H "cookie: argocd.token=${DevtronApiToken}") 
    if [ $(echo $temp | jq \'.code\') == "403" ]; then
        echo "Error: User does not authorized for target environments. Exiting..."
        exit 1
    elif [ $(echo $temp | jq \'.code\') == "401" ]; then
        echo "Error: Enter the correct Devtron API token. Exiting..."
        exit 1
    elif [ $(echo $temp | jq \'.code\') -ge 400 ] && [ $(echo $temp | jq \'.code\') -lt 500 ]; then
        echo "Error: Failed to delete the target environment configMaps. Exiting..."
        exit 1
    fi
}

# Fetch JSON response from API
sourceJsonResponse=$(curl -s "${DevtronEndpoint}/orchestrator/config/environment/cm/${appId}/${sourceEnvId}" -H "cookie: argocd.token=${DevtronApiToken}")
if [ $(echo $sourceJsonResponse | jq \'.code\') == "403" ]; then
    echo "Error: User does not authorized for the source config environments. Exiting..."
    exit 1
elif [ $(echo $sourceJsonResponse | jq \'.code\') == "401" ]; then
    echo "Error: Enter the correct Devtron API token. Exiting..."
    exit 1
elif [ $(echo $sourceJsonResponse | jq \'.code\') -ge 400 ] && [ $(echo $sourceJsonResponse | jq \'.code\') -lt 500 ]; then
    echo "Error: Failed to get the source environment configMaps. Exiting..."
    exit 1
fi

targetJsonResponse=$(curl -s "${DevtronEndpoint}/orchestrator/config/environment/cm/${appId}/${targetEnvId}" -H "cookie: argocd.token=${DevtronApiToken}")
if [ $(echo $targetJsonResponse | jq \'.code\') == "403" ]; then
    echo "Error: User does not authorized for the target environments configMaps. Exiting..."
    exit 1
elif [ $(echo $targetJsonResponse | jq \'.code\') == "401" ]; then
    echo "Error: Enter the correct Devtron API token. Exiting..."
    exit 1
elif [ $(echo $targetJsonResponse | jq \'.code\') -ge 400 ] && [ $(echo $targetJsonResponse | jq \'.code\') -lt 500 ]; then
    echo "Error: Failed to get the target environment configMaps. Exiting..."
    exit 1
fi

sourceCmId=$(echo $sourceJsonResponse | jq \'.result.id\')
if [ -z "$sourceCmId" ]; then
    echo "Error: Could not get source configMap id. Exiting... "
    exit 1
fi

targetCmId=$(echo $targetJsonResponse | jq \'.result.id\')
if [ -z "$targetCmId" ]; then
    echo "Error: Could not get target configMap id. Exiting... "
    exit 1
fi

# Parse config maps from JSON response
configMapsData=$(echo "$sourceJsonResponse" | jq -c \'.result.configData\')
if [ -z "$configMapsData" ]; then
    echo "Error: Could not get source configMap data. Exiting... "
    exit 1
fi


# Loop through each config map and delete it
echo "$configMapsData" | jq -c \'.[]\' | while read -r cm; do
    configName=$(echo "$cm" | jq -r \'.name // empty\')

    # Check if required fields are present
    if [[ -z "$configName" ]]; then
        echo "Empty Config Map for target environment. Exiting..."
        exit 1
    elif [[ -z "$targetCmId" || -z "$appId" || -z "$targetEnvId" ]]; then
        echo "Error: Missing required fields in config map data. Exiting..."
        exit 1
    else
        # Delete the config map
        delete_config_map "$targetCmId" "$appId" "$targetEnvId" "$configName"
    fi

    # Send the config map data as a JSON object in the array
    echo "Syncing config map in the target environment: $configName"
    temp=$(curl -s "${DevtronEndpoint}/orchestrator/config/environment/cm" -H "cookie: argocd.token=${DevtronApiToken}" --data-raw "{"\'"id"\'": $targetCmId, "\'"appId"\'": $appId, "\'"environmentId"\'": $targetEnvId, "\'"configData"\'": [$cm]}")
    if [ $(echo $temp | jq \'.code\') == "403" ]; then
        echo "Error: User does not authorized for the creating target environments configMaps. Exiting..."
        exit 1
    elif [ $(echo $temp | jq \'.code\') == "401" ]; then
        echo "Error: Enter the correct Devtron API token. Exiting..."
        exit 1
    elif [ $(echo $temp | jq \'.code\') -ge 400 ] && [ $(echo $temp | jq \'.code\') -lt 500 ]; then
        echo "Error: Failed to create the target environment configMaps. Exiting..."
        exit 1
    fi
done 


echo "=========================================== SECRETS ========================================================"

# Function to delete a config map
delete_secrets() {
    local id=$1
    local appId=$2
    local targetEnvId=$3
    local configName=$4

    # Construct the DELETE request
    temp=$(curl -s -X DELETE "${DevtronEndpoint}/orchestrator/config/environment/cs/${appId}/${targetEnvId}/${id}?name=${configName}" -H "cookie: argocd.token=${DevtronApiToken}") 
    if [ $(echo $temp | jq \'.code\') == "403" ]; then
        echo "Error: User does not authorized for target environments. Exiting..."
        exit 1
    elif [ $(echo $temp | jq \'.code\') == "401" ]; then
        echo "Error: Enter the correct Devtron API token. Exiting..."
        exit 1
    elif [ $(echo $temp | jq \'.code\') -ge 400 ] && [ $(echo $temp | jq \'.code\') -lt 500 ]; then
        echo "Error: Failed to delete the target environment secrets. Exiting..."
        exit 1
    fi
}

# Fetch JSON response from API
sourceJsonResponse=$(curl -s "${DevtronEndpoint}/orchestrator/config/environment/cs/${appId}/${sourceEnvId}" -H "cookie: argocd.token=${DevtronApiToken}")
if [ $(echo $sourceJsonResponse | jq \'.code\') == "403" ]; then
    echo "Error: User does not authorized for the source secrets environments. Exiting..."
    exit 1
elif [ $(echo $sourceJsonResponse | jq \'.code\') == "401" ]; then
    echo "Error: Enter the correct Devtron API token. Exiting..."
    exit 1
elif [ $(echo $sourceJsonResponse | jq \'.code\') -ge 400 ] && [ $(echo $sourceJsonResponse | jq \'.code\') -lt 500 ]; then
    echo "Error: Failed to get the source environment secrets. Exiting..."
    exit 1
fi

targetJsonResponse=$(curl -s "${DevtronEndpoint}/orchestrator/config/environment/cs/${appId}/${targetEnvId}" -H "cookie: argocd.token=${DevtronApiToken}")
if [ $(echo $targetJsonResponse | jq \'.code\') == "403" ]; then
    echo "Error: User does not authorized for the target environments secrets. Exiting..."
    exit 1
elif [ $(echo $targetJsonResponse | jq \'.code\') == "401" ]; then
    echo "Error: Enter the correct Devtron API token. Exiting..."
    exit 1
elif [ $(echo $targetJsonResponse | jq \'.code\') -ge 400 ] && [ $(echo $targetJsonResponse | jq \'.code\') -lt 500 ]; then
    echo "Error: Failed to get the target environment secrets. Exiting..."
    exit 1
fi

sourceCsId=$(echo $sourceJsonResponse | jq \'.result.id\')
if [ -z "$sourceCsId" ]; then
    echo "Error: Could not get source secrets id. Exiting... "
    exit 1
fi

targetCsId=$(echo $targetJsonResponse | jq \'.result.id\')
if [ -z "$targetCsId" ]; then
    echo "Error: Could not get target secrets id. Exiting... "
    exit 1
fi

# Parse config maps from JSON response
configSecretData=$(echo "$sourceJsonResponse" | jq -c \'.result.configData\')
if [ -z "$configSecretData" ]; then
    echo "Error: Could not get source secrets data. Exiting... "
    exit 1
fi

# Loop through each config map and delete it
echo "$configSecretData" | jq -c \'.[]\' | while read -r cs; do
    secretName=$(echo "$cs" | jq -r \'.name // empty\')

    # Check if required fields are present
    if [[ -z "$secretName" ]]; then
        echo "Empty Config Map for target environment. Exiting..."
        exit 1
    elif [[ -z "$targetCsId" || -z "$appId" || -z "$targetEnvId" ]]; then
        echo "Error: Missing required fields in secrets data. Exiting..."
        exit 1
    else
        # Delete the config map
        delete_secrets "$targetCsId" "$appId" "$targetEnvId" "$secretName"
    fi

    # Fetching particular secret data.
    tempSecretData=$(curl -s "${DevtronEndpoint}/orchestrator/config/environment/cs/edit/$appId/$sourceEnvId/$sourceCsId?name=$secretName" -H "cookie: argocd.token=${DevtronApiToken}")
    cs=$(echo $tempSecretData | jq \'.result.configData[0]\' )

    # Send the config map data as a JSON object in the array
    echo "Syncing secret in the target environment: $secretName"
    temp=$(curl -s "${DevtronEndpoint}/orchestrator/config/environment/cs" -H "cookie: argocd.token=${DevtronApiToken}" --data-raw "{"\'"id"\'": $targetCsId, "\'"appId"\'": $appId, "\'"environmentId"\'": $targetEnvId, "\'"configData"\'": [$cs]}")
    if [ $(echo $temp | jq \'.code\') == "403" ]; then
        echo "Error: User does not authorized for the creating target environments secrets. Exiting..."
        exit 1
    elif [ $(echo $temp | jq \'.code\') == "401" ]; then
        echo "Error: Enter the correct Devtron API token. Exiting..."
        exit 1
    elif [ $(echo $temp | jq \'.code\') -ge 400 ] && [ $(echo $temp | jq \'.code\') -lt 500 ]; then
        echo "Error: Failed to create the target environment secrets. Exiting..."
        exit 1
    fi
done 
echo "Successfully Sync the Deployment Template, ConfigMaps and Secrets"
echo "============================== PLUGIN STEP SUCCESSFULLY COMPLETE ==========================================="',

    'SHELL',
    'f',
    'now()',
    1,
    'now()',
    1
);


INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='AutoSync Environmment Configuration v1.0.0'),'Step 1','Step 1 - AutoSync Environmment Configuration v1.0.0','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES 
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='AutoSync Environmment Configuration v1.0.0' and ps."index"=1 and ps.deleted=false),'SourceEnvironmentName','STRING','Enter source environment name','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='AutoSync Environmment Configuration v1.0.0' and ps."index"=1 and ps.deleted=false),'TargetEnvironmentName','STRING','Enter target environment name','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='AutoSync Environmment Configuration v1.0.0' and ps."index"=1 and ps.deleted=false),'DevtronEndpoint','STRING','Enter host url','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='AutoSync Environmment Configuration v1.0.0' and ps."index"=1 and ps.deleted=false),'DevtronApiToken','STRING','Enter the devtron api token','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);









