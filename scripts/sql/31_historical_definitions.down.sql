DROP TABLE "public"."config_map_global_history" CASCADE;
DROP TABLE "public"."config_map_env_history" CASCADE;
DROP TABLE "public"."charts_global_history" CASCADE;
DROP TABLE "public"."charts_env_history" CASCADE;
DROP TABLE "public"."installed_app_history" CASCADE;
DROP TABLE "public"."ci_script_history" CASCADE;
DROP TABLE "public"."cd_config_history" CASCADE;
ALTER TABLE "public"."config_map_app_level" DROP CONSTRAINT config_map_app_level_pkey;
ALTER TABLE "public"."config_map_env_level" DROP CONSTRAINT config_map_env_level_pkey;