CREATE SEQUENCE IF NOT EXISTS id_seq_ci_template_history;

CREATE TABLE public.ci_template_history (
    id integer NOT NULL DEFAULT nextval('id_seq_ci_template_history'::regclass),
    ci_template_id integer,
    app_id integer,
    docker_registry_id character varying(250),
    docker_repository character varying(250),
    dockerfile_path character varying(250),
    args text,
    before_docker_build text,
    after_docker_build text,
    template_name character varying(250),
    version character varying(250),
    target_platform VARCHAR(1000) NOT NULL DEFAULT '',
    docker_build_options text,
    active boolean,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL,
    git_material_id integer,
    ci_build_config_id   integer,
    build_meta_data_type  varchar(100),
    build_metadata       text,
    PRIMARY KEY ("id")
);

ALTER TABLE public.ci_template OWNER TO postgres;