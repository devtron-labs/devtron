/*
 * Copyright (c) 2025. Devtron Inc.
 */

ALTER TABLE public.docker_artifact_store
    DROP COLUMN IF EXISTS assume_role_arn;
