BEGIN;
-- gitops_config modifications
-- Step 1: Drop the new columns for allow_custom_repository
ALTER TABLE public.gitops_config
    DROP COLUMN IF EXISTS allow_custom_repository;

-- installed_apps modifications
-- Step 2: Create a new columns for is_custom_repository
ALTER TABLE public.installed_apps
    DROP COLUMN IF EXISTS is_custom_repository bool;

ALTER TABLE public.installed_apps
    RENAME COLUMN git_ops_repo_url TO git_ops_repo_name;
UPDATE installed_apps set git_ops_repo_name = REPLACE(REVERSE(SPLIT_PART(REVERSE(git_ops_repo_name), '/', 1)), '.git', '');


-- charts modifications
-- Step 3: Create a new columns for is_custom_repository
ALTER TABLE public.charts
    DROP COLUMN IF EXISTS is_custom_repository bool;

UPDATE charts set git_repo_url = '' where git_repo_url = 'NOT_CONFIGURED';
END;
