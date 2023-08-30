ALTER TABLE cluster
    ADD COLUMN is_ssh_tunnel_configured  boolean,
    ADD COLUMN ssh_tunnel_user           VARCHAR(100),
    ADD COLUMN ssh_tunnel_password       text,
    ADD COLUMN ssh_tunnel_auth_key       text,
    ADD COLUMN ssh_tunnel_server         VARCHAR(200),
    ADD COLUMN ssh_tunnel_remote_address VARCHAR(200),
    ADD COLUMN ssh_tunnel_remote_port    integer;
