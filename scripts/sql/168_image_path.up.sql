CREATE TABLE "public"."image_path_reservation"(
    id            serial PRIMARY KEY,
    custom_tag_id int,
    image_path    text,
    active        boolean default true,
    FOREIGN KEY (custom_tag_id) REFERENCES custom_tag (id)
);

CREATE INDEX IF NOT EXISTS image_path_index ON image_path_reservation (image_path);

ALTER TABLE ci_workflow
    ADD column IF NOT EXISTS image_path_reservation_id int;