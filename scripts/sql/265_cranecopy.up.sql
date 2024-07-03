INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'CraneCopy','The Crane Copy plugin can be used to copy container images from one registry to another.The Plugin can only be used in Post build Stage.','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/cranecopy.png',false,'now()',1,'now()',1);

INSERT INTO plugin_stage_mapping (id,plugin_id,stage_type,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_stage_mapping'),(SELECT id from plugin_metadata where name='CraneCopy'), 0,'now()',1,'now()',1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (
     nextval('id_seq_plugin_pipeline_script'),
        $$#!/bin/sh 
set -eo pipefail 

type=$(echo $CI_CD_EVENT | jq -r '.type')
if [[ "$type" == "CD" ]]; then
    echo "You are in Deployment stage,the plugin can only be used in Post Build Stage"
    exit 1
fi

echo "##################################################
#                                                 #
#         CRANE COPY PLUGIN IS RUNNING...         #
#                                                 #
###################################################
"
targetRegistry="$TargetRegistry"
targetRepo="${targetRegistry#*/}"
username="$RegistryUsername"
password="$RegistryPassword"
sourcerepo=$(echo "$CI_CD_EVENT" | jq -r '.commonWorkflowRequest.dockerRepository')
sourceregistry=$(echo "$CI_CD_EVENT" | jq -r '.commonWorkflowRequest.dockerRegistryURL')
export sourcekey="$(echo "$CI_CD_EVENT" | jq -r '.commonWorkflowRequest.accessKey')"
export sourcepassd="$(echo "$CI_CD_EVENT" | jq -r '.commonWorkflowRequest.secretKey')"
export sourceregion="$(echo "$CI_CD_EVENT" | jq -r '.commonWorkflowRequest.awsRegion')"
Tag=$(echo "$CI_CD_EVENT" | jq -r '.commonWorkflowRequest.dockerImageTag')
sourcepass=$(echo "$CI_CD_EVENT" | jq -r '.commonWorkflowRequest.dockerPassword')
sourceuser=$(echo "$CI_CD_EVENT" | jq -r '.commonWorkflowRequest.dockerUsername')
source_is_ecr=false
target_is_ecr=false
if [[ "$sourceregistry" == *"amazonaws.com"* ]]; then
    source_is_ecr=true
fi

if [[ "$targetRegistry" == *"amazonaws.com"* ]]; then
    target_is_ecr=true
fi

if [[ "$targetRegistry" == *"pkg.dev"* ]]; then
    echo $RegistryPassword > output.txt
    cat output.txt| base64 -d > key.json  
    auth=$(docker run --rm --name gcloud-config -v "$(pwd)/key.json":/key.json gcr.io/google.com/cloudsdktool/google-cloud-cli:alpine /bin/bash -c 'gcloud auth login --no-launch-browser --cred-file="/key.json" && gcloud auth print-access-token' | tail -n1)
    username=oauth2accesstoken
    password=$auth
fi

if $source_is_ecr && $target_is_ecr; then
    region="${targetRegistry##*.dkr.ecr.}"
    region="${region%%.*}"
    export region
    export AWS_ACCESS_KEY_ID="$username"
    export AWS_SECRET_ACCESS_KEY="$password"
    aws_auth=$(docker run --rm -e AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" -e AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" -e AWS_DEFAULT_REGION="$region" amazon/aws-cli ecr get-login-password --region "$region")
    aws_pass=$(docker run --rm -e AWS_ACCESS_KEY_ID="$sourcekey" -e AWS_SECRET_ACCESS_KEY="$sourcepassd" -e AWS_DEFAULT_REGION="$sourceregion" amazon/aws-cli ecr get-login-password --region "$sourceregion")
    docker run --rm --entrypoint /busybox/sh gcr.io/go-containerregistry/crane:debug -c " \
        mkdir ytr && \
        crane auth login -u AWS -p '$aws_pass' '$sourceregistry' && \
        crane pull $sourceregistry/$sourcerepo:$Tag /ytr --platform=all --format=oci && \
        crane auth login -u AWS -p '$aws_auth' '${targetRegistry%%/*}' && \
        crane push /ytr '$targetRegistry':$Tag && \
        echo -e '\nSuccessfully copied image from $sourceregistry/$sourcerepo:$Tag to $targetRegistry:$Tag' && \
        echo -e '\nImage Details:' && \
        echo -e 'Repository: $targetRepo' && \
        echo -e 'Tag: $Tag' && \
        echo -e 'Image Digest:' && \
        crane digest '$targetRegistry:$Tag'"
    docker login -u AWS -p $aws_pass $sourceregistry
elif $target_is_ecr; then
    region="${targetRegistry##*.dkr.ecr.}"
    region="${region%%.*}"
    export region
    export AWS_ACCESS_KEY_ID="$username"
    export AWS_SECRET_ACCESS_KEY="$password"
    aws_auth=$(docker run --rm -e AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" -e AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" -e AWS_DEFAULT_REGION="$region" amazon/aws-cli ecr get-login-password --region "$region")
    docker run --rm --entrypoint /busybox/sh -v /root/.docker:/root/.docker gcr.io/go-containerregistry/crane:debug -c " \
        mkdir ytr && \
        crane pull $sourceregistry/$sourcerepo:$Tag /ytr --platform=all --format=oci && \
        echo  "${targetRegistry%%/*}"  && \
        crane auth login -u AWS -p '$aws_auth' '${targetRegistry%%/*}' && \
        crane push /ytr '$targetRegistry':$Tag && \
        echo -e '\nSuccessfully copied image from $sourceregistry/$sourcerepo:$Tag to $targetRegistry:$Tag' && \
        echo -e '\nImage Details:' && \
        echo -e 'Repository: $targetRepo' && \
        echo -e 'Tag: $Tag' && \
        echo -e 'Image Digest:' && \
        crane digest '$targetRegistry:$Tag'"
    if [[ "$sourceregistry" == *"pkg.dev"* ]]; then
    echo "$CI_CD_EVENT" | jq -r .commonWorkflowRequest.dockerPassword | tr -d "'" > gcld.json
    wauth=$(docker run --rm --name gcloud-config -v "$(pwd)/gcld.json":/gcld.json gcr.io/google.com/cloudsdktool/google-cloud-cli:alpine /bin/bash -c 'gcloud auth login --no-launch-browser --cred-file="/gcld.json" && gcloud auth print-access-token' | tail -n1)
    docker login -u oauth2accesstoken -p $wauth $sourceregistry
    else
    docker login -u $sourceuser -p $sourcepass $sourceregistry
    fi

elif $source_is_ecr; then
    export sourcekey="$(echo "$CI_CD_EVENT" | jq -r '.commonWorkflowRequest.accessKey')"
    export sourcepassd="$(echo "$CI_CD_EVENT" | jq -r '.commonWorkflowRequest.secretKey')"
    export sourceregion="$(echo "$CI_CD_EVENT" | jq -r '.commonWorkflowRequest.awsRegion')"
    aws_pass=$(docker run --rm -e AWS_ACCESS_KEY_ID="$sourcekey" -e AWS_SECRET_ACCESS_KEY="$sourcepassd" -e AWS_DEFAULT_REGION="$sourceregion" amazon/aws-cli ecr get-login-password --region "$sourceregion" )
    docker run --rm --entrypoint /busybox/sh gcr.io/go-containerregistry/crane:debug -c " \
        mkdir ytr && \
        crane auth login -u AWS -p '$aws_pass' '$sourceregistry' && \
        crane pull $sourceregistry/$sourcerepo:$Tag /ytr --platform=all --format=oci && \
        crane auth login -u "$username" -p "$password" "${targetRegistry%%/*}" && \
        crane push /ytr \"$targetRegistry\":$Tag && \
        echo -e '\nSuccessfully copied image from $sourceregistry/$sourcerepo:$Tag to $targetRegistry:$Tag' && \
        echo -e '\nImage Details:' && \
        echo -e 'Repository: $targetRepo' && \
        echo -e 'Tag: $Tag' && \
        echo -e 'Image Digest:' && \
        crane digest \"$targetRegistry:$Tag\""
    docker login -u AWS -p $aws_pass $sourceregistry
else
    docker run --rm --entrypoint /busybox/sh -v /root/.docker:/root/.docker gcr.io/go-containerregistry/crane:debug -c " \
        mkdir ytr && \
        crane pull $sourceregistry/$sourcerepo:$Tag /ytr --platform=all --format=oci && \
        crane auth login -u "$username" -p "$password" "${targetRegistry%%/*}" && \
        crane push /ytr \"$targetRegistry\":$Tag && \
        echo -e '\nSuccessfully copied image from $sourceregistry/$sourcerepo:$Tag to $targetRegistry:$Tag' && \
        echo -e '\nImage Details:' && \
        echo -e 'Repository: $targetRepo' && \
        echo -e 'Tag: $Tag' && \
        echo -e 'Image Digest:' && \
        crane digest \"$targetRegistry:$Tag\""
    if [[ "$sourceregistry" == *"pkg.dev"* ]]; then
    echo "$CI_CD_EVENT" | jq -r .commonWorkflowRequest.dockerPassword | tr -d "'" > gcld.json
    wauth=$(docker run --rm --name gcloud-config -v "$(pwd)/gcld.json":/gcld.json gcr.io/google.com/cloudsdktool/google-cloud-cli:alpine /bin/bash -c 'gcloud auth login --no-launch-browser --cred-file="/gcld.json" && gcloud auth print-access-token' | tail -n1)
    docker login -u oauth2accesstoken -p $wauth $sourceregistry
    else
    docker login -u $sourceuser -p $sourcepass $sourceregistry
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
VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='CraneCopy'),'Step 1','Step 1 - CraneCopy','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='CraneCopy' and ps."index"=1 and ps.deleted=false),'TargetRegistry','STRING','The target registry to push the image.In the format taregtregistry.com/repo','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='CraneCopy' and ps."index"=1 and ps.deleted=false),'RegistryUsername','STRING','The username for authentication.(Provide AWS Access key ID in case of ECR)','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='CraneCopy' and ps."index"=1 and ps.deleted=false),'RegistryPassword','STRING','The password to the registry for authentication.(Provide AWS Secret Access key in case of ECR).','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);
