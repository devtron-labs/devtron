INSERT INTO module(name, version, status, updated_on)
SELECT 'security.clair', version, status, now()
FROM module
WHERE name = 'cicd' and status = 'installed' and not exists (SELECT 1 FROM module where name = 'security.clair')
