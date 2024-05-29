/*
 * Copyright (c) 2024. Devtron Inc.
 */

ALTER TABLE cluster
    ADD COLUMN IF NOT EXISTS proxy_url text;