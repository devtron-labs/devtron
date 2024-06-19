CREATE SEQUENCE IF NOT EXISTS id_seq_plugin_parent_metadata;

--  Table Definition
CREATE TABLE IF NOT EXISTS public.plugin_parent_metadata
(
    "id"                    int4         NOT NULL DEFAULT nextval('id_seq_plugin_parent_metadata'::regclass),
    "name"                  TEXT         NOT NULL,
    "identifier"            TEXT         NOT NULL,
    "deleted"               BOOL         NOT NULL,
    "created_on"            timestamptz  NOT NULL,
    "created_by"            int4         NOT NULL,
    "updated_on"            timestamptz,
    "updated_by"            int4,
    PRIMARY KEY ("id"),
);
