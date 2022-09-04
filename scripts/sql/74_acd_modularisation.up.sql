INSERT INTO module(name, version, status, updated_on)
SELECT 'argo-cd', version, status, now()
FROM module
WHERE name = 'cicd' and status = 'installed' and not exists (SELECT 1 FROM module where name = 'argo-cd')
