DELETE FROM plugin_step_variable WHERE plugin_step_id =(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Copy container image' and ps."index"=1 and ps.deleted=false);
DELETE FROM plugin_step WHERE plugin_id=(SELECT id FROM plugin_metadata WHERE name='Copy container image');
DELETE FROM plugin_stage_mapping WHERE plugin_id =(SELECT id FROM plugin_metadata WHERE name='Copy container image');
DELETE FROM pipeline_stage_step_variable WHERE pipeline_stage_step_id in (SELECT id FROM pipeline_stage_step where ref_plugin_id =(SELECT id from plugin_metadata WHERE name ='Copy container image'));
DELETE FROM pipeline_stage_step where ref_plugin_id in (SELECT id from plugin_metadata WHERE name ='Copy container image');
DELETE FROM plugin_metadata WHERE name ='Copy container image';


ALTER TABLE custom_tag DROP COLUMN enabled;
ALTER TABLE ci_artifact DROP COLUMN credentials_source_type ;
ALTER TABLE ci_artifact DROP COLUMN credentials_source_value ;
ALTER TABLE ci_artifact DROP COLUMN  component_id;
ALTER TABLE ci_workflow DROP COLUMN image_path_reservation_ids;
ALTER TABLE cd_workflow_runner DROP COLUMN image_path_reservation_ids;
ALTER TABLE image_path_reservation  DROP CONSTRAINT image_path_reservation_custom_tag_id_fkey;