

CREATE TABLE IF NOT EXISTS public.chart_ref_schema
(
    id SERIAL PRIMARY KEY,
    "name" varchar(250) NOT NULL,
    "type" int NOT NULL,
    "schema" text,
    "created_on"     timestamptz,
    "created_by"     int4,
    "updated_on"     timestamptz,
    "updated_by"     int4,
    "active"         bool
    );

ALTER TABLE public.chart_ref_schema OWNER TO postgres;

INSERT INTO devtron_resource_searchable_key(name, is_removed, created_on, created_by, updated_on, updated_by)
VALUES ('CHART_REF_ID', false, now(), 1, now(), 1);