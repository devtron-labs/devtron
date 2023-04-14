ALTER TABLE app ADD COLUMN IF NOT EXISTS app_type integer not null DEFAULT 0;
UPDATE app SET app_type = CASE WHEN app_store = false THEN 0 WHEN app_store = true THEN 1 ELSE app_type  END;
ALTER TABLE app ADD COLUMN IF NOT EXISTS description text;
ALTER TABLE app ADD COLUMN IF NOT EXISTS display_name varchar(250);
ALTER TABLE ci_artifact ADD COLUMN IF NOT EXISTS is_artifact_uploaded BOOLEAN DEFAULT FALSE;
UPDATE ci_artifact SET is_artifact_uploaded = true;


