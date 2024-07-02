ALTER TABLE installed_apps ADD COLUMN IF NOT EXISTS is_manifest_scan_enabled bool;
