CREATE SEQUENCE IF NOT EXISTS id_seq_lock_configuration;

CREATE TABLE IF NOT EXISTS public.lock_configuration
(
    "id"                           int4         NOT NULL DEFAULT nextval('id_seq_lock_configuration'::regclass),
    "config" text,
    "active" bool,
    "created_on"                   timestamptz  NOT NULL,
    "created_by"                   int4         NOT NULL,
    "updated_on"                   timestamptz  NOT NULL,
    "updated_by"                   int4         NOT NULL,
    PRIMARY KEY ("id")
    );
