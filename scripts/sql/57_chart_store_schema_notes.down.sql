ALTER TABLE app_store_application_version
DROP COLUMN IF EXISTS schema_json,
    DROP COLUMN IF EXISTS notes;