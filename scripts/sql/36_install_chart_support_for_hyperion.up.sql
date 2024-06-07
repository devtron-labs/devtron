/*
 * Copyright (c) 2024. Devtron Inc.
 */

ALTER TABLE app
ADD COLUMN IF NOT EXISTS app_offering_mode varchar(50) NOT NULL DEFAULT 'FULL';