CREATE TABLE pg_stat_activity_history (
    datid OID,
    datname NAME,
    pid INTEGER,
    usesysid OID,
    usename NAME,
    application_name TEXT,
    client_addr INET,
    client_hostname TEXT,
    client_port INTEGER,
    backend_start TIMESTAMP,
    backend_xmin XID,
    query_start TIMESTAMP,
    state_change TIMESTAMP,
    wait_event_type TEXT,
    wait_event TEXT,
    state TEXT,
    backend_type TEXT,
    query TEXT
);

CREATE OR REPLACE FUNCTION capture_pg_stat_activity_history()
RETURNS VOID AS $$
BEGIN
    INSERT INTO pg_stat_activity_history (
        datid,
        datname,
        pid,
        usesysid,
        usename,
        application_name,
        client_addr,
        client_hostname,
        client_port,
        backend_start,
        backend_xmin,
        query_start,
        state_change,
        wait_event_type,
        wait_event,
        state,
        backend_type,
        query
    )
    SELECT
        datid,
        datname,
        pid,
        usesysid,
        usename,
        application_name,
        client_addr,
        client_hostname,
        client_port,
        backend_start,
        backend_xmin,
        query_start,
        state_change,
        wait_event_type,
        wait_event,
        state,
        backend_type,
        query
    FROM pg_stat_activity WHERE usename = '$db_user';
END;
 $$ LANGUAGE plpgsql; 

CREATE VIEW git_provider_view AS SELECT id,name,url,user_name,access_token,auth_mode,active FROM git_provider;
CREATE VIEW gitops_config_view AS SELECT id,provider,username,github_org_id,host,active,created_on,created_by,updated_on,updated_by,gitlab_group_id,azure_project,bitbucket_workspace_id,bitbucket_project_key,email_id,allow_custom_repository FROM gitops_config;
CREATE VIEW cluster_view AS SELECT id,cluster_name,active,created_on,created_by,updated_on,updated_by,server_url,prometheus_endpoint,cd_argo_setup,p_username,p_password,p_tls_client_cert,p_tls_client_key,agent_installation_stage,k8s_version,error_in_connecting,is_virtual_cluster,insecure_skip_tls_verify,proxy_url,to_connect_with_ssh_tunnel,ssh_tunnel_user,ssh_tunnel_password,ssh_tunnel_auth_key,ssh_tunnel_server_address,description FROM cluster;
CREATE VIEW sso_login_config_view AS SELECT id,name,label,url,created_on,created_by,updated_on,updated_by,active FROM sso_login_config;
CREATE VIEW docker_artifact_store_view AS SELECT id,plugin_id,registry_url,registry_type,aws_accesskey_id,aws_secret_accesskey,aws_region,username,is_default,active,created_on,created_by,updated_on,updated_by,connection,cert,is_oci_compliant_registry FROM docker_artifact_store;
CREATE VIEW ses_config_view AS SELECT id,region,session_token,from_email,config_name,description,created_on,updated_on,created_by,updated_by,owner_id,deleted FROM ses_config;
CREATE VIEW slack_config_view AS SELECT id,config_name,description,created_on,updated_on,created_by,updated_by,owner_id,team_id,deleted FROM slack_config;
CREATE VIEW smtp_config_view AS SELECT id,port,host,auth_type,auth_user,from_email,config_name,description,owner_id,deleted,created_on,created_by,updated_on,updated_by FROM smtp_config;
CREATE VIEW webhook_config_view AS SELECT id,config_name,header,payload,description,owner_id,active,deleted,created_on,created_by,updated_on,updated_by FROM webhook_config;