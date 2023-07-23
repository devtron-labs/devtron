-- app_store modifications
-- Step 1: Drop column for docker_artifact_store_id foreign key
ALTER TABLE public.app_store
    DROP COLUMN IF EXISTS docker_artifact_store_id;

-- Step 2: Drop foreign key constraint for docker_artifact_store_id column
ALTER TABLE public.app_store
    DROP CONSTRAINT IF EXISTS fk_app_store_docker_artifact_store;

-- Step 3: Drop the unique constraint for the combination of name, chart_repo_id, and docker_artifact_store_id
ALTER TABLE public.app_store
    DROP CONSTRAINT IF EXISTS app_store_unique;

-- Step 4: Revert app_store_unique constraint to the combination of name and chart_repo_id
ALTER TABLE ONLY public.app_store
    ADD CONSTRAINT app_store_unique UNIQUE (name, chart_repo_id);

-- oci_registry_config modifications
-- Step 1: Drop columns for repository_list and 161_virtual_cluster_v3.up.sqlis_public
ALTER TABLE public.oci_registry_config
    DROP COLUMN IF EXISTS repository_list,
    DROP COLUMN IF EXISTS is_chart_pull_active,
    DROP COLUMN IF EXISTS is_public;