BEGIN;
-- gitops_config modifications
-- Step 1: Drop the new columns allow_custom_repository
ALTER TABLE public.gitops_config
    DROP COLUMN IF EXISTS allow_custom_repository;

-- installed_apps modifications
-- Step 2: Drop the new columns is_custom_repository
ALTER TABLE public.installed_apps
    DROP COLUMN IF EXISTS is_custom_repository;

-- Step 3: Drop the new columns git_ops_repo_url
ALTER TABLE public.installed_apps
    DROP COLUMN IF EXISTS git_ops_repo_url;

-- charts modifications
-- Step 4: Drop the new columns is_custom_repository
ALTER TABLE public.charts
    DROP COLUMN IF EXISTS is_custom_repository;

UPDATE charts SET git_repo_url = '' WHERE git_repo_url = 'NOT_CONFIGURED';
END;
