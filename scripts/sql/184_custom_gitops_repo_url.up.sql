-- gitops_config modifications
-- Step 1: Create a new columns for allow_custom_repository
ALTER TABLE public.gitops_config
    ADD COLUMN IF NOT EXISTS allow_custom_repository bool DEFAULT FALSE;