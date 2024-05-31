/*
 * Copyright (c) 2024. Devtron Inc.
 */

ALTER TABLE ci_workflow
    ADD column IF NOT EXISTS executor_type varchar(50);