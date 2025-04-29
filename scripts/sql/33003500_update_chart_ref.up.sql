-- below update script will add name for all rollout charts.
update chart_ref set name='Rollout Deployment' where name is null and location like 'reference%' and created_by=1;

ALTER TABLE chart_ref ALTER COLUMN name SET NOT NULL;
