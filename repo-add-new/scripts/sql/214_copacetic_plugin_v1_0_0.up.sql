INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'Copacetic v1.0.0','This plugin is used to patch the container image vulnerabilities (Patching for Multi Architecture Builds not supported currently).','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/copa-plugin-icon.png',false,'now()',1,'now()',1);

INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_tag_relation'), (SELECT id FROM plugin_tag WHERE name='Security'), (SELECT id FROM plugin_metadata WHERE name='Copacetic v1.0.0'),'now()', 1, 'now()', 1);
INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_tag_relation'), (SELECT id FROM plugin_tag WHERE name='DevSecOps'), (SELECT id FROM plugin_metadata WHERE name='Copacetic v1.0.0'),'now()', 1, 'now()', 1);

INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_stage_mapping'),(SELECT id from plugin_metadata where name='Copacetic v1.0.0'), 0,'now()',1,'now()',1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (
     nextval('id_seq_plugin_pipeline_script'),
        $$#!/bin/sh

export appName=$(echo $CI_CD_EVENT | jq --raw-output .commonWorkflowRequest.appName)
export registry=$(echo $CI_CD_EVENT | jq --raw-output .commonWorkflowRequest.dockerRegistryURL)
export repo=$(echo $CI_CD_EVENT | jq --raw-output .commonWorkflowRequest.dockerRepository)
export tag=$(echo $CI_CD_EVENT | jq --raw-output .commonWorkflowRequest.dockerImageTag)
export platform=$(echo $CI_CD_EVENT | jq --raw-output .commonWorkflowRequest.ciBuildConfig.dockerBuildConfig.targetPlatform)

if [[ $platform == "linux/arm64,linux/amd64" ]] ; then
    echo "platform = $platform"
    echo "------------------------------------------------------------------------------------------------------------------------"
    echo "######### Patching Multi Architecture Image not supported #########"
    echo "------------------------------------------------------------------------------------------------------------------------"
    echo "error: can't use copacetic plugin for muti-architecture builds"
    exit 1;
else
    curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b /usr/local/bin v0.46.1

    uname_arch() {
    arch=$(uname -m)
    case $arch in
        x86_64) arch="amd64" ;;
        aarch64) arch="arm64" ;;
    esac
    echo ${arch}
    }
    os=$(uname | tr "[:upper:]" "[:lower:]")
    uname_arch
    wget https://github.com/project-copacetic/copacetic/releases/download/v0.5.1/copa_0.5.1_${os}_${arch}.tar.gz
    tar -xvzf copa_0.5.1_${os}_${arch}.tar.gz
    mv copa /usr/local/bin/

    trivy image --vuln-type os --ignore-unfixed $registry/$repo:$tag | grep -i total
    trivy image --vuln-type os --ignore-unfixed -f json -o $appName.json $registry/$repo:$tag

    export BUILDKIT_VERSION=v0.12.0
    docker run \
        --detach \
        --rm \
        --privileged \
        --name buildkitd \
        --entrypoint buildkitd \
        "moby/buildkit:$BUILDKIT_VERSION"

    copa patch -i $registry/$repo:$tag -r $appName.json -t $tag --addr docker-container://buildkitd --timeout "$CopaTimeout"
    if [ $? -eq 0 ] ; then
        trivy image --vuln-type os --ignore-unfixed $registry/$repo:$tag | grep -i total
        docker push $registry/$repo:$tag
    else
        echo "------------------------------------------------------------------------------------------------------------------------"
        echo "BUILD FAILED: non zero exit status during copa patch..."
        echo "------------------------------------------------------------------------------------------------------------------------"
        exit 1;
    fi
fi
$$,
        'SHELL',
        'f',
        'now()',
        1,
        'now()',
        1
);

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Copacetic v1.0.0'),'Step 1','Step 1 - Copacetic v1.0.0','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Copacetic v1.0.0' and ps."index"=1 and ps.deleted=false),'CopaTimeout','STRING','Timeout for copa patch command, default timeout is 5 minutes. For ex: 10m','t','t','5m',null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);
