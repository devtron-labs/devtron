ALTER TABLE environment ADD COLUMN virtual_environment BOOLEAN;

update environment set virtual_environment=false;

ALTER TABLE cluster ADD COLUMN is_virtual_cluster BOOLEAN;

