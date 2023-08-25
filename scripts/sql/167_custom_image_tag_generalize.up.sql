ALTER TABLE custom_tag
    ADD COLUMN entity_key varchar(30);
ALTER TABLE custom_tag
    ADD COLUMN entity_value varchar(100);

ALTER TABLE custom_tag
    ADD COLUMN metadata jsonb;

