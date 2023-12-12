BEGIN;
-- gitops_config modifications
-- Step 1: Create a new columns for allow_custom_repository
ALTER TABLE public.gitops_config
    ADD COLUMN IF NOT EXISTS allow_custom_repository bool DEFAULT FALSE;

-- installed_apps modifications
-- Step 2: Create a new columns for is_custom_repository
ALTER TABLE public.installed_apps
    ADD COLUMN IF NOT EXISTS is_custom_repository bool DEFAULT FALSE;

ALTER TABLE public.installed_apps
    RENAME COLUMN git_ops_repo_name TO git_ops_repo_url;


-- charts modifications
-- Step 3: Create a new columns for is_custom_repository
ALTER TABLE public.charts
    ADD COLUMN IF NOT EXISTS is_custom_repository bool DEFAULT FALSE;

UPDATE charts SET git_repo_url = 'NOT_CONFIGURED' WHERE git_repo_url IS NULL OR git_repo_url = '';
END;
