ALTER TABLE "public"."external_ci_pipeline" ALTER COLUMN "access_token" DROP NOT NULL;
ALTER TABLE "public"."external_ci_pipeline" ALTER COLUMN "ci_pipeline_id" DROP NOT NULL;

ALTER TABLE "public"."ci_artifact" ADD COLUMN "external_ci_pipeline_id" int4;
ALTER TABLE "public"."ci_artifact" ADD FOREIGN KEY ("external_ci_pipeline_id") REFERENCES "public"."external_ci_pipeline" ("id");

ALTER TABLE "public"."external_ci_pipeline" ADD COLUMN "app_id" int4;
ALTER TABLE "public"."external_ci_pipeline" ADD FOREIGN KEY ("app_id") REFERENCES "public"."app" ("id");

ALTER TABLE "public"."ci_artifact" ADD COLUMN "payload_schema" text;