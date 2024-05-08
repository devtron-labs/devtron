DELETE FROM plugin_step_variable WHERE plugin_step_id=(SELECT ps.id FROM plugin_metadata p INNER JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Copacetic v1.0.0' and ps."index"=1 and ps.deleted=false);
DELETE FROM plugin_step WHERE plugin_id=(SELECT id FROM plugin_metadata WHERE name='Copacetic v1.0.0');
DELETE FROM plugin_stage_mapping WHERE plugin_id=(SELECT id FROM plugin_metadata WHERE name='Copacetic v1.0.0');
DELETE FROM pipeline_stage_step_variable WHERE pipeline_stage_step_id in (SELECT pipeline_stage_id FROM pipeline_stage_step WHERE name='Copacetic v1.0.0');
DELETE FROM pipeline_stage_step_variable WHERE pipeline_stage_step_id in (SELECT id FROM pipeline_stage_step WHERE name='Copacetic v1.0.0');
DELETE FROM pipeline_stage_step WHERE name ='Copacetic v1.0.0';
DELETE FROM plugin_tag_relation WHERE plugin_id=(SELECT id FROM plugin_metadata WHERE name='Copacetic v1.0.0');
DELETE FROM plugin_metadata WHERE name='Copacetic v1.0.0';