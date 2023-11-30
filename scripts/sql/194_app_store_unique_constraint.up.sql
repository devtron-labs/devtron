ALTER TABLE public.app_store
DROP CONSTRAINT IF EXISTS app_store_unique;
CREATE UNIQUE INDEX IF NOT EXISTS app_store_unique_oci_repo ON public.app_store
    (name, docker_artifact_store_id) where active=true;
CREATE UNIQUE INDEX IF NOT EXISTS app_store_unique_chart_repo ON public.app_store
    (name, chart_repo_id) where active=true;
CREATE UNIQUE INDEX IF NOT EXISTS app_store_application_version_unique ON public.app_store_application_version
      (version,app_store_id);