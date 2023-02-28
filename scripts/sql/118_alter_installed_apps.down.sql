update installed_apps set deployment_app_delete_request=false
where active=false AND deployment_app_type='argo_cd';