DROP TABLE IF EXISTS "public"."ci_env_mapping";
DROP SEQUENCE IF EXISTS public.id_seq_ci_env_mapping;

DROP TABLE IF EXISTS "public"."ci_env_mapping_history";
DROP SEQUENCE IF EXISTS public.id_seq_ci_env_mapping_history;

ALTER TABLE config_map_env_level DROP COLUMN IF EXISTS deleted;
ALTER TABLE ci_workflow DROP COLUMN IF EXISTS environment_id;
ALTER TABLE ci_workflow DROP COLUMN IF EXISTS environment_name;