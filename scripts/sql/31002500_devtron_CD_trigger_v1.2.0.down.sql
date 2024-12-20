DELETE FROM plugin_step_variable WHERE plugin_step_id=(SELECT id FROM plugin_metadata WHERE name='Devtron CD Trigger' and plugin_version='1.2.0');
DELETE FROM plugin_step where plugin_id=(SELECT id FROM plugin_metadata WHERE name='Devtron CD Trigger' and plugin_version='1.2.0');
DELETE FROM plugin_pipeline_script where id=(SELECT id FROM plugin_metadata WHERE name='Devtron CD Trigger');
DELETE FROM plugin_stage_mapping where plugin_id=(SELECT id from plugin_metadata where name='Devtron CD Trigger');
DELETE FROM plugin_metadata where name='Devtron CD Trigger';
UPDATE plugin_metadata SET is_latest = true WHERE id = (SELECT id FROM plugin_metadata WHERE name= 'Devtron CD Trigger' and is_latest= false and plugin_version='1.1.0');
