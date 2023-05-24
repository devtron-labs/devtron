ALTER TABLE "public"."pipeline" ADD COLUMN deployment_app_delete_request bool DEFAULT false;
ALTER TABLE "public"."installed_apps" ADD COLUMN deployment_app_delete_request bool DEFAULT false;

update pipeline set deployment_app_delete_request=true
where deleted=true AND deployment_app_type='argo_cd' AND deployment_app_created=true;

update installed_apps set deployment_app_delete_request=true
where active=true AND deployment_app_type='argo_cd';