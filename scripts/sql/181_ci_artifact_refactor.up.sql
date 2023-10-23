ALTER TABLE ci_artifact ADD COLUMN credentials_source_type VARCHAR(50);
ALTER TABLE ci_artifact ADD COLUMN credentials_source_value VARCHAR(50);
ALTER TABLE ci_artifact ADD COLUMN  component_id integer;
