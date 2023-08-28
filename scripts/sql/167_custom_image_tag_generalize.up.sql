ALTER TABLE custom_tag
    ADD COLUMN entity_key int;
ALTER TABLE custom_tag
    ADD COLUMN entity_value varchar(100);

CREATE INDEX IF NOT EXISTS entity_key_value ON custom_tag (entity_key, entity_value);

ALTER TABLE custom_tag
     ADD CONSTRAINT constraint_name UNIQUE (entity_key, entity_value)


