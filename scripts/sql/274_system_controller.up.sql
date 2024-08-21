CREATE SEQUENCE IF NOT EXISTS id_seq_system_network_controller_config;

CREATE TABLE IF NOT EXISTS "public"."system_network_controller_config"
(
    "id"                             int NOT NULL DEFAULT nextval('id_seq_system_network_controller_config'::regclass),
    "ip"                             VARCHAR(50),
    "username"                       VARCHAR(100),
    "password"                       VARCHAR(100),
    "active"                          bool,
    "action_link_json"              jsonb,
    "created_on"                     timestamptz,
    "created_by"                     integer,
    "updated_on"                     timestamptz,
    "updated_by"                     integer,
    PRIMARY KEY ("id")
    );
