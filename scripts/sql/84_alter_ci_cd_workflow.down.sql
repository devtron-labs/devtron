/*
 * Copyright (c) 2024. Devtron Inc.
 */

ALTER TABLE ci_workflow DROP COLUMN pod_name;

ALTER TABLE cd_workflow_runner DROP COLUMN pod_name;
