DROP TABLE "public"."ci_template_override" CASCADE;

ALTER TABLE "public"."ci_pipeline" DROP COLUMN "is_docker_config_overridden";