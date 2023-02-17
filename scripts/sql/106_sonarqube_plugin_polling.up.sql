do $$
BEGIN
    IF NOT EXISTS (SELECT * FROM plugin_step_variable WHERE name = 'UsePropertiesFileFromProject' AND plugin_step_id = 2) THEN
        INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
        VALUES(nextval('id_seq_plugin_step_variable'),2,'UsePropertiesFileFromProject','BOOL','Boolean value - true or false. Set true to use source code sonar-properties file.','t','f',false,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);
    END IF;
END;
$$;

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES(nextval('id_seq_plugin_step_variable'),2,'CheckForSonarAnalysisReport','BOOL','Boolean value - true or false. Set true to poll for generated report from sonarqube.','t','f',false,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);

INSERT INTO plugin_step_variable (id,plugin_step_id,name,format,description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by) 
VALUES(nextval('id_seq_plugin_step_variable'),2,'AbortPipelineOnPolicyCheckFailed','BOOL','Boolean value - true or false. Set true to abort on report check failed.','t','f',false,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);


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
fi' WHERE id=2;