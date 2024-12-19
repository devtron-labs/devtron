ALTER TABLE "public"."pipeline" ADD COLUMN IF NOT EXISTS "pre_stage_cache_config" varchar(50);
ALTER TABLE "public"."pipeline" ADD COLUMN IF NOT EXISTS "post_stage_cache_config" varchar(50);
ALTER TABLE "public"."ci_pipeline" ADD COLUMN IF NOT EXISTS "workflow_cache_config" varchar(50);