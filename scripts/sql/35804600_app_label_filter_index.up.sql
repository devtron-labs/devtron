/*
 * Copyright (c) 2024. Devtron Inc.
 */

-- Index to support app listing tag filters based on app labels.
-- This keeps EXISTS/NOT EXISTS predicates fast for app_id + key lookups,
-- and also helps equality checks on value.
CREATE INDEX IF NOT EXISTS idx_app_label_app_id_key_value
    ON public.app_label USING BTREE (app_id, key, value);
