-- app_store modifications
-- Step 1: Create a new column for docker_artifact_store_id foreign key
ALTER TABLE public.app_store
    ADD COLUMN IF NOT EXISTS docker_artifact_store_id varchar(250);

-- Step 2: Add a foreign key constraint for docker_artifact_store_id column
ALTER TABLE public.app_store
    ADD CONSTRAINT fk_app_store_docker_artifact_store
        FOREIGN KEY (docker_artifact_store_id) REFERENCES docker_artifact_store (id);

-- Step 3: Drop the unique constraint for the combination of name, chart_repo_id
ALTER TABLE public.app_store
    DROP CONSTRAINT IF EXISTS app_store_unique;

-- Step 3: Create a new unique constraint with the combination of name, chart_repo_id, and docker_artifact_store_id
ALTER TABLE public.app_store
    ADD CONSTRAINT app_store_unique
        UNIQUE (name, chart_repo_id, docker_artifact_store_id);

-- oci_registry_config modifications
-- Step 1: Create a new columns for repository_list and 161_virtual_cluster_v3.up.sqlis_public
ALTER TABLE public.oci_registry_config
    ADD COLUMN IF NOT EXISTS repository_list text,
    ADD COLUMN IF NOT EXISTS is_public bool DEFAULT FALSE;