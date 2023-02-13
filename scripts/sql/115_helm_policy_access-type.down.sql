DELETE FROM "public"."default_auth_role"
WHERE access_type='helm-app' AND entity ="apps";

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


