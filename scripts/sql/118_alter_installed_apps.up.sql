
update pipeline set deployment_app_delete_request=true
where deleted=true AND deployment_app_type='argo_cd' AND deployment_app_created=false;

update installed_apps set deployment_app_delete_request=false
where active=true AND deployment_app_type='argo_cd';

update installed_apps set deployment_app_delete_request=true
where active=false AND deployment_app_type='argo_cd';