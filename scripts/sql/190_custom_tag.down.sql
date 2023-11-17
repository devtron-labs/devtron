ALTER TABLE custom_tag DROP COLUMN enabled;
ALTER TABLE ci_artifact DROP COLUMN credentials_source_type ;
ALTER TABLE ci_artifact DROP COLUMN credentials_source_value ;
ALTER TABLE ci_artifact DROP COLUMN  component_id;
ALTER TABLE ci_workflow DROP COLUMN image_path_reservation_ids;
ALTER TABLE cd_workflow_runner DROP COLUMN image_path_reservation_ids;
ALTER TABLE image_path_reservation  DROP CONSTRAINT image_path_reservation_custom_tag_id_fkey;