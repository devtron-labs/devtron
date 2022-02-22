CREATE SEQUENCE IF NOT EXISTS public.id_seq_plugin_scripts;

CREATE TABLE public.plugin_scripts
(
    "id"                        INT4 NOT NULL DEFAULT NEXTVAL('id_seq_plugin_scripts'::
     regclass),
    "name"                      VARCHAR(250),
    "description"               VARCHAR(250),
    "body"                      TEXT,
    "step_template_language"    VARCHAR(100),
    "step_template"             TEXT,
    PRIMARY KEY ("id")
);

CREATE TABLE public.plugin_inputs
(
    "id"            INT4 NOT NULL,
    "name"          VARCHAR(250),
    "default_value" VARCHAR(250),
    "description"   VARCHAR(250)
);