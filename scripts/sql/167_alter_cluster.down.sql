ALTER TABLE cluster
    DROP COLUMN is_ssh_tunnel_configured,
    DROP COLUMN ssh_tunnel_user,
    DROP COLUMN ssh_tunnel_password,
    DROP COLUMN ssh_tunnel_auth_key,
    DROP COLUMN ssh_tunnel_server,
    DROP COLUMN ssh_tunnel_remote_address,
    DROP COLUMN ssh_tunnel_remote_port;
