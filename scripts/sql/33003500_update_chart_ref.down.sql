ALTER TABLE chart_ref ALTER COLUMN name DROP NOT NULL;

update chart_ref set name=NULL where name like 'Rollout Deployment' and location like 'reference%' and created_by=1;