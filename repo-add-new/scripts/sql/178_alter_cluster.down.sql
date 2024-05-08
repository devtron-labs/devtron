ALTER TABLE cluster
    DROP COLUMN to_connect_with_ssh_tunnel,
    DROP COLUMN ssh_tunnel_user,
    DROP COLUMN ssh_tunnel_password,
    DROP COLUMN ssh_tunnel_auth_key,
    DROP COLUMN ssh_tunnel_server_address;
