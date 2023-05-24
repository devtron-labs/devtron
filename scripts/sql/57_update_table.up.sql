ALTER TABLE "public"."installed_apps" ADD COLUMN "deployment_app_type" varchar(50);

update installed_apps set deployment_app_type='helm' WHERE app_id in (SELECT id from app WHERE app_offering_mode='EA_ONLY');

update installed_apps set deployment_app_type='argo_cd' WHERE app_id in (SELECT id from app WHERE app_offering_mode!='EA_ONLY');