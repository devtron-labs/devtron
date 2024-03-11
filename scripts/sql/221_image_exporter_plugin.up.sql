-- Plugin metadata
INSERT INTO "plugin_metadata" ("id", "name", "description","type","icon","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_metadata'), 'Container Image Exporter v1.0.0','Generate a Tarball Archive for Your Docker Images','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/docker-logo.png','f', 'now()', 1, 'now()', 1);

-- Plugin Stage Mapping

INSERT INTO "plugin_stage_mapping" ("plugin_id","stage_type","created_on", "created_by", "updated_on", "updated_by")
VALUES ((SELECT id FROM plugin_metadata WHERE name='Container Image Exporter v1.0.0'),0,'now()', 1, 'now()', 1);

-- Plugin Script

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (
    nextval('id_seq_plugin_pipeline_script'),
    E'#!/bin/bash
    set -eo pipefail

export platform=$(echo $CI_CD_EVENT | jq --raw-output .commonWorkflowRequest.ciBuildConfig.dockerBuildConfig.targetPlatform)
echo $platform
arch
if [[ $platform == "linux/arm64,linux/amd64" ]] ; then
    platform=$Platform
elif [[ $platform == "linux/arm64" ]]
then
    platform="arm64"
elif [[ $platform == "linux/amd64" ]]  
then
    platform="amd64"
else
    if [[ arch == "x86_64" ]]
    then
        platform="amd64"
    else
        platform="arm64"
    fi    
fi        
echo $platform   
CloudProvider=$(echo "$CloudProvider" | awk \'{print tolower($0)}\')
current_timestamp=$(date +%s)
if [[ -z $FilePrefix ]]
then
    file=$ContainerImage-$platform.tar
    file=$(echo $file | tr \'/\' \'_\')
else
    file=$FilePrefix-$platform-$ContainerImage.tar
    file=$(echo $file | tr \'/\' \'_\')
fi   
echo $file     
future_timestamp=$((current_timestamp + $Expiry * 60))
future_date=$(date -u  -d@"$future_timestamp" +"%Y-%m-%dT%H:%M:%SZ")
aws_secs=$(($Expiry * 60))
docker pull --platform linux/$platform $ContainerImage
docker save $ContainerImage > $file
ls
if [ $CloudProvider == "azure" ]
then
    docker run  --rm  -v $(pwd):/data  mcr.microsoft.com/azure-cli  /bin/bash -c " az storage blob upload --account-name $AzureAccountName --account-key $AzureAccountKey  --container-name $BucketName --name $file --file data/$file"
    echo "docker run  --rm   mcr.microsoft.com/azure-cli  /bin/bash -c " az storage blob generate-sas --account-name $AzureAccountName --account-key $AzureAccountKey --container-name $BucketName --name $file --permissions r --expiry $future_date""
    sas_token=$(docker run  --rm   mcr.microsoft.com/azure-cli  /bin/bash -c " az storage blob generate-sas --account-name $AzureAccountName --account-key $AzureAccountKey --container-name $BucketName --name $file --permissions r --expiry $future_date")
    token=$sas_token
    echo $token 
    token=$(echo $sas_token| tr -d \'"\')
    echo $token
    link=https://$AzureAccountName.blob.core.windows.net/$BucketName/$file?$token
fi
if [ $CloudProvider == "aws" ]
then
    echo "aws command"
    docker run --rm -v $(pwd):/data -e AWS_ACCESS_KEY_ID=$AwsAccessKey -e AWS_SECRET_ACCESS_KEY=$AwsSecretKey amazon/aws-cli:latest  s3 cp /data/$file s3://$BucketName --region $AswRegion
    link=$(docker run --rm -v $(pwd):/data -e AWS_ACCESS_KEY_ID=$AwsAccessKey -e AWS_SECRET_ACCESS_KEY=$AwsSecretKey amazon/aws-cli:latest s3 presign s3://$BucketName/$file --region $AswRegion --expires-in $aws_secs )
fi
echo "***Copy the below link to download the tar file***"
echo $link
',
    'SHELL',
    'f',
    'now()',
    1,
    'now()',
    1
);


--Plugin Step

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Container Image Exporter v1.0.0'),'Step 1','Creating tarball of image','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);


-- Input Variables

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Container Image Exporter v1.0.0' and ps."index"=1 and ps.deleted=false), 'Platform','STRING',' Specify the platform architecture of the image being exported (arm64 or amd64).',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Container Image Exporter v1.0.0' and ps."index"=1 and ps.deleted=false), 'AwsRegion','STRING',' Specify the AWS region where your S3 bucket is located ',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Container Image Exporter v1.0.0' and ps."index"=1 and ps.deleted=false), 'AwsAccessKey','STRING',' Provide your AWS access key ID. ',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Container Image Exporter v1.0.0' and ps."index"=1 and ps.deleted=false), 'AwsSecretKey','STRING',' Provide your AWS secret access key. Keep this key confidential.',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Container Image Exporter v1.0.0' and ps."index"=1 and ps.deleted=false), 'AzureAccountKey','STRING',' Provide the access key for your Azure storage account.',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Container Image Exporter v1.0.0' and ps."index"=1 and ps.deleted=false), 'AzureAccountName','STRING',' Specify the name of your Azure storage account.',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Container Image Exporter v1.0.0' and ps."index"=1 and ps.deleted=false), 'FilePrefix','STRING','If youd like to add a prefix to the exported image files name,  enter it here.',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Container Image Exporter v1.0.0' and ps."index"=1 and ps.deleted=false), 'CloudProvider','STRING',' Provide which cloud storage provider you want to use: "aws" for Amazon S3 or "azure" for Azure Blob Storage.',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Container Image Exporter v1.0.0' and ps."index"=1 and ps.deleted=false), 'Expiry','STRING','Must be a whole number between 1 and 720 minutes',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Container Image Exporter v1.0.0' and ps."index"=1 and ps.deleted=false), 'BucketName','STRING',' Enter the name of the storage container where you want to upload the exported image file.',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Container Image Exporter v1.0.0' and ps."index"=1 and ps.deleted=false), 'ContainerImage','STRING','Provide the image from system variable',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);
