/*
 * Copyright (c) 2024. Devtron Inc.
 */

ALTER TABLE chart_ref
    ADD COLUMN chart_description text DEFAULT '',
    ADD COLUMN user_uploaded boolean DEFAULT false ;
