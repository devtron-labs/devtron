ALTER TABLE "public"."pipeline" ADD COLUMN deployment_app_delete_request bool DEFAULT false;
ALTER TABLE "public"."installed_apps" ADD COLUMN deployment_app_delete_request bool DEFAULT false;
