DELETE FROM "public"."default_auth_role"
WHERE access_type='helm-app' AND entity ="apps";


UPDATE "public"."default_auth_policy"
SET role_type='clusterEdit'
WHERE role_type = 'edit' AND  entity = 'cluster';

UPDATE "public"."default_auth_policy"
SET role_type='clusterView'
WHERE role_type = 'view' AND  entity = 'cluster';

UPDATE "public"."default_auth_policy"
SET role_type='clusterAdmin'
WHERE role_type = 'admin' AND  entity = 'cluster';


ALTER TABLE "public"."default_auth_role"
DROP COLUMN access_type;

ALTER TABLE "public"."default_auth_role"
DROP COLUMN entity;

DELETE FROM "public"."default_auth_policy"
WHERE access_type = 'helm-app' AND entity ='apps';

ALTER TABLE "public"."default_auth_policy"
DROP COLUMN access_type;

ALTER TABLE "public"."default_auth_policy"
DROP COLUMN entity;


