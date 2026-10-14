ALTER TABLE roles DROP COLUMN "release";
ALTER TABLE roles DROP COLUMN "release_track";
DELETE from rbac_role_resource_detail where resource in ('release','release-track');
DELETE from rbac_policy_resource_detail where resource in ('release','release-track');
