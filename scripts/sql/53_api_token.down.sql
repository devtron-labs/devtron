DROP TABLE "public"."api_token" CASCADE;

DROP TABLE "public"."api_token_secret" CASCADE;

---- DROP sequence
DROP SEQUENCE IF EXISTS public.id_seq_api_token;

DROP SEQUENCE IF EXISTS public.id_seq_api_token_secret;