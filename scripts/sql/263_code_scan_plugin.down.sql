DELETE FROM plugin_step_variable WHERE plugin_step_id =(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Code Scan' and ps."index"=1 and ps.deleted=false);
DELETE FROM plugin_step WHERE plugin_id = (SELECT id FROM plugin_metadata WHERE name='Code Scan');
DELETE FROM plugin_stage_mapping WHERE plugin_id =(SELECT id FROM plugin_metadata WHERE name='Code Scan');
DELETE FROM plugin_tag_relation WHERE plugin_id =(SELECT id FROM plugin_metadata WHERE name='Code Scan');
DELETE FROM pipeline_stage_step where ref_plugin_id in (SELECT id from plugin_metadata WHERE name ='Code Scan');
DELETE from plugin_pipeline_script where id = (SELECT script_id from plugin_step WHERE plugin_id=(SELECT id FROM plugin_metadata WHERE name='Code Scan'));
DELETE FROM plugin_metadata WHERE name ='Code Scan';