-- Drop the message column from the "public"."installed_app_version_history" table
ALTER TABLE "public"."installed_app_version_history"
    DROP COLUMN IF EXISTS message;