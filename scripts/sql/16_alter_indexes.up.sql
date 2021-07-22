CREATE INDEX "app_name_idx" ON "public"."app" USING BTREE ("app_name");
CREATE INDEX "ci_artifact_id_idx" ON "public"."pipeline_config_override" USING BTREE ("ci_artifact_id");
CREATE INDEX "environment_id_idx" ON "public"."pipeline" USING BTREE ("environment_id");
CREATE INDEX "app_id_idx" ON "public"."pipeline" USING BTREE ("app_id");