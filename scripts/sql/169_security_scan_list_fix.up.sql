ALTER TABLE image_scan_deploy_info ADD COLUMN IF NOT EXISTS is_latest_image_scanned BOOLEAN DEFAULT TRUE;
