DELETE FROM plugin_step_variable WHERE name = 'UsePropertiesFileFromProject';

DELETE FROM plugin_step_variable WHERE name = 'CheckForSonarAnalysisReport';

DELETE FROM plugin_step_variable WHERE name = 'AbortPipelineOnPolicyCheckFailed';

UPDATE plugin_pipeline_script SET script=E'PathToCodeDir=/devtroncd$CheckoutPath
cd $PathToCodeDir
if [[ -z "$UsePropertiesFileFromProject" ]]
then
  echo "sonar.projectKey=$SonarqubeProjectKey" > sonar-project.properties
  docker run \\
  --rm \\
  -e SONAR_HOST_URL=$SonarqubeEndpoint \\
  -e SONAR_LOGIN=$SonarqubeApiKey \\
  -v "/$PWD:/usr/src" \\
  sonarsource/sonar-scanner-cli
elif [[ $UsePropertiesFileFromProject == false ]]
 then
 echo "sonar.projectKey=$SonarqubeProjectKey" > sonar-project.properties
 docker run \\
 --rm \\
 -e SONAR_HOST_URL=$SonarqubeEndpoint \\
 -e SONAR_LOGIN=$SonarqubeApiKey \\
 -v "/$PWD:/usr/src" \\
 sonarsource/sonar-scanner-cli

 if [[ $CheckForSonarAnalysisReport == true || ! -z "$CheckForSonarAnalysisReport" ]]
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
 fi
fi' WHERE id=2;