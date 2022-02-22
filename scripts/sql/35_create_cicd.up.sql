CREATE SEQUENCE IF NOT EXISTS public.id_seq_plugin_scripts;

CREATE TABLE public.plugin_scripts
(
    "id"            INT4 NOT NULL DEFAULT NEXTVAL('id_seq_plugin_scripts'::
     regclass),
    "name"          VARCHAR(250),
    "description"   VARCHAR(250),
    PRIMARY KEY ("id")
);

CREATE TABLE public.plugin_inputs
(
    "id"            INT4 NOT NULL,
    "key"          VARCHAR(250),
    "value"   VARCHAR(250),
);