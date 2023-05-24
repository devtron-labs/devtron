CREATE SEQUENCE IF NOT EXISTS id_seq_git_material_history;


CREATE TABLE public.git_material_history (
     "id" integer  NOT NULL DEFAULT nextval('id_seq_git_material_history'::regclass),
     "app_id" integer,
     "git_provider_id" integer,
     "git_material_id" integer,
     "active" boolean NOT NULL,
     "name" character varying(250),
     "url" character varying(250),
     "created_on" timestamp with time zone NOT NULL,
     "created_by" integer NOT NULL,
     "updated_on" timestamp with time zone NOT NULL,
     "updated_by" integer NOT NULL,
     "checkout_path" character varying(250),
     "fetch_submodules" boolean NOT NULL,
     PRIMARY KEY ("id"),
     CONSTRAINT git_material_history_git_material_id_fkey
         FOREIGN KEY(git_material_id)
             REFERENCES public.git_material(id)
);

ALTER TABLE public.git_material_history OWNER TO postgres;


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
    trigger character varying(100),
    PRIMARY KEY ("id"),
    CONSTRAINT ci_template_history_ci_template_id_fkey
        FOREIGN KEY (ci_template_id)
            REFERENCES public.ci_template(id),
    CONSTRAINT ci_template_history_docker_registry_id_fkey
        FOREIGN KEY(docker_registry_id)
            REFERENCES public.docker_artifact_store(id),
    CONSTRAINT ci_template_history_app_id_fkey
        FOREIGN KEY(app_id)
            REFERENCES public.app(id),
    CONSTRAINT ci_template_git_material_history_id_fkey
        FOREIGN KEY(git_material_id)
            REFERENCES public.git_material(id)
);


ALTER TABLE public.ci_template OWNER TO postgres;

CREATE SEQUENCE IF NOT EXISTS id_seq_ci_pipeline_history;

CREATE TABLE public.ci_pipeline_history(
   id integer NOT NULL default nextval('id_seq_ci_pipeline_history'::regclass),
   ci_pipeline_id integer,
   ci_template_override_history text,
   ci_pipeline_material_history text,
   scan_enabled boolean,
   manual boolean,
   trigger character varying(100),
   PRIMARY KEY ("id"),
   CONSTRAINT ci_pipeline_history_ci_pipeline_id_fk
       FOREIGN KEY (ci_pipeline_id)
           REFERENCES public.ci_pipeline(id)
);

