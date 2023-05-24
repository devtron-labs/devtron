DELETE FROM plugin_step_variable WHERE name = 'UsePropertiesFileFromProject';

DELETE FROM plugin_step_variable WHERE name = 'CheckForSonarAnalysisReport';

DELETE FROM plugin_step_variable WHERE name = 'AbortPipelineOnPolicyCheckFailed';

UPDATE plugin_pipeline_script SET script=E'PathToCodeDir=/devtroncd$CheckoutPath
cd $PathToCodeDir
echo "sonar.projectKey=$SonarqubeProjectKey" > sonar-project.properties
docker run
--rm
-e SONAR_HOST_URL=$SonarqubeEndpoint
-e SONAR_LOGIN=$SonarqubeApiKey
-v "/$PWD:/usr/src"
sonarsource/sonar-scanner-cli' WHERE id = 2;