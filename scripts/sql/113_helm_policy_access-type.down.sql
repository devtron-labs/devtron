DELETE FROM "public"."default_auth_role"
WHERE access_type ='helm-app';

ALTER TABLE "public"."default_auth_role"
DROP COLUMN access_type;

DELETE FROM "public"."default_auth_policy"
WHERE access_type = 'helm-app';

ALTER TABLE "public"."default_auth_policy"
DROP COLUMN access_type;


