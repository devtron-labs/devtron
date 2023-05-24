INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'Jira Issue Updater','This plugin extends the capabilities of Devtron CI and can update issues in JIRA by adding pipeline status and metadata as comment on the tickets.','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/plugin-jira.png',false,'now()',1,'now()',1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (
   nextval('id_seq_plugin_pipeline_script'),
   '#!/bin/sh
if [[ $UpdateWithBuildStatus == true ]]
then
	# step-1 -> updating the jira issue with build status
	echo -e "\033[1m======== Updating the Jira issue with build status ========"
    buildStatusMessage="Failed"
    if [[ $BuildSuccess == true ]]
    then
        buildStatusMessage="Succeeded"
    fi
	curl -u $JiraUsername:$JiraPassword -X PUT $JiraBaseUrl/rest/api/2/issue/$JiraId -H "Content-Type: application/json" -d ''{"update": {"comment": [{"add":{"body":"''"Build status : $buildStatusMessage"''"}}]}}''

	if [ $? != 0 ]; then
	   echo -e "\033[1m======== Updating the jira Jira with build status failed ========"
	   exit 1
	fi
fi

if [[ $UpdateWithDockerImageId == true && $BuildSuccess == true ]]
then
	# step-2 -> updating the jira issue with docker image Id
	echo -e "\033[1m======== Updating the Jira issue with docker image Id ========"
	curl -u $JiraUsername:$JiraPassword -X PUT $JiraBaseUrl/rest/api/2/issue/$JiraId -H "Content-Type: application/json" -d ''{"update": {"comment": [{"add":{"body":"''"Image built : $DockerImage"''"}}]}}''

	if [ $? != 0 ]; then
	   echo -e "\033[1m======== Updating the jira Jira with docker image Id failed ========"
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
VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Jira Issue Updater'),'Step 1','Step 1 - Jira Issue Updater','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Jira Issue Updater' and ps."index"=1 and ps.deleted=false), 'JiraUsername','STRING','Username of Jira account',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Jira Issue Updater' and ps."index"=1 and ps.deleted=false), 'JiraPassword','STRING','Password of Jira account',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Jira Issue Updater' and ps."index"=1 and ps.deleted=false), 'JiraBaseUrl','STRING','Base Url of Jira account',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "default_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Jira Issue Updater' and ps."index"=1 and ps.deleted=false), 'UpdateWithDockerImageId','BOOL','If true - Jira Issue will be updated with docker image Id in comment. Default: true',true,true, true, 'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Jira Issue Updater' and ps."index"=1 and ps.deleted=false), 'UpdateWithBuildStatus','BOOL','If true - Jira Issue will be updated with build status in comment. Default: true',true,true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Jira Issue Updater' and ps."index"=1 and ps.deleted=false), 'JiraId','STRING','Jira Id',false,true,3,'INPUT','GLOBAL',1 ,'JIRA_ID','f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Jira Issue Updater' and ps."index"=1 and ps.deleted=false), 'DockerImage','STRING','Docker Image',false,true,3,'INPUT','GLOBAL',1 ,'DOCKER_IMAGE','f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Jira Issue Updater' and ps."index"=1 and ps.deleted=false), 'BuildSuccess','BOOL','Build Success',false,true,3,'INPUT','GLOBAL',1 ,'BUILD_SUCCESS','f','now()', 1, 'now()', 1);