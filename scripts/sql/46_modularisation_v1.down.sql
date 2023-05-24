DROP TABLE "public"."server_action_audit_log" CASCADE;

DROP TABLE "public"."module_action_audit_log" CASCADE;

DROP TABLE "public"."module" CASCADE;

---- DROP sequence
DROP SEQUENCE IF EXISTS public.id_seq_module;

---- DROP sequence
DROP SEQUENCE IF EXISTS public.id_seq_module_action_audit_log;

---- DROP sequence
DROP SEQUENCE IF EXISTS public.id_seq_server_action_audit_log;