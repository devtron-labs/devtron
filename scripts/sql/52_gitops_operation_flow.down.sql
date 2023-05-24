ALTER TABLE "public"."charts" ALTER COLUMN "chart_location" SET NOT NULL;

ALTER TABLE "public"."charts" ALTER COLUMN "git_repo_url" SET NOT NULL;

ALTER TABLE "public"."pipeline" DROP COLUMN "deployment_app_created";