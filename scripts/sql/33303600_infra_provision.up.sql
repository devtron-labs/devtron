BEGIN;
-- Sequence
CREATE SEQUENCE IF NOT EXISTS id_seq_chart_category;

-- DROP TABLE IF EXISTS public.chart_category;

CREATE TABLE IF NOT EXISTS public.chart_category
(
    id integer NOT NULL DEFAULT nextval('id_seq_chart_category'::regclass),
    name character varying(250) COLLATE pg_catalog."default" NOT NULL,
    description text COLLATE pg_catalog."default" NOT NULL,
    deleted boolean NOT NULL,
    created_on timestamp with time zone,
    created_by integer,
    updated_on timestamp with time zone,
                             updated_by integer,
                             CONSTRAINT chart_category_pkey PRIMARY KEY (id)
    );


-- Sequence
CREATE SEQUENCE IF NOT EXISTS id_seq_chart_category_mapping;


-- DROP TABLE IF EXISTS public.chart_category_mapping;

CREATE TABLE IF NOT EXISTS public.chart_category_mapping
(
    id integer NOT NULL DEFAULT nextval('id_seq_chart_category_mapping'::regclass),
    app_store_id integer,
    chart_category_id integer,
    deleted boolean NOT NULL,
    created_on timestamp with time zone,
    created_by integer,
    updated_on timestamp with time zone,
                             updated_by integer,
                             CONSTRAINT chart_category_mapping_pkey PRIMARY KEY (id),
    CONSTRAINT chart_category_mapping_app_store_id_fkey FOREIGN KEY (app_store_id)
    REFERENCES public.app_store (id) MATCH SIMPLE
                         ON UPDATE NO ACTION
                         ON DELETE NO ACTION,
    CONSTRAINT chart_category_mapping_chart_category_id_fkey FOREIGN KEY (chart_category_id)
    REFERENCES public.chart_category (id) MATCH SIMPLE
                         ON UPDATE NO ACTION
                         ON DELETE NO ACTION
    );



CREATE SEQUENCE IF NOT EXISTS id_seq_infrastructure_installation;

CREATE TABLE infrastructure_installation (
                                            id int4 NOT NULL DEFAULT nextval('id_seq_infrastructure_installation'::regclass),
                                            installation_type VARCHAR(255),
                                            installed_entity_type VARCHAR(64),
                                            installed_entity_id INT ,
                                            installation_name VARCHAR(128),
                                            "created_on" timestamptz,
                                            "created_by" integer,
                                            "updated_on" timestamptz,
                                            "updated_by" integer,
                                            "active" boolean,
                                            PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_infrastructure_installation_versions;

CREATE TABLE infrastructure_installation_versions (
                                           id int4 NOT NULL DEFAULT nextval('id_seq_infrastructure_installation_versions'::regclass),
                                           infrastructure_installation_id INT,
                                           installation_config TEXT,
                                           action INT ,
                                           apply_status VARCHAR(100),
                                           apply_status_message VARCHAR(200),
                                           "created_on" timestamptz,
                                           "created_by" integer,
                                           "updated_on" timestamptz,
                                           "updated_by" integer,
                                           "active" boolean,
                                           PRIMARY KEY ("id"),
                                           CONSTRAINT infrastructure_installation_id_fkey
                                           FOREIGN KEY("infrastructure_installation_id")
                                           REFERENCES"public"."infrastructure_installation" ("id")
);


COMMIT;