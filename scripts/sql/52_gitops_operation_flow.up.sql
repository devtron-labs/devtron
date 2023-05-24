ALTER TABLE "public"."charts" ALTER COLUMN "git_repo_url" DROP NOT NULL;

ALTER TABLE "public"."charts" ALTER COLUMN "chart_location" DROP NOT NULL;

ALTER TABLE "public"."pipeline" ADD COLUMN "deployment_app_created" bool NOT NULL DEFAULT 'false';

update pipeline set deployment_app_created=true where deleted=false;