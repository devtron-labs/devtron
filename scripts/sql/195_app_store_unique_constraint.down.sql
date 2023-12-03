ALTER TABLE public.app_store
DROP CONSTRAINT IF EXISTS app_store_unique_oci_repo;

ALTER TABLE public.app_store
DROP CONSTRAINT IF EXISTS app_store_unique_chart_repo;

ALTER TABLE public.app_store
DROP CONSTRAINT IF EXISTS app_store_application_version_unique;
