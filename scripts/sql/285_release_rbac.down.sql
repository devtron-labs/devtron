ALTER TABLE roles DROP COLUMN "release";
ALTER TABLE roles DROP COLUMN "release-track";
DELETE from rbac_role_resource_detail where resource in ('release','release-track');
DELETE from rbac_policy_resource_detail where resource in ('release','release-track');
ALTER TABLE rbac_role_data ALTER COLUMN access_type SET NOT NULL;
ALTER TABLE rbac_policy_data ALTER COLUMN access_type SET NOT NULL;