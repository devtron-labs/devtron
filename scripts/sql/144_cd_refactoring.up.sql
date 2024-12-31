ALTER TABLE environment ADD COLUMN is_virtual_environment BOOLEAN;

update environment set is_virtual_environment=false;

ALTER TABLE cluster ADD COLUMN is_virtual_cluster BOOLEAN;

