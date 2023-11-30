ALTER TABLE public.ci_pipeline
    ADD COLUMN IF NOT EXISTS is_git_required bool;

UPDATE public.ci_pipeline SET is_git_required= true where ci_pipeline_type !='CI_JOB';
UPDATE public.ci_pipeline SET is_git_required= false where ci_pipeline_type='CI_JOB';
