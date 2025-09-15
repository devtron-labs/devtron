/*
 * Copyright (c) 2025. Devtron Inc.
 */


-- Update the existing entry is_latest false for terrfaorm plugin
UPDATE plugin_metadata SET is_latest = false WHERE id = (SELECT id FROM plugin_metadata WHERE name= 'Terraform CLI v1.0.0' and is_latest= true);

-- Insert an entry in plugin_metadata for new version v1.0.1
INSERT INTO "plugin_metadata" ("id", "name", "description","deleted", "created_on", "created_by", "updated_on", "updated_by","plugin_parent_metadata_id","plugin_version","is_deprecated","is_latest")
VALUES (nextval('id_seq_plugin_metadata'), 'Terraform CLI v1.0.0','Terraform: Simplify infrastructure as code, manage resources effortlessly, and deploy with ease using this plugins','f', 'now()', 1, 'now()', 1, (SELECT id FROM plugin_parent_metadata WHERE identifier='terraform-cli-v1-0-0'),'1.0.1', false, true);

-- Insert the plugin stage mapping 
INSERT INTO "plugin_stage_mapping" ("plugin_id","stage_type","created_on", "created_by", "updated_on", "updated_by")
VALUES ((SELECT id FROM plugin_metadata WHERE plugin_version='1.0.1' and name='Terraform CLI v1.0.0' and deleted= false),0,'now()', 1, 'now()', 1);

-- Insert updated script 
INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_pipeline_script'),E'
export DEFAULT_TF_IMAGE=docker.io/hashicorp/terraform:latest
if [ -n "$TERRAFORM_IMAGE" ]; then
    echo "Using $TERRAFORM_IMAGE as the custom image."
    DEFAULT_TF_IMAGE="$TERRAFORM_IMAGE"
else
    echo "Using the default image --> $DEFAULT_TF_IMAGE"
fi

# RUNNING Terraform init 
if [ $RUN_TERRAFORM_INIT == "TRUE" ]; then 
    #RUNNING Terraform init 
    docker run -v $PWD:$PWD -w $PWD/$WORKINGDIR  -e HTTP_PROXY=$HTTP_PROXY -e HTTPS_PROXY=$HTTPS_PROXY -e NO_PROXY=$NO_PROXY $DEFAULT_TF_IMAGE init
fi


# exporting all the env variables 
echo "$ADDITIONALPARAMS" > devtron-custom-values.env

# RUNNING Terraform command 
echo "docker run -v $PWD:$PWD -w $PWD/$WORKINGDIR --env-file devtron-custom-values.env -e HTTP_PROXY=$HTTP_PROXY -e HTTPS_PROXY=$HTTPS_PROXY -e NO_PROXY=$NO_PROXY $DEFAULT_TF_IMAGE $ARGS"
docker run -v $PWD:$PWD -w $PWD/$WORKINGDIR  --env-file devtron-custom-values.env -e HTTP_PROXY=$HTTP_PROXY -e HTTPS_PROXY=$HTTPS_PROXY -e NO_PROXY=$NO_PROXY $DEFAULT_TF_IMAGE $ARGS','SHELL','f','now()',1,'now()',1);

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Terraform CLI v1.0.0' and plugin_version='1.0.1),'Step 1','Step 1 - Terraform CLI v1.0.0','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Terraform CLI v1.0.0' and ps."index"=1 and p.plugin_version='1.0.1' and ps.deleted=false),'HTTP_PROXY','STRING','Specify the HTTP proxy server for non-SSL requests.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Terraform CLI v1.0.0' and ps."index"=1 and p.plugin_version='1.0.1' and ps.deleted=false),'HTTPS_PROXY','STRING','Specify the HTTPS proxy server for SSL requests.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Terraform CLI v1.0.0' and ps."index"=1 and p.plugin_version='1.0.1' and ps.deleted=false),'NO_PROXY','STRING','Opt out of proxying HTTP/HTTPS requests. Use this to specify hosts that should bypass the proxy.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Terraform CLI v1.0.0' and ps."index"=1 and p.plugin_version='1.0.1' and ps.deleted=false),'TERRAFORM_IMAGE','STRING','Specify a custom Terraform container image to use for the execution of Terraform commands.','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Terraform CLI v1.0.0' and ps."index"=1 and p.plugin_version='1.0.1' and ps.deleted=false),'WORKINGDIR','STRING','Set the source directory for Terraform execution.Example: /path/to/terraform/project','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Terraform CLI v1.0.0' and ps."index"=1 and p.plugin_version='1.0.1' and ps.deleted=false),'ARGS','STRING','Specifies Terraform CLI commands to run. Example: plan -var-file=myvars.tfvars','t','t','--help',null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Terraform CLI v1.0.0' and ps."index"=1 and p.plugin_version='1.0.1' and ps.deleted=false),'RUN_TERRAFORM_INIT','BOOL','Determines whether to run the Terraform initialization command.','t','t','TRUE',null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Terraform CLI v1.0.0' and ps."index"=1 and p.plugin_version='1.0.1' and ps.deleted=false),'ADDITIONALPARAMS','STRING','Provides key-value pairs to inject into the Terraform container.Example: VAR1=value1 VAR2=value2c','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);



