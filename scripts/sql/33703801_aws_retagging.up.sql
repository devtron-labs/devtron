INSERT INTO "plugin_parent_metadata" ("id", "name","identifier", "description","type","icon","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_parent_metadata'), 'AWS Retag','aws-retag','AWS Retag is a functionality to retag','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/aws-signer-image.jpeg','f', 'now()', 1, 'now()', 1);


INSERT INTO "plugin_metadata" ("id", "name", "description","deleted", "created_on", "created_by", "updated_on", "updated_by","plugin_parent_metadata_id","plugin_version","is_deprecated","is_latest")
VALUES (nextval('id_seq_plugin_metadata'), 'AWS Retag','Re tag your image with AWS Retag','f', 'now()', 1, 'now()', 1, (SELECT id FROM plugin_parent_metadata WHERE identifier='aws-retag'),'1.0.0', false, true);

INSERT INTO "plugin_stage_mapping" ("plugin_id","stage_type","created_on", "created_by", "updated_on", "updated_by")
VALUES ((SELECT id FROM plugin_metadata WHERE plugin_version='1.0.0' and name='AWS Retag' and deleted= false),0,'now()', 1, 'now()', 1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")VALUES (
    nextval('id_seq_plugin_pipeline_script'),
    E'
    #!/bin/sh 
    set -eo pipefail 
    #set -v  ## uncomment this to debug the script 
    pipeline_type=$(echo $CI_CD_EVENT | jq -r \'.type\')
    if [[ "$pipeline_type" == "CI" ]]; then
        image_repo=$(echo $CI_CD_EVENT | jq -r  .commonWorkflowRequest.dockerRepository)
        image_tag=$(echo $CI_CD_EVENT | jq -r .commonWorkflowRequest.dockerImageTag)
        MANIFEST=$(aws ecr batch-get-image --repository-name $image_repo --image-ids imageTag=$image_tag --region ap-south-1 --output text --query \'images[].imageManifest\')
        aws ecr put-image --repository-name $image_repo --image-tag=$CustomTag --image-manifest "$MANIFEST" --region ap-south-1
    elif [[ "$pipeline_type" == "CD" ]]; then
        image_repo=$(echo $CI_CD_EVENT | jq -r .commonWorkflowRequest.ciArtifactDTO.image | cut -d\'/\' -f2 | cut -d':' -f1)
        image_tag=$(echo $CI_CD_EVENT | jq -r .commonWorkflowRequest.ciArtifactDTO.image | cut -d \':\' -f2)
        MANIFEST=$(aws ecr batch-get-image --repository-name $image_repo --image-ids imageTag=$image_tag --region ap-south-1 --output text --query \'images[].imageManifest\')
        aws ecr put-image --repository-name $image_repo --image-tag=$CustomTag --image-manifest "$MANIFEST" --region ap-south-1
    else
        echo "No able to re-tag the image"
    fi      
    
    ' ,'SHELL','f','now()',1,'now()',1);


INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by") 
VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='AWS Retag' AND plugin_version='1.0.0' AND deleted= false),'Step 1','Runnig the plugin','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='AWS Retag' and p.plugin_version='1.0.0' and ps."index"=1 and ps.deleted=false),'CustomTag','STRING','Provide the  tag for retagging','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);