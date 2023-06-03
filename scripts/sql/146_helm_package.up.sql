ALTER TABLE cd_workflow_runner ADD COLUMN helm_reference_chart bytea;
alter table cluster drop column is_virtual_cluster;
