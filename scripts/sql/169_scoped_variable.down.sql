
DROP TABLE IF EXISTS "public"."variable_data";
DROP TABLE IF EXISTS "public"."variable_scope";
DROP TABLE IF EXISTS "public"."variable_definition";

DROP SEQUENCE IF EXISTS public.variable_data_seq;
DROP SEQUENCE IF EXISTS public.variable_scope_seq;
DROP SEQUENCE IF EXISTS public.variable_definition_seq;

DROP TABLE IF EXISTS "public"."variable_entity_mapping";
DROP TABLE IF EXISTS "public"."variable_snapshot_history";
DROP SEQUENCE IF EXISTS public.variable_entity_mapping_seq;
DROP SEQUENCE IF EXISTS public.variable_snapshot_history_seq;