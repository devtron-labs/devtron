
ALTER TABLE ci_workflow
    ADD COLUMN IF NOT EXISTS target_image_location text default null;
CREATE INDEX IF NOT EXISTS target_image_path ON ci_workflow (target_image_location);


CREATE TABLE "public"."custom_tag"
(
    id serial PRIMARY KEY,
    ci_pipeline_id int NOT NULL UNIQUE,
    custom_tag_format text,
    auto_increasing_number int DEFAULT 0,

    FOREIGN KEY (ci_pipeline_id) REFERENCES ci_pipeline (id)
);