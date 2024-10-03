DELETE FROM plugin_step_variable WHERE plugin_step_id=(SELECT id FROM plugin_metadata WHERE name='Devtron CI Trigger');
DELETE FROM plugin_step where plugin_id=(SELECT id FROM plugin_metadata WHERE name='Devtron CI Trigger');
DELETE FROM plugin_pipeline_script where id=(SELECT id FROM plugin_metadata WHERE name='Devtron CI Trigger');
DELETE FROM plugin_stage_mapping where plugin_id=(SELECT id from plugin_metadata where name='Devtron CI Trigger');
DELETE FROM plugin_metadata where name='Devtron CI Trigger';
UPDATE plugin_metadata SET is_latest = true WHERE id = (SELECT id FROM plugin_metadata WHERE name= 'Devtron CI Trigger v1.0.0' and is_latest= false);
