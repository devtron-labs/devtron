ALTER TABLE public.ci_workflow
    ADD COLUMN IF NOT EXISTS is_artifact_uploaded BOOLEAN;
ALTER TABLE public.cd_workflow_runner
    ADD COLUMN IF NOT EXISTS is_artifact_uploaded BOOLEAN,
    ADD COLUMN IF NOT EXISTS cd_artifact_location varchar(256);
