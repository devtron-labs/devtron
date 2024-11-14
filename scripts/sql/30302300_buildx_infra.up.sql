ALTER TABLE infra_profile_configuration ADD COLUMN IF NOT EXISTS value_string text;
ALTER TABLE infra_profile_configuration ADD COLUMN IF NOT EXISTS platform varchar(50);