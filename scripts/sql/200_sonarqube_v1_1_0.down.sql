DELETE FROM plugin_step_variable WHERE name = 'SonarqubeProjectPrefixName';
DELETE FROM plugin_step_variable WHERE name = 'SonarqubeBranchName';
DELETE FROM plugin_step_variable WHERE name = 'SonarqubeProjectKey';
DELETE FROM plugin_step_variable WHERE name = 'CheckForSonarAnalysisReport';
DELETE FROM plugin_step_variable WHERE name = 'AbortPipelineOnPolicyCheckFailed';
DELETE FROM plugin_step_variable WHERE name = 'GIT_MATERIAL_REQUEST';
DELETE FROM plugin_step_variable WHERE name = 'UsePropertiesFileFromProject';
DELETE FROM plugin_step_variable WHERE name = 'SonarqubeEndpoint';
DELETE FROM plugin_step_variable WHERE name = 'CheckoutPath';
DELETE FROM plugin_step_variable WHERE name = 'SonarqubeApiKey';
DELETE FROM plugin_step_variable WHERE name = 'TotalSonarqubeIssues';
DELETE FROM plugin_step_variable WHERE name = 'SonarqubeHighHotspots';
DELETE FROM plugin_step_variable WHERE name = 'SonarqubeProjectStatus';
DELETE FROM plugin_step_variable WHERE name = 'SonarqubeVulnerabilities';


DELETE FROM plugin_step_variable WHERE plugin_step_id =(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube-v1.1.0' and ps."index"=1 and ps.deleted=false);
DELETE FROM plugin_step WHERE plugin_id=(SELECT id FROM plugin_metadata WHERE name='Sonarqube-v1.1.0');
DELETE FROM plugin_stage_mapping WHERE plugin_id =(SELECT id FROM plugin_metadata WHERE name='Sonarqube-v1.1.0');
DELETE from pipeline_stage_step_variable where pipeline_stage_step_id =(select pipeline_stage_id from pipeline_stage_step where name='Sonarqube-v1.1.0');
DELETE from pipeline_stage_step where name ='Sonarqube-v1.1.0';
DELETE from plugin_tag_relation where plugin_id=(SELECT id FROM plugin_metadata WHERE name='Sonarqube-v1.1.0');
DELETE FROM plugin_metadata WHERE name ='Sonarqube-v1.1.0';
