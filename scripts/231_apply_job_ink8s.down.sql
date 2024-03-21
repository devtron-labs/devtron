DELETE FROM plugin_step_variable WHERE plugin_step_id =(SELECT id FROM plugin_metadata WHERE name='Apply JOB in k8s v1.0.0');
DELETE FROM plugin_step WHERE plugin_id=(SELECT id FROM plugin_metadata WHERE name='Apply JOB in k8s v1.0.0');
DELETE FROM plugin_pipeline_script WHERE id=(SELECT id FROM plugin_metadata WHERE name='Apply JOB in k8s v1.0.0');
DELETE FROM plugin_stage_mapping WHERE plugin_id=(SELECT id FROM plugin_metadata WHERE name='Apply JOB in k8s v1.0.0');
DELETE FROM plugin_metadata WHERE name ='Apply JOB in k8s v1.0.0';