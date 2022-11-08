CREATE SEQUENCE IF NOT EXISTS id_seq_ci_pipeline_history;

CREATE TABLE public.ci_pipeline_history(
    id integer NOT NULL default nextval('id_seq_ci_pipeline_history'::regclass),
    ci_pipeline_id integer,
    ci_template_override_history text,
    ci_pipeline_material_history text
);