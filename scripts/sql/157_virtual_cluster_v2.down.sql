-- Dropping the sequence
DROP SEQUENCE IF EXISTS id_seq_push_config;

-- Dropping table manifest_push_config
DROP TABLE IF EXISTS manifest_push_config;

-- Dropping the sequence
DROP SEQUENCE IF EXISTS id_seq_oci_config;

-- Dropping the unique index
DROP INDEX IF EXISTS idx_unique_repositories;

-- Dropping table oci_registry_config
DROP TABLE IF EXISTS oci_registry_config;


-- Dropping the is_oci_compliant_registry Column from docker_artifact_store Table
ALTER TABLE docker_artifact_store DROP COLUMN IF EXISTS is_oci_compliant_registry;