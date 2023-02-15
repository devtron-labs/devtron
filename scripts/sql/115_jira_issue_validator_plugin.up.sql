INSERT INTO plugin_metadata (id,name,description,type,icon,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_metadata'),'Jira Issue Validator','Checks for a valid Jira Issue','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/plugin-jira-validation.png',false,'now()',1,'now()',1);

INSERT INTO "plugin_pipeline_script" ("id", "script","type","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (
   nextval('id_seq_plugin_pipeline_script'),
   '#!/bin/sh
# step-1 -> find the jira issue
echo "Finding the jira issue"
curl -u $JiraUsername:$JiraPassword $JiraBaseUrl/rest/api/2/issue/$JiraId > jira_issue_search_result.txt

if [ $? != 0 ]; then
   echo "Finding the jira issue failed"
   exit 1
fi

# step-2 -> converting to JSON
echo "Converting to json result"
cat jira_issue_search_result.txt | jq > jira_issue_search_result_json.txt

if [ $? != 0 ]; then
   echo "Converting to json result failed"
   exit 1
fi

# step-3 -> Find the error message from JSON result
echo "Finding the error message from JSON result"
jq ".errorMessages" jira_issue_search_result_json.txt > error_message.txt

if [ $? != 0 ]; then
   echo "Finding the error message from JSON result failed"
   exit 1
fi

# step-4 -> check if error message if null or not
echo "checking if error message is not null"

if [ null == "$(cat error_message.txt)" ] ;then
    echo "jira issue exists"
else
    echo "jira issue does not exist"
    exit 1
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
VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Jira Issue Validator'),'Step 1','Step 1 - Jira Issue Validator','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Jira Issue Validator' and ps."index"=1 and ps.deleted=false), 'JiraUsername','STRING','Username of Jira account',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Jira Issue Validator' and ps."index"=1 and ps.deleted=false), 'JiraPassword','STRING','Password of Jira account',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1),
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Jira Issue Validator' and ps."index"=1 and ps.deleted=false), 'JiraBaseUrl','STRING','Base Url of Jira account',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Jira Issue Validator' and ps."index"=1 and ps.deleted=false), 'JiraId','STRING','Jira Id',false,true,3,'INPUT','GLOBAL',1 ,'description_jiraId','f','now()', 1, 'now()', 1);