Begin;

ALTER TABLE public.chart_ref_schema
    ADD COLUMN IF NOT EXISTS resourceType INT NOT NULL DEFAULT 0;

ALTER TABLE public.chart_ref_schema
    ALTER COLUMN resourceType DROP DEFAULT;

ALTER TABLE public.chart_ref_schema
    ADD COLUMN IF NOT EXISTS resourceValue TEXT;

CREATE UNIQUE INDEX chart_ref_schema_unique_active_idx
    ON public.chart_ref_schema (name, resourceType)
    WHERE active = TRUE;

End;