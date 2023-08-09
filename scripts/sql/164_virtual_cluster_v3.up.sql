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
-- Step 1: Create a new columns for repository_list, is_chart_pull_active and is_public
ALTER TABLE public.oci_registry_config
    ADD COLUMN IF NOT EXISTS repository_list text,
    ADD COLUMN IF NOT EXISTS is_chart_pull_active bool,
    ADD COLUMN IF NOT EXISTS is_public bool DEFAULT FALSE;

-- docker_registry_ips_config modifications
-- Step 1: Add active column
ALTER TABLE public.docker_registry_ips_config
    ADD COLUMN IF NOT EXISTS active bool DEFAULT TRUE;

-- Migration Script
BEGIN;
INSERT INTO public.oci_registry_config ("docker_artifact_store_id", "repository_type", "repository_action","created_on", "created_by", "updated_on", "updated_by","deleted")
    SELECT id, 'CONTAINER', 'PULL/PUSH', 'now()', 1, 'now()', 1, 'f' from docker_artifact_store WHERE registry_type != 'gcr'AND is_oci_compliant_registry IS FALSE;
UPDATE docker_artifact_store set is_oci_compliant_registry = TRUE WHERE registry_type != 'gcr';
END;