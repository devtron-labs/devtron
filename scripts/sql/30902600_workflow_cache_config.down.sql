ALTER TABLE "public"."pipeline" DROP COLUMN IF EXISTS "pre_stage_cache_config";
ALTER TABLE "public"."pipeline" DROP COLUMN IF EXISTS "post_stage_cache_config";
ALTER TABLE "public"."ci_pipeline" DROP COLUMN IF EXISTS "workflow_cache_config";
