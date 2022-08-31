INSERT INTO module(name, version, status, updated_on)
SELECT 'argo-cd', version, status, now()
FROM module
WHERE name = 'cicd' and status = 'installed';