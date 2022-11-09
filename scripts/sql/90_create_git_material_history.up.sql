
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
    PRIMARY KEY ("id")
);

ALTER TABLE public.git_material_history OWNER TO postgres;

