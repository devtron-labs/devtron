ALTER TABLE public.terminal_access_templates DROP constraint terminal_access_template_name_unique;

DROP TABLE IF EXISTS "public"."terminal_access_templates";
DROP TABLE IF EXISTS "public"."user_terminal_access_data";

DROP SEQUENCE IF EXISTS public.id_seq_terminal_access_templates;
DROP SEQUENCE IF EXISTS public.id_seq_user_terminal_access_data;

delete from attributes where key = 'DEFAULT_TERMINAL_IMAGE_LIST';