INSERT INTO "public"."plugin_metadata" ("id", "name", "description","type","icon","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_metadata'), 'Sonarqube PR Based','Enhance Your Workflow with Continuous Code Quality & Code Security For Pull Request.','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/sonarqube-plugin-icon.png','f', 'now()', '1', 'now()', '1');

INSERT INTO "public"."plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_tag_relation'), (SELECT id FROM plugin_tag WHERE name='Code quality'),(SELECT id FROM plugin_metadata WHERE name='Sonarqube PR Based'),'now()', '1', 'now()', '1'),
(nextval('id_seq_plugin_tag_relation'), (SELECT id FROM plugin_tag WHERE name='Security'),(SELECT id FROM plugin_metadata WHERE name='Sonarqube PR Based'),'now()', '1', 'now()', '1');

INSERT INTO "public"."plugin_pipeline_script" ("id", "script", "type","deleted","created_on", "created_by", "updated_on", "updated_by") VALUES(nextval('id_seq_plugin_pipeline_script'), E'PathToCodeDir=/devtroncd$CheckoutPath
cd $PathToCodeDir
if [[ -z "$UsePropertiesFileFromProject" || $UsePropertiesFileFromProject == false ]]
then
  echo "sonar.projectKey=$SonarqubeProjectKey" > sonar-project.properties
fi

echo "sonar.projectKey=$SonarqubeProjectKey" >> sonar-project.properties
export PR_SOURCE_BRANCH=$(echo $CI_CD_EVENT |jq \'.ciRequest.ciProjectDetails[0].WebhookData.Data."source branch name"\')
export PR_TARGET_BRANCH=$(echo $CI_CD_EVENT |jq \'.ciRequest.ciProjectDetails[0].WebhookData.Data."target branch name"\')
export PR_NUMBER=$(echo $CI_CD_EVENT |jq \'.ciRequest.ciProjectDetails[0].WebhookData.Data."pull request number"\')
export REPO_NAME=$(echo $CI_CD_EVENT |jq \'.ciRequest.ciProjectDetails[0].WebhookData.Data."github repo name"\')
echo sonar.pullrequest.key=$PR_NUMBER >> sonar-project.properties
echo sonar.pullrequest.branch=$PR_SOURCE_BRANCH >> sonar-project.properties
echo sonar.pullrequest.base=$PR_TARGET_BRANCH >> sonar-project.properties
echo sonar.pullrequest.github.repository=$REPO_NAME >> sonar-project.properties

docker run \\
--rm \\
-e SONAR_HOST_URL=$SonarqubeEndpoint \\
-e SONAR_TOKEN=$SonarqubeApiKey \\
-v "/$PWD:/usr/src" \\
sonarsource/sonar-scanner-cli

if [[ $CheckForSonarAnalysisReport == true && ! -z "$CheckForSonarAnalysisReport" ]]
then
 status=$(curl -u ${SonarqubeApiKey}:  -sS ${SonarqubeEndpoint}/api/qualitygates/project_status?projectKey=${SonarqubeProjectKey}&branch=master)
 project_status=$(echo $status | jq -r  ".projectStatus.status")
 export SonarScanResult=$project_status
 echo "*********  SonarQube Policy Report  *********"
 echo $status
 if [[ $AbortPipelineOnPolicyCheckFailed == true && $project_status == "ERROR" ]]
 then
  echo "*********  SonarQube Policy Violated *********"
  echo "*********  Exiting Build *********"
  exit
 elif [[ $AbortPipelineOnPolicyCheckFailed == true && $project_status == "OK" ]]
 then
  echo "*********  SonarQube Policy Passed *********"
 fi
fi','SHELL','f', 'now()', '1', 'now()', '1');

INSERT INTO "public"."plugin_step" (id,plugin_id,name,description,index,step_type,script_id,ref_plugin_id,output_directory_path,dependent_on_step,deleted,created_on,created_by,updated_on,updated_by)
VALUES (nextval('id_seq_plugin_step'),(SELECT id FROM plugin_metadata WHERE name='Sonarqube PR Based'),'Step 1','Step 1 for Sonarqube',1,'INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),null,null,null,false,'now()',1,'now()',1);

INSERT INTO "public"."plugin_step_variable" (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube PR Based' and ps."index"=1 and ps.deleted=false),'SonarqubeProjectKey','STRING','project key of grafana sonarqube account.','t','f',false,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube PR Based' and ps."index"=1 and ps.deleted=false),'SonarqubeApiKey','STRING','api key of sonarqube account.','t','f',false,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube PR Based' and ps."index"=1 and ps.deleted=false),'SonarqubeEndpoint','STRING','api endpoint of sonarqube account.','t','f',false,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube PR Based' and ps."index"=1 and ps.deleted=false),'CheckoutPath','STRING','checkout path of git material.','t','f',false,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube PR Based' and ps."index"=1 and ps.deleted=false),'UsePropertiesFileFromProject','BOOL','Boolean value - true or false. Set true to use source code sonar-properties file.','t','f',false,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube PR Based' and ps."index"=1 and ps.deleted=false),'CheckForSonarAnalysisReport','BOOL','Boolean value - true or false. Set true to poll for generated report from sonarqube.','t','f',false,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube PR Based' and ps."index"=1 and ps.deleted=false),'AbortPipelineOnPolicyCheckFailed','BOOL','Boolean value - true or false. Set true to abort on report check failed.','t','f',false,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

INSERT INTO "public"."plugin_stage_mapping" ("plugin_id", "stage_type","created_on","created_by","updated_on","updated_by")
VALUES ((SELECT id FROM plugin_metadata WHERE name='Sonarqube PR Based'),0,now(),1,now(),1);