BEGIN;
-- drop sequences
DROP SEQUENCE IF EXISTS public.resource_filter_audit_seq;
DROP SEQUENCE IF EXISTS public.resource_filter_evaluation_audit_seq;

-- drop tables
DROP TABLE IF EXISTS resource_filter_audit;
DROP TABLE IF EXISTS resource_filter_evaluation_audit;
COMMIT;