-- Drop the unique index if it exists
DROP INDEX IF EXISTS public.chart_ref_schema_unique_active_idx;

-- Drop columns if they exist
ALTER TABLE public.chart_ref_schema
DROP COLUMN IF EXISTS resource_type,
DROP COLUMN IF EXISTS resource_value;
--- hard Deleting the added resourceQualifier entries for the hibernationPatch
DELETE FROM public.resource_qualifier_mapping where qualifier_id=9 and resource_type=10;