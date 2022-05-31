DROP TABLE "public"."api_token" CASCADE;

DROP TABLE "public"."api_token_secret" CASCADE;

---- DROP sequence
DROP SEQUENCE IF EXISTS public.id_seq_api_token;

DROP SEQUENCE IF EXISTS public.id_seq_api_token_secret;

-- drop column
ALTER TABLE "public"."users" DROP COLUMN IF EXISTS "user_type";
ALTER TABLE "public"."users" DROP COLUMN IF EXISTS "last_used_at";
ALTER TABLE "public"."users" DROP COLUMN IF EXISTS "last_used_by_ip";