ALTER TABLE public.ci_workflow
    DROP COLUMN IF EXISTS is_artifact_uploaded;
ALTER TABLE public.cd_workflow_runner
    DROP COLUMN IF EXISTS is_artifact_uploaded,
    DROP COLUMN IF EXISTS cd_artifact_location;
