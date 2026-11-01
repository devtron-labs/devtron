-- Add message column to "public"."installed_app_version_history" table
ALTER TABLE "public"."installed_app_version_history"
    ADD COLUMN IF NOT EXISTS message TEXT;
