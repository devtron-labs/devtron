INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'DockerSlim v1.1.0','This plugin is used to Slim the docker images (Currently this plugin can be used only for docker images not for docker buildx images).','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/dockerslim-plugin-icon.png',false,'now()',1,'now()',1);

INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_stage_mapping'),(SELECT id from plugin_metadata where name='DockerSlim v1.1.0'), 0,'now()',1,'now()',1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (
     nextval('id_seq_plugin_pipeline_script'),
        $$#!/bin/sh
httpProbe=$(echo "$HTTPProbe" | tr "[:upper:]" "[:lower:]")
includeFilePath=$IncludePathFile

export tag=$(echo $CI_CD_EVENT | jq --raw-output .commonWorkflowRequest.dockerImageTag)
export repo=$(echo $CI_CD_EVENT | jq --raw-output .commonWorkflowRequest.dockerRepository)

cd /devtroncd

docker pull dslim/slim

if [ "$httpProbe" == "true" ]; then
    if [ -n "$includeFilePath" ]; then
        docker run -i --rm -v /var/run/docker.sock:/var/run/docker.sock -v $PWD:$PWD dslim/slim build --http-probe=true --target $repo:$tag --tag $repo:$tag --continue-after=2 --include-path-file $includeFilePath
    else
        docker run -i --rm -v /var/run/docker.sock:/var/run/docker.sock -v $PWD:$PWD dslim/slim build --http-probe=true --target $repo:$tag --tag $repo:$tag --continue-after=2
    fi
elif [ -n "$includeFilePath" ]; then
    docker run -i --rm -v /var/run/docker.sock:/var/run/docker.sock -v $PWD:$PWD dslim/slim build --http-probe=false --target $repo:$tag --tag $repo:$tag --continue-after=2 --include-path-file $includeFilePath
else
    docker run -i --rm -v /var/run/docker.sock:/var/run/docker.sock -v $PWD:$PWD dslim/slim build --http-probe=false --target $repo:$tag --tag $repo:$tag --continue-after=2
fi

# Check the exit code of the last command
if [ $? -eq 0 ]; then
    echo "-----------***** Success: Docker-slim images built successfully *****-----------"
else
    echo "-----------***** Error: Docker-slim build failed, we are pushing original image to the container registry *****-----------"
fi$$,
        'SHELL',
        'f',
        'now()',
        1,
        'now()',
        1
);

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='DockerSlim v1.1.0'),'Step 1','Step 1 - DockerSlim','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='DockerSlim v1.1.0' and ps."index"=1 and ps.deleted=false),'HTTPProbe','BOOL','Is port expose or not in Dockerfile','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='DockerSlim v1.1.0' and ps."index"=1 and ps.deleted=false),'IncludePathFile','STRING','File path contains including path for dockerslim build flag --include-path-file','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);