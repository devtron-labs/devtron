ALTER TABLE "public"."external_ci_pipeline" ALTER COLUMN "ci_pipeline_id" SET NOT NULL;

ALTER TABLE "public"."external_ci_pipeline" ALTER COLUMN "access_token" SET NOT NULL;

ALTER TABLE "public"."external_ci_pipeline" DROP COLUMN "app_id";

ALTER TABLE "public"."ci_artifact" DROP COLUMN "external_ci_pipeline_id";

ALTER TABLE "public"."ci_artifact" DROP COLUMN "payload_schema";