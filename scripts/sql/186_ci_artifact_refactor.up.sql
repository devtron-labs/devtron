ALTER TABLE ci_artifact ADD COLUMN credentials_source_type VARCHAR(50);
ALTER TABLE ci_artifact ADD COLUMN credentials_source_value VARCHAR(50);
ALTER TABLE ci_artifact ADD COLUMN  component_id integer;

ALTER TABLE ci_workflow ADD COLUMN image_reservation_ids integer[];

UPDATE ci_workflow set image_reservation_ids=ARRAY[image_path_reservation]

ALTER TABLE ci_workflow_runner ADD COLUMN image_reservation_ids integer[];