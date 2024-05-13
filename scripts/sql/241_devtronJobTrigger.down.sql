DELETE FROM plugin_step_variable WHERE plugin_step_id=(SELECT id FROM plugin_metadata WHERE name='Devtron Job Trigger v1.0.0');
DELETE FROM plugin_step where plugin_id=(SELECT id FROM plugin_metadata WHERE name='Devtron Job Trigger v1.0.0' );
DELETE FROM plugin_pipeline_script where id=(SELECT id FROM plugin_metadata WHERE name='Devtron Job Trigger v1.0.0');
DELETE FROM plugin_stage_mapping where plugin_id=(SELECT id from plugin_metadata where name='Devtron Job Trigger v1.0.0');
DELETE FROM plugin_metadata where name='Devtron Job Trigger v1.0.0';
