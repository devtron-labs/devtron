CREATE SEQUENCE IF NOT EXISTS public.id_seq_create_cicd;

CREATE TABLE public.id_create_cicd
(
    "id"         INT4 NOT NULL DEFAULT NEXTVAL('id_seq_create_cicd'::
     regclass),
    "name"       VARCHAR(250),
    PRIMARY KEY ("id")
);