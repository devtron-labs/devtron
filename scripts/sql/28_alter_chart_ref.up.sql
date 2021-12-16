ALTER TABLE "public"."chart_ref" ADD COLUMN IF NOT EXISTS "name" varchar(250);

ALTER TABLE "public"."chart_ref" ADD COLUMN IF NOT EXISTS "chart_data" bytea;