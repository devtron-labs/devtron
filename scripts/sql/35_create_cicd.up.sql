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
    "plugin_id"     INT4 NOT NULL,
    "key_name"          VARCHAR(250),
    "default_value" VARCHAR(250),
    "plugin_key_description"   VARCHAR(250)
);