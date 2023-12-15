DELETE FROM plugin_step_variable WHERE plugin_step_id IN (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='GCS Create Bucket' and ps."index"=1 and ps.deleted=false);

DELETE FROM plugin_step WHERE plugin_id IN (SELECT id FROM plugin_metadata WHERE name='GCS Create Bucket' AND deleted=false);

DELETE FROM plugin_pipeline_script WHERE id=(SELECT last_value FROM id_seq_plugin_pipeline_script);

DELETE FROM plugin_stage_mapping WHERE plugin_id IN (SELECT id FROM plugin_metadata WHERE name='GCS Create Bucket' AND deleted=false);

DELETE FROM plugin_tag_relation WHERE plugin_id in (SELECT id FROM plugin_metadata WHERE name='GCS Create Bucket' AND deleted=false);

DELETE FROM plugin_tag WHERE name IN ('cloud','gcs');

DELETE FROM plugin_metadata WHERE name='GCS Create Bucket' AND deleted=false;