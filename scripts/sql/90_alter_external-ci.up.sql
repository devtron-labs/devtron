ALTER TABLE "public"."external_ci_pipeline" ALTER COLUMN "access_token" DROP NOT NULL;

ALTER TABLE "public"."pipeline" ALTER COLUMN "ci_pipeline_id" DROP NOT NULL;