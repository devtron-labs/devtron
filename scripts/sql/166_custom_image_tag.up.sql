CREATE TABLE "public"."custom_tag"
(
    id                     serial PRIMARY KEY,
    custom_tag_format      text,
    auto_increasing_number int     DEFAULT 0,
    entity_key             int,
    entity_value           text,
    active                 boolean DEFAULT true,
    metadata               jsonb
);

CREATE INDEX IF NOT EXISTS entity_key_value ON custom_tag (entity_key, entity_value);

ALTER TABLE custom_tag
    ADD CONSTRAINT unique_entity_key_entity_value UNIQUE (entity_key, entity_value);


CREATE TABLE IF not exists "public"."image_path_reservation"
(
    id            serial PRIMARY KEY,
    custom_tag_id int,
    image_path    text,
    active        boolean default true,
    FOREIGN KEY (custom_tag_id) REFERENCES custom_tag (id)
);

CREATE INDEX IF NOT EXISTS image_path_index ON image_path_reservation (image_path);

ALTER TABLE ci_workflow
    ADD column IF NOT EXISTS image_path_reservation_id int;
ALTER TABLE ci_workflow
    ADD CONSTRAINT fk_image_path_reservation_id FOREIGN KEY (image_path_reservation_id) REFERENCES image_path (id);