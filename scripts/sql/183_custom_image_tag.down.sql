DROP TABLE IF EXISTS custom_tag;

DROP INDEX IF EXISTS entity_key_value;

ALTER TABLE custom_tag
    DROP CONSTRAINT unique_entity_key_entity_value;

DROP TABLE IF EXISTS image_path_reservation;

DROP INDEX IF EXISTS image_path_index;

ALTER TABLE ci_workflow
    DROP column IF EXISTS image_path_reservation_id;
ALTER TABLE ci_workflow
    DROP CONSTRAINT fk_image_path_reservation_id;