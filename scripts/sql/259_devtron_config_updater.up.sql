-- Plugin metadata
INSERT INTO "plugin_metadata"("id", "name", "description","type","icon","deleted", "created_on", "created_by", "updated_on", "updated_by")VALUES (nextval('id_seq_plugin_metadata'), 'Devtron Config Updater v1.0.0','Update the configurations for the environment','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/devtron-logo-plugin.png','f', 'now()', 1, 'now()', 1);

-- Plugin Stage Mapping

INSERT INTO "plugin_stage_mapping" ("plugin_id","stage_type","created_on", "created_by", "updated_on", "updated_by")VALUES ((SELECT id FROM plugin_metadata WHERE name='Devtron Config Updater v1.0.0'),0,'now()', 1, 'now()', 1);

-- Plugin Script

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")VALUES (
    nextval('id_seq_plugin_pipeline_script'),
    E'
    #!/bin/bash
pipeline_type=$(echo $CI_CD_EVENT | jq -r \'.type\')
if [ $pipeline_type != "CD" ]; then
    echo "Plugin only works in Pre/Post CD"
    exit 1
fi
echo "$ExecutionScript" > execution_file.sh
chmod +x execution_file.sh
echo "Executing the ExecutionScript"
cat execution_file.sh
bash execution_file.sh

ResourceType=$(echo "$ResourceType" | tr \'[:upper:]\' \'[:lower:]\')
MountAs=$(echo "$MountAs" | tr \'[:upper:]\' \'[:lower:]\')

if [ "$MountAs" == "volume" ]; then
    if [ ! "$path" ]; then
        echo "Error: Please provide path variable..."
        exit 1
    fi
    if [ ! $FilePermission ]; then
        echo "File permissions not provided, using the default: 0644"
    fi
    if [ ! "$FileNameToBeMounted" ]; then
        echo "FileNameToBeMounted is empty, taking same as FileName"
        FileNameToBeMounted=$FileName
    fi
fi

if [ "$ResourceType" == "cm" ]; then

    json_data=$(cat $FileName);
    if [ ! "$json_data" ]; then
        echo "Error: No such file named $FileName found or the file is empty...."
        exit 1
    fi
    # app_id=$(echo "$CI_CD_EVENT" | jq ".commonWorkflowRequest.appId")
    # env_id=$(echo "$CI_CD_EVENT" | jq ".commonWorkflowRequest.Env.Id")
    app_id=1
    env_id=1

    cm_data=$(curl \\
    -s "${DashboardUrl}/orchestrator/config/environment/cm/${app_id}/${env_id}" \\
    -H "token: ${DevtronApiToken}")
    if [ ! "$cm_data" ]; then
        echo \'Please check the "DashboardUrl" value\'
        exit 1
    fi
    return_code=$(echo "$cm_data" | jq ".code")
    if [ "$return_code" -eq 200 ]; then

        cm_id=$(echo "$cm_data" | jq ".result.id")
        if [ "$MountAs" == "volume" ]; then
            payload=$(jq -n \\
            --arg FilePermission $FilePermission \\
            --arg ResourceName $ResourceName  \\
            --arg path $path \\
            --arg FileNameToBeMounted $FileNameToBeMounted \\
            --arg json_data "$json_data" \\
            --argjson cm_id "$cm_id" \\
            --argjson app_id "$app_id" \\
            --argjson env_id "$env_id" \\
            --argjson setSubpath "$setSubpath" \\
            \'{
                id: $cm_id,
                appId: $app_id,
                environmentId: $env_id,
                configData: [
                    {
                    name: $ResourceName,
                    type: "volume",
                    external: false,
                    data: {
                        ($FileNameToBeMounted) : $json_data
                    },
                    mountPath: $path,
                    subPath: $setSubpath,
                    filePermission : $FilePermission
                    }
                ]
            }\')

        elif [ "$MountAs" == "environment" ]; then

            payload=$(jq -n \\
            --arg ResourceName $ResourceName \\
            --arg FileName $FileName \\
            --argjson json_data "$json_data" \\
            --argjson cm_id "$cm_id" \\
            --argjson app_id "$app_id" \\
            --argjson env_id "$env_id" \\
            \'{
                id: $cm_id,
                appId: $app_id,
                environmentId: $env_id,
                configData: [{
                        name: $ResourceName,
                        type: "environment",
                        external: false,
                        data: $json_data
                }]
            }\')
        else
            echo "Please check the value of MountAs";
        fi
        echo "$payload" | jq "."

        cm_trigger=$(curl \\
        -s "${DashboardUrl}/orchestrator/config/environment/cm" \\
        -H "token: ${DevtronApiToken}" \\
        -H "Content-Type: application/json" \\
        --data-raw "$payload")

        cm_trigger_rescode=$(echo "$cm_trigger" | jq ".code")
        if [ "$cm_trigger_rescode" -eq 200 ]; then
            echo "The data has been patched successfully in cm."
        else
            echo "There was an error: "
            echo "Make sure the values in the json file are in double quotes and the values are in key : value format only."
            echo "$cm_trigger" | jq "."
        fi
    else
        echo "$cm_data" | jq "."
        exit 1
    fi


elif [ "$ResourceType" == "cs" ]; then

    json_data=$(cat $FileName);
    if [ ! "$json_data" ]; then
        echo "Error: No such file named $FileName found or the file is empty...."
        exit 1
    fi
    # app_id=$(echo "$CI_CD_EVENT" | jq ".commonWorkflowRequest.appId")
    # env_id=$(echo "$CI_CD_EVENT" | jq ".commonWorkflowRequest.Env.Id")
    app_id=1
    env_id=1

    cs_data=$(curl \\
    -s "${DashboardUrl}/orchestrator/config/environment/cs/${app_id}/${env_id}" \\
    -H "token: ${DevtronApiToken}")
    if [ ! "$cs_data" ];then
        echo "Please check the DashboardUrl value"
        exit
    fi
    return_code=$(echo "$cs_data" | jq ".code")
    if [ "$return_code" -eq 200 ]; then
        json_data=$(echo "$json_data" | base64)

        cs_id=$(echo "$cs_data" | jq ".result.id")
        if [ "$MountAs" == "volume" ]; then
            payload=$(jq -n \\
            --arg FilePermission $FilePermission \\
            --arg ResourceName $ResourceName  \\
            --arg path $path \\
            --arg FileNameToBeMounted $FileNameToBeMounted \\
            --arg json_data "$json_data" \\
            --argjson cs_id "$cs_id" \\
            --argjson app_id "$app_id" \\
            --argjson env_id "$env_id" \\
            --argjson setSubpath "$setSubpath" \\
            \'{
                id: $cs_id,
                appId: $app_id,
                environmentId: $env_id,
                configData: [
                    {
                    name: $ResourceName,
                    type: "volume",
                    external: false,
                    data: {
                        ($FileNameToBeMounted): $json_data
                    },
                    mountPath: $path,
                    subPath: $setSubpath,
                    filePermission: $FilePermission
                    }
                ]
            }\')

        elif [ "$MountAs" == "environment" ]; then

            json_object=$(cat $FileName)
            json_object=$(echo "$json_object" | jq ".")
            if [ ! "$json_object" ]; then
                echo "Could not get the valid JSON data. File should contain valid JSON object to add as environment"
                exit 1
            fi
            echo "$json_object"
            base64_encode() {
            echo -n "$1" | base64
            }

            # Use jq to process the JSON and encode the values
            encoded_json=$(echo "$json_object" | jq \'with_entries(
            .value |= (
                if type == "string" or type == "number" then
                @base64
                elif type == "boolean" then
                tostring | @base64
                else
                .
                end
            )
            )\')

            # Output the new JSON object with Base64 encoded values
            echo "$encoded_json"

            payload=$(jq -n \\
            --arg ResourceName $ResourceName \\
            --arg FileName $FileName \\
            --argjson json_data "$encoded_json" \\
            --argjson cs_id "$cs_id" \\
            --argjson app_id "$app_id" \\
            --argjson env_id "$env_id" \\
            \'{
                id: $cs_id,
                appId: $app_id,
                environmentId: $env_id,
                configData: [{
                        name: $ResourceName,
                        type: "environment",
                        external: false,
                        data: $json_data
                }]
            }\')
        else
            echo "Please check the value of MountAs";
        fi
        echo "$payload" | jq "."


        cs_trigger=$(curl \\
        -s "${DashboardUrl}/orchestrator/config/environment/cs" \\
        -H "token: ${DevtronApiToken}" \\
        -H "Content-Type: application/json" \\
        --data-raw "$payload")

        cs_trigger_rescode=$(echo "$cs_trigger" | jq ".code")
        if [ "$cs_trigger_rescode" -eq 200 ]; then
            echo "The data has been patched successfully in cs."
        else
            echo "There was an error: "
            echo "Make sure the values in the json file are in double quotes and the values are in key : value format only."
            echo "$cs_trigger" | jq "."
        fi
    else
        echo "$cs_data" | jq "."
        exit 1
    fi

else
    echo "Unknown resource type."
    exit 1
fi
',
    'SHELL',
    'f',
    'now()',
    1,
    'now()',
    1
);


--Plugin Step

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by")VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Devtron Config Updater v1.0.0'),'Step 1','Updating the configs','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);


-- Input Variables

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Config Updater v1.0.0' and ps."index"=1 and ps.deleted=false), 'ResourceType','STRING','Specify the type of resource the file is to be mounted as configMap or COnfigSecret. Options: cm/cs',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Config Updater v1.0.0' and ps."index"=1 and ps.deleted=false), 'DashboardUrl','STRING','Dashboard url of Devtron for eg. https://previw.devtron.ai',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Config Updater v1.0.0' and ps."index"=1 and ps.deleted=false), 'DevtronApiToken','STRING','Devtron API token with required permissions.',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Config Updater v1.0.0' and ps."index"=1 and ps.deleted=false), 'FileName','STRING','Name of the file to be mounted.',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Config Updater v1.0.0' and ps."index"=1 and ps.deleted=false), 'FileNameToBeMounted','STRING','Name of the file to be mounted as volume. Default is same as FileName',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Config Updater v1.0.0' and ps."index"=1 and ps.deleted=false), 'path','STRING','Path on which the file is to be mounted.',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Config Updater v1.0.0' and ps."index"=1 and ps.deleted=false), 'setSubpath','STRING','true or false, true if you want to set subpath. Default is false',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Config Updater v1.0.0' and ps."index"=1 and ps.deleted=false), 'MountAs','STRING','How do you want to mount this file? Options: volume/environment.',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Config Updater v1.0.0' and ps."index"=1 and ps.deleted=false), 'ResourceName','STRING','Name of the resource to be created.',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Config Updater v1.0.0' and ps."index"=1 and ps.deleted=false), 'FilePermission','STRING','Set the permission of the file after mounting as a volume.',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Devtron Config Updater v1.0.0' and ps."index"=1 and ps.deleted=false), 'ExecutionScript','STRING','Provide the script to create the config/secret file.',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);
