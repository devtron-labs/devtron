ALTER TABLE environment ADD COLUMN is_virtual_environment BOOLEAN;
update environment set is_virtual_environment=false;

ALTER TABLE cluster ADD COLUMN is_virtual_cluster BOOLEAN;

ALTER TABLE cd_workflow_runner ADD COLUMN helm_reference_chart bytea;

