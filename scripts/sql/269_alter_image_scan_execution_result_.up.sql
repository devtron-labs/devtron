ALTER TABLE "image_scan_execution_result" ADD COLUMN IF NOT EXISTS "version" text;
ALTER TABLE "image_scan_execution_result" ADD COLUMN IF NOT EXISTS "fixed_version" text;