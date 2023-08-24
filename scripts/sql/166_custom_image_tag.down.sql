ALTER TABLE ci_workflow
    DROP COLUMN IF EXISTS target_image_location;

DROP INDEX IF EXISTS target_image_path;

DROP TABLE IF EXISTS custom_tag;
