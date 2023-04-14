CREATE SEQUENCE IF NOT EXISTS id_seq_global_cm_cs;

CREATE TABLE public.global_cm_cs (
"id"                            integer NOT NULL DEFAULT nextval('id_seq_smtp_config'::regclass),
"config_type"                   text,
"name"                          text,
"data"                          text,
"mount_path"                    text,
"deleted"                       bool NOT NULL DEFAULT FALSE,
"created_on"                    timestamptz,
"created_by"                    int4,
"updated_on"                    timestamptz,
"updated_by"                    int4,
PRIMARY KEY ("id")
);