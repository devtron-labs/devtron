Begin;

ALTER TABLE public.chart_ref_schema
    ADD COLUMN IF NOT EXISTS resource_type INT NOT NULL DEFAULT 0;

ALTER TABLE public.chart_ref_schema
    ALTER COLUMN resource_type DROP DEFAULT;

ALTER TABLE public.chart_ref_schema
    ADD COLUMN IF NOT EXISTS resource_value TEXT;

-- Migrate data from old column "schema" to "resourceValue"
UPDATE public.chart_ref_schema
SET resource_value = "schema",
    updated_by = 1,
    updated_on = now()
WHERE "schema" IS NOT NULL;

CREATE UNIQUE INDEX chart_ref_schema_unique_active_idx
    ON public.chart_ref_schema (name, resource_type)
    WHERE active = TRUE;

End;