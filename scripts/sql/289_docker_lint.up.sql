INSERT INTO "plugin_parent_metadata" ("id", "name","identifier", "description","type","icon","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_parent_metadata'), 'DOCKER LINT','docker-lint','This is used to analyze the Dockerfile and offer suggestions for improvements','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/devtron-logo-plugin.png','f', 'now()', 1, 'now()', 1);


UPDATE plugin_metadata SET is_latest = false WHERE id = (SELECT id FROM plugin_metadata WHERE name= 'DOCKER LINT' and is_latest= true);


INSERT INTO "plugin_metadata" ("id", "name", "description","deleted", "created_on", "created_by", "updated_on", "updated_by","plugin_parent_metadata_id","plugin_version","is_deprecated","is_latest")
VALUES (nextval('id_seq_plugin_metadata'), 'DOCKER LINT','This is used to analyze the Dockerfile and offer suggestions for improvements','f', 'now()', 1, 'now()', 1, (SELECT id FROM plugin_parent_metadata WHERE identifier='docker-lint'),'1.0.0', false, true);


INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_tag_relation'),(SELECT id FROM plugin_tag WHERE name='Security') , (SELECT id FROM plugin_metadata WHERE plugin_version='1.0.0' and name='DOCKER LINT' and deleted= false),'now()', 1, 'now()', 1);


INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_tag_relation'),(SELECT id FROM plugin_tag WHERE name='DevSecOps') , (SELECT id FROM plugin_metadata WHERE plugin_version='1.0.0' and name='DOCKER LINT' and deleted= false),'now()', 1, 'now()', 1);


INSERT INTO "plugin_stage_mapping" ("plugin_id","stage_type","created_on", "created_by", "updated_on", "updated_by")
VALUES ((SELECT id FROM plugin_metadata WHERE plugin_version='1.0.0' and name='DOCKER LINT' and deleted= false),3,'now()', 1, 'now()', 1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (
    nextval('id_seq_plugin_pipeline_script'),
    E'
    set -ex
    arch=$(uname -m)
    os=$(uname -s)
    echo $arch
    echo $os
    command=$(wget https://github.com/hadolint/hadolint/releases/download/v2.12.0/hadolint-$os-$arch)
    echo $command
    docker_path="Dockerfile"
    echo $docker_path
    if [ ! -z  "$DOCKER_FILE_PATH" ]
    then
        echo "hello"
        docker_path=$DOCKER_FILE_PATH
    fi  
    echo $docker_path   
    cp hadolint-Linux-x86_64 hadolint
    chmod +x hadolint
    if [[ $FAIL_ON_ERROR == "true" ]]
    then
        ./hadolint "/devtroncd/$docker_path"
    else
         ./hadolint "/devtroncd/$docker_path"  --no-fail
    fi       

',
    'SHELL',
    'f',
    'now()',
    1,
    'now()',
    1
);





INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step'),(SELECT id FROM plugin_metadata WHERE plugin_version='1.0.0' and name='DOCKER LINT' and deleted= false),'Step 1','Step 1 - Triggering DOCKER LINT','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);


INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type","default_value", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.plugin_version='1.0.0' and p.name='DOCKER LINT' and p.deleted=false and ps."index"=1 and ps.deleted=false), 'DOCKER_FILE_PATH','STRING','Specify the file path to the Dockerfile for linting or processing. By default path is Dockerfile',true,true,'INPUT','NEW','',1 ,'f','now()', 1, 'now()', 1);


INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type","default_value", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.plugin_version='1.0.0' and p.name='DOCKER LINT' and p.deleted=false and ps."index"=1 and ps.deleted=false), 'FAIL_ON_ERROR','STRING','Pass true/false to fail the pipeline or not ',true,false,'INPUT','NEW','false',1 ,'f','now()', 1, 'now()', 1);

