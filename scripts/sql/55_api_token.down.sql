DROP TABLE "public"."api_token" CASCADE;

DROP TABLE "public"."user_audit" CASCADE;

---- DROP sequence
DROP SEQUENCE IF EXISTS public.id_seq_api_token;

DROP SEQUENCE IF EXISTS public.id_seq_user_audit;

---- DROP index
DROP INDEX IF EXISTS public.user_audit_user_id_IX;

-- drop column
ALTER TABLE "public"."users" DROP COLUMN IF EXISTS "user_type";

-- delete apiTokenSecret from attributes
DELETE FROM attributes WHERE key = 'apiTokenSecret';