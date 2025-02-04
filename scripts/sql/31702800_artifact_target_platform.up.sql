ALTER TABLE public.ci_artifact ADD COLUMN IF NOT EXISTS target_platforms varchar(200) NULL;
