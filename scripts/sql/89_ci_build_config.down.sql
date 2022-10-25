ALTER TABLE ci_template DROP COLUMN ci_build_config_id;
ALTER TABLE ci_template_override DROP COLUMN ci_build_config_id;

DROP TABLE IF EXISTS "public"."ci_build_config";