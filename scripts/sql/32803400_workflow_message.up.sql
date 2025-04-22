BEGIN;

-- cd_workflow_runner.message has a limit of 1000 characters. migrating to
ALTER TABLE "public"."cd_workflow_runner"
    ALTER COLUMN "message" TYPE TEXT;

-- ci_workflow.message has a limit of 250 characters. migrating to
ALTER TABLE "public"."ci_workflow"
    ALTER COLUMN "message" TYPE TEXT;

COMMIT;