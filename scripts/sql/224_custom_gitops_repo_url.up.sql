BEGIN;
-- gitops_config modifications
-- Step 1: Create a new columns for allow_custom_repository
ALTER TABLE public.gitops_config
    ADD COLUMN IF NOT EXISTS allow_custom_repository bool DEFAULT FALSE;

-- installed_apps modifications
-- Step 2: Create a new columns for is_custom_repository
ALTER TABLE public.installed_apps
    ADD COLUMN IF NOT EXISTS is_custom_repository bool DEFAULT FALSE;

-- Step 3: Create a new columns for git_ops_repo_url
ALTER TABLE public.installed_apps
    ADD COLUMN IF NOT EXISTS git_ops_repo_url varchar(255);

-- charts modifications
-- Step 4: Create a new columns for is_custom_repository
ALTER TABLE public.charts
    ADD COLUMN IF NOT EXISTS is_custom_repository bool DEFAULT FALSE;

UPDATE charts SET git_repo_url = 'NOT_CONFIGURED' WHERE git_repo_url IS NULL OR git_repo_url = '';
END;
