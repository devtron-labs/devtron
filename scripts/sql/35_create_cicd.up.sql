CREATE SEQUENCE IF NOT EXISTS public.id_seq_plugin_scripts;
CREATE SEQUENCE IF NOT EXISTS public.id_seq_plugin_tags;
CREATE SEQUENCE IF NOT EXISTS public.id_seq_plugin_steps;
CREATE SEQUENCE IF NOT EXISTS public.id_seq_plugin_steps_seq;

CREATE TABLE public.plugin_scripts
(
    "id"                               INT4 NOT NULL DEFAULT NEXTVAL('id_seq_plugin_scripts'::
     regclass),
    "plugin_id"                        INT4 NOT NULL,
    "plugin_name"                      VARCHAR(250),
    "plugin_description"               VARCHAR(250),
    "plugin_body"                      TEXT,
    "plugin_version"                   INT4,
    PRIMARY KEY ("id")
);

CREATE TABLE public.plugin_fields
(
    "plugin_id"                     INT4 NOT NULL,
    "steps_id"                      INT4 NOT NULL,
    "key_name"                      VARCHAR(250),
    "default_value"                 VARCHAR(250),
    "plugin_key_description"        VARCHAR(250),
    "plugin_field_type"             VARCHAR(100)
);

CREATE TABLE public.plugin_tags
(
    "tag_id"                        INT4 NOT NULL DEFAULT NEXTVAL('id_seq_plugin_tags'::
     regclass),
    "tag_name"                      VARCHAR(250),
    PRIMARY KEY ("tag_id")
);

CREATE TABLE public.plugin_tags_map
(
    "tag_id"                        INT4 NOT NULL,
    "plugin_id"                     INT4 NOT NULL
);

CREATE TABLE public.plugin_steps
(
    "steps_id"                  INT4 NOT NULL DEFAULT NEXTVAL('id_seq_plugin_steps'::
     regclass),
    "steps_name"                VARCHAR(100),
    "steps_template_language"   VARCHAR(100),
    "steps_template"            TEXT,
    "depends_on"                INT4[],
    PRIMARY KEY ("steps_id")
);

CREATE TABLE public.plugin_steps_sequence
(
    "sequence_id"                   INT4 NOT NULL DEFAULT NEXTVAL('id_seq_plugin_steps_seq'::
     regclass),
    "steps_id"                      INT4 NOT NULL,
    "plugin_id"                     INT4 NOT NULL,
    PRIMARY KEY ("sequence_id")
);