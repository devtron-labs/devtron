-- Drop the unique index if it exists
DROP INDEX IF EXISTS public.chart_ref_schema_unique_active_idx;

-- Drop columns if they exist
ALTER TABLE public.chart_ref_schema
DROP COLUMN IF EXISTS resource_type,
DROP COLUMN IF EXISTS resource_value;