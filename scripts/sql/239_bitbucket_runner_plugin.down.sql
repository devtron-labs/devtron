DELETE FROM plugin_step_variable WHERE plugin_step_id =(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='BitBucket Runner Trigger v1.0.0' and ps."index"=1 and ps.deleted=false);
DELETE FROM plugin_step WHERE plugin_id=(SELECT id FROM plugin_metadata WHERE name='BitBucket Runner Trigger v1.0.0');
DELETE FROM pipeline_stage_step_variable where pipeline_stage_step_id in (select id from pipeline_stage_step where name ='BitBucket Runner Trigger v1.0.0');
DELETE from pipeline_stage_step where name ='BitBucket Runner Trigger v1.0.0';
DELETE FROM plugin_metadata WHERE name ='BitBucket Runner Trigger v1.0.0';


DELETE FROM plugin_step_variable WHERE plugin_step_id IN (SELECT id FROM plugin_step WHERE plugin_id = (SELECT id FROM plugin_metadata WHERE name = 'BitBucket Runner Trigger v1.0.0'));

DELETE FROM plugin_step WHERE plugin_id = (SELECT id FROM plugin_metadata WHERE name = 'BitBucket Runner Trigger v1.0.0');

DELETE FROM plugin_pipeline_script WHERE id = (SELECT script_id FROM plugin_step WHERE plugin_id = (SELECT id FROM plugin_metadata WHERE name = 'BitBucket Runner Trigger v1.0.0'));

DELETE FROM plugin_stage_mapping WHERE plugin_id = (SELECT id FROM plugin_metadata WHERE name = 'BitBucket Runner Trigger v1.0.0');

DELETE FROM plugin_metadata WHERE name = 'BitBucket Runner Trigger v1.0.0';


