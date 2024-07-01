ALTER TABLE "image_scan_execution_history" ADD COLUMN IF NOT EXISTS "parent_id" integer;
ALTER TABLE "image_scan_execution_history" ADD COLUMN IF NOT EXISTS "is_latest" boolean;
