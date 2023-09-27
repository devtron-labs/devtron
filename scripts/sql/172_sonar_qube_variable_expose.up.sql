
/* adding a variable */

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES
(nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube' and ps."index"=1 and ps.deleted=false),'SonarAnalysisReport','STRING','Analysis report of source code','t','f',false,null,'OUTPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

/* changing the sonarqube script */

UPDATE plugin_pipeline_script SET script=E'PathToCodeDir=/devtroncd$CheckoutPath
cd $PathToCodeDir
if [[ -z "$UsePropertiesFileFromProject" || $UsePropertiesFileFromProject == false ]]
then
  echo "sonar.projectKey=$SonarqubeProjectKey" > sonar-project.properties
fi
docker run \\
--rm \\
-e SONAR_HOST_URL=$SonarqubeEndpoint \\
-e SONAR_LOGIN=$SonarqubeApiKey \\
-v "/$PWD:/usr/src" \\
sonarsource/sonar-scanner-cli

if [[ $CheckForSonarAnalysisReport == true && ! -z "$CheckForSonarAnalysisReport" ]]
then
 status=$(curl -u ${SonarqubeApiKey}:  -sS ${SonarqubeEndpoint}/api/qualitygates/project_status?projectKey=${SonarqubeProjectKey}&branch=master)
 project_status=$(echo $status | jq -r  ".projectStatus.status")
 echo "*********  SonarQube Policy Report  *********"
 export SonarAnalysisReport=$status
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
fi' WHERE id=(select script_id from  plugin_step inner join plugin_metadata on plugin_step.plugin_id=plugin_metadata.id  where plugin_metadata.name='Sonarqube');


/* changing webhook trigger in git_sensor */


INSERT INTO git_host_webhook_event_selectors(event_id,name,selector,to_show,to_show_in_ci_filter,is_active,created_on,updated_on,to_use_in_ci_env_variable) 
VALUES(2,'pull request id','pullrequest.id',true,false,true,now(),now(),true);

