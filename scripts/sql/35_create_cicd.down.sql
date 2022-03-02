DELETE SEQUENCE IF EXISTS public.id_seq_plugin_scripts;
DELETE SEQUENCE IF EXISTS public.id_seq_plugin_tags;
DELETE SEQUENCE IF EXISTS public.id_seq_plugin_steps;
DELETE SEQUENCE IF EXISTS public.id_seq_plugin_steps_seq;

DROP TABLE "public"."plugin_scripts" CASCADE;

DROP TABLE "public"."plugin_fields" CASCADE;

DROP TABLE "public"."plugin_tags" CASCADE;

DROP TABLE "public"."plugin_tags_map" CASCADE;

DROP TABLE "public"."plugin_steps" CASCADE;

DROP TABLE "public"."plugin_steps_sequence" CASCADE;