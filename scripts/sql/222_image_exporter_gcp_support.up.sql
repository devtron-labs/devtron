UPDATE plugin_pipeline_script SET script=E'#!/bin/bash
    set -eo pipefail
if [[ $GoogleServiceAccount ]] 
then   
    echo $GoogleServiceAccount > output.txt
    cat output.txt| base64 -d > gcloud.json    
fi
architecture=$(uname -m)
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
    if [[ $architecture == "x86_64" ]]
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
    file=$ContainerImage.tar
    file=$(echo $file | tr \'/\' \'_\')
else
    file=$FilePrefix-$ContainerImage.tar
    file=$(echo $file | tr \'/\' \'_\')
fi   
echo $file     
future_timestamp=$((current_timestamp + $Expiry * 60))
future_date=$(date -u  -d@"$future_timestamp" +"%Y-%m-%dT%H:%M:%SZ")
aws_secs=$(($Expiry * 60))
gcp_secs="${Expiry}m"
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
    docker run --network=host --rm -v $(pwd):/data -e AWS_ACCESS_KEY_ID=$AwsAccessKey -e AWS_SECRET_ACCESS_KEY=$AwsSecretKey public.ecr.aws/aws-cli/aws-cli:latest  s3 cp /data/$file s3://$BucketName --region $AwsRegion
    link=$(docker  run  --network=host --rm -v $(pwd):/data -e AWS_ACCESS_KEY_ID=$AwsAccessKey -e AWS_SECRET_ACCESS_KEY=$AwsSecretKey public.ecr.aws/aws-cli/aws-cli:latest s3 presign s3://$BucketName/$file --region $AwsRegion --expires-in $aws_secs )
fi
if [ $CloudProvider == "gcp" ]
then
    echo "gcp command"
    docker run  --rm  -v $(pwd):/data  kamal1109/gcloud:latest  /bin/bash -c "gcloud auth activate-service-account --key-file=data/gcloud.json;gcloud config set project $GcpProjectName; gcloud storage ls;gsutil cp data/$file gs://$BucketName/ ; gcloud storage ls gs://$BucketName/;"
    link=$(docker run  --rm  -v $(pwd):/data  kamal1109/gcloud:latest  /bin/bash -c "gcloud auth activate-service-account --key-file=data/gcloud.json;gcloud config set project $GcpProjectName; gsutil signurl -d $gcp_secs data/gcloud.json gs://$BucketName/$file " )
fi
echo "***Copy the below link to download the tar file***"
echo $link
' WHERE id=(select script_id  from plugin_step where id=(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Container Image Exporter v1.0.0' and ps."index"=1 and ps.deleted=false));


INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Container Image Exporter v1.0.0' and ps."index"=1 and ps.deleted=false), 'GoogleServiceAccount','STRING','Provide Google service account Creds/Use Scope Variable ',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Container Image Exporter v1.0.0' and ps."index"=1 and ps.deleted=false), 'GcpProjectName','STRING','Specify Google Account Project Name',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

UPDATE plugin_step_variable SET description='Provide which cloud storage provider you want to use: "aws" for Amazon S3 or "azure" for Azure Blob Storage or "gcp" for Google Cloud Storage' WHERE name='CloudProvider';