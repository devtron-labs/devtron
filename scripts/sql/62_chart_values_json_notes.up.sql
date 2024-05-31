/*
 * Copyright (c) 2024. Devtron Inc.
 */

ALTER TABLE app_store_application_version ADD COLUMN values_schema_json TEXT;

ALTER TABLE app_store_application_version ADD COLUMN notes TEXT;