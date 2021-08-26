CREATE INDEX "cdwf_pipeline_id_idx" ON "public"."cd_workflow" USING BTREE ("pipeline_id");
CREATE INDEX "pco_pipeline_id_idx" ON "public"."pipeline_config_override" USING BTREE ("pipeline_id");
CREATE INDEX "cdwfr_cd_workflow_id_idx" ON "public"."cd_workflow_runner" USING BTREE ("cd_workflow_id");