CREATE SEQUENCE IF NOT EXISTS id_seq_ci_build_config;

-- Table Definition
CREATE TABLE IF NOT EXISTS public.ci_build_config
(
    "id"                      integer NOT NULL DEFAULT nextval('id_seq_ci_build_config'::regclass),
    "type"                    varchar(100),
    "ci_template_id"          integer,
    "ci_template_override_id" integer,
    "build_metadata"          text,
    "created_on"              timestamptz,
    "created_by"              integer,
    "updated_on"              timestamptz,
    "updated_by"              integer,
    PRIMARY KEY ("id")
);


ALTER TABLE ci_template
    ADD COLUMN IF NOT EXISTS ci_build_config_id integer;

ALTER TABLE ONLY public.ci_template
    ADD CONSTRAINT ci_template_ci_build_config_id_fkey FOREIGN KEY (ci_build_config_id) REFERENCES public.ci_build_config(id);


ALTER TABLE ci_template_override
    ADD COLUMN IF NOT EXISTS ci_build_config_id integer;

ALTER TABLE ONLY public.ci_template_override
    ADD CONSTRAINT ci_template_override_ci_build_config_id_fkey FOREIGN KEY (ci_build_config_id) REFERENCES public.ci_build_config(id);

ALTER TABLE ci_workflow
    ADD COLUMN IF NOT EXISTS ci_build_type varchar(100);
