BEGIN;
ALTER TABLE rbac_role_resource_detail DROP COLUMN IF EXISTS "role_resource_version";

DELETE FROM rbac_policy_resource_detail where resource ='user/entity/accessType';
DELETE FROM rbac_role_resource_detail where resource in ('action','subAction');

ALTER TABLE roles DROP COLUMN IF EXISTS "subaction";

COMMIT;