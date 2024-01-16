DELETE FROM plugin_step_variable WHERE plugin_step_id =(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Container Image Exporter' and ps."index"=1 and ps.deleted=false);
DELETE from plugin_pipeline_script where id = (SELECT script_id from plugin_step WHERE plugin_id=(SELECT id FROM plugin_metadata WHERE name='Jenkins'));
DELETE FROM plugin_step WHERE plugin_id=(SELECT id FROM plugin_metadata WHERE name='Container Image Exporter');
DELETE FROM plugin_stage_mapping WHERE plugin_id =(SELECT id FROM plugin_metadata WHERE name='Container Image Exporter');
DELETE FROM plugin_metadata WHERE name ='Container Image Exporter';