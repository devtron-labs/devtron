ALTER TABLE "public"."roles" DROP COLUMN IF EXISTS "access_type";

---- DROP Index
DROP INDEX IF EXISTS "public"."role_unique";