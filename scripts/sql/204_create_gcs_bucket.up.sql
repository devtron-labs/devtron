INSERT INTO "plugin_metadata" ("id", "name", "description","type","icon","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_metadata'), 'GCS Create Bucket','Plugin to create GCS bucket','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/gcs-bucket.svg','f', 'now()', 1, 'now()', 1);

INSERT INTO "plugin_tag" ("id", "name", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_tag'), 'cloud','f', 'now()',1, 'now()', 1);
INSERT INTO "plugin_tag" ("id", "name", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_tag'), 'gcs','f', 'now()',1, 'now()', 1);

INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_tag_relation'),(SELECT id FROM plugin_tag WHERE name='cloud') , (SELECT id FROM plugin_metadata WHERE name='GCS Create Bucket'),'now()', 1, 'now()', 1);

INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_tag_relation'),(SELECT id FROM plugin_tag WHERE name='gcs') , (SELECT id FROM plugin_metadata WHERE name='GCS Create Bucket'),'now()', 1, 'now()', 1);

INSERT INTO "plugin_stage_mapping" ("plugin_id","stage_type","created_on", "created_by", "updated_on", "updated_by")
VALUES ((SELECT id FROM plugin_metadata WHERE name='GCS Create Bucket'),0,'now()', 1, 'now()', 1);

INSERT INTO "plugin_pipeline_script" ("id","type","script","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_pipeline_script'),'SHELL',E'cat << "EOF" > /devtroncd/create-gs-bucket.sh
#!/bin/bash 
set -e pipefail 
echo $ServiceAccountCred | base64 -d > /tmp/cred.json
CRED_PATH=/tmp/cred.json
if [[ -f "$CRED_PATH" ]]
then
GOOGLE_APPLICATION_CREDENTIALS="$CRED_PATH"
fi
if [[ "${GOOGLE_APPLICATION_CREDENTIALS}" != "" ]]
then
echo "GOOGLE_APPLICATION_CREDENTIALS is set, activating Service Account..."
gcloud auth activate-service-account --key-file=${GOOGLE_APPLICATION_CREDENTIALS}
fi

if [[ "$LocationType" != "dual-region" ]]
then 
    if [[ "$Location" != "" ]]
    then
        LOCATION="-l $Location"
    else
        LOCATION="-l us"
    fi
elif [[ "$LocationType" == "dual-region" && "$Location" != "" ]]
then
    LOCATION="--placement $Location"
fi

if [[ "$EnableAutoClass" == "true" ]]
then
    STORAGECLASS="--autoclass"
elif [[ "$EnableAutoClass" == "false" && "$StorageClass" != "" ]]
then
    STORAGECLASS="-c $StorageClass"
else
    STORAGECLASS="-c standard"
fi

if [[ "$UniformAccess" == "true" ]]
then
    ACCESSCONTROL="-b on"
else
    ACCESSCONTROL="-b off"
fi

if [[ "$EnableBucketPrefix" == true ]]
then
    export BucketName="$BucketName-$(date +%d-%m-%Y-%H-%M)"
else
    export BucketName="$BucketName"
fi

echo "Creating $BucketName in $Project Project"

echo "gsutil mb -p $Project $ACCESSCONTROL $LOCATION $STORAGECLASS gs://$BucketName"
gsutil mb -p $Project $ACCESSCONTROL $LOCATION $STORAGECLASS gs://$BucketName

if [[ $? == 0 ]]
then
    echo "Bucket created succeefully"
else 
    exit
fi
EOF

docker run -e ServiceAccountCred=$ServiceAccountCred \\
    -e LocationType=$LocationType \\
    -e Location=$Location \\
    -e EnableAutoClass=$EnableAutoClass \\
    -e StorageClass=$StorageClass \\
    -e UniformAccess=$UniformAccess \\
    -e Project=$Project \\
    -e BucketName=$BucketName \\
    -e EnableBucketPrefix=$EnableBucketPrefix \\
    -v /devtroncd:/mnt google/cloud-sdk bash -c "bash /mnt/create-gs-bucket.sh"
','f','now()',1,'now()',1);

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='GCS Create Bucket'),'Step 1','GCS Create Bucket','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GCS Create Bucket' and ps."index"=1 and ps.deleted=false), 'ServiceAccountCred','STRING','base64 encoded credentials of service account key file.',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GCS Create Bucket' and ps."index"=1 and ps.deleted=false), 'Project','STRING','The project with which your bucket will be associated.',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","default_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GCS Create Bucket' and ps."index"=1 and ps.deleted=false), 'LocationType','STRING','Defines the geographic placement of your data(Possible values:-region,multi-region,dual-region. default multi-region). Cannot be changed later..',true,true,'multi-region','INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GCS Create Bucket' and ps."index"=1 and ps.deleted=false), 'Location','STRING','The Region, for the new bucket default US.',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","default_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GCS Create Bucket' and ps."index"=1 and ps.deleted=false), 'EnableAutoClass','BOOL','If enabled automatically transitions each object to Standard or Nearline class based on object-level activity, to optimize for cost and latency. default disabled',true,true,'false','INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GCS Create Bucket' and ps."index"=1 and ps.deleted=false), 'EnableBucketPrefix','STRING','If enabled bucket name will be used as prefix.',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GCS Create Bucket' and ps."index"=1 and ps.deleted=false), 'BucketName','STRING','The bucket name to be created. If EnableBucketPrefix is enabled, a random string will be suffixed to the name.',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GCS Create Bucket' and ps."index"=1 and ps.deleted=false), 'StorageClass','STRING','The storage class for the new bucket. standard, nearline, coldline, or archive.',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GCS Create Bucket' and ps."index"=1 and ps.deleted=false), 'UniformAccess','STRING','Set this to "true" if the bucket should be created with bucket-level permissions instead of Access Control Lists.',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO plugin_step_variable ("id","plugin_step_id","name","format","description","is_exposed","allow_empty_value","default_value","value","variable_type","value_type","previous_step_index","variable_step_index","variable_step_index_in_plugin","reference_variable_name","deleted","created_on","created_by","updated_on","updated_by")
VALUES (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GCS Create Bucket' and ps."index"=1 and ps.deleted=false),'BucketName','STRING','The name of the bucket createed.','t','f',false,null,'OUTPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);