-- below update script will add name for all rollout charts.
update chart_ref set name='Rollout Deployment' where name is null and location like 'reference%';