/*
 * Copyright (c) 2026. Devtron Inc.
 */

-- Widen gitops_config.token to TEXT.
-- The encrypted form of long credentials (e.g. Atlassian/Bitbucket Cloud API tokens, ~190 chars
-- plaintext) exceeds varchar(250), causing "value too long for type character varying(250)" on save.
ALTER TABLE public.gitops_config
    ALTER COLUMN "token" TYPE varchar(2000);
