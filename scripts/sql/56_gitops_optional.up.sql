ALTER TABLE "public"."pipeline" ADD COLUMN "deployment_app_type" varchar(50);
ALTER TABLE "public"."charts" ADD COLUMN "reference_chart" bytea;
UPDATE "public"."pipeline" SET "deployment_app_type" = 'argo_cd';