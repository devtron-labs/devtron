INSERT INTO module(name, version, status, updated_on)
SELECT 'notifier', version, status, now()
FROM module
WHERE name = 'cicd' and status = 'installed' and not exists (SELECT 1 FROM module where name = 'notifier');

INSERT INTO module(name, version, status, updated_on)
SELECT 'monitoring.grafana', version, status, now()
FROM module
WHERE name = 'cicd' and status = 'installed' and not exists (SELECT 1 FROM module where name = 'monitoring.grafana');