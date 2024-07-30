CREATE SEQUENCE IF NOT EXISTS id_seq_plugin_parent_metadata;

--  Table Definition
CREATE TABLE IF NOT EXISTS public.plugin_parent_metadata
(
    "id"                    integer         NOT NULL DEFAULT nextval('id_seq_plugin_parent_metadata'::regclass),
    "name"                  TEXT         NOT NULL,
    "identifier"            TEXT         NOT NULL,
    "deleted"               BOOL         NOT NULL,
    "description"           TEXT,
    "type"                  varchar(255),  -- SHARED, PRESET etc
    "icon"                  TEXT,
    "created_on"            timestamptz  NOT NULL,
    "created_by"            int4         NOT NULL,
    "updated_on"            timestamptz NOT NULL,
    "updated_by"            int4 NOT NULL,
    PRIMARY KEY ("id"),
    UNIQUE("identifier")
);
