BEGIN;
ALTER TABLE rbac_role_resource_detail DROP COLUMN IF EXISTS "role_resource_version";

DELETE FROM rbac_policy_resource_detail where resource ='user/entity/accessType';
DELETE FROM rbac_role_resource_detail where resource in ('action','subAction');
DELETE FROM default_rbac_role_data where role = 'accessManager';

ALTER TABLE roles DROP COLUMN IF EXISTS "subaction";

COMMIT;