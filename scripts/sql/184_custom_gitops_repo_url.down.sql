-- gitops_config modifications
-- Step 1: Drop the new columns for allow_custom_repository
ALTER TABLE public.gitops_config
    DROP COLUMN IF EXISTS allow_custom_repository;