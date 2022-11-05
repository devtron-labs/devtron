
CREATE TABLE public.ci_template_history (
    id integer NOT NULL,
    app_id integer,
    docker_registry_id character varying(250),
    docker_repository character varying(250),
    dockerfile_path character varying(250),
    args text,
    before_docker_build text,
    after_docker_build text,
    template_name character varying(250),
    version character varying(250),
    active boolean,
    created_on timestamp with time zone NOT NULL,
    created_by integer NOT NULL,
    updated_on timestamp with time zone NOT NULL,
    updated_by integer NOT NULL,
    git_material_id integer
)

ALTER TABLE public.ci_template OWNER TO postgres;