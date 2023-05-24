ALTER TABLE ci_template DROP COLUMN IF EXISTS ci_build_config_id;
ALTER TABLE ci_template_override DROP COLUMN IF EXISTS ci_build_config_id;

DROP TABLE IF EXISTS "public"."ci_build_config";

ALTER TABLE ci_workflow
    DROP COLUMN IF EXISTS ci_build_type;