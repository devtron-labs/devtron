INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'Cosign v1.0.0','This plugin is used to Cosign to sign the docker images created in CI.','PRESET','https://github.com/jatin-jangir-0220/test/blob/main/CDBA2B21-4D52-4A1A-8BF4-C9F66B9CF2FF.png?raw=true',false,'now()',1,'now()',1);

INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_stage_mapping'),(SELECT id from plugin_metadata where name='Cosign v1.0.0'), 0,'now()',1,'now()',1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (
     nextval('id_seq_plugin_pipeline_script'),
        $$#!/bin/sh 
set -eo pipefail 
apk update

# Install 
apk add cosign

export COSIGN_PASSWORD=$CosignPassword
# Verify the installation
cosign version
echo $DOCKER_IMAGE

if [ -z "$VariableAsPrivateKey" ]; then
    echo "VariableAsPrivateKey is not set. VariableAsPrivateKey must be present."
    if [ -z "$PreCommand" ]; then
        echo " PreCommand must be present."
        if [ -z "$PrivateKeyFile" ]; then
            echo "PrivateKeyFile must be present."
            exit 1
        else
            echo "in PrivateKeyFile"
            cosign sign --yes=true --key $PrivateKeyFile $DOCKER_IMAGE $ExtraArguments
        fi
    else
        if [ -z "$PrivateKeyFile" ]; then
            echo " PreCommand is  set but PrivateKeyFile is not, We must define PrivateKeyFile ."
            exit 1
        else
            echo "in PreCommand"
            $PreCommand
            cosign sign --yes=true --key $PrivateKeyFile $DOCKER_IMAGE $ExtraArguments
        fi
    fi
else
    echo "in VariableAsPrivateKey"
    echo $VariableAsPrivateKey | base64 -d > cosign_ci.key
    cosign sign  --yes=true --key cosign_ci.key $DOCKER_IMAGE $ExtraArguments
fi

$PostCommand
echo "Cosign completed"$$,
        'SHELL',
        'f',
        'now()',
        1,
        'now()',
        1
);






INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Cosign v1.0.0'),'Step 1','Step 1 - Cosign v1.0.0','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);


INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Cosign v1.0.0' and ps."index"=1 and ps.deleted=false),'DOCKER_IMAGE','STRING','docker image','f','t',null,null,'INPUT','GLOBAL',null,1,null,'DOCKER_IMAGE','f','now()',1,'now()',1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Cosign v1.0.0' and ps."index"=1 and ps.deleted=false),'CosignPassword','STRING','password for cosign private key','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);


INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Cosign v1.0.0' and ps."index"=1 and ps.deleted=false),'VariableAsPrivateKey','STRING','base64 encoded private-key (use scope variable)[highest priority]','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Cosign v1.0.0' and ps."index"=1 and ps.deleted=false),'PreCommand','STRING','run command to get required conditions to run cosign sign command. (also required PrivateKeyFile)','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);


INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Cosign v1.0.0' and ps."index"=1 and ps.deleted=false),'PrivateKeyFile','STRING','path of key in git repo. [lowest priority]','t','t','cosign.key',null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Cosign v1.0.0' and ps."index"=1 and ps.deleted=false),'PostCommand','STRING','command to run after cosign sign.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Cosign v1.0.0' and ps."index"=1 and ps.deleted=false),'ExtraArguments','STRING','arguments for cosign sign command','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);
