/*
 * Copyright (c) 2025. Devtron Inc.
 */

-- Add assume_role_arn column to docker_artifact_store for cross-account ECR access via STS AssumeRole
ALTER TABLE public.docker_artifact_store
    ADD COLUMN IF NOT EXISTS assume_role_arn VARCHAR(300) DEFAULT '';
