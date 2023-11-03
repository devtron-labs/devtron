ALTER TABLE ci_artifact ADD COLUMN credentials_source_type VARCHAR(50);
ALTER TABLE ci_artifact ADD COLUMN credentials_source_value VARCHAR(50);
ALTER TABLE ci_artifact ADD COLUMN  component_id integer;

ALTER TABLE ci_workflow ADD COLUMN image_path_reservation_ids integer[];

UPDATE ci_workflow set image_path_reservation_ids=ARRAY["image_path_reservation_id"] where image_path_reservation_id is not NULL;

ALTER TABLE cd_workflow_runner ADD COLUMN image_path_reservation_ids integer[];