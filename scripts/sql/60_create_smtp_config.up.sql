CREATE SEQUENCE IF NOT EXISTS id_seq_smtp_config;

CREATE TABLE public.smtp_config (
"id"                          integer NOT NULL DEFAULT nextval('id_seq_smtp_config'::regclass),
"port"                        text,
"host"                        text,
"auth_type"                   text,
"auth_user"                   text,
"auth_password"               text,
"from_email"                  text,
"config_name"                 text,
"description"                 text,
"owner_id"                    int4,
"default"                     bool,
"deleted"                     bool NOT NULL DEFAULT FALSE,
"created_on"                  timestamptz,
"created_by"                  int4,
"updated_on"                  timestamptz,
"updated_by"                  int4,
PRIMARY KEY ("id")
);