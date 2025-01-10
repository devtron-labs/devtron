INSERT INTO devtron_resource_searchable_key (name, is_removed, created_on, created_by, updated_on, updated_by)
SELECT 'APP_NAME', false, now(), 1, now(), 1
WHERE NOT EXISTS (SELECT 1 FROM devtron_resource_searchable_key WHERE name = 'APP_NAME');

INSERT INTO devtron_resource_searchable_key (name, is_removed, created_on, created_by, updated_on, updated_by)
SELECT 'ENV_NAME', false, now(), 1, now(), 1
WHERE NOT EXISTS (SELECT 1 FROM devtron_resource_searchable_key WHERE name = 'ENV_NAME');

INSERT INTO devtron_resource_searchable_key (name, is_removed, created_on, created_by, updated_on, updated_by)
SELECT 'CLUSTER_NAME', false, now(), 1, now(), 1
WHERE NOT EXISTS (SELECT 1 FROM devtron_resource_searchable_key WHERE name = 'CLUSTER_NAME');
