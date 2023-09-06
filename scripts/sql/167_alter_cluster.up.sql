ALTER TABLE cluster
    ADD COLUMN to_connect_with_ssh_tunnel boolean,
    ADD COLUMN ssh_tunnel_user            VARCHAR(100),
    ADD COLUMN ssh_tunnel_password        text,
    ADD COLUMN ssh_tunnel_auth_key        text,
    ADD COLUMN ssh_tunnel_server_address  VARCHAR(250);
