INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'Github Pull Request Updater','This plugin extends the capabilities of Devtron CI and can update pull requests in GITHUB by adding pipeline status and metadata as comment.','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/plugin-github-pr.png',false,'now()',1,'now()',1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (
   nextval('id_seq_plugin_pipeline_script'),
   '#!/bin/sh
if [[ $UpdateWithBuildStatus == true ]]
then
	# step-1 -> updating the PR with build status
	echo -e "\033[1m======== Commenting build status in PR ========"
    buildStatusMessage="Failed"
    if [[ $BuildSuccess == true ]]
    then
        buildStatusMessage="Succeeded"
    fi
	curl -X POST -H "Accept: application/vnd.github+json" -H "Authorization: Bearer $AccessToken" -H "X-GitHub-Api-Version: 2022-11-28" $CommentsUrl -d ''{"body": "''"Build status : $buildStatusMessage"''"}''

	if [ $? != 0 ]; then
	   echo -e "\033[1m======== Updating the PR with build status failed ========"
	   exit 1
	fi
fi

if [[ $UpdateWithDockerImageId == true && $BuildSuccess == true ]]
then
	# step-2 -> updating the PR with docker image Id
	echo -e "\033[1m======== Commenting docker image Id in PR ========"
    curl -X POST -H "Accept: application/vnd.github+json" -H "Authorization: Bearer $AccessToken" -H "X-GitHub-Api-Version: 2022-11-28" $CommentsUrl -d ''{"body": "''"Image built : $DockerImage"''"}''

	if [ $? != 0 ]; then
	   echo -e "\033[1m======== Updating the PR with docker image Id failed ========"
	   exit 1
	fi
fi
'
   ,
   'SHELL',
   'f',
   'now()',
   1,
   'now()',
   1
);

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Github Pull Request Updater'),'Step 1','Step 1 - Github Pull Request Updater','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Github Pull Request Updater' and ps."index"=1 and ps.deleted=false), 'AccessToken','STRING','Personal access token which will be used to authenticating to Github APIs for this plugin',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "default_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Github Pull Request Updater' and ps."index"=1 and ps.deleted=false), 'UpdateWithDockerImageId','BOOL','If true - PR will be updated with docker image Id in comment. Default: true',true,true, true, 'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Github Pull Request Updater' and ps."index"=1 and ps.deleted=false), 'UpdateWithBuildStatus','BOOL','If true - PR will be updated with build status in comment. Default: true',true,true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Github Pull Request Updater' and ps."index"=1 and ps.deleted=false), 'CommentsUrl','STRING','Comments url',false,true,3,'INPUT','GLOBAL',1 ,'COMMENTS_URL','f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Github Pull Request Updater' and ps."index"=1 and ps.deleted=false), 'DockerImage','STRING','Docker Image',false,true,3,'INPUT','GLOBAL',1 ,'DOCKER_IMAGE','f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Github Pull Request Updater' and ps."index"=1 and ps.deleted=false), 'BuildSuccess','BOOL','Build Success',false,true,3,'INPUT','GLOBAL',1 ,'BUILD_SUCCESS','f','now()', 1, 'now()', 1);