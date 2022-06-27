UPDATE plugin_pipeline_script
SET script = 'PathToCodeDir=/devtroncd$CheckoutPath

cd $PathToCodeDir
if [ $UsePropertiesFileFromProject == false ]
then
    echo "sonar.projectKey=$SonarqubeProjectKey" > sonar-project.properties
fi
docker run \
    --rm \
    -e SONAR_HOST_URL=$SonarqubeEndpoint \
    -e SONAR_LOGIN=$SonarqubeApiKey \
    -v "/$PWD:/usr/src" \
    sonarsource/sonar-scanner-cli'
FROM plugin_step JOIN plugin_metadata ON plugin_step.plugin_id=plugin_metadata.id
WHERE plugin_step.script_id = plugin_pipeline_script.id and plugin_metadata.name='Sonarqube';



INSERT INTO "public"."plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "default_value", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
((nextval('id_seq_plugin_step_variable')),(SELECT id from plugin_metadata WHERE name='Sonarqube'),'UsePropertiesFileFromProject','BOOL','Boolean value - true or false. For knowing if sonar-project.properties file is to be used from the project.','t','f','INPUT','NEW','false', '1', 'f','now()', '1', 'now()', '1');
