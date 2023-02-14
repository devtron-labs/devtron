DELETE FROM "public"."default_auth_role"
WHERE access_type='helm-app' AND entity ="apps";


UPDATE "public"."default_auth_role"
SET role_type='entitySpecificView'
WHERE role_type = 'view' AND entity='chart-group';

UPDATE "public"."default_auth_role"
SET role_type='roleSpecific'
WHERE role_type = 'update' AND entity='chart-group';

UPDATE "public"."default_auth_role"
SET role_type='entitySpecificAdmin'
WHERE role_type = 'admin' AND entity='chart-group';

UPDATE "public"."default_auth_role"
SET role_type='clusterEdit'
WHERE role_type = 'edit' AND  entity = 'cluster';

UPDATE "public"."default_auth_role"
SET role_type='clusterView'
WHERE role_type = 'view' AND  entity = 'cluster';

UPDATE "public"."default_auth_role"
SET role_type='clusterAdmin'
WHERE role_type = 'admin' AND  entity = 'cluster';

ALTER TABLE "public"."default_auth_role"
DROP COLUMN access_type;

ALTER TABLE "public"."default_auth_role"
DROP COLUMN entity;

DELETE FROM "public"."default_auth_policy"
WHERE access_type = 'helm-app' AND entity ='apps';

UPDATE "public"."default_auth_policy"
SET role_type='clusterEdit'
WHERE role_type = 'edit' AND  entity = 'cluster';

UPDATE "public"."default_auth_policy"
SET role_type='clusterView'
WHERE role_type = 'view' AND  entity = 'cluster';

UPDATE "public"."default_auth_policy"
SET role_type='clusterAdmin'
WHERE role_type = 'admin' AND  entity = 'cluster';

UPDATE "public"."default_auth_policy"
SET role_type='entitySpecific'
WHERE role_type = 'update' AND entity = 'chart-group';

UPDATE "public"."default_auth_policy"
SET  role_type='entityView'
WHERE role_type = 'view' AND entity = 'chart-group';

UPDATE "public"."default_auth_policy"
SET role_type='entityAll'
WHERE role_type = 'admin' AND entity = 'chart-group';

ALTER TABLE "public"."default_auth_policy"
DROP COLUMN access_type;

ALTER TABLE "public"."default_auth_policy"
DROP COLUMN entity;


