ALTER TABLE ONLY public.app DROP CONSTRAINT app_app_name_key;

ALTER TABLE "public"."deployment_status" ADD COLUMN "active" bool NOT NULL DEFAULT 'true';