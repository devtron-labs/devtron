CREATE TABLE public.app_status
(
    "app_id" integer,
    "env_id" integer,
    "status" varchar(50),
    "updated_on" timestamp with time zone NOT NULL,
    PRIMARY KEY ("app_id","env_id"),
    CONSTRAINT app_status_app_id_fkey
        FOREIGN KEY(app_id)
            REFERENCES public.app(id),
    CONSTRAINT app_status_env_id_fkey
        FOREIGN KEY(env_id)
            REFERENCES public.environment(id)

)