-- Migration Script virtual cluster V3
INSERT INTO public.oci_registry_config ("docker_artifact_store_id", "repository_type", "repository_action","created_on", "created_by", "updated_on", "updated_by","deleted")
    SELECT docker_artifact_store.id, 'CONTAINER', 'PULL/PUSH', 'now()', 1, 'now()', 1, 'f' from docker_artifact_store LEFT JOIN oci_registry_config orc on orc.docker_artifact_store_id = docker_artifact_store.id
        WHERE orc.id IS NULL AND docker_artifact_store.is_oci_compliant_registry = TRUE;

UPDATE public.oci_registry_config set deleted = TRUE WHERE docker_artifact_store_id IN (SELECT id FROM docker_artifact_store where active = FALSE);